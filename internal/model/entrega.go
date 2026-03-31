package model

import (
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
)

type ItemParaInserir struct {
	ID_epi     int64 `json:"id_epi" binding:"required"`
	ID_tamanho int64 `json:"id_tamanho" binding:"required"`
	Quantidade int   `json:"quantidade" binding:"required,numeric,gt=0"`
}

type EntregaParaInserir struct {
	ID_funcionario     int64             `json:"id_funcionario" binding:"required"`
	Id_user            int               `json:"id_user" binding:"required,numeric"`
	Data_entrega       configs.DataBr    `json:"data_entrega" binding:"required"`
	IdTroca            *int              `json:"idTroca"`
	Assinatura_Digital string            `json:"assinatura_digital" binding:"required"`
	Itens              []ItemParaInserir `json:"itens" binding:"required,min=1,dive"`
}

type ItemEntregueDto struct {
	Id         int64       `json:"id"`
	Epi        EpiResponse `json:"epi"`
	Tamanho    TamanhoDto  `json:"tamanho"`
	Quantidade int         `json:"quantidade"`
}

type EntregaDto struct {
	Id                 int64             `json:"id"`
	Id_user            int               `json:"id_user"`
	Funcionario        Funcionario_Dto   `json:"funcionario"`
	Data_entrega       configs.DataBr    `json:"data_entrega"`
	Assinatura_Digital string            `json:"assinatura_digital"`
	Itens              []ItemEntregueDto `json:"itens,omitempty"`
}

type EntregaDashbord struct {
	Id             int            `json:"id"`
	IdFuncionario  int            `json:"idFuncionario"`
	Data_entrega   configs.DataBr `json:"data_entrega"`
	Assinatura     string         `json:"assinatura"`
	TokenValidacao string         `json:"token_validacao"`
}

type EntregaItensDashBord struct {
	Id         int `json:"id"`
	IdEntrega  int `json:"idEntrega"`
	IdEpi      int `json:"idEpi"`
	IdTamanho  int `json:"idTamanho"`
	Quantidade int `json:"quantidade"`
}

// 1. A versão da Entrega feita para morar dentro do Funcionário
type EntregaDoFuncionarioDto struct {
	Id                 int64             `json:"id"`
	Data_entrega       configs.DataBr    `json:"data_entrega"`
	Assinatura_Digital string            `json:"assinatura_digital"`
	Itens              []ItemEntregueDto `json:"itens,omitempty"`
}

type ItemEntregueFuncionario struct {
	Id         int64       `json:"id"`
	Quantidade int         `json:"quantidade"`
	Tamanho    TamanhoDto  `json:"tamanho"`
	Epi        EpiResponse `json:"epi"`
}