package controller

import (
	"context"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/service"
)

type DevolucaoService interface {
	SalvarDevolucao(ctx context.Context, modelDevolucao model.DevolucaoInserir, tenantId int32) error
	CancelarDevolucao(ctx context.Context, id, iduser, tenantId int) error
	ListarDevolucoes(ctx context.Context, f service.FiltroDevolucao, tenantId int32) (service.DevolucaoPaginada, error)
}


