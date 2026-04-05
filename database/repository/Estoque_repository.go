package repository

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EstoqueRepository struct {
	q  *Queries
	db *pgxpool.Pool
}

func NewEstoqueRepository(pool *pgxpool.Pool) *EstoqueRepository {
	return &EstoqueRepository{
		q:  New(pool),
		db: pool,
	}
}

// SomaQuantidade retorna a visão geral de quantos itens existem por cada EPI (agrupado)
func (e *EstoqueRepository) SomaQuantidade(ctx context.Context, args ListarEstoqueAtualParams) ([]ListarEstoqueAtualRow, error) {
	quantidades, err := e.q.ListarEstoqueAtual(ctx, args)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return quantidades, nil
}

// SaldoAtual retorna o valor financeiro e a quantidade física (geralmente usado em relatórios)
func (e *EstoqueRepository) SaldoAtual(ctx context.Context, arg ListarSaldoEstoqueParams) ([]ListarSaldoEstoqueRow, error) {
	saldo, err := e.q.ListarSaldoEstoque(ctx, arg)
	if err != nil {
		return nil, helper.TraduzErroPostgres(err)
	}
	return saldo, nil
}

// DICA: Como agora você tem a lógica de "AdicionarEntregaItem" no outro repository,
// você pode usar o e.q.AbaterEstoqueLote dentro de uma transação para manter
// o estoque sempre atualizado em tempo real.