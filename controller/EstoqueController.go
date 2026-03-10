package controller

import (
	"context"
	"net/http"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/service"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
)

type EstoqueService interface {
	MostrarQuantidadeTotais(ctx context.Context, f service.FiltroEstoqueAtual, tenantId int32) (service.EstoqueAtualPaginado, error)
	MostrarSaldoAtual(ctx context.Context, f service.FiltroEstoqueSaldo, tenantId int32) (service.EstoqueSaldoPaginado, error)
}

type EstoqueController struct {
	service EstoqueService
}

func NewEstoqueController(serv EstoqueService) *EstoqueController {

	return &EstoqueController{service: serv}
}

func (e *EstoqueController) MostrarQuantidades() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var filtro service.FiltroEstoqueAtual

		if err := ctx.ShouldBindQuery(&filtro); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error":    "parametros de busca invalidos",
				"detalhes": err.Error(),
			})
			return
		}
		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {

			ctx.JSON(500, gin.H{"error": "erro interno de tenant"})
			return
		}

		if filtro.Pagina <= 0 {
			filtro.Pagina = 1
		}
		if filtro.Quantidade <= 0 {
			filtro.Quantidade = 4 // Padrão de 4 itens se não informar
		}

		quantidades, err := e.service.MostrarQuantidadeTotais(ctx, filtro,tenantId)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, quantidades)
	}
}

func (e *EstoqueController) MostrarSaldo() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var filtro service.FiltroEstoqueSaldo

		if err := ctx.ShouldBindQuery(&filtro); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error":    "parametros de busca invalidos",
				"detalhes": err.Error(),
			})
			return
		}
		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {

			ctx.JSON(500, gin.H{"error": "erro interno de tenant"})
			return
		}

		if filtro.Pagina <= 0 {
			filtro.Pagina = 1
		}
		if filtro.Quantidade <= 0 {
			filtro.Quantidade = 4 // Padrão de 4 itens se não informar
		}

		saldos, err := e.service.MostrarSaldoAtual(ctx, filtro, tenantId)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, saldos)
	}
}
