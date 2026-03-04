package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FuncionarioRepository interface {
	Adicionar(ctx context.Context, args repository.AddFuncionarioParams) error
	ListarFuncionario(ctx context.Context, arg repository.BuscaFuncionarioParams) (repository.BuscaFuncionarioRow, error)
	ListarFuncionarios(ctx context.Context, args repository.BuscarTodosFuncionariosParams) ([]repository.BuscarTodosFuncionariosRow, error)
	CancelarFuncionario(ctx context.Context, arg repository.DeletarFuncionarioParams) (int64, error)
	AtualizarFuncionarioNome(ctx context.Context, arg repository.UpdateFuncionarioNomeParams, qtx *repository.Queries) (int64, error)
	AtualizarFuncionarioDepartamento(ctx context.Context, arg repository.UpdateFuncionarioDepartamentoParams, qtx *repository.Queries) (int64, error)
	AtualizarFuncionarioFuncao(ctx context.Context, arg repository.UpdateFuncionarioFuncaoParams, qtx *repository.Queries) (int64, error)
}

type FuncionarioService struct {
	repo    FuncionarioRepository
	db      *pgxpool.Pool
	queries *repository.Queries
}

func NewFuncionarioService(f FuncionarioRepository, pool *pgxpool.Pool) *FuncionarioService {
	return &FuncionarioService{repo: f, db: pool, queries: repository.New(pool)}
}

func (f *FuncionarioService) SalvarFuncionario(ctx context.Context, model model.FuncionarioINserir, tenantId int32) error {

	model.Nome = strings.TrimSpace(model.Nome)

	args := repository.AddFuncionarioParams{
		Nome:           model.Nome,
		Matricula:      model.Matricula,
		Iddepartamento: int32(model.ID_departamento),
		Idfuncao:       int32(model.ID_funcao),
		TenantID:       tenantId,
	}
	err := f.repo.Adicionar(ctx, args)
	if err != nil {

		if errors.Is(err, helper.ErrConflitoIntegridade) {

			return fmt.Errorf("departamento ou função nao encontrado")

		}

		if errors.Is(err, helper.ErrDadoDuplicado) {

			return fmt.Errorf("matricula ja cadastrada")
		}

		return err

	}

	return nil
}

func (f *FuncionarioService) ListarFuncionario(ctx context.Context, matricula string, tenantId int32) (model.Funcionario_Dto, error) {

	if matricula <= "" {

		return model.Funcionario_Dto{}, helper.ErrId
	}
	funcionario, err := f.repo.ListarFuncionario(ctx, repository.BuscaFuncionarioParams{
		Matricula: matricula,
		TenantID:  tenantId,
	})
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {

			return model.Funcionario_Dto{}, helper.ErrNaoEncontrado
		}
		return model.Funcionario_Dto{}, err
	}

	funcDto := model.Funcionario_Dto{
		ID:        int(funcionario.ID),
		Nome:      funcionario.Nome,
		Matricula: funcionario.Matricula,
		Funcao: model.FuncaoDto{
			ID:     int(funcionario.Idfuncao),
			Funcao: funcionario.FuncaoNome,
			Departamento: model.DepartamentoDto{
				ID:           int(funcionario.Iddepartamento),
				Departamento: funcionario.DepartamentoNome,
			},
		},
	}

	return funcDto, nil

}

type FiltroFuncionario struct {
	ID_funcionario int32  `form:"idFuncionario"`
	Matricula      string `form:"matricula"`
	Nome           string `form:"nome"`
	Cancelados     bool   `form:"cancelados"`
	FiltroPaginacao
}

type FuncionarioPaginado struct {
	Funcionarios []model.Funcionario_Dto `json:"funcionario"`
	Total        int64                   `json:"total"`
	Pagina       int32                   `json:"pagina"`
	PaginaFinal  int32                   `json:"paginaFinal"`
}

func (funcio *FuncionarioService) ListaTodosFuncionarios(ctx context.Context, f FiltroFuncionario, tenantId int32) (FuncionarioPaginado, error) {

	p := Paginacao(f.FiltroPaginacao)

	filtro := repository.BuscarTodosFuncionariosParams{
		Limit:      p.Limit,
		Offset:     p.Offset,
		TenantID:   tenantId,
		ID:         pgtype.Int4{Int32: f.ID_funcionario, Valid: f.ID_funcionario > 0},
		Matricula:  pgtype.Text{String: f.Matricula, Valid: f.Matricula != ""},
		Nome:       pgtype.Text{String: f.Nome, Valid: f.Nome != ""},
		Cancelados: f.Cancelados,
	}

	funcionarios, err := funcio.repo.ListarFuncionarios(ctx, filtro)
	if err != nil {

		return FuncionarioPaginado{}, err
	}

	dto := make([]model.Funcionario_Dto, 0, len(funcionarios))

	for _, funcionario := range funcionarios {

		funcs := model.Funcionario_Dto{
			ID:        int(funcionario.ID),
			Nome:      funcionario.Nome,
			Matricula: funcionario.Matricula,
			Funcao: model.FuncaoDto{
				ID:     int(funcionario.Idfuncao),
				Funcao: funcionario.FuncaoNome,
				Departamento: model.DepartamentoDto{
					ID:           int(funcionario.Iddepartamento),
					Departamento: funcionario.DepartamentoNome,
				},
			},
		}

		dto = append(dto, funcs)
	}

	var total int64
	if len(funcionarios) > 0 {
		total = funcionarios[0].TotalGeral
	}

	//numero da ultima pagina
	ultimaPagina := int32(math.Ceil(float64(total) / float64(p.Limit)))

	return FuncionarioPaginado{

		Funcionarios: dto,
		Total:        total,
		Pagina:       p.PaginaAtual,
		PaginaFinal:  ultimaPagina,
	}, nil
}

