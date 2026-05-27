package model

import (
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
	"github.com/shopspring/decimal"
)

type EmpresaInserir struct {
   
    NomeFantasia string          `json:"nome_fantasia" binding:"required"`
    Cnpj         string          `json:"cnpj" binding:"cnpj,required,max=40"`
    Responsavel  string          `json:"responsavel" binding:"required,max=40"`
    Email        string          `json:"email" binding:"required,max=40"`
  
    Telefone     string          `json:"telefone" binding:"max=40"`
    
    Plano        string          `json:"plano" binding:"required"`
    Status       string          `json:"status" binding:"required"`
    Mensalidade  decimal.Decimal `json:"mensalidade" binding:"required"`
    Vencimento   configs.DataBr  `json:"vencimento" binding:"required"`
    
    
    Observacoes  string          `json:"observacoes" binding:"lte=150"`
    Subdominio   string          `json:"-"`
}
