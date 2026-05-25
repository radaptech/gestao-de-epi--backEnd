package controller

import (
	"context"
	"net/http"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/gin-gonic/gin"
)

type PlanosServices interface {
	SalvarPlanos(ctx context.Context, model model.Plano) (int32, error)
	MostrarPlanos(ctx context.Context) ([]model.Plano, error)
}

type PlanosController struct {
	service PlanosServices
}

func NewPlanoController(serv PlanosServices) *PlanosController {

	return &PlanosController{
		service: serv,
	}
}

func (p *PlanosController) SalvarPlano() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.Plano

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error": err.Error(),
			})
			return
		}

		idPlano, err := p.service.SalvarPlanos(ctx, input)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, idPlano)
	}
}

func (p *PlanosController) MostrarPlanos() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		planos, err := p.service.MostrarPlanos(ctx)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro interno ao listar funcoes",
			})
			return
		}

		ctx.JSON(http.StatusOK, planos)
	}
}
