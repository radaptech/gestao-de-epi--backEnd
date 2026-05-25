package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlanosRepository struct {
	q  *Queries
	db *pgxpool.Pool
}

func NewPlanosRepository(pool *pgxpool.Pool) *PlanosRepository {

	return &PlanosRepository{
		q:  New(pool),
		db: pool,
	}
}


func (p *PlanosRepository) Adicionar(ctx context.Context, arg AddPlanoParams)(int32, error){

	planoId, err:= p.q.AddPlano(ctx, arg)
	if err != nil {
		return  0, helper.TraduzErroPostgres(err)
	}

	return planoId, nil
}


func (p *PlanosRepository) MostrarPlanos(ctx context.Context)([]BuscaPlanosRow, error){

	planos,err:= p.q.BuscaPlanos(ctx)
	if err != nil {

		return []BuscaPlanosRow{}, helper.TraduzErroPostgres(err)
	}

	return planos, err
}