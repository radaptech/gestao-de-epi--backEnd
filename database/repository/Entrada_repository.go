package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)


type EntradaRepository struct {

	q *	Queries
	db *pgxpool.Pool
}

func NewEntradaRepository(pool *pgxpool.Pool) *EntradaRepository {

	return &EntradaRepository{
		q: New(pool),
		db: pool,
	}
}

func (e *EntradaRepository) Adicionar(ctx context.Context, args AddEntradaEpiParams) error {

	err:= e.q.AddEntradaEpi(ctx, args)
	if err != nil {

		return helper.TraduzErroPostgres(err)
	}

	return err
}



func (e *EntradaRepository) ListarEntradas(ctx context.Context, args ListarEntradasParams) ([]ListarEntradasRow, error) {

	entradas,err:= e.q.ListarEntradas(ctx, args)
	if err != nil {

		return []ListarEntradasRow{}, helper.TraduzErroPostgres(err)
	}

	return entradas, nil

}

func (e *EntradaRepository) CancelarEntrada(ctx context.Context, args CancelarEntradaParams) (int64, error) {


	linhasAfetadas,err:= e.q.CancelarEntrada(ctx, args)
	if err != nil {

		return 0 ,err
	}

	return linhasAfetadas, nil
}

func (e *EntradaRepository) TotalEntradas(ctx context.Context, args ContarEntradasFiltradasParams) (int64, error){

	total, err:= e.q.ContarEntradasFiltradas(ctx,args )
	if err != nil {
		return 0, helper.TraduzErroPostgres(err)
	}

	return  total, nil
}

func (e *EntradaRepository) BuscaEntradaDashbord(ctx context.Context, tenant int32) ([]EntradaDashbordRow, error){

	entradas, err:= e.q.EntradaDashbord(ctx, tenant)
	if err != nil {

		return []EntradaDashbordRow{}, helper.TraduzErroPostgres(err)
	}

	return entradas, nil
}

func (e *EntradaRepository) EntradaEstoque(ctx context.Context, tenant int32) ([]EntradaEpiEstoqueRow, error){

	entradas, err:= e.q.EntradaEpiEstoque(ctx, tenant)
	if err != nil {

		return []EntradaEpiEstoqueRow{}, helper.TraduzErroPostgres(err)
	} 

	return entradas, nil
}