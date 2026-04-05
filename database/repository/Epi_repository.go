package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EpiRepository struct {
	q  *Queries
	db *pgxpool.Pool
}

func NewEpiRepository(pool *pgxpool.Pool) *EpiRepository {
	return &EpiRepository{
		q:  New(pool),
		db: pool,
	}
}

// ExecInTx facilita rodar operações de escrita (AddEpi + AddTamanhos) na mesma transação
func (e *EpiRepository) ExecInTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := e.db.Begin(ctx)
	if err != nil {
		return helper.TraduzErroPostgres(err)
	}
	defer tx.Rollback(ctx)

	qtx := e.q.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// --- MÉTODOS DE ESCRITA (Suportam Transação) ---

func (e *EpiRepository) Adicionar(ctx context.Context, qtx *Queries, epi AddEpiParams) (int32, error) {
	id, err := qtx.AddEpi(ctx, epi)
	if err != nil {
		return 0, helper.TraduzErroPostgres(err)
	}
	return id, nil
}

func (e *EpiRepository) AdicionarTamanho(ctx context.Context, qtx *Queries, arg AddEpiTamanhoParams) error {
	err := qtx.AddEpiTamanho(ctx, arg)
	if err != nil {
		return helper.TraduzErroPostgres(err)
	}
	return nil
}

func (e *EpiRepository) CancelarEpi(ctx context.Context, qtx *Queries, arg DeletarEpiParams) (int64, error) {
	linhasAfetadas, err := qtx.DeletarEpi(ctx, arg)
	if err != nil {
		return 0, helper.TraduzErroPostgres(err)
	}
	return linhasAfetadas, nil
}

// --- MÉTODOS DE LEITURA (Usam o Pool padrão) ---

func (e *EpiRepository) ListarEpi(ctx context.Context, arg BuscarEpiParams) (BuscarEpiRow, error) {
	epi, err := e.q.BuscarEpi(ctx, arg)
	if err != nil {
		// Retornar a struct vazia e o erro traduzido
		return BuscarEpiRow{}, helper.TraduzErroPostgres(err)
	}
	return epi, nil
}

func (e *EpiRepository) ListarEpis(ctx context.Context, args BuscarTodosEpisPaginadoParams) ([]BuscarTodosEpisPaginadoRow, error) {
	epis, err := e.q.BuscarTodosEpisPaginado(ctx, args)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return epis, nil
}

func (e *EpiRepository) AtualizaEpi(ctx context.Context, epi UpdateEpiCampoParams) (int64, error) {
	linhasAfetadas, err := e.q.UpdateEpiCampo(ctx, epi)
	if err != nil {
		return 0, helper.TraduzErroPostgres(err)
	}
	return linhasAfetadas, nil
}

func (e *EpiRepository) BuscaEpiDashbord(ctx context.Context, tenant int32) ([]BuscaEpiDashbordRow, error) {
	epis, err := e.q.BuscaEpiDashbord(ctx, tenant)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return epis, nil
}

func (e *EpiRepository) BuscaEpiTenant(ctx context.Context, tenant int32) ([]BuscaTodosItensEntreguesDoTenantRow, error) {
	// Chamada da "Chave Mestra" refatorada anteriormente
	epi, err := e.q.BuscaTodosItensEntreguesDoTenant(ctx, tenant)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return epi, nil
}