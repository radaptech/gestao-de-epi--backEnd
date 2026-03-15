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

type TipoProtecaoService interface {
	SalvarProtecao(ctx context.Context, m model.TipoProtecao, tenantId int32) (model.TipoProtecaoDto, error)
	ListarProtecao(ctx context.Context, id int, tenatId int32) (model.TipoProtecaoDto, error)
	ListarProtecoes(ctx context.Context, tenantId int32) ([]model.TipoProtecaoDto, error)
	DeletarProtecao(ctx context.Context, id int, tenantId int32) error
}

type TipoProtecaoController struct {
	service TipoProtecaoService
}

func NewTipoProtecaoController(service TipoProtecaoService) *TipoProtecaoController {

	return &TipoProtecaoController{
		service: service,
	}
}

func (t *TipoProtecaoController) AdicionarProtecao() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.TipoProtecao

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		protec := model.TipoProtecao{
			Nome: input.Nome,
		}

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": "erro interno de tenant",
			})
			return
		}

		tp,err := t.service.SalvarProtecao(ctx, protec, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrDadoDuplicado) {

				ctx.JSON(http.StatusConflict, gin.H{

					"erro": "Esse tipo de proteção ja está no sistema",
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{

			"mensagem": "proteção cadastrada",
			"protecao":tp,
		})
	}
}

func (t *TipoProtecaoController) ListarProtecoes() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {

			ctx.JSON(500, gin.H{"error": "erro interno de tenant"})
			return
		}

		protecs, err := t.service.ListarProtecoes(ctx, tenantId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, protecs)
	}
}
func (t *TipoProtecaoController) ListarProtecaoPorId() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		idString := ctx.Param("id")
		id, err := strconv.Atoi(idString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "id deve ser um numero",
			})
		}

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		protec, err := t.service.ListarProtecao(ctx, id, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": "protecao nao encontrada",
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, protec)
	}
}

func (t *TipoProtecaoController) DeletarProtecao() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		idString := ctx.Param("id")
		id, err := strconv.Atoi(idString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "id deve ser um numero",
			})
		}

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		err = t.service.DeletarProtecao(ctx, id, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": " protecao nao encontrada",
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
