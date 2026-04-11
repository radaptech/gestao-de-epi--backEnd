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
	Idfornecedor       int32                   `json:"idfornecedor" binding:"required,gt=0"`
	Nota_fiscal_numero string                  `json:"nota_fiscal_numero" binding:"required,max=50"`
	Nota_fiscal_serie  string                  `json:"nota_fiscal_serie" binding:"required,max=20"`
	Data_emissao       configs.DataBr          `json:"data_emissao" binding:"required"`
	Id_user            int32                   `json:"-"`
	Itens              []EntradaEpiItemInserir `json:"itens" binding:"required,dive"` // "dive" valida cada item da lista
}

// EntradaEpiDto representa o retorno detalhado (Listagem)
type EntradaEpiDto struct {
	ID           int `json:"id"`
	IDEpi        int `json:"id_epi"`
	IDTamanho    int `json:"id_tamanho"`
	IDFornecedor int `json:"id_fornecedor"`

	// 2. Linha do tempo
	DataEntrada configs.DataBr `json:"data_entrada"`

	// 3. Valores e informações da operação
	Quantidade       int             `json:"quantidade"`
	QuantidadeAtual  int             `json:"quantidade_atual"`
	ValorUnitario    decimal.Decimal `json:"valor_unitario"`
	Lote             string          `json:"lote"`
	NotaFiscalNumero string          `json:"nota_fiscal_numero"`
	NotaFiscalSerie  string          `json:"nota_fiscal_serie"`
	UsuarioCriacao   string          `json:"usuario"`

	// 4. Objetos completos (Relacionamentos)
	Epi        EpiSimples        `json:"epi"`
	Tamanho    TamanhoSimples    `json:"tamanho"`
	Fornecedor FornecedorSimples `json:"fornecedor"`
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
	DataEntrada     configs.DataBr  `json:"data_entrada"`
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

type EpiSimples struct {
	ID         int    `json:"id"`
	Nome       string `json:"nome"`
	Fabricante string `json:"fabricante"`
	CA         string `json:"ca"`
}

type TamanhoSimples struct {
	ID      int    `json:"id"`
	Tamanho string `json:"tamanho"`
}

type FornecedorSimples struct {
	ID           int    `json:"id"`
	NomeFantasia string `json:"nome_fantasia"`
	RazaoSocial  string `json:"razao_social"`
}
