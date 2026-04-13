package model

import (
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"
)

// ItemParaInserir representa cada EPI saindo do estoque
type ItemParaInserir struct {
	ID_epi        int32 `json:"id_epi" binding:"required"`
	ID_tamanho    int32 `json:"id_tamanho" binding:"required"`
	ID_entrada_item int32 `json:"id_entrada_item"` // 🔑 NOVO: Precisamos saber de qual lote saiu!
	Quantidade    int32 `json:"quantidade" binding:"required,gt=0"`
}

// EntregaParaInserir é o DTO que a Paloma envia do Frontend
type EntregaParaInserir struct {
	ID_funcionario     int32             `json:"id_funcionario" binding:"required"`
	Id_user            int32             `json:"id_user"`
	Data_entrega       configs.DataBr    `json:"data_entrega" binding:"required"`
	IdTroca            *int32            `json:"id_troca"` // Usei ponteiro para aceitar nulo
	Assinatura_Digital string            `json:"assinatura_digital" binding:"required"`
	Itens              []ItemParaInserir `json:"itens" binding:"required,min=1,dive"`
}

// EntregaDto representa o retorno completo para a listagem
type EntregaDto struct {
	Id                 int32             `json:"id"`
	Id_user            int32             `json:"id_user"`
	Funcionario        Funcionario_Dto   `json:"funcionario"`
	Data_entrega       configs.DataBr    `json:"data_entrega"`
	Assinatura_Digital string            `json:"assinatura_digital"`
	Token_validacao    string            `json:"token_validacao"`
	Itens              []ItemEntregueDto `json:"itens,omitempty"`
}

// ItemEntregueDto detalha cada item em uma entrega já realizada
type ItemEntregueDto struct {
	Id         int32       `json:"id"`
	Quantidade int32       `json:"quantidade"`
	Epi        EpiResponse `json:"epi"`
	Tamanho    TamanhoDto  `json:"tamanho"`
}

// Modelos para Dashboard (Otimizados para serem leves)
type EntregaDashbord struct {
	Id             int32          `json:"id"`
	IdFuncionario  int32          `json:"id_funcionario"`
	Data_entrega   configs.DataBr `json:"data_entrega"`
	Assinatura     string         `json:"assinatura"`
	TokenValidacao string         `json:"token_validacao"`
}

type EntregaItensDashBord struct {
	Id                 int32 `json:"id"`
	IdEntregaCabecalho int32 `json:"id_entrega_cabecalho"`
	IdEpi              int32 `json:"id_epi"`
	IdTamanho          int32 `json:"id_tamanho"`
	Quantidade         int32 `json:"quantidade"`
}

type EntregaDoFuncionarioDto struct {
	Id                 int64             `json:"id"`
	Data_entrega       configs.DataBr    `json:"data_entrega"`
	Assinatura_Digital string            `json:"assinatura_digital"`
	Itens              []ItemEntregueDto `json:"itens,omitempty"` // <-- AQUI ELA!
}