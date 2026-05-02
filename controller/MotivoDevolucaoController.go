package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
)

type MotivoService interface {
	Salvar(ctx context.Context, modelM model.MotivoDevolucao, tenantId int32) (model.MotivoDevolucaoEpiDto, error)
	ListarMotivos(ctx context.Context, tenantId int32) ([]model.MotivoDevolucaoEpiDto, error)
}

type MotivoController struct {
	service MotivoService
}

func NewMotivoController(service MotivoService) *MotivoController {

	return &MotivoController{
		service: service,
	}
}

func (m *MotivoController) Salvar() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.MotivoDevolucao

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		motivos, err:= m.service.Salvar(ctx, input, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrDadoDuplicado) {
				ctx.JSON(http.StatusConflict, gin.H{

					"error":   "motivo ja registrado",
					"detalhe": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, motivos)
	}
}

func (m *MotivoController) ListarMotivo() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {

			ctx.JSON(500, gin.H{"error": "erro interno de tenant"})
			return
		}

		motivos, err := m.service.ListarMotivos(ctx, tenantId)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":"erro ao mostrar os motivos da devolucao",
				"detalhes": err.Error(),
			})
			return
		}


		ctx.JSON(http.StatusOK, motivos)
	}
}
