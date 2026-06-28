package model

import (
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
)

type EmpresaInserir struct {
	NomeFantasia string `json:"nome_fantasia" binding:"required"`
	Cnpj         string `json:"cnpj" binding:"cnpj,required,max=40"`
	Responsavel  string `json:"responsavel" binding:"required,max=40"`
	Email        string `json:"email" binding:"required,max=40"`

	Telefone string `json:"telefone" binding:"max=40"`

	Plano      string         `json:"plano" binding:"required"`
	Status     string         `json:"status" binding:"required"`
	Vencimento configs.DataBr `json:"vencimento" binding:"required"`

	Observacoes string `json:"observacoes" binding:"lte=150"`
	Subdominio  string `json:"-"`
}

type ResumoDashboard struct {
	EmpresasAtivas     int     `json:"empresasAtivas"`
	EmpresasBloqueadas int     `json:"empresasBloqueadas"`
	EmpresasEmTeste    int     `json:"empresasEmTeste"`
	TotalEmpresas      int     `json:"totalEmpresas"`
	TotalFuncionarios  int     `json:"totalFuncionarios"`
	TotalEpis          int     `json:"totalEpis"`
	TotalEntregas      int     `json:"totalEntregas"`
	ReceitaMensal      float64 `json:"receitaMensal"`
}

type EmpresaRecente struct {
	ID           int     `json:"id"`
	Nome         string  `json:"nome"`
	Subdominio   string  `json:"subdominio"`
	Responsavel  string  `json:"responsavel"`
	Status       string  `json:"status"`
	Plano        string  `json:"plano"`
	Funcionarios int     `json:"funcionarios"`
	Epis         int     `json:"epis"`
	Mensalidade  float64 `json:"mensalidade"`
}

type Empresa struct {
	ID           int64          `json:"id"`
	Nome         string         `json:"nome"`
	CNPJ         string         `json:"cnpj"`
	Responsavel  string         `json:"responsavel"`
	Email        string         `json:"email"`
	Telefone     string         `json:"telefone"`
	Plano        string         `json:"plano"`
	Funcionarios int            `json:"funcionarios"`
	EPIs         int            `json:"epis"`
	Mensalidade  float64        `json:"mensalidade"`
	Vencimento   configs.DataBr `json:"vencimento"`
	Status       string         `json:"status"`
	Observacoes  string         `json:"observacoes"`
}

type EditarEmpresaRequest struct {
	Nome        string         `json:"nome" binding:"required"`
	Cnpj        string         `json:"cnpj"`
	Responsavel string         `json:"responsavel" binding:"required"`
	Email       string         `github:"email" json:"email" binding:"required,email"`
	Telefone    string         `json:"telefone"`
	PlanoID     int64          `json:"planoId" binding:"required"`
	Vencimento  configs.DataBr `json:"vencimento"`
	Status      string         `json:"status" binding:"required"`
	Observacoes string         `json:"observacoes"`
}
