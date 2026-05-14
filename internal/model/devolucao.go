package model

import "github.com/davi-fernandesx/sistema-de-gestao-de-epi/configs"

type DevolucaoInserir struct {
    // Campos Obrigatórios (binding:"required")
    IdFuncionario       int            `json:"idFuncionario" binding:"required"`
    IdEpi               int            `json:"idEpi" binding:"required"`
    IdMotivo            int            `json:"idMotivo" binding:"required"`
    IdTamanho           int            `json:"idTamanho" binding:"required"`
    DataDevolucao       configs.DataBr `json:"data_devolucao" binding:"required"`
    QuantidadeADevolver int            `json:"quantidadeADevolver" binding:"required,numeric,gt=0"`
    
	Troca               bool           `json:"houve_troca"`
    // Campos de Troca (Podem ser nulos no JS)
    IdEpiNovo           *int           `json:"idEpiNovo"`
    IdTamanhoNovo       *int           `json:"idTamanhoNovo"`
    NovaQuantidade      *int           `json:"quantidadeNova"`
    
    // Outros campos do Payload
    AssinaturaDigital   string         `json:"assinatura_digital" binding:"required"`
    TokenValidacao      string         `json:"-"`
    Observacao          *string        `json:"observacao"`
    
    // Você mencionou 'IdUser' na struct anterior, mas ele não está no payload do JS.
    // DICA: Pegue o ID do usuário logado direto do Token JWT no Backend por segurança!
}

type DevolucaoDto struct {
    Id                  int                   `json:"id"`
    // O JS vai procurar por 'id_funcionario', se encontrar um objeto, ele pega o .id dele
    Funcionario         Funcionario_Dto       `json:"funcionario"` 
    Epi                 EpiDto                `json:"epi"`
    Motivo              MotivoDevolucaoEpiDto `json:"motivo"` // Mudei para 'motivo' para bater com a busca do JS
    DataDevolucao       configs.DataBr        `json:"data_devolucao"` // 'data_devolucao' é o que o filtro JS usa primeiro
    QuantidadeADevolver int                   `json:"quantidade_a_devolver"`
    AssinaturaDigital   string                `json:"assinatura_digital"`
    
    // Troca
    EpiNovo             *EpiDto               `json:"epi_novo"` // JS busca por 'epi_novo' ou 'epiNovo'
    Tamanho             *TamanhoDto           `json:"tamanho"`
    TamanhoNovo         *TamanhoDto           `json:"tamanho_novo"` // Adicione este se houver no banco
    NovaQuantidade      *int                  `json:"quantidade_nova"`
}