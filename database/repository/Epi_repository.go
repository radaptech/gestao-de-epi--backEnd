package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)


type EpiRepository struct {

	q *Queries
	db *pgxpool.Pool
}


func NewEpiRepository(pool *pgxpool.Pool) *EpiRepository {

	return  &EpiRepository{q: New(pool), db: pool}
}

func (e *EpiRepository) Adicionar(ctx context.Context,qtx *Queries, epi AddEpiParams)(int32, error){

	id,err:= qtx.AddEpi(ctx, epi)
	if err != nil {
		return 0,helper.TraduzErroPostgres(err)
	}
	return id, nil
}

func (e *EpiRepository) ListarEpi(ctx context.Context, arg BuscarEpiParams) (BuscarEpiRow, error){

	epi, err:= e.q.BuscarEpi(ctx, arg)
	if err != nil {

		return BuscarEpiRow{},err
	}

	return epi, nil
}

func (e *EpiRepository) ListarEpis(ctx context.Context, args BuscarTodosEpisPaginadoParams) ([]BuscarTodosEpisPaginadoRow, error){


	epis, err:= e.q.BuscarTodosEpisPaginado(ctx, args)
	if err != nil {

		return []BuscarTodosEpisPaginadoRow{},helper.TraduzErroPostgres(err)
	}

	return epis, nil
}

func (e *EpiRepository) CancelarEpi(ctx context.Context, qtx *Queries ,arg DeletarEpiParams)(int64, error){

	linhasAfetadas, err:= qtx.DeletarEpi(ctx, arg)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return  linhasAfetadas, nil
}


func (e *EpiRepository) AtualizaEpi(ctx context.Context , epi  UpdateEpiCampoParams)(int64, error) {

	linhasAfetadas, err:= e.q.UpdateEpiCampo(ctx, epi)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
} 

func (e *EpiRepository) BuscaEpiDashbord(ctx context.Context, tenant int32) ([]BuscaEpiDashbordRow, error){

	epis, err:= e.q.BuscaEpiDashbord(ctx, tenant)
	if err != nil {
		return []BuscaEpiDashbordRow{}, helper.TraduzErroPostgres(err)
	}

	return epis, err
}


func (e *EpiRepository) BuscaEpiTenant(ctx context.Context, tenant int32) ([]BuscaTodosItensEntreguesDoTenantRow, error){

	epi, err:= e.q.BuscaTodosItensEntreguesDoTenant(ctx, tenant)
	if err != nil {

		return []BuscaTodosItensEntreguesDoTenantRow{},err
	}

	return epi, nil
}