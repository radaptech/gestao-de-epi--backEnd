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

// ExecInTx é um helper para rodar múltiplas operações na mesma transação.
// Isso evita que você precise passar *Queries manualmente em cada chamada de service.
func (e *EntregaRepository) ExecInTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := e.db.Begin(ctx)
	if err != nil {
		return helper.TraduzErroPostgres(err)
	}
	defer tx.Rollback(ctx)

	qtx := e.q.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err // O erro já deve vir traduzido da função interna
	}

	return tx.Commit(ctx)
}

// --- MÉTODOS DE ESCRITA (Geralmente usados dentro de transação) ---

func (e *EntregaRepository) AdicionarEntrega(ctx context.Context, qtx *Queries, args AddEntregaEpiParams) (int32, error) {
	id, err := qtx.AddEntregaEpi(ctx, args)
	if err != nil {
		return 0, helper.TraduzErroPostgres(err)
	}
	return id, nil
}

func (e *EntregaRepository) AdicionarEntregaItem(ctx context.Context, qtx *Queries, arg AddItemEntregueParams) (AddItemEntregueRow, error) {
	ids, err := qtx.AddItemEntregue(ctx, arg)
	if err != nil {
		return AddItemEntregueRow{}, helper.TraduzErroPostgres(err)
	}
	return ids, nil
}

func (e *EntregaRepository) AbaterEstoqueEntrada(ctx context.Context, qtx *Queries, args AbaterEstoqueLoteParams) (int64, error) {
	linhasAfetadas, err := qtx.AbaterEstoqueLote(ctx, args)
	if err != nil {
		return 0, helper.TraduzErroPostgres(err)
	}
	return linhasAfetadas, nil
}

// --- MÉTODOS DE LEITURA (Usam a conexão padrão do pool) ---

func (e *EntregaRepository) ListarEntregas(ctx context.Context, args ListarEntregasParams) ([]ListarEntregasRow, error) {
	entregas, err := e.q.ListarEntregas(ctx, args)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return entregas, nil
}

func (e *EntregaRepository) ListarLotesParaConsumo(ctx context.Context, args ListarLotesParaConsumoParams) ([]ListarLotesParaConsumoRow, error) {
	lotes, err := e.q.ListarLotesParaConsumo(ctx, args)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return lotes, nil
}

func (e *EntregaRepository) ListasEntregasPorMatricula(ctx context.Context, args ListarHistoricoEntregasPorMatriculaParams) ([]ListarHistoricoEntregasPorMatriculaRow, error) {
	entrega, err := e.q.ListarHistoricoEntregasPorMatricula(ctx, args)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return entrega, nil
}

func (e *EntregaRepository) BuscaEntregaDashbord(ctx context.Context, tenant int32) ([]EntregaDashbordRow, error) {
	entregas, err := e.q.EntregaDashbord(ctx, tenant)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return entregas, nil
}

// --- MÉTODOS DE CANCELAMENTO ---

func (e *EntregaRepository) Cancelar(ctx context.Context, qtx *Queries, args CancelarEntregaParams) (int32, error) {
	id, err := qtx.CancelarEntrega(ctx, args)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, helper.ErrNaoEncontrado
		}
		return 0, helper.TraduzErroPostgres(err)
	}
	return id, nil
}

func (e *EntregaRepository) CancelarEntregaItem(ctx context.Context, qtx *Queries, arg CancelaItemEntregueParams) ([]CancelaItemEntregueRow, error) {
	itemsCancelados, err := qtx.CancelaItemEntregue(ctx, arg)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return itemsCancelados, nil
}

func (e *EntregaRepository) ReporEstoqueEntrada(ctx context.Context, qtx *Queries, args ReporEstoqueLoteParams) (int64, error) {
	linhasAfetadas, err := qtx.ReporEstoqueLote(ctx, args)
	if err != nil {
		return 0, helper.TraduzErroPostgres(err)
	}
	return linhasAfetadas, nil
}

// BuscaEntregaItensDashbord implements [service.EntregaRepository].
func (e *EntregaRepository) BuscaEntregaItensDashbord(ctx context.Context, tenant int32) ([]EntregaItensDashbordRow, error) {
	
	entregaDash,err:= e.q.EntregaItensDashbord(ctx, tenant)
	if err != nil {
		return []EntregaItensDashbordRow{}, helper.TraduzErroPostgres(err)
	}

	return entregaDash, nil
}

// BuscaTodasEntregasDoTenant implements [service.EntregaRepository].
func (e *EntregaRepository) BuscaTodasEntregasDoTenant(ctx context.Context, tenantId int32) ([]BuscaTodasEntregasDoTenantRow, error) {
	
	entregasTenant, err:= e.q.BuscaTodasEntregasDoTenant(ctx, tenantId)
	if err != nil {

		return []BuscaTodasEntregasDoTenantRow{}, helper.TraduzErroPostgres(err)
	}

	return entregasTenant, nil
}

// ListarEntregasDisponiveis implements [service.EntregaRepository].
func (e *EntregaRepository) ListarEntregasDisponiveis(ctx context.Context, qtx *Queries, args ListarLotesParaConsumoParams) ([]ListarLotesParaConsumoRow, error) {
	
	listaLotes, err:= e.q.ListarLotesParaConsumo(ctx, args)
	if err != nil {

		return  []ListarLotesParaConsumoRow{}, helper.TraduzErroPostgres(err)
	}

	return listaLotes, nil
}

// ListarEpisEntreguesCancelados implements [service.EntregaRepository].
func (e *EntregaRepository) ListarEpisEntreguesCancelados(ctx context.Context, qtx *Queries, arg ListarItensEntregueCanceladosParams) ([]ListarItensEntregueCanceladosRow, error) {
	
	episEntreguesCanc, err:= e.q.ListarItensEntregueCancelados(ctx, arg)
	if err != nil {
		return []ListarItensEntregueCanceladosRow{}, helper.TraduzErroPostgres(err)
	}

	return  episEntreguesCanc, nil
}