func (f *FuncionarioService) DeletarFuncionario(ctx context.Context, id int, tenantId int32) error {

	linhas, err := f.repo.CancelarFuncionario(ctx, repository.DeletarFuncionarioParams{
		ID:       int32(id),
		TenantID: tenantId,
	})
	if err != nil {

		return fmt.Errorf("erro ao deletar funcionario, %w, funcionario ja pode estar inativo", err)
	}

	if linhas == 0 {

		return helper.ErrNaoEncontrado
	}

	return nil

}

func (f *FuncionarioService) AtualizaNomeFuncionario(ctx context.Context, id int, nome string, tenantId int32, qtx *repository.Queries) error {

	if id <= 0 {
		return helper.ErrId
	}

	nomeLimpo := strings.TrimSpace(nome)

	if len(nomeLimpo) < 2 {

		return helper.ErrNomeCurto
	}
	args := repository.UpdateFuncionarioNomeParams{
		ID:       int32(id),
		Nome:     nomeLimpo,
		TenantID: tenantId,
	}

	linha, err := f.repo.AtualizarFuncionarioNome(ctx, args, qtx)
	if err != nil {

		return fmt.Errorf("erro tecnico ao realizar o update: %w", err)
	}

	if linha == 0 {
		return helper.ErrNaoEncontrado
	}

	return nil
}

func (f *FuncionarioService) AtualizaDepartamentoFuncionario(ctx context.Context, id, iddepartamento, tenantId int, qtx *repository.Queries) error {

	if id <= 0 {
		return helper.ErrId
	}

	args := repository.UpdateFuncionarioDepartamentoParams{
		ID:             int32(id),
		Iddepartamento: int32(iddepartamento),
		TenantID:       int32(tenantId),
	}

	linha, err := f.repo.AtualizarFuncionarioDepartamento(ctx, args, qtx)
	if err != nil {

		return fmt.Errorf("erro tecnico ao realizar o update: %w", err)
	}

	if linha == 0 {
		return helper.ErrNaoEncontrado
	}

	return nil
}

func (f *FuncionarioService) AtualizaFuncaoFuncionario(ctx context.Context, id, idfuncao, tenantID int, qtx *repository.Queries) error {

	if id <= 0 {
		return helper.ErrId
	}

	args := repository.UpdateFuncionarioFuncaoParams{
		ID:       int32(id),
		Idfuncao: int32(idfuncao),
		TenantID: int32(tenantID),
	}
	linha, err := f.repo.AtualizarFuncionarioFuncao(ctx, args, qtx)
	if err != nil {

		return fmt.Errorf("erro tecnico ao realizar o update: %w", err)
	}

	if linha == 0 {
		return helper.ErrNaoEncontrado
	}

	return nil
}

func (f *FuncionarioService) AtualizarFuncionarioCompleto(ctx context.Context, id int, req model.UpdateFuncionarioRequest, tenantId int) error {

	// Validação básica de ID acontece uma vez só
	if id <= 0 {
		return helper.ErrId
	}
	tx, err := f.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := f.queries.WithTx(tx)
	// 1. Atualiza Nome (se foi enviado)
	if req.Nome != nil {
		// Reutiliza sua lógica existente que já valida tamanho, trim, etc.
		err := f.AtualizaNomeFuncionario(ctx, id, *req.Nome, int32(tenantId), qtx)
		if err != nil {
			return err
		}
	}

	// 2. Atualiza Departamento (se foi enviado)
	if req.IdDepartamento != nil {
		err := f.AtualizaDepartamentoFuncionario(ctx, id, *req.IdDepartamento, tenantId, qtx)
		if err != nil {
			return err
		}
	}

	// 3. Atualiza Função (se foi enviado)
	if req.IdFuncao != nil {
		err := f.AtualizaFuncaoFuncionario(ctx, id, *req.IdFuncao, tenantId, qtx)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
