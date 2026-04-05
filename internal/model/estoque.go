package model

import "github.com/shopspring/decimal"

// EstoqueTotalDto é usado na listagem simples de "Quanto eu tenho de cada EPI?"
type EstoqueTotalDto struct {
	IDEpi           int32  `json:"id_epi"`
	NomeEpi         string `json:"nome_epi"`
	QuantidadeTotal int64  `json:"quantidade_total"`
}

// EstoqueSaldoTotalDto é o DTO para relatórios de inventário e valor de mercado
type EstoqueSaldoTotalDto struct {
	IDEpi           int32           `json:"id_epi"`
	NomeEpi         string          `json:"nome_epi"`
	QuantidadeAtual int32           `json:"quantidade_atual"`
	SaldoTotal      decimal.Decimal `json:"saldo_total"`
}

// Dica: Adicione este DTO se precisar detalhar por tamanho no Frontend da Paloma
type EstoquePorTamanhoDto struct {
	IDEpi           int32  `json:"id_epi"`
	IDTamanho       int32  `json:"id_tamanho"`
	TamanhoNome     string `json:"tamanho_nome"`
	QuantidadeAtual int32  `json:"quantidade_atual"`
}