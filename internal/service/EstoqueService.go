package service

import (
	"context"
	"math"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// --- INTERFACE DO REPOSITORY ---
type EstoqueRepository interface {
	SomaQuantidade(ctx context.Context, args repository.ListarEstoqueAtualParams) ([]repository.ListarEstoqueAtualRow, error)
	SaldoAtual(ctx context.Context, arg repository.ListarSaldoEstoqueParams) ([]repository.ListarSaldoEstoqueRow, error)
}

type EstoqueService struct {
	repo EstoqueRepository
}

func NewEstoqueService(r EstoqueRepository) *EstoqueService {
	return &EstoqueService{repo: r}
}

// --- ESTRUTURAS DE FILTRO E PAGINAÇÃO ---

type FiltroEstoqueAtual struct {
	FiltroPaginacao
}

type EstoqueAtualPaginado struct {
	EstoqueAtual []model.EstoqueTotalDto `json:"estoque_atual"`
	Total        int64                   `json:"total"`
	Pagina       int32                   `json:"pagina"`
	PaginaFinal  int32                   `json:"pagina_final"`
}

type FiltroEstoqueSaldo struct {
	Fabricante string `form:"fabricante"`
	FiltroPaginacao
}

type EstoqueSaldoPaginado struct {
	EstoqueSaldo []model.EstoqueSaldoTotalDto `json:"saldo_atual"`
	Total        int64                        `json:"total"`
	Pagina       int32                        `json:"pagina"`
	PaginaFinal  int32                        `json:"pagina_final"`
}

// --- MÉTODOS DO SERVICE ---

// MostrarQuantidadeTotais retorna o saldo quantitativo de cada EPI (Quantas unidades existem)
func (e *EstoqueService) MostrarQuantidadeTotais(ctx context.Context, f FiltroEstoqueAtual, tenantId int32) (EstoqueAtualPaginado, error) {
	p := Paginacao(f.FiltroPaginacao)

	filtro := repository.ListarEstoqueAtualParams{
		Limit:    p.Limit,
		Offset:   p.Offset,
		TenantID: tenantId,
	}

	quantidades, err := e.repo.SomaQuantidade(ctx, filtro)
	if err != nil {
		return EstoqueAtualPaginado{}, err
	}

	dto := make([]model.EstoqueTotalDto, 0, len(quantidades))
	for _, q := range quantidades {
		// Ajuste: Batendo com os campos da model.EstoqueTotalDto (IDEpi, NomeEpi, QuantidadeTotal)
		item := model.EstoqueTotalDto{
			IDEpi:           q.IDEpi,
			NomeEpi:         q.NomeEpi,
			QuantidadeTotal: q.QuantidadeTotal,
		}
		dto = append(dto, item)
	}

	var total int64
	if len(quantidades) > 0 {
		total = quantidades[0].TotalGeral
	}

	return EstoqueAtualPaginado{
		EstoqueAtual: dto,
		Total:        total,
		Pagina:       p.PaginaAtual,
		PaginaFinal:  int32(math.Ceil(float64(total) / float64(p.Limit))),
	}, nil
}

// MostrarSaldoAtual retorna o valor financeiro em estoque (Quantidade * Valor Unitário)
func (e *EstoqueService) MostrarSaldoAtual(ctx context.Context, f FiltroEstoqueSaldo, tenantId int32) (EstoqueSaldoPaginado, error) {
	p := Paginacao(f.FiltroPaginacao)

	filtro := repository.ListarSaldoEstoqueParams{
		Limit:      p.Limit,
		Offset:     p.Offset,
		TenantID:   tenantId,
		Fabricante: pgtype.Text{String: f.Fabricante, Valid: f.Fabricante != ""},
	}

	saldos, err := e.repo.SaldoAtual(ctx, filtro)
	if err != nil {
		return EstoqueSaldoPaginado{}, err
	}

	dto := make([]model.EstoqueSaldoTotalDto, 0, len(saldos))
	for _, s := range saldos {
		// Ajuste: Batendo com os campos da model.EstoqueSaldoTotalDto (IDEpi, NomeEpi, QuantidadeAtual, SaldoTotal)
		item := model.EstoqueSaldoTotalDto{
			IDEpi:           s.IDEpi,
			NomeEpi:         s.NomeEpi,
			QuantidadeAtual: s.QuantidadeAtual,
			SaldoTotal:      decimal.NewFromFloat(s.SaldoAtual),
		}
		dto = append(dto, item)
	}

	var total int64
	if len(saldos) > 0 {
		total = saldos[0].TotalGeral
	}

	return EstoqueSaldoPaginado{
		EstoqueSaldo: dto,
		Total:        total,
		Pagina:       p.PaginaAtual,
		PaginaFinal:  int32(math.Ceil(float64(total) / float64(p.Limit))),
	}, nil
}