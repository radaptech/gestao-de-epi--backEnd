package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/service"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
	// Imports da V2 do Maroto
)

type EntregasService interface {
	Salvar(ctx context.Context, model model.EntregaParaInserir, tenantid int32) error
	ListaEntregas(ctx context.Context, f service.FiltroEntregas, tenantId int32) (service.EntregaPaginada, error)
	CancelarEntrega(ctx context.Context, tenantId, id, iduser int) error
	GerarDadosPdfService(ctx context.Context, matricula string, tenantId int32) (helper.DadosPdf, error)
}

type EntregaController struct {
	Service EntregasService
}

func NewEntregaController(service EntregasService) *EntregaController {

	return &EntregaController{
		Service: service,
	}
}

func (e *EntregaController) Adicionar() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.EntregaParaInserir

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno de tenant",
			})
			return
		}

		err := e.Service.Salvar(ctx, input, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{
					"error":    "id do funcionario, usuario ou entrega, não encontrado",
					"detalhes": err.Error(),
				})
				return
			}

			if strings.Contains(err.Error(), "estoque insuficiente") {

				ctx.JSON(http.StatusUnprocessableEntity, gin.H{
					"error": err.Error(),
				})
				return

			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"detalhes": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"mensagem": "entrega cadastrada com sucesso"})

	}
}

func (e *EntregaController) ListarEntregas() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var filtro service.FiltroEntregas

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
			filtro.Quantidade = 10 // Padrão de 10 itens se não informar
		}

		entregas, err := e.Service.ListaEntregas(ctx.Request.Context(), filtro, tenantId)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": "erro ao realizar buscar das entregas de epi",
			})
			return
		}

		ctx.JSON(http.StatusOK, entregas)
	}
}

func (e *EntregaController) CancelarEntrega() gin.HandlerFunc {

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

		err = e.Service.CancelarEntrega(ctx, int(tenantId), id, int(idUser.(uint)))
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error":    "entrega não encontrada",
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

func (e *EntregaController) GerarFichaEpiPDF() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		matricula := ctx.Param("matricula")
		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno de tenant",
			})
			return
		}

		responsavel:= ctx.GetString("user_nome") //pegando o responsavel logado no sistema

		auditoria := helper.Auditoria{
			DadosServidor: time.Now().Format("02/01/2006 às 15:04:05"),
			Ip:            ctx.ClientIP(),
		}

		fmt.Printf("DEBUG: Matricula do Param: '%s' | Tenant do Middleware: %d\n", matricula, tenantId)
		entregaDadosPdf, err := e.Service.GerarDadosPdfService(ctx.Request.Context(), matricula, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{
					"error":    "dados nao encontrados",
					"detalhes": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    err.Error(),
				"detalhes": "dados não obtidos para gerar o pdf",
			})
			return
		}

		documento, err := helper.CreatePdf(entregaDadosPdf, auditoria, responsavel)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    err.Error(),
				"detalhes": "erro na geração do pdf",
			})
			return

		}

		// Baixa o arquivo direto no navegador
		ctx.Header("Content-Disposition", "attachment; filename=Ficha_EPI_"+matricula+".pdf")
		ctx.Data(http.StatusOK, "application/pdf", documento.GetBytes()) // Extrai os bytes limpos!
	}
}
