package model

import (
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/shopspring/decimal"
)

type EntradaEpiInserir struct {
	ID_epi             int             `json:"id_epi" binding:"required,numeric"`
	Id_tamanho         int             `json:"id_tamanho" binding:"required,numeric"`
	Id_user            int             `json:"id_user" binding:"required,numeric"`
	Data_entrada       configs.DataBr  `json:"data_entrada" binding:"required"`
	Quantidade_Atual   int             `json:"quantidade_Atual" binding:"required,numeric,gt=0"`
	Quantidade         int             `json:"quantidade" binding:"required,numeric,gt=0"`
	DataFabricacao     configs.DataBr  `json:"data_fabricacao" binding:"required"`
	DataValidade       configs.DataBr  `json:"data_validade" binding:"required"`
	Lote               string          `json:"lote" binding:"required,numeric,max=6"`
	Id_fornecedor      int             `json:"id_fornecedor" binding:"required,max=50"`
	Nota_fiscal_serie  string          `json:"notaFiscalSerie" binding:"required,max=20,numeric"`
	Nota_fiscal_numero string          `json:"notaFiscalNumero" binding:"required,max=10,numeric"`
	ValorUnitario      decimal.Decimal `json:"valorUnitario" binding:"required"`
}

type EntradaEpiDto struct {
	ID                         int                 `json:"id"`
	UsuarioEntrada             RecuperaUserEntrada `json:"usuario"`
	Epi                        EpiDto              `json:"epi"`
	Data_entrada               configs.DataBr      `json:"data_entrada"`
	Quantidade                 int                 `json:"quantidade"`
	Quantidade_Atual           int                 `json:"quantidade_Atual"`
	Lote                       string              `json:"lote"`
	Fornecedor                 FornecedorDto       `json:"fornecedor"`
	Nota_fiscal_serie          string              `json:"notaFicalSerie"`
	Nota_fiscal_numero         string              `json:"notaFiscalNumero"`
	UsuarioEntradaCancelamento RecuperaUserEntrada `json:"usuario_Cancelamento"`
	ValorUnitario              decimal.Decimal     `json:"valor_unitario"`
}

type EntradaDashbord struct {
	Id              int             `json:"id"`
	IdEpi           int             `json:"idEpi"`
	IdTamanho       int             `json:"idTamanho"`
	QuantidadeAtual int             `json:"quantidadeAtual"`
	valorUnitario   decimal.Decimal `json:"valor_unitario"`
	Quantidade      int             `json:"quantidade"`
	DataEntrada     configs.DataBr  `json:"data_entrada"`
	Lote            string          `json:"lote"`
}
