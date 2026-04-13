package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
)

type TamanhoService interface {
	SalvarTamanho(ctx context.Context, model model.Tamanhos, tenantId int32) error
	ListarTamanho(ctx context.Context, id int, tenantId int32) (model.TamanhoDto, error)
	ListarTodosTamanhos(ctx context.Context, tenantId int32) ([]model.TamanhoDto, error)
	CancelarTamanho(ctx context.Context, id int, tenantId int32) error
	ListarTamanhoPorIdEpi(ctx context.Context, idEpi, tenantId int32) ([]model.TamanhoEntregaDto, error)
}

type TamanhoController struct {
	service TamanhoService
}

func NewTamanhoControle(service TamanhoService) *TamanhoController {

	return &TamanhoController{
		service: service,
	}
}

func (t *TamanhoController) Adicionar() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.Tamanhos

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		tamanho := model.Tamanhos{
			Tamanho: input.Tamanho,
		}

		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		err := t.service.SalvarTamanho(ctx, tamanho, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrDadoDuplicado) {
				ctx.JSON(http.StatusConflict, gin.H{

					"error": "tamanho ja registrado",
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{

			"mensagem": "tamanho cadastrado",
		})
	}
}

func (t *TamanhoController) ListarTodosTamanhos() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {

			ctx.JSON(500, gin.H{"error": "erro interno de tenant"})
			return
		}

		tamanhos, err := t.service.ListarTodosTamanhos(ctx, tenantId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, tamanhos)
	}
}

func (t *TamanhoController) ListarTamanhoPorId() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		idString := ctx.Param("id")
		id, err := strconv.Atoi(idString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "id deve ser um numero",
			})
			return 
		}

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		tamanho, err := t.service.ListarTamanho(ctx, id, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": "tamanho nao encontrado",
				})

				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, tamanho)
	}
}

func (t *TamanhoController) DeletarTamanho() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		idString := ctx.Param("id")
		id, err := strconv.Atoi(idString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "id deve ser um numero",
			})
			return 
		}

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		err = t.service.CancelarTamanho(ctx, id, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado){

				ctx.JSON(http.StatusNotFound, gin.H{
					
					"error":" tamanho nao encontrado",
				})

				return 
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			}) 

			return 
		}


		ctx.Status(http.StatusNoContent)
	}

}

func (t *TamanhoController) ListarTamanhoPorIdEpi() gin.HandlerFunc{

	return  func(ctx *gin.Context) {

		idString := ctx.Param("id")
		id, err := strconv.Atoi(idString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "id deve ser um numero",
			})
			return 
		}

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		tamanho, err := t.service.ListarTamanhoPorIdEpi(ctx, int32(id), tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": "tamanho nao encontrado",
					"detalhes": err.Error(),
				})

				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, tamanho)
	}

}