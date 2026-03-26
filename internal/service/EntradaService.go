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
	"github.com/shopspring/decimal"
)

type EntradaRepository interface {
	Adicionar(ctx context.Context, args repository.AddEntradaEpiParams) error
	ListarEntradas(ctx context.Context, args repository.ListarEntradasParams) ([]repository.ListarEntradasRow, error)
	CancelarEntrada(ctx context.Context, args repository.CancelarEntradaParams) (int64, error)
	TotalEntradas(ctx context.Context, args repository.ContarEntradasFiltradasParams) (int64, error)
	BuscaEntradaDashbord(ctx context.Context, tenant int32) ([]repository.EntradaDashbordRow, error)
}

type EntradaService struct {
	repo EntradaRepository
}

func NewEntradaService(e EntradaRepository) *EntradaService {

	return &EntradaService{repo: e}
}

func (e *EntradaService) Adicionar(ctx context.Context, model model.EntradaEpiInserir, tenantID int32) error {

	//data de entrada menor que a atual
	hoje := time.Now().Truncate(24 * time.Hour)
	if model.Data_entrada.Time().Truncate(24 * time.Hour).Before(hoje) {

		return helper.ErrDataMenor
	}
	//data de validade igual a de fabricacao
	if model.DataValidade.Time().Equal(model.DataFabricacao.Time()) {

		return helper.ErrDataIgual
	}
	//data de validade menor a de fabricacao
	if model.DataValidade.Time().Before(model.DataFabricacao.Time()) {
		return helper.ErrDataMenorValidade
	}

	model.Lote = strings.TrimSpace(model.Lote)

	model.Nota_fiscal_numero = strings.TrimSpace(model.Nota_fiscal_numero)
	model.Nota_fiscal_serie = strings.TrimSpace(model.Nota_fiscal_serie)

	var vm pgtype.Numeric
	err := vm.Scan(model.ValorUnitario.String())
	if err != nil {
		return err
	}

	err = e.repo.Adicionar(ctx, repository.AddEntradaEpiParams{
		Idepi:            int32(model.ID_epi),
		Idtamanho:        int32(model.Id_tamanho),
		DataEntrada:      pgtype.Date{Time: model.Data_entrada.Time(), Valid: true},
		Quantidade:       int32(model.Quantidade),
		Quantidadeatual:  int32(model.Quantidade_Atual),
		DataFabricacao:   pgtype.Date{Time: model.DataFabricacao.Time(), Valid: true},
		DataValidade:     pgtype.Date{Time: model.DataValidade.Time(), Valid: true},
		Idfornecedor:     int32(model.Id_fornecedor),
		Lote:             model.Lote,
		ValorUnitario:    vm,
		NotaFiscalNumero: model.Nota_fiscal_numero,
		NotaFiscalSerie:  pgtype.Text{String: model.Nota_fiscal_serie, Valid: true},
		IDUsuarioCriacao: pgtype.Int4{Int32: int32(model.Id_user), Valid: true},
		TenantID:         tenantID,
	})
	if err != nil {

		return err
	}

	return nil
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

func (e *EntradaService) ListarEntradas(ctx context.Context, f FiltroEntradas, tenatId int32) (EntradaPaginada, error) {

	limit := f.Quantidade
	if limit <= 0 {
		limit = 1
	}
	paginaAtual := f.Pagina
	if paginaAtual <= 0 {
		paginaAtual = 1
	}
	offset := max((paginaAtual-1)*limit, 0)

	filtro := repository.ListarEntradasParams{
		Canceladas: f.Canceladas,
		IDEpi:      pgtype.Int4{Int32: f.EpiID, Valid: f.EpiID > 0},
		IDEntrada:  pgtype.Int4{Int32: f.EntradaID, Valid: f.EntradaID > 0},
		DataInicio: pgtype.Date{Time: f.DataInicio.Time(), Valid: !f.DataInicio.IsZero()},
		DataFim:    pgtype.Date{Time: f.DataFim.Time(), Valid: !f.DataFim.IsZero()},
		NotaFiscal: pgtype.Text{String: f.NotaFiscal, Valid: f.NotaFiscal != ""},
		Limit:      limit,
		Offset:     offset,
		TenantID:   tenatId,
	}

	entradas, err := e.repo.ListarEntradas(ctx, filtro)
	if err != nil {

		return EntradaPaginada{}, err
	}

	dto := make([]model.EntradaEpiDto, 0, len(entradas))

	for _, entrada := range entradas {

		var valorDecimal decimal.Decimal
		if fVal, err := entrada.ValorUnitario.Float64Value(); err == nil {
			valorDecimal = decimal.NewFromFloat(fVal.Float64)
		}

		var idUsuario int
		if entrada.IDUsuarioCriacao.Valid {
			idUsuario = int(entrada.IDUsuarioCriacao.Int32)
		} else {
			idUsuario = 0 // ou algum valor padrão
		}

		var idUsuarioCancelamento int
		if entrada.IDUsuarioCriacaoCancelamento.Valid {
			idUsuarioCancelamento = int(entrada.IDUsuarioCriacaoCancelamento.Int32)
		} else {
			idUsuarioCancelamento = 0
		}

		nomeCriacao := "" // Valor padrão se vier nulo do banco
		if entrada.UsuarioCriacaoNome.Valid {
			nomeCriacao = entrada.UsuarioCriacaoNome.String
		}

		// 2. Tratamento para Usuario de Cancelamento (se houver essa coluna)
		nomeCancelamento := ""
		if entrada.UsuarioCancelamentoNome.Valid {
			nomeCancelamento = entrada.UsuarioCancelamentoNome.String
		}

		e := model.EntradaEpiDto{
			ID: int(entrada.ID),
			Epi: model.EpiDto{
				Id:         int(entrada.Idepi),
				Nome:       entrada.EpiNome,
				Fabricante: entrada.Fabricante,
				CA:         entrada.Ca,
				Tamanho: []model.TamanhoDto{
					{
						ID:      int(entrada.Idtamanho),
						Tamanho: entrada.TamanhoNome,
					},
				},
				Descricao:      entrada.EpiDescricao,
				DataValidadeCa: configs.DataBr(entrada.ValidadeCa.Time),
				Protecao: model.TipoProtecaoDto{
					ID:   int64(entrada.Idtipoprotecao),
					Nome: entrada.ProtecaoNome,
				},
			},
			Data_entrada:       *configs.NewDataBrPtr(entrada.DataEntrada.Time),
			Quantidade:         int(entrada.Quantidade),
			Quantidade_Atual:   int(entrada.Quantidadeatual),
			Lote:               entrada.Lote,
			Fornecedor:         model.FornecedorDto{
				ID: int(entrada.Idfornecedor),
				RazaoSocial: entrada.RazaoSocial,
				NomeFantasia: entrada.NomeFantasia,
				CNPJ: entrada.Cnpj,
				InscricaoEstadual: entrada.InscricaoEstadual,
			},
			Nota_fiscal_serie:  entrada.NotaFiscalSerie.String,
			Nota_fiscal_numero: entrada.NotaFiscalNumero,
			ValorUnitario:      valorDecimal,
			UsuarioEntrada: model.RecuperaUserEntrada{
				Id:   idUsuario,
				Nome: nomeCriacao,
			},
			UsuarioEntradaCancelamento: model.RecuperaUserEntrada{
				Id:   idUsuarioCancelamento,
				Nome: nomeCancelamento,
			},
		}

		dto = append(dto, e)
	}

	total, err := e.repo.TotalEntradas(ctx, repository.ContarEntradasFiltradasParams{
		Canceladas: filtro.Canceladas,
		IDEpi:      filtro.IDEpi,
		IDEntrada:  filtro.IDEntrada,
		DataInicio: filtro.DataInicio,
		DataFim:    filtro.DataFim,
		NotaFiscal: filtro.NotaFiscal,
		TenantID:   tenatId,
	})
	if err != nil {
		return EntradaPaginada{}, err
	}

	paginaFinal := int32(math.Ceil(float64(total) / float64(limit)))

	return EntradaPaginada{
		Entradas:    dto,
		Total:       total,
		Pagina:      paginaAtual,
		PaginaFinal: paginaFinal,
	}, nil
}

func (e *EntradaService) CancelarEntrada(ctx context.Context, id, idUser, tenantid int) (int64, error) {

	if id <= 0 {

		return 0, helper.ErrId
	}

	arg := repository.CancelarEntradaParams{
		ID:                           int32(id),
		IDUsuarioCriacaoCancelamento: pgtype.Int4{Int32: int32(idUser), Valid: true},
		TenantID:                     int32(tenantid),
	}
	linhasAfetadas, err := e.repo.CancelarEntrada(ctx, arg)
	if err != nil {

		return 0, fmt.Errorf("erro técnico ao cancelar: %w", err)
	}

	if linhasAfetadas == 0 {

		return 0, helper.ErrNaoEncontrado
	}

	return linhasAfetadas, nil
}


func (e *EntradaService) EntradaDashbordBusca(ctx context.Context, tenantId int32)([]model.EntradaDashbord, error){


	entradas, err:= e.repo.BuscaEntradaDashbord(ctx, tenantId)
	if err != nil {

		return  []model.EntradaDashbord{}, err
	}


	dto := make([]model.EntradaDashbord, 0, len(entradas))

	for _, ee:= range entradas {

		var valorDecimal decimal.Decimal
		if fVal, err := ee.ValorUnitario.Float64Value(); err == nil {
			valorDecimal = decimal.NewFromFloat(fVal.Float64)
		}
		ee:= model.EntradaDashbord {

			Id: int(ee.ID),
			IdEpi: int(ee.Idepi),
			IdTamanho: int(ee.Idtamanho),
			QuantidadeAtual: int(ee.Quantidadeatual),
			ValorUnitario: valorDecimal,
			Quantidade: int(ee.Quantidade),
			DataEntrada: *configs.NewDataBrPtr(ee.DataEntrada.Time),
			Lote: ee.Lote,
		}

		dto = append(dto, ee)
	}


	return dto, err
}