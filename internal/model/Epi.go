package model

import "github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"

type EpiInserir struct {
	Nome           string         `json:"nome" binding:"required"`
	Fabricante     string         `json:"fabricante" binding:"required,max=50"`
	CA             string         `json:"ca" binding:"required,numeric,min=1,max=10"`
	Descricao      string         `json:"descricao" binding:"lte=250"`
	DataValidadeCa configs.DataBr `json:"data_validade_ca" binding:"required"`
	Idtamanho      []int          `json:"id_tamanho" binding:"required,min=1"`
	IDprotecao     int            `json:"id_protecao" binding:"required,numeric"`
	AlertaMinimo   int            `json:"alerta_minimo" binding:"required,gte=0"`
}

type EpiDto struct {
	Id             int             `json:"id"`
	Nome           string          `json:"nome"`
	Fabricante     string          `json:"fabricante"`
	CA             string          `json:"CA"`
	Tamanho        []TamanhoDto    `json:"-"`
	Descricao      string          `json:"descricao"`
	DataValidadeCa configs.DataBr  `json:"validade_CA"`
	Protecao       TipoProtecaoDto `json:"protecao"`
	AlertaMinimo   int             `json:"alerta_minimo"`
}

type EpiDtoEstoque struct {
	Id             int             `json:"id"`
	Nome           string          `json:"nome"`
	Fabricante     string          `json:"fabricante"`
	CA             string          `json:"ca"`
	Descricao      string          `json:"descricao"`
	DataValidadeCa configs.DataBr  `json:"validade_ca"`
	AlertaMinimo   int             `json:"alerta_minimo"`
	Protecao       TipoProtecaoDto `json:"protecao"`
}

type EpiDashBord struct {
	Id           int    `json:"id"`
	Nome         string `json:"nome"`
	AlertaMinimo int    `json:"alerta_minimo"`
}

type UpdateEpiInput struct {
	Nome       *string         `json:"nome"`
	Fabricante *string         `json:"fabricante"`
	CA         *string         `json:"ca"`
	Descricao  *string         `json:"descricao"`
	ValidadeCa *configs.DataBr `json:"validadeCa"`
	Tamanhos   []int32         `json:"tamanhos"` // Novos IDs de tamanhos
}

type EpiResponse struct {
	Id             int             `json:"id"`
	Nome           string          `json:"nome"`
	Fabricante     string          `json:"fabricante"`
	CA             string          `json:"ca"`
	Descricao      string          `json:"descricao"`
	DataValidadeCa configs.DataBr  `json:"validade_ca"`
	AlertaMinimo   int             `json:"alerta_minimo"`
	Protecao       TipoProtecaoDto `json:"protecao"`
}
