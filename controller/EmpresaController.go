package controller

import (
	"context"
	"net/http"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/gin-gonic/gin"
)

type EmpresaService interface {
	Salvar(ctx context.Context, model model.EmpresaInserir) error
}

type EmpresaController struct {
	service EmpresaService
}

func NewEmpresaController(serv EmpresaService) *EmpresaController {

	return &EmpresaController{
		service: serv,
	}
}

func (e *EmpresaController) Salvar() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input model.EmpresaInserir

		
		if err := ctx.ShouldBindJSON(&input); err != nil {
			
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "Dados inválidos ou formato incorreto.",
				"detalhes": err.Error(),
			})
			return
		}

		
		err := e.service.Salvar(ctx, input)
		if err != nil {
			
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":    "Erro ao salvar empresa no banco de dados.",
				"detalhes": err.Error(),
			})
			return
		}

		
		ctx.JSON(http.StatusCreated, gin.H{
			"sucesso": "Empresa cadastrada com sucesso",
		})
	}
}
