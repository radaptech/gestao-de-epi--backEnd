package service

import (
	"context"
	
	"fmt"
	"math"
	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	
	"github.com/jackc/pgx/v5/pgtype"
)

type FuncaoRepository interface {
	Adicionar(ctx context.Context, args repository.AddFuncaoParams) error
	ListarFuncoes(ctx context.Context, args repository.BuscarTodasFuncoesParams) ([]repository.BuscarTodasFuncoesRow, error)
	CancelarFuncao(ctx context.Context, arg repository.DeletarFuncaoParams) (int64, error)
	AtualizarFuncao(ctx context.Context, arg repository.UpdateFuncaoParams) (int64, error)
}

type FuncaoService struct {
	repo FuncaoRepository
}

func NewFuncaoService(f FuncaoRepository) *FuncaoService {
	return &FuncaoService{repo: f}
}

func (f *FuncaoService) SalvarFuncao(ctx context.Context, model model.Funcao, tenantid int32) error {

	model.Funcao = strings.TrimSpace(model.Funcao)

	F := repository.AddFuncaoParams{
		Nome:           model.Funcao,
		Iddepartamento: int32(model.IdDepartamento),
		TenantID:       tenantid,
	}
	if err := f.repo.Adicionar(ctx, F); err != nil {

		return fmt.Errorf("erro ao salvar funcao, %w", err)
	}

	return nil
}



type FiltroFuncao struct {
	Id_funcao        int32  `form:"idFuncao"`
	Nome             string `form:"nome"`
	NomeDepartamento string `form:"nome_departamento"`
	Cancelado        bool   `form:"cancelado"`
	FiltroPaginacao
}

type FuncaoPaginado struct {
	Funcao      []model.FuncaoDto `json:"funcoes"`
	Total       int64             `json:"total"`
	Pagina      int32             `json:"pagina"`
	PaginaFinal int32             `json:"paginaFinal"`
}

func (fu *FuncaoService) ListasTodasFuncao(ctx context.Context, f FiltroFuncao, tenantId int32) (FuncaoPaginado, error) {

	p := Paginacao(f.FiltroPaginacao)

	filtro := repository.BuscarTodasFuncoesParams{
		Limit:            p.Limit,
		Offset:           p.Offset,
		TenantID:         tenantId,
		Nome:             pgtype.Text{String: f.Nome, Valid: f.Nome != ""},
		ID:               pgtype.Int4{Int32: f.Id_funcao, Valid: f.Id_funcao > 0},
		Cancelados:       f.Cancelado,
		NomeDepartamento: pgtype.Text{String: f.NomeDepartamento, Valid: f.NomeDepartamento != ""},
	}

	funcoes, err := fu.repo.ListarFuncoes(ctx, filtro)
	if err != nil {
		return FuncaoPaginado{}, err
	}

	dto := make([]model.FuncaoDto, 0, len(funcoes))

	for _, funcs := range funcoes {

		f := model.FuncaoDto{
			ID:     int(funcs.ID),
			Funcao: funcs.Nome,
			Departamento: model.DepartamentoDto{
				ID:           int(funcs.Iddepartamento),
				Departamento: funcs.DepartamentoNome,
			},
		}

		dto = append(dto, f)
	}

	var total int64
	if len(funcoes) > 0 {
		total = funcoes[0].TotalGeral
	}

	ultimaPagina := int32(math.Ceil(float64(total) / float64(p.Limit)))

	return FuncaoPaginado{
		Funcao: dto,
		Total: total,
		Pagina: p.PaginaAtual,
		PaginaFinal: ultimaPagina,
	}, nil

}

func (f *FuncaoService) DeletarFuncao(ctx context.Context, id int, tenantId int32) error {

	if id <= 0 {

		return helper.ErrId
	}

	linha, err := f.repo.CancelarFuncao(ctx, repository.DeletarFuncaoParams{
		ID:       int32(id),
		TenantID: tenantId,
	})
	if err != nil {

		return fmt.Errorf("erro ao deletar a funcao, %w", err)
	}

	if linha == 0 {

		return helper.ErrNaoEncontrado
	}

	return nil
}

func (f *FuncaoService) AtualizarFuncao(ctx context.Context, id int, funcao string, tenantId int32) error {

	if id <= 0 {
		return helper.ErrId
	}

	funcaoLimpa := strings.TrimSpace(funcao)

	if len(funcaoLimpa) < 2 {

		return helper.ErrNomeCurto
	}

	arg := repository.UpdateFuncaoParams{
		ID:       int32(id),
		Nome:     funcaoLimpa,
		TenantID: tenantId,
	}

	linha, err := f.repo.AtualizarFuncao(ctx, arg)
	if err != nil {

		return fmt.Errorf("erro tecnico ao realizar o update: %w", err)
	}

	if linha == 0 {
		return helper.ErrNaoEncontrado
	}

	return nil
}
