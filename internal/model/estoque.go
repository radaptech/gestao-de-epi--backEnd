package model

import "github.com/shopspring/decimal"

type EstoqueTotalDto struct {
	Id              int    `json:"id"`
	Nome            string `json:"nome"`
	QuantidadeAtual int    `json:"quantidade"`
}

type EstoqueSaldoTotalDto struct {
	Id              int             `json:"id"`
	Nome            string          `json:"nome"`
	QuantidadeAtual int             `json:"quantidade_atual"`
	SaldoTotal      decimal.Decimal `json:"saldo_total"`
}
