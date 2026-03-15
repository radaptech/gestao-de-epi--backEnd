package service

import (
	"context"
	"errors"
	"math"

	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

//go:generate mockery --name=DepartamentoRepository --output=../../mocks --outpkg=mocks --with-expecter
type DepartamentoRepository interface {
	Adicionar(ctx context.Context, departamento repository.CriaDepartamentoParams) (repository.Departamento, error)
	ListarDepartamentos(ctx context.Context, args repository.BuscarTodosDepartamentosParams) ([]repository.BuscarTodosDepartamentosRow, error)
	CancelarDepartamento(ctx context.Context, arg repository.DeletarDepartamentoParams) (int64, error)
	AtualizarDepartamento(ctx context.Context, arg repository.UpdateDepartamentoParams) (int64, error)
}

type DepartamentoService struct {
	repo DepartamentoRepository
}

func NewDepartamentoService(r DepartamentoRepository) *DepartamentoService {
	return &DepartamentoService{repo: r}
}

func (d *DepartamentoService) SalvarDepartamento(ctx context.Context, tenantId int32, m model.Departamento) (model.DepartamentoDto, error) {

	m.Departamento = strings.TrimSpace(m.Departamento)

	if len(m.Departamento) < 2 {

		return model.DepartamentoDto{}, helper.ErrNomeCurto

	}

	dep, err := d.repo.Adicionar(ctx, repository.CriaDepartamentoParams{
		TenantID: tenantId,
		Nome:     m.Departamento,
	})
	if err != nil {

		if errors.Is(err, helper.ErrDadoDuplicado) {

			return model.DepartamentoDto{}, err
		}

		return model.DepartamentoDto{}, err
	}

	return model.DepartamentoDto{ID: int(dep.ID), Departamento: dep.Nome}, nil
}

type FiltroDepartamento struct {
	Id_departamento int32  `form:"idDepartamento"`
	Nome            string `form:"nome"`
	Cancelado       bool   `form:"cancelado"`
	FiltroPaginacao
}

type DepartamentoPaginado struct {
	Departamento []model.DepartamentoDto `json:"departamentos"`
	Total        int64                   `json:"total"`
	Pagina       int32                   `json:"pagina"`
	PaginaFinal  int32                   `json:"paginaFinal"`
}

func (d *DepartamentoService) ListarTodosDepartamentos(ctx context.Context, f FiltroDepartamento, tenantId int32) (DepartamentoPaginado, error) {

	p := Paginacao(f.FiltroPaginacao)

	filtro := repository.BuscarTodosDepartamentosParams{
		Limit:      p.Limit,
		Offset:     p.Offset,
		TenantID:   tenantId,
		Nome:       pgtype.Text{String: f.Nome, Valid: f.Nome != ""},
		ID:         pgtype.Int4{Int32: f.Id_departamento, Valid: f.Id_departamento > 0},
		Cancelados: f.Cancelado,
	}

	departamentos, err := d.repo.ListarDepartamentos(ctx, filtro)
	if err != nil {
		return DepartamentoPaginado{}, err
	}

	dto := make([]model.DepartamentoDto, 0, len(departamentos))
	for _, departamento := range departamentos {

		d := model.DepartamentoDto{
			ID:           int(departamento.ID),
			Departamento: departamento.Departamento,
		}

		dto = append(dto, d)
	}

	var total int64
	if len(departamentos) > 0 {
		total = departamentos[0].TotalGeral
	}
	//numero da ultima pagina
	ultimaPagina := int32(math.Ceil(float64(total) / float64(p.Limit)))

	return DepartamentoPaginado{

		Departamento: dto,
		Total:        total,
		Pagina:       p.PaginaAtual,
		PaginaFinal:  ultimaPagina,
	}, nil

}

func (d *DepartamentoService) DeletarDepartamento(ctx context.Context, id int, tenantId int32) error {

	if id <= 0 {

		return helper.ErrId
	}

	idDep, err := d.repo.CancelarDepartamento(ctx, repository.DeletarDepartamentoParams{
		ID:       int32(id),
		TenantID: tenantId,
	})
	if err != nil {

		return helper.ErrInternal
	}

	if idDep == 0 {

		return helper.ErrNaoEncontrado
	}

	return nil
}

func (d *DepartamentoService) AtualizarDepartamento(ctx context.Context, id int32, novoNome string, tenantId int32) error {

	novoNome = strings.TrimSpace(novoNome)
	if len(novoNome) < 2 {
		return helper.ErrNomeCurto
	}

	arg := repository.UpdateDepartamentoParams{
		ID:       id,
		Nome:     novoNome,
		TenantID: tenantId,
	}

	linha, errDep := d.repo.AtualizarDepartamento(ctx, arg)
	if errDep != nil {

		return errDep
	}

	if linha == 0 {

		return helper.ErrNaoEncontrado
	}

	return nil

}
