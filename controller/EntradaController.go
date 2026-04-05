package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/service"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
)

type EntradaService interface {
	Adicionar(ctx context.Context, model model.EntradaEpiInserir, tenantID int32) error
	ListarEntradas(ctx context.Context, f service.FiltroEntradas, tenatId int32) (service.EntradaPaginada, error)
	CancelarEntrada(ctx context.Context, id, idUser, tenantid int) (int64, error)
	EntradaDashbordBusca(ctx context.Context, tenantId int32) ([]model.EntradaDashbord, error)
	BuscaEntradaEstoque(ctx context.Context, tenantId int32) ([]model.EntradaEstoqueDto, error)
}

type EntradaController struct {
	service EntradaService
}

func NewEntradaController(service EntradaService) *EntradaController {

	return &EntradaController{
		service: service,
	}
}

func (e *EntradaController) AdicionarEntrada() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.EntradaEpiInserir

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{

				"error":    "dados invalidos, ",
				"detalhes": err.Error(),
			})
			return
		}

		// 1. Remove espaços extras no começo/fim
		// 2. Transforma tudo em MAIÚSCULO para padronizar
		

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno de tenant",
			})
			return
		}

		err := e.service.Adicionar(ctx, input, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrDataMenor) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{
					"error":    "data de entrada inferior a data atual",
					"detalhes": err.Error(),
				})
				return
			}

			if errors.Is(err, helper.ErrDataIgual) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error":    "data da validade é igual a data de fabricação",
					"detalhes": err.Error(),
				})
				return
			}

			if errors.Is(err, helper.ErrDataMenorValidade) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{
					"error":    "data de validade inferior a data de fabricação",
					"detalhes": err.Error(),
				})
				return
			}

			if errors.Is(err, helper.ErrConflitoIntegridade) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error":    "epi,tamanho ou fornecedor nao encontrado",
					"detalhes": err.Error(),
				})
				return

			}

			if errors.Is(err, helper.ErrDadoDuplicado) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error":    "NF repetida, NF ja cadastrada no banco de dados, por favor verifique.",
					"detalhes": err.Error(),
				})
				return
			}
		}

		ctx.JSON(http.StatusOK, gin.H{

			"mensagem": "entrada cadastrada",
		})
	}
}

func (e *EntradaController) ListarEntradas() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var filtro service.FiltroEntradas

		if err := ctx.ShouldBindQuery(&filtro); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error":    "parametros de busca invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro ao receber tenantId",
			})
			return
		}

		if filtro.Pagina <= 0 {
			filtro.Pagina = 1
		}
		if filtro.Quantidade <= 0 {
			filtro.Quantidade = 1000 // Padrão de 10 itens se não informar
		}

		entradas, err := e.service.ListarEntradas(ctx, filtro, tenantId)
		if err != nil {

			fmt.Printf("Erro ao listar entradas: %v\n", err)

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": "erro ao realizar buscar das entradas de epi",
			})
			return
		}

		ctx.JSON(http.StatusOK, entradas)
	}
}

func (e *EntradaController) CancelarEntrada() gin.HandlerFunc {

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

		idUser, existe := ctx.Get("userId")
		if !existe {
			ctx.JSON(http.StatusUnauthorized, gin.H{

				"error": "Token inválido ou sem id",
			})

			return
		}

		_, err = e.service.CancelarEntrada(ctx, id, int(idUser.(uint)), int(tenantId))
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error":    "entrada não encontrada",
					"detalhes": err.Error(),
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

func (e *EntradaController) BuscaEntradaDashbord() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}


		entradas, err:= e.service.EntradaDashbordBusca(ctx, tenantId)
		if err != nil {

			 ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			 })
			 return 
		}


		ctx.JSON(http.StatusOK, entradas)
	}
}

func (e *EntradaController) BuscaEntradaEstoque() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}


		entradas, err:= e.service.BuscaEntradaEstoque(ctx, tenantId)
		if err != nil {

			 ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			 })
			 return 
		}


		ctx.JSON(http.StatusOK, entradas)
	}
}