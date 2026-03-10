package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)



type EstoqueRepository struct {

	q *Queries
	db *pgxpool.Pool
}


func NewEstoqueRepository(pool *pgxpool.Pool) *EstoqueRepository {

	return  &EstoqueRepository{
		q: New(pool),
		db: pool,
	}
}

func (e *EstoqueRepository) SomaQuantidade(ctx context.Context, args ListarEstoqueAtualParams) ([]ListarEstoqueAtualRow, error) {

	quantidades, err:= e.q.ListarEstoqueAtual(ctx, args)
	if err != nil {

		return  []ListarEstoqueAtualRow{}, helper.TraduzErroPostgres(err)
	}

	return  quantidades, nil
}


func (e *EstoqueRepository) SaldoAtual(ctx context.Context, arg ListarSaldoEstoqueParams) ([]ListarSaldoEstoqueRow, error) {

	saldo, err:= e.q.ListarSaldoEstoque(ctx, arg)
	if err != nil {

		return  []ListarSaldoEstoqueRow{}, helper.TraduzErroPostgres(err)
	}

	return  saldo, nil
}