package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EntradaRepository struct {
	q  *Queries
	db *pgxpool.Pool
}

func NewEntradaRepository(pool *pgxpool.Pool) *EntradaRepository {
	return &EntradaRepository{
		q:  New(pool),
		db: pool,
	}
}

// ⚠️ MUDANÇA IMPORTANTE: Agora recebe os dados da NF e uma lista de ITENS
func (e *EntradaRepository) AdicionarCompleta(ctx context.Context, nfArgs CreateEntradaNFParams, itens []CreateEntradaEpiItemParams) error {
	// 1. Inicia a Transação
	tx, err := e.db.Begin(ctx)
	if err != nil {
		return helper.TraduzErroPostgres(err)
	}
	// 2. Garante o Rollback se algo der errado antes do Commit
	defer tx.Rollback(ctx)

	// 3. Cria uma instância de queries que usa a transação
	qtx := e.q.WithTx(tx)

	// 4. Salva o cabeçalho da Nota Fiscal e pega o ID gerado
	nfID, err := qtx.CreateEntradaNF(ctx, nfArgs)
	if err != nil {
		return helper.TraduzErroPostgres(err)
	}

	// 5. Loop para salvar cada item vinculado a essa NF
	for _, item := range itens {
		item.EntradaNfID = nfID // 🔑 Aqui vinculamos o item ao ID da NF que acabou de ser criada
		err := qtx.CreateEntradaEpiItem(ctx, item)
		if err != nil {
			return helper.TraduzErroPostgres(err)
		}
	}

	// 6. Finaliza a transação gravando tudo no banco
	if err := tx.Commit(ctx); err != nil {
		return helper.TraduzErroPostgres(err)
	}

	return nil
}

// Os outros métodos seguem o padrão, apenas ajustando os tipos gerados pelo sqlc
func (e *EntradaRepository) ListarEntradas(ctx context.Context, args ListarEntradasParams) ([]ListarEntradasRow, error) {
	entradas, err := e.q.ListarEntradas(ctx, args)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return entradas, nil
}

func (e *EntradaRepository) CancelarEntrada(ctx context.Context, args CancelarEntradaParams) (int64, error) {
	linhasAfetadas, err := e.q.CancelarEntrada(ctx, args)
	if err != nil {
		return 0, helper.TraduzErroPostgres(err)
	}
	return linhasAfetadas, nil
}

func (e *EntradaRepository) TotalEntradas(ctx context.Context, args ContarEntradasFiltradasParams) (int64, error) {
	total, err := e.q.ContarEntradasFiltradas(ctx, args)
	if err != nil {
		return 0, helper.TraduzErroPostgres(err)
	}
	return total, nil
}

func (e *EntradaRepository) BuscaEntradaDashbord(ctx context.Context, tenant int32) ([]EntradaDashbordRow, error) {
	entradas, err := e.q.EntradaDashbord(ctx, tenant)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return entradas, nil
}

func (e *EntradaRepository) EntradaEstoque(ctx context.Context, tenant int32) ([]EntradaEpiEstoqueRow, error) {
	entradas, err := e.q.EntradaEpiEstoque(ctx, tenant)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return entradas, nil
}