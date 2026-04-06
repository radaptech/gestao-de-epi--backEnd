package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// --- INTERFACE DO REPOSITORY ---
type EntradaRepository interface {
	// Agora precisamos de uma função que lide com a transação para NF + Itens
	AdicionarCompleta(ctx context.Context, nfArgs repository.CreateEntradaNFParams, itens []repository.CreateEntradaEpiItemParams) error
	ListarEntradas(ctx context.Context, args repository.ListarEntradasParams) ([]repository.ListarEntradasRow, error)
	CancelarEntrada(ctx context.Context, args repository.CancelarEntradaParams) (int64, error)
	TotalEntradas(ctx context.Context, args repository.ContarEntradasFiltradasParams) (int64, error)
	BuscaEntradaDashbord(ctx context.Context, tenant int32) ([]repository.EntradaDashbordRow, error)
	EntradaEstoque(ctx context.Context, tenant int32) ([]repository.EntradaEpiEstoqueRow, error)
}

type EntradaService struct {
	repo EntradaRepository
	db   *pgxpool.Pool
}

func NewEntradaService(e EntradaRepository, db *pgxpool.Pool) *EntradaService {
	return &EntradaService{repo: e, db: db}
}

// Adicionar: Processa a entrada de uma Nota Fiscal e todos os seus EPIs vinculados
func (e *EntradaService) Adicionar(ctx context.Context, input model.EntradaEpiInserir, tenantID int32) error {
	// 1. Validação de Datas
	hoje := time.Now().Truncate(24 * time.Hour)
	if input.Data_emissao.Time().After(hoje) {
		return fmt.Errorf("data de emissão não pode ser futura")
	}

	// 2. Prepara os parâmetros da Nota Fiscal (Cabeçalho)
	nfArgs := repository.CreateEntradaNFParams{
		TenantID:         tenantID,
		Idfornecedor:       input.Idfornecedor,
		NotaFiscalNumero: strings.TrimSpace(input.Nota_fiscal_numero),
		NotaFiscalSerie:  pgtype.Text{String: strings.TrimSpace(input.Nota_fiscal_serie), Valid: true},
		DataEmissao:      pgtype.Date{Time: input.Data_emissao.Time(), Valid: true},
		IDUsuarioCriacao: int32(input.Id_user), // Auditoria da Nota
	}

	// 3. Prepara a lista de Itens (EPIs no estoque)
	var itensParams []repository.CreateEntradaEpiItemParams
	for _, item := range input.Itens {
		// Valida se validade é menor que fabricação
		if item.DataValidade.Time().Before(item.DataFabricacao.Time()) {
			return helper.ErrDataMenorValidade
		}

		var vm pgtype.Numeric
		vm.Scan(item.ValorUnitario.String())

		itensParams = append(itensParams, repository.CreateEntradaEpiItemParams{
			TenantID:         tenantID,
			IDEpi:            int32(item.ID_epi),
			IDTamanho:        int32(item.Id_tamanho),
			Quantidade:       int32(item.Quantidade),
			QuantidadeAtual:  int32(item.Quantidade), // Inicialmente igual ao total
			DataFabricacao:   pgtype.Date{Time: item.DataFabricacao.Time(), Valid: true},
			DataValidade:     pgtype.Date{Time: item.DataValidade.Time(), Valid: true},
			Lote:             strings.TrimSpace(item.Lote),
			ValorUnitario:    vm,
			IDUsuarioCriacao: int32(input.Id_user), // Auditoria do Item
		})
	}

	// 4. Chama o repo que executa tudo em uma Transação SQL
	return e.repo.AdicionarCompleta(ctx, nfArgs, itensParams)
}

type FiltroEntradas struct {
	Canceladas bool           `form:"canceladas"`
	EpiID      int32          `form:"epi_id"`
	EntradaID  int32          `form:"entrada_id"`
	DataInicio configs.DataBr `form:"data_inicio"` // O Gin tentará converter a string para seu tipo Custom
	DataFim    configs.DataBr `form:"data_fim"`
	NotaFiscal string         `form:"nota_fiscal"`
	Pagina     int32          `form:"pagina"`
	Quantidade int32          `form:"quantidade"`
}

type EntradaPaginada struct {
	Entradas    []model.EntradaEpiDto `json:"entradas"`
	Total       int64                 `json:"total"`
	Pagina      int32                 `json:"pagina"`
	PaginaFinal int32                 `json:"pagina_final"`
}

// ListarEntradas: Busca paginada com filtros de EPI, Data e Nota Fiscal
func (e *EntradaService) ListarEntradas(ctx context.Context, f FiltroEntradas, tenantId int32) (EntradaPaginada, error) {
	limit := f.Quantidade
	if limit <= 0 {
		limit = 10
	}
	paginaAtual := f.Pagina
	if paginaAtual <= 0 {
		paginaAtual = 1
	}
	offset := (paginaAtual - 1) * limit

	// Filtro mapeado para o sqlc
	filtro := repository.ListarEntradasParams{
		Canceladas: f.Canceladas,
		IDEpi:      pgtype.Int4{Int32: f.EpiID, Valid: f.EpiID > 0},
		IDEntrada:  pgtype.Int4{Int32: f.EntradaID, Valid: f.EntradaID > 0},
		DataInicio: pgtype.Date{Time: f.DataInicio.Time(), Valid: !f.DataInicio.IsZero()},
		DataFim:    pgtype.Date{Time: f.DataFim.Time(), Valid: !f.DataFim.IsZero()},
		NotaFiscal: pgtype.Text{String: f.NotaFiscal, Valid: f.NotaFiscal != ""},
		Limit:      limit,
		Offset:     offset,
		TenantID:   tenantId,
	}

	entradas, err := e.repo.ListarEntradas(ctx, filtro)
	if err != nil {
		return EntradaPaginada{}, err
	}

	dto := make([]model.EntradaEpiDto, 0, len(entradas))
	for _, ent := range entradas {
		var valorDecimal decimal.Decimal
		if fVal, err := ent.ValorUnitario.Float64Value(); err == nil {
			valorDecimal = decimal.NewFromFloat(fVal.Float64)
		}

		dto = append(dto, model.EntradaEpiDto{
			ID: int(ent.ID),
			IDEpi: int(ent.ID),
			IDTamanho: int(ent.IDTamanho),
			IDFornecedor: int(ent.Idfornecedor),
			DataEntrada: *configs.NewDataBrPtr(ent.DataEntrada.Time),
			Quantidade: int(ent.Quantidade),
			QuantidadeAtual: int(ent.QuantidadeAtual),
			ValorUnitario: valorDecimal,
			Lote: ent.Lote,
			NotaFiscalNumero: ent.NotaFiscalNumero,
			NotaFiscalSerie: ent.NotaFiscalSerie.String,
			Epi: model.EpiSimples{
				ID: int(ent.IDEpi),
				Nome: ent.EpiNome,
				Fabricante: ent.Fabricante,
				CA: ent.Ca,
			},
			Tamanho: model.TamanhoSimples{
				ID: int(ent.IDTamanho),
				Tamanho: ent.TamanhoNome,
			},
			Fornecedor: model.FornecedorSimples{
				ID: int(ent.Idfornecedor),
				NomeFantasia: ent.NomeFantasia,
				RazaoSocial: ent.RazaoSocial,
			},

		})
	}

	total, _ := e.repo.TotalEntradas(ctx, repository.ContarEntradasFiltradasParams{
		Canceladas: filtro.Canceladas,
		IDEpi:      filtro.IDEpi,
		DataInicio: filtro.DataInicio,
		NotaFiscal: filtro.NotaFiscal,
		TenantID:   tenantId,
	})

	return EntradaPaginada{
		Entradas:    dto,
		Total:       total,
		Pagina:      paginaAtual,
		PaginaFinal: int32(math.Ceil(float64(total) / float64(limit))),
	}, nil
}

