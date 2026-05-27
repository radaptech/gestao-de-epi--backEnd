package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)


type EmpresaRepository struct {

	q *Queries
	db *pgxpool.Pool
}

func NewEmpresaRepository(pool *pgxpool.Pool) *EmpresaRepository {

	return  &EmpresaRepository{
		q: New(pool),
		db: pool,
	}
}

func (e *EmpresaRepository) Salvar(ctx context.Context, arg CriarEmpresaParams) error {

	err:= e.q.CriarEmpresa(ctx, arg)
	if err != nil {

		return helper.TraduzErroPostgres(err)
	}

	return nil
}