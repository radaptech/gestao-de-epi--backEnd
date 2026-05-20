package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
)

type DevolucaoService interface {
	SalvarDevolucao(ctx context.Context, modelDevolucao model.DevolucaoInserir, tenantId int32, token string) error
	CancelarDevolucao(ctx context.Context, id, iduser, tenantId int) error
	ListarDevolucoes(ctx context.Context, tenantId int32) ([]model.DevolucaoResponse, error)
	TokenDevolucao(ctx context.Context, tenantId, Idfuncionario int32) (string, error)
	GerarDadosPdf(ctx context.Context, idDevolucao, tenantId int32) (helper.DadosDevolucaoPdf, error)
}

type DevolucaoController struct {
	service DevolucaoService
}

func NewDevolucaoController(service DevolucaoService) *DevolucaoController {

	return &DevolucaoController{
		service: service,
	}
}

func (d *DevolucaoController) Adicionar() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		var input model.DevolucaoInserir

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno de tenant"})
			return
		}

		userId, ok := middleware.GetUserID(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"erro": "erro ao setar usuario",
			})
			return
		}

		token, err := d.service.TokenDevolucao(ctx, tenantId, int32(input.IdFuncionario))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao gerar token de auditoria"})
			return
		}

		urlAssinatura, err := helper.UploadAssinaturaSupabase(input.AssinaturaDigital, token, "devolucao")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":    "falha ao salvar assinatura digital",
				"detalhes": err.Error(),
			})
			return
		}

		input.AssinaturaDigital = urlAssinatura
		input.Iduser = int(userId)
		err = d.service.SalvarDevolucao(ctx, input, tenantId, token)
		if err != nil {
			// Tratamento de erros específicos
			if errors.Is(err, helper.ErrNaoEncontrado) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "funcionario ou registro não encontrado"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao salvar entrega", "detalhes": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"mensagem": "devolucao cadastrada com sucesso",
		})
	}
}

func (d *DevolucaoController) Listar() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro ao receber tenantId",
			})
			return
		}

		devolucoes, err := d.service.ListarDevolucoes(ctx, tenantId)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    "erro ao realizar buscar das entregas de epi",
				"detalhes": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, devolucoes)
	}
}

func (d *DevolucaoController) GerarFichaPDF() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		iddevolucaoStr := ctx.Param("id")
		idDevolucao, err := strconv.Atoi(iddevolucaoStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "id da entrega inválido"})
			return
		}

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno de tenant",
			})
			return
		}

		responsavel := ctx.GetString("user_nome") //pegando o responsavel logado no sistema

		auditoria := helper.Auditoria{
			DadosServidor: time.Now().Format("02/01/2006 às 15:04:05"),
			Ip:            ctx.ClientIP(),
		}

		devolucaoDados, err := d.service.GerarDadosPdf(ctx, int32(idDevolucao), tenantId)
		if err != nil {

			ctx.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":    "dados nao encontrados",
				"detalhes": err.Error(),
			})
			return
		}

		pdf, err := helper.CreatePdfDevolucao(devolucaoDados, auditoria, responsavel)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    err.Error(),
				"detalhes": "erro na geração do pdf",
			})
			return
		}

		ctx.Header("Content-Disposition", "attachment; filename=Ficha_devolucao_"+iddevolucaoStr+".pdf")
		ctx.Data(http.StatusOK, "application/pdf", pdf.GetBytes()) // Extrai os bytes limpos!
	}
}
