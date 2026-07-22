package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)


type FuncaoRepository struct {

	q *Queries
	db *pgxpool.Pool
}


func NewFuncaoRepository(pool *pgxpool.Pool) *FuncaoRepository {

	return &FuncaoRepository{
		q: New(pool),
		db: pool,
	}
}

func (f *FuncaoRepository) Adicionar(ctx context.Context, args AddFuncaoParams) error {

	err :=f.q.AddFuncao(ctx, args)
	if err != nil {

		return helper.TraduzErroPostgres(err)
	}

	return nil	
}


func (f *FuncaoRepository) ListarFuncoes(ctx context.Context, args BuscarTodasFuncoesParams)([]BuscarTodasFuncoesRow, error) {

	funcoes, err:= f.q.BuscarTodasFuncoes(ctx, args)
	if err != nil {

		return []BuscarTodasFuncoesRow{}, helper.TraduzErroPostgres(err)
	}

	return  funcoes, nil
}

func (f *FuncaoRepository) CancelarFuncao(ctx context.Context, arg DeletarFuncaoParams) (int64, error){

	linhasAfetadas,err:= f.q.DeletarFuncao(ctx, arg)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
}

func (f *FuncaoRepository) AtualizarFuncao(ctx context.Context, arg UpdateFuncaoParams) (int64, error){

	linhasAfetadas,err:= f.q.UpdateFuncao(ctx, arg)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
}


func (f *FuncaoRepository) MapDepartamentosParaFuncao(ctx context.Context, tenantId int32) ([]BuscaDepartamentosMapRow, error){

	deps, err:= f.q.BuscaDepartamentosMap(ctx, tenantId)
	if err != nil {

		return  []BuscaDepartamentosMapRow{}, helper.TraduzErroPostgres(err)
	}

	return deps, nil

}