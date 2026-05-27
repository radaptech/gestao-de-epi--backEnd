package service

import (
	"context"
	

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5/pgtype"
)

type EmpresaRepository interface {
	Salvar(ctx context.Context, arg repository.CriarEmpresaParams) error
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

	var mensalidadePg pgtype.Numeric
	err := mensalidadePg.Scan(model.Mensalidade.String())
	if err != nil {
		return err
	}

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
		Responsavel:  pgtype.Text{String: model.Responsavel, Valid: model.Responsavel != ""},
		Email:        pgtype.Text{String: model.Email, Valid: model.Email != ""},
		Telefone:     pgtype.Text{String: model.Telefone, Valid: model.Telefone != ""},
		Observacoes:  pgtype.Text{String: model.Observacoes, Valid: model.Observacoes != ""},
		
		PlanoID:      pgtype.Int4{Int32: plano.ID, Valid: true},
		Status:       model.Status,
		Mensalidade:  mensalidadePg,
		Subdominio:   subdominio,

		// 👇 Supondo que Vencimento seja do tipo time.Time padrão do Go:
		Vencimento: pgtype.Date{Time: model.Vencimento.Time(), Valid: !model.Vencimento.IsZero()},
	})

	return err
}
