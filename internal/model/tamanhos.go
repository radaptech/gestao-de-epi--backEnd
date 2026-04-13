package model

type Tamanhos struct {
	Tamanho string `json:"tamanho" binding:"required"`
}

type TamanhoDto struct {
	ID      int    `json:"id"`
	Tamanho string `json:"tamanho"`
}

type TamanhoEntregaDto struct {
	ID      int    `json:"id"`
	Tamanho string `json:"tamanho"`
	Id_epi  int32  `json:"id_epi"`
}
