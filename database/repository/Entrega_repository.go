package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EntregaRepository struct {
	q  *Queries
	db *pgxpool.Pool
}

func NewEntregaRepository(pool *pgxpool.Pool) *EntregaRepository {

	return &EntregaRepository{
		q:  New(pool),
		db: pool,
	}

}

func (e *EntregaRepository) AdicionarEntrega(ctx context.Context, qtx *Queries, args AddEntregaEpiParams) (int32, error) {

	id, err := qtx.AddEntregaEpi(ctx, args)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return id, nil
}

func (e *EntregaRepository) AdicionarEntregaItem(ctx context.Context, qtx *Queries, arg AddItemEntregueParams) (AddItemEntregueRow, error) {

	ids, err := qtx.AddItemEntregue(ctx, arg)

	// LOG 2: Verificar o resultado do banco
	if err != nil {

		return AddItemEntregueRow{}, helper.TraduzErroPostgres(err)
	}

	return ids, nil
}

func (e *EntregaRepository) ListarEntregas(ctx context.Context, args ListarEntregasParams) ([]ListarEntregasRow, error) {

	entregas, err := e.q.ListarEntregas(ctx, args)
	if err != nil {
		return []ListarEntregasRow{}, helper.TraduzErroPostgres(err)
	}
	return entregas, nil
}

func (e *EntregaRepository) Cancelar(ctx context.Context, qtx *Queries, args CancelarEntregaParams) (int32, error) {

	id, err := qtx.CancelarEntrega(ctx, args)
	if err != nil {
		if err == pgx.ErrNoRows {

			return 0, helper.ErrNaoEncontrado
		}
		return 0, err
	}

	return id, nil
}

func (e *EntregaRepository) CancelarEntregaItem(ctx context.Context, qtx *Queries, arg CancelaItemEntregueParams) ([]CancelaItemEntregueRow, error) {

	itemsCancelados, err := qtx.CancelaItemEntregue(ctx, arg)
	if err != nil {

		return []CancelaItemEntregueRow{}, helper.TraduzErroPostgres(err)
	}

	return itemsCancelados, nil
}

func (e *EntregaRepository) AbaterEstoqueEntrada(ctx context.Context, qtx *Queries, args AbaterEstoqueLoteParams) (int64, error) {

	linhasAfetadas, err := qtx.AbaterEstoqueLote(ctx, args)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
}

func (e *EntregaRepository) ReporEstoqueEntrada(ctx context.Context, qtx *Queries, args ReporEstoqueLoteParams) (int64, error) {

	linhasAfetadas, err := qtx.ReporEstoqueLote(ctx, args)
	if err != nil {

		return 0, helper.TraduzErroPostgres(err)
	}

	return linhasAfetadas, nil
}
func (e *EntregaRepository) ListarEntregasDisponiveis(ctx context.Context, qtx *Queries, args ListarLotesParaConsumoParams) ([]ListarLotesParaConsumoRow, error) {

	lotes, err := qtx.ListarLotesParaConsumo(ctx, args)
	if err != nil {

		return []ListarLotesParaConsumoRow{}, helper.TraduzErroPostgres(err)
	}

	return lotes, nil
}

func (e *EntregaRepository) ListarEpisEntreguesCancelados(ctx context.Context, qtx *Queries, arg ListarItensEntregueCanceladosParams) ([]ListarItensEntregueCanceladosRow, error) {

	cancelados, err := qtx.ListarItensEntregueCancelados(ctx, arg)
	if err != nil {

		return []ListarItensEntregueCanceladosRow{}, err
	}

	return cancelados, nil
}

func (e *EntregaRepository) ListasEntregasPorMatricula(ctx context.Context, args ListarHistoricoEntregasPorMatriculaParams) ([]ListarHistoricoEntregasPorMatriculaRow, error) {

	
	entrega, err := e.q.ListarHistoricoEntregasPorMatricula(ctx, args)
	if err != nil {

		return []ListarHistoricoEntregasPorMatriculaRow{}, helper.TraduzErroPostgres(err)
	}

	
	return entrega, nil
}

func (e *EntregaRepository) BuscaEntregaDashbord(ctx context.Context, tenant int32) ([]EntregaDashbordRow, error){

	entregas, err:= e.q.EntregaDashbord(ctx, tenant)
	if err != nil {

		return []EntregaDashbordRow{}, helper.TraduzErroPostgres(err)
	}


	return entregas, nil
}

func (e *EntregaRepository) BuscaEntregaItensDashbord(ctx context.Context, tenant int32) ([]EntregaItensDashbordRow, error){

	itens, err:= e.q.EntregaItensDashbord(ctx, tenant)
	if err != nil {

		return []EntregaItensDashbordRow{}, helper.TraduzErroPostgres(err)
	}


	return itens, nil
}

// Repositorio (ex: repository/funcionario_repository.go)

func (e *EntregaRepository) BuscaTodasEntregasDoTenant(ctx context.Context, tenantId int32) ([]BuscaTodasEntregasDoTenantRow, error) {

    
    entregas, err := e.q.BuscaTodasEntregasDoTenant(ctx, tenantId)
    if err != nil {
        return nil, helper.TraduzErroPostgres(err)
    }
    
    return entregas, nil
}