package model

type MotivoDevolucao struct {
	Motivo string `json:"motivo" binding:"required"`
}

type MotivoDevolucaoEpiDto struct {
	Id     int    `json:"id"`
	Motivo string `json:"motivo"`
}
