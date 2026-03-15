package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	
	"github.com/jackc/pgx/v5/pgxpool"
)


type ProtecaoRepository struct {

	q *Queries
	db *pgxpool.Pool
}

func NewProtecaoRepository(pool *pgxpool.Pool) *ProtecaoRepository {

	return &ProtecaoRepository{
		q: New(pool),
		db: pool,
	}
}

func (p *ProtecaoRepository) Adicionar(ctx context.Context, nome AddProtecaoParams) (TipoProtecao, error) {

	tp,err := p.q.AddProtecao(ctx, nome)
	if err != nil {
		return  TipoProtecao{},helper.TraduzErroPostgres(err)
	}

	return  tp, nil
}

func (p *ProtecaoRepository) ListarProtecao(ctx context.Context, arg BuscarProtecaoParams) (BuscarProtecaoRow, error){


	return p.q.BuscarProtecao(ctx, arg)
}

func (p *ProtecaoRepository) ListarProtecoes(ctx context.Context, tenantId int32) ([]BuscarTodasProtecoesRow, error){

	protc, err:= p.q.BuscarTodasProtecoes(ctx, tenantId)
	if err != nil {

		return []BuscarTodasProtecoesRow{}, helper.TraduzErroPostgres(err)
	}

	return protc, nil
}

func (p *ProtecaoRepository) CancelarProtecao(ctx context.Context, arg DeletarProtecaoParams) (int64, error){

	linhasAfetadas,err:= p.q.DeletarProtecao(ctx, arg)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
}

