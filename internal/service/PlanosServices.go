package service

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type PlanosRepository interface {
	Adicionar(ctx context.Context, arg repository.AddPlanoParams) (int32, error)
	MostrarPlanos(ctx context.Context)([]repository.BuscaPlanosRow, error)
}


type PlanosService struct {

	repo PlanosRepository

}

func NewPlanoService(p PlanosRepository) *PlanosService {

	return &PlanosService{
		repo: p,
	}
}


func (p *PlanosService) SalvarPlanos(ctx context.Context, model model.Plano) (int32, error) {

	// 1. Conversão segura de Decimal para pgtype.Numeric
	// Usar o Scan() passando a string garante que os centavos não se percam
	var mensalidadePg pgtype.Numeric
	err := mensalidadePg.Scan(model.Mensalidade.String())
	if err != nil {
		return 0, err
	}

	// 2. Função anônima (helper) para converter *int em pgtype.Int4 com segurança
	// Se for nil (ilimitado), ele devolve Valid: false, salvando NULL no banco.
	toPgInt4 := func(val *int32) pgtype.Int4 {
		if val == nil {
			return pgtype.Int4{Valid: false}
		}
		return pgtype.Int4{Int32: int32(*val), Valid: true}
	}

	// 3. Chamada ao repositório
	planoId, err := p.repo.Adicionar(ctx, repository.AddPlanoParams{
		Nome:               model.Nome,
		Mensalidade:        mensalidadePg,
		LimiteFuncionarios: toPgInt4(model.LimiteFuncionarios),
		LimiteUsuarios:     toPgInt4(model.LimiteUsuarios),
		LimiteEpis:         toPgInt4(model.LimiteEpis),
		Status:             pgtype.Text{String: model.Status, Valid: model.Status != ""},
		Descricao:          model.Descricao,
	})
	
	if err != nil {
		return 0, err
	}

	return planoId, nil
}


func (p *PlanosService) MostrarPlanos(ctx context.Context) ([]model.Plano, error) {

	planos, err := p.repo.MostrarPlanos(ctx)
	if err != nil {
		return []model.Plano{}, err
	}

	dto := make([]model.Plano, 0, len(planos))

	// Função auxiliar (Closure) para converter de pgtype.Int4 para *int32 respeitando o NULL
	toPtr := func(pgInt pgtype.Int4) *int32 {
		if !pgInt.Valid {
			return nil // Retorna nulo para o JSON (Ilimitado)
		}
		val := pgInt.Int32
		return &val
	}

	for _, plano := range planos {

		// Conversão exata de pgtype.Numeric para decimal.Decimal (sem passar por float)
		var valorDecimal decimal.Decimal
		if plano.Mensalidade.Valid && plano.Mensalidade.Int != nil {
			valorDecimal = decimal.NewFromBigInt(plano.Mensalidade.Int, plano.Mensalidade.Exp)
		}

		pp := model.Plano{
			ID:                 int(plano.ID),
			Nome:               plano.Nome,
			Mensalidade:        valorDecimal,
			LimiteFuncionarios: toPtr(plano.LimiteFuncionarios),
			LimiteUsuarios:     toPtr(plano.LimiteUsuarios),
			LimiteEpis:         toPtr(plano.LimiteEpis),
			Status:             plano.Status.String,
			Descricao:          plano.Descricao,
		}

		dto = append(dto, pp)
	}

	return dto, nil
}
