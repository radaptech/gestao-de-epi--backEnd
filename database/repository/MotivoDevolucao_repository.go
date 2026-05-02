package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)


type MotivoDevolucaoRepository struct {
	q *Queries
	db *pgxpool.Pool
}


func NewMotivoDevolucaoRepository(pool *pgxpool.Pool) *MotivoDevolucaoRepository {

	return &MotivoDevolucaoRepository{
		q: New(pool),
		db: pool,
	}
}

func (m *MotivoDevolucaoRepository) Adicionar(ctx context.Context, arg AddMotivoDevolucaoParams) (AddMotivoDevolucaoRow, error) {

	motivo, err:= m.q.AddMotivoDevolucao(ctx,arg )
	if err != nil {
		return AddMotivoDevolucaoRow{}, helper.TraduzErroPostgres(err)
	}

	return motivo, nil
}

func (m *MotivoDevolucaoRepository) ListarMotivo(ctx context.Context, arg BuscaMotivoDevolucaoParams) (BuscaMotivoDevolucaoRow, error){

	motivo, err:= m.q.BuscaMotivoDevolucao(ctx, arg)
	if err != nil {

		 return BuscaMotivoDevolucaoRow{},helper.TraduzErroPostgres(err)
	}

	return  motivo, err
}

func (m *MotivoDevolucaoRepository) ListarMotivos(ctx context.Context, tenantId int32) ([]BuscaTodosMotivosDevolucaoRow, error){

	motivos, err:= m.q.BuscaTodosMotivosDevolucao(ctx, tenantId)
	if err != nil {

		 return []BuscaTodosMotivosDevolucaoRow{},helper.TraduzErroPostgres(err)
	}

	return  motivos, err
}

func (m *MotivoDevolucaoRepository) CancelarMotivoDevolucao(ctx context.Context, arg DeleteMotivoDevolucaoParams) (int64, error) {

	linhasAfetadas,err:= m.q.DeleteMotivoDevolucao(ctx, arg)
	if err != nil {

		return 0,helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
}