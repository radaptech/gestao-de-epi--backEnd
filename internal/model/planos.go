package model

import "github.com/shopspring/decimal"

type Plano struct {
	ID                 int             `json:"id"`
	Nome               string          `json:"nome" binding:"required"`
	Mensalidade        decimal.Decimal `json:"mensalidade" binding:"required"`
	LimiteFuncionarios *int32          `json:"limite_funcionarios"`
	LimiteUsuarios     *int32          `json:"limite_usuarios"`
	LimiteEpis         *int32          `json:"limite_epis"`
	Status             string          `json:"status" binding:"required"`
	Descricao          string          `json:"descricao" binding:"required"`
}

type AtualizarPlanoParams struct {
	Nome               *string          `json:"nome"`
	Mensalidade        *decimal.Decimal `json:"mensalidade"` // ou pgtype.Numeric
	Descricao          *string          `json:"descricao"`
	LimiteFuncionarios *int32           `json:"limite_funcionarios"`
	LimiteUsuarios     *int32           `json:"limite_usuarios"`
	LimiteEpis         *int32           `json:"limite_epis"`
	Status             *string          `json:"status"`
	ID                 int32            `json:"id"`
}

type PlanoNome struct {
	Id   int
	Nome string
}
