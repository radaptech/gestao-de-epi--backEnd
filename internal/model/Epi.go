package model

import (
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
)

// EpiInserir representa o formulário de cadastro de um novo modelo de EPI
type EpiInserir struct {
	Nome           string         `json:"nome" binding:"required"`
	Fabricante     string         `json:"fabricante" binding:"required,max=100"`
	CA             string         `json:"ca" binding:"required,max=20"`
	Descricao      string         `json:"descricao" binding:"lte=250"`
	DataValidadeCa configs.DataBr `json:"data_validade_ca" binding:"required"`
	IdTamanho      []int32        `json:"id_tamanho" binding:"required,min=1"` // Lista de tamanhos permitidos para este EPI
	IDProtecao     int32          `json:"id_protecao" binding:"required,numeric"`
	AlertaMinimo   int32          `json:"alerta_minimo" binding:"required,gte=0"`
}

// EpiDto é o retorno completo usado na listagem de cadastros
type EpiDto struct {
	Id             int32           `json:"id"`
	Nome           string          `json:"nome"`
	Fabricante     string          `json:"fabricante"`
	CA             string          `json:"ca"`
	Tamanhos       []TamanhoDto    `json:"tamanhos"` // Slive de objetos Tamanho (id e nome)
	Descricao      string          `json:"descricao"`
	DataValidadeCa configs.DataBr  `json:"validade_ca"`
	Protecao       TipoProtecaoDto `json:"protecao"`
	AlertaMinimo   int32           `json:"alerta_minimo"`
}

type EpiDtoEntrega struct {
	Id             int32               `json:"id"`
	Nome           string              `json:"nome"`
	Fabricante     string              `json:"fabricante"`
	CA             string              `json:"ca"`
	Tamanhos       []TamanhoEntregaDto `json:"tamanhos"` // Slive de objetos Tamanho (id e nome)
	Descricao      string              `json:"descricao"`
	DataValidadeCa configs.DataBr      `json:"validade_ca"`
	Protecao       TipoProtecaoDto     `json:"protecao"`
	AlertaMinimo   int32               `json:"alerta_minimo"`
}
type EpiDtoDevolucao struct {
	Id             int32           `json:"id"`
	Nome           string          `json:"nome"`
	Fabricante     string          `json:"fabricante"`
	CA             string          `json:"ca"`
	Tamanhos       []TamanhoDto    `json:"tamanhos"` // Slive de objetos Tamanho (id e nome)
	Descricao      string          `json:"descricao"`
	DataValidadeCa configs.DataBr  `json:"validade_ca"`
	Protecao       TipoProtecaoDto `json:"protecao"`
	AlertaMinimo   int32           `json:"alerta_minimo"`
	SaldoAtual     int32           `json:"saldoAtual"`
}

// UpdateEpiInput usa ponteiros para permitir atualização parcial (PATCH)
type UpdateEpiInput struct {
	Nome           *string         `json:"nome"`
	Fabricante     *string         `json:"fabricante"`
	CA             *string         `json:"ca"`
	Descricao      *string         `json:"descricao"`
	DataValidadeCa *configs.DataBr `json:"validade_ca"`
	IdProtecao     *int32          `json:"id_protecao"`
	AlertaMinimo   *int32          `json:"alerta_minimo"`
	Tamanhos       []int32         `json:"tamanhos"` // Se enviado, você deve resetar os tamanhos no banco
}

// EpiResponse é a versão simplificada para compor outros DTOs (como o de Entrega)
type EpiResponse struct {
	Id             int32           `json:"id"`
	Nome           string          `json:"nome"`
	Fabricante     string          `json:"fabricante"`
	CA             string          `json:"ca"`
	Descricao      string          `json:"descricao"`
	DataValidadeCa configs.DataBr  `json:"validade_ca"`
	Protecao       TipoProtecaoDto `json:"protecao"`
	AlertaMinimo   int             `json:"alerta_minimo"`
}

// EpiDashBord é o resumo para os cards do dashboard
type EpiDashBord struct {
	Id           int32  `json:"id"`
	Nome         string `json:"nome"`
	AlertaMinimo int32  `json:"alerta_minimo"`
}

type EpiDtoEstoque struct {
	Id             int32           `json:"id"`
	Nome           string          `json:"nome"`
	Fabricante     string          `json:"fabricante"`
	Descricao      string          `json:"descricao"`
	DataValidadeCa configs.DataBr  `json:"validadeCa"`
	Ca             string          `json:"ca"`
	AlertaMinimo   int             `json:"alertaMinimo"`
	Protecao       TipoProtecaoDto `json:"protecao"`
}
