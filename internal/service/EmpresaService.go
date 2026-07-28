package service

import (
	"context"
	"time"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5/pgtype"
)

type EmpresaRepository interface {
	Salvar(ctx context.Context, arg repository.CriarEmpresaParams) error
	ResumoDashboard(ctx context.Context) (int64, int64, int64, int64, int64, int64, int64, float64, error)
	EmpresasRecentes(ctx context.Context) ([]repository.EmpresasRecentesRow, error)
	DadosEmpresas(ctx context.Context) ([]repository.DadosEmpresasRow, error)
	EditarEmpresa(ctx context.Context, arg repository.EditarEmpresaParams) error
}

type EmpresaService struct {
	repo      EmpresaRepository
	repoPlano PlanosRepository
}

func NewEmpresaService(repo EmpresaRepository, repoPlano PlanosRepository) *EmpresaService {

	return &EmpresaService{
		repo:      repo,
		repoPlano: repoPlano,
	}
}

func (e *EmpresaService) Salvar(ctx context.Context, model model.EmpresaInserir) error {

	plano, err := e.repoPlano.BuscarPlanoPorNome(ctx, model.Plano)
	if err != nil {
		return err
	}

	subdominio := slug.Make(model.NomeFantasia)

	err = e.repo.Salvar(ctx, repository.CriarEmpresaParams{
		NomeFantasia: model.NomeFantasia,
		RazaoSocial:  model.NomeFantasia,
		Cnpj:         model.Cnpj,

		// 👇 Só envia pro Postgres se realmente tiver texto!
		Responsavel: pgtype.Text{String: model.Responsavel, Valid: model.Responsavel != ""},
		Email:       pgtype.Text{String: model.Email, Valid: model.Email != ""},
		Telefone:    pgtype.Text{String: model.Telefone, Valid: model.Telefone != ""},
		Observacoes: pgtype.Text{String: model.Observacoes, Valid: model.Observacoes != ""},

		PlanoID:    pgtype.Int4{Int32: plano.ID, Valid: true},
		Status:     model.Status,
		Subdominio: subdominio,

		// Supondo que Vencimento seja do tipo time.Time padrão do Go:
		Vencimento: pgtype.Date{Time: model.Vencimento.Time(), Valid: !model.Vencimento.IsZero()},
	})

	return err
}

func (e *EmpresaService) EmpresaDashboard(ctx context.Context) (model.ResumoDashboard, error) {

	empresaA, empresaB, empresaT, Te, Tf, tep, tee, RM, err := e.repo.ResumoDashboard(ctx)

	if err != nil {

		return model.ResumoDashboard{}, err
	}

	return model.ResumoDashboard{
		EmpresasAtivas:     int(empresaA),
		EmpresasBloqueadas: int(empresaB),
		EmpresasEmTeste:    int(empresaT),
		TotalEmpresas:      int(Te),
		TotalFuncionarios:  int(Tf),
		TotalEpis:          int(tep),
		TotalEntregas:      int(tee),
		ReceitaMensal:      RM,
	}, nil
}

func (e *EmpresaService) EmpresaRecentes(ctx context.Context) ([]model.EmpresaRecente, error) {

	empresas, err := e.repo.EmpresasRecentes(ctx)
	if err != nil {
		return []model.EmpresaRecente{}, err
	}

	dto := make([]model.EmpresaRecente, 0, len(empresas))

	for _, empresa := range empresas {

		e := model.EmpresaRecente{
			ID:           int(empresa.ID),
			Nome:         empresa.NomeFantasia,
			Subdominio:   empresa.Subdominio,
			Responsavel:  empresa.Responsavel.String,
			Status:       empresa.Status,
			Plano:        empresa.PlanoNome,
			Funcionarios: int(empresa.Funcionarios),
			Epis:         int(empresa.Epis),
			Mensalidade:  empresa.PMensalidade,
		}

		dto = append(dto, e)
	}

	return dto, nil
}

func (e *EmpresaService) DadosEmpresas(ctx context.Context) ([]model.Empresa, error) {

	empresas, err := e.repo.DadosEmpresas(ctx)
	if err != nil {
		return []model.Empresa{}, err

	}

	dto := make([]model.Empresa, 0, len(empresas))

	for _, empresa := range empresas {

		ee := model.Empresa{
			ID:           int64(empresa.ID),
			Nome:         empresa.Nome,
			CNPJ:         empresa.Cnpj,
			Responsavel:  empresa.Responsavel.String,
			Email:        empresa.Email.String,
			Telefone:     empresa.Telefone.String,
			Plano:        empresa.PlanoNome,
			Funcionarios: int(empresa.Funcionarios),
			EPIs:         int(empresa.Epis),
			Mensalidade:  empresa.PMensalidade,
			Vencimento:   *configs.NewDataBrPtr(empresa.Vencimento.Time),
			Status:       empresa.Status,
			Observacoes:  empresa.Observacoes.String,
		}

		dto = append(dto, ee)
	}

	return dto, nil
}

func (e *EmpresaService) EditarEmpresa(ctx context.Context,  id int32, model model.EditarEmpresaRequest) error {

	err := e.repo.EditarEmpresa(ctx, repository.EditarEmpresaParams{
		NomeFantasia: model.Nome,
		RazaoSocial:  model.Nome,
		Cnpj:         model.Cnpj,
		Responsavel:  pgtype.Text{String: model.Responsavel, Valid: model.Responsavel != ""},
		Email:        pgtype.Text{String: model.Email, Valid: model.Email != ""},
		Telefone:     pgtype.Text{String: model.Telefone, Valid: model.Telefone != ""},
		PlanoID:      pgtype.Int4{Int32: int32(model.PlanoID), Valid: int32(model.PlanoID) != 0},
		Status:       model.Status,
		Vencimento:   pgtype.Date{Time: time.Time(*configs.NewDataBrPtr(model.Vencimento.Time())), Valid: true},
		Observacoes: pgtype.Text{String: model.Observacoes, Valid: model.Observacoes != ""},
		ID: id,
	})
	if err != nil {

		return err
	}


	return nil

}