// CancelarEntrada: Soft delete que registra quem cancelou
func (e *EntradaService) CancelarEntrada(ctx context.Context, id, idUser, tenantId int) (int64, error) {
	if id <= 0 {
		return 0, helper.ErrId
	}

	arg := repository.CancelarEntradaParams{
		ID:                    int32(id),
		IDUsuarioCancelamento: pgtype.Int4{Int32: int32(idUser), Valid: true},
		TenantID:              int32(tenantId),
	}

	return e.repo.CancelarEntrada(ctx, arg)
}

// BuscaEntradaEstoque: Usado para o almoxarife selecionar o lote na hora da entrega
func (e *EntradaService) BuscaEntradaEstoque(ctx context.Context, tenantId int32) ([]model.EntradaEstoqueDto, error) {
	entradas, err := e.repo.EntradaEstoque(ctx, tenantId)
	if err != nil {
		return []model.EntradaEstoqueDto{}, err
	}

	dto := make([]model.EntradaEstoqueDto, 0, len(entradas))
	for _, ee := range entradas {
		var valorDecimal decimal.Decimal
		if fVal, err := ee.ValorUnitario.Float64Value(); err == nil {
			valorDecimal = decimal.NewFromFloat(fVal.Float64)
		}

		dto = append(dto, model.EntradaEstoqueDto{
			Id:              int(ee.ID),
			Lote:            ee.Lote,
			Quantidade:      int(ee.QuantidadeInicial),
			QuantidadeAtual: int(ee.QuantidadeAtual),
			ValorUnitario:   valorDecimal,
			DataValidade:    configs.DataBr(ee.DataValidade.Time),
			Tamanho:         model.TamanhoDto{ID: int(ee.IDTamanho), Tamanho: ee.TamanhoNome},
			Epi: model.EpiDtoEstoque{
				Id: ee.IDEpi, Nome: ee.EpiNome, Fabricante: ee.Fabricante,
				Ca: ee.Ca, Descricao: ee.Descricao, DataValidadeCa: configs.DataBr(ee.DataValidade.Time),
				AlertaMinimo: int(ee.AlertaMinimo),
				Protecao:     model.TipoProtecaoDto{ID: int64(ee.Idtipoprotecao), Nome: ee.ProtecaoNome},
			},
		})
	}
	return dto, nil
}

// EntradaDashbordBusca retorna um resumo das entradas para popular os gráficos do Dashboard
func (e *EntradaService) EntradaDashbordBusca(ctx context.Context, tenantId int32) ([]model.EntradaDashbord, error) {
	// Busca os dados brutos do repositório (referente aos itens de entrada)
	entradas, err := e.repo.BuscaEntradaDashbord(ctx, tenantId)
	if err != nil {
		// Retorna slice vazia em vez de nil para evitar problemas no JSON do Frontend
		return []model.EntradaDashbord{}, err
	}

	dto := make([]model.EntradaDashbord, 0, len(entradas))

	for _, ee := range entradas {
		// Conversão segura do tipo Numeric do Postgres para Decimal do Go
		var valorDecimal decimal.Decimal
		if fVal, err := ee.ValorUnitario.Float64Value(); err == nil {
			valorDecimal = decimal.NewFromFloat(fVal.Float64)
		}

		// Mapeia para o DTO do Dashboard (Note o uso de int32 vindo do sqlc)
		d := model.EntradaDashbord{
			Id:              int(ee.ID),
			IdEpi:           int(ee.IDEpi),
			IdTamanho:       int(ee.IDTamanho),
			QuantidadeAtual: int(ee.QuantidadeAtual),
			ValorUnitario:   valorDecimal,
			Quantidade:      int(ee.Quantidade),
			// Usa o helper para garantir o ponteiro da data formatada
			DataEntrada: configs.DataBr(ee.DataEntrada.Time),
			Lote:        ee.Lote,
		}

		dto = append(dto, d)
	}

	return dto, nil
}
