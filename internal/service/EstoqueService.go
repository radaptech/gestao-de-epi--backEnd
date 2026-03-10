package service

import (
	"context"
	"math"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

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

type FiltroEstoqueAtual struct {
	FiltroPaginacao
}

type EstoqueAtualPaginado struct {
	EstoqueAtual []model.EstoqueTotalDto `json:"estoque_atual"`
	Total        int64                   `json:"total"`
	Pagina       int32                   `json:"pagina"`
	PaginaFinal  int32                   `json:"paginaFinal"`
}

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

	for _, quantidade := range quantidades {

		e := model.EstoqueTotalDto{
			Id:              int(quantidade.Idepi),
			Nome:            quantidade.NomeEpi,
			QuantidadeAtual: int(quantidade.QuantidadeTotal),
		}

		dto = append(dto, e)
	}

	var total int64
	if len(quantidades) > 0 {
		total = quantidades[0].TotalGeral
	}

	ultimaPagina := int32(math.Ceil(float64(total) / float64(p.Limit)))

	return EstoqueAtualPaginado{
		EstoqueAtual: dto,
		Total: total,
		Pagina: p.PaginaAtual,
		PaginaFinal: ultimaPagina,
	}, nil
}

type FiltroEstoqueSaldo struct {
	Fabricante string `form:"fabricante"`
	FiltroPaginacao
}

type EstoqueSaldoPaginado struct {
	EstoqueSaldo []model.EstoqueSaldoTotalDto `json:"saldo_Atual"`
	Total        int64                        `json:"total"`
	Pagina       int32                        `json:"pagina"`
	PaginaFinal  int32                        `json:"paginaFinal"`
}

func (e *EstoqueService) MostrarSaldoAtual(ctx context.Context, f FiltroEstoqueSaldo, tenantId int32) (EstoqueSaldoPaginado, error) {

	p := Paginacao(f.FiltroPaginacao)

	filtro := repository.ListarSaldoEstoqueParams{
		Limit:      p.Limit,
		Offset:     p.Offset,
		TenantID:   tenantId,
		Fabricante: pgtype.Text{String: f.Fabricante, Valid: f.Fabricante != ""},
	}

	saldo, err := e.repo.SaldoAtual(ctx, filtro)
	if err != nil {

		return EstoqueSaldoPaginado{}, err
	}

	dto := make([]model.EstoqueSaldoTotalDto, 0, len(saldo))

	for _, s := range saldo {

		sa := model.EstoqueSaldoTotalDto{
			Id:              int(s.Idepi),
			Nome:            s.NomeEpi,
			QuantidadeAtual: int(s.QuantidadeAtual),
			SaldoTotal:      decimal.NewFromFloat(s.SaldoAtual),
		}

		dto = append(dto, sa)
	}

	var total int64
	if len(saldo) > 0 {
		total = saldo[0].TotalGeral
	}

	ultimaPagina := int32(math.Ceil(float64(total) / float64(p.Limit)))

	return EstoqueSaldoPaginado{
		EstoqueSaldo: dto,
		Total:        total,
		Pagina:       p.PaginaAtual,
		PaginaFinal:  ultimaPagina,
	}, nil
}
