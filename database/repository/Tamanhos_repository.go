package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)


type TamanhosRepository struct {

	q *Queries
	db *pgxpool.Pool
}

func NewTamanhoRepository(pool *pgxpool.Pool) *TamanhosRepository {

	return &TamanhosRepository{q: New(pool), db: pool}
}

func (t *TamanhosRepository) Adicionar(ctx context.Context, tamanho AddTamanhoParams) error {

	err := t.q.AddTamanho(ctx, tamanho)
	if err != nil {

		return helper.TraduzErroPostgres(err)
	}

	return nil
}

func (t *TamanhosRepository) ListarTamanho(ctx context.Context, arg BuscarTamanhoParams) (BuscarTamanhoRow, error){


	return  t.q.BuscarTamanho(ctx, arg)

}

func (t *TamanhosRepository) ListarTamanhos(ctx context.Context, tenantId int32) ([]BuscarTodosTamanhosRow, error){

	tamanhos, err:= t.q.BuscarTodosTamanhos(ctx, tenantId)
	if err != nil {

		return  []BuscarTodosTamanhosRow{}, helper.TraduzErroPostgres(err)
	}

	return tamanhos, nil

}

func (t *TamanhosRepository) CancelarTamanho(ctx context.Context, arg DeletarTamanhoParams) (int64, error) {

	linhasAfetadas, err:= t.q.DeletarTamanho(ctx, arg)
		if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
}

func (t *TamanhosRepository) BuscarTamanhoPorIdePI(ctx context.Context, arg BuscarTamanhosPorIdEpiParams) ([]BuscarTamanhosPorIdEpiRow, error){

	Tamanhos, err:= t.q.BuscarTamanhosPorIdEpi(ctx, arg)
	if err != nil {

		return  []BuscarTamanhosPorIdEpiRow{}, helper.TraduzErroPostgres(err)
	}

	return Tamanhos, nil
}