package model

type MotivoDevolucao struct {
	Motivo   string `json:"motivo" binding:"required"`
	Descaste bool   `json:"gera_descarte"`
}

type MotivoDevolucaoEpiDto struct {
	Id     int    `json:"id"`
	Motivo string `json:"motivo"`
}
