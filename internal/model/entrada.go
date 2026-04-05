package model

import (
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/shopspring/decimal"
)

// EntradaEpiItemInserir representa cada produto dentro de uma nota
type EntradaEpiItemInserir struct {
	ID_epi         int             `json:"id_epi" binding:"required,numeric"`
	Id_tamanho     int             `json:"id_tamanho" binding:"required,numeric"`
	Quantidade     int             `json:"quantidade" binding:"required,numeric,gt=0"`
	DataFabricacao configs.DataBr  `json:"data_fabricacao" binding:"required"`
	DataValidade   configs.DataBr  `json:"data_validade" binding:"required"`
	Lote           string          `json:"lote" binding:"required,max=50"`
	ValorUnitario  decimal.Decimal `json:"valor_unitario" binding:"required"`
}

// EntradaEpiInserir é o DTO Mestre (O que o Frontend envia)
type EntradaEpiInserir struct {
	Fornecedor         string                  `json:"fornecedor" binding:"required,max=100"`
	Nota_fiscal_numero string                  `json:"nota_fiscal_numero" binding:"required,max=50"`
	Nota_fiscal_serie  string                  `json:"nota_fiscal_serie" binding:"default=1"`
	Data_emissao       configs.DataBr          `json:"data_emissao" binding:"required"`
	Id_user            int32                   `json:"idUser"`
	Itens              []EntradaEpiItemInserir `json:"itens" binding:"required,dive"` // "dive" valida cada item da lista
}

// EntradaEpiDto representa o retorno detalhado (Listagem)
type EntradaEpiDto struct {
	ID                 int             `json:"id"`
	EpiNome            string          `json:"epi_nome"`
	TamanhoNome        string          `json:"tamanho_nome"`
	Quantidade         int             `json:"quantidade"`
	Quantidade_Atual   int             `json:"quantidade_atual"`
	Lote               string          `json:"lote"`
	Fornecedor         string          `json:"fornecedor"`
	Nota_fiscal_numero string          `json:"nota_fiscal_numero"`
	Data_entrada       configs.DataBr  `json:"data_entrada"`
	ValorUnitario      decimal.Decimal `json:"valor_unitario"`
	CanceladaEm        *configs.DataBr `json:"cancelada_em,omitempty"`
}

// EntradaEstoqueDto (Usado para o almoxarife selecionar o lote na hora da entrega)
type EntradaEstoqueDto struct {
	Id              int             `json:"id"`
	Lote            string          `json:"lote"`
	Quantidade      int             `json:"quantidade_inicial"`
	QuantidadeAtual int             `json:"quantidade_atual"`
	ValorUnitario   decimal.Decimal `json:"valor_unitario"`
	DataValidade    configs.DataBr  `json:"data_validade"`
	Tamanho         TamanhoDto      `json:"tamanho"`
	Epi             EpiDtoEstoque   `json:"epi"`
}

type EntradaDashbord struct {
	Id              int
	IdEpi           int
	IdTamanho       int
	QuantidadeAtual int
	ValorUnitario   decimal.Decimal
	Quantidade      int
	// Usa o helper para garantir o ponteiro da data formatada
	DataEntrada configs.DataBr
	Lote        string
}
