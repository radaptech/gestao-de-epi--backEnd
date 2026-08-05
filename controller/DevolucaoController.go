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

// Adicionar godoc
// @Summary      Adicionar devolução
// @Description  Registra uma nova devolução de EPI com assinatura digital
// @Tags         Devoluções
// @Accept       json
// @Produce      json
// @Param        devolucao body model.DevolucaoInserir true "Dados da devolução"
// @Success      200  {object}  map[string]string "Sucesso"
// @Failure      400  {object}  helper.HTTPError "Dados inválidos"
// @Failure      422  {object}  helper.HTTPError "Entidade não processável"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /devolucao [post]
// @Security     BearerAuth
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

// Listar godoc
// @Summary      Listar devoluções
// @Description  Retorna a lista de todas as devoluções registradas para o tenant
// @Tags         Devoluções
// @Produce      json
// @Success      200  {array}   model.DevolucaoResponse
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /devolucoes [get]
// @Security     BearerAuth
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

// GerarFichaPDF godoc
// @Summary      Gerar ficha de devolução PDF
// @Description  Gera e faz o download da ficha de devolução em formato PDF
// @Tags         Devoluções
// @Param        id   path      int  true  "ID da devolução"
// @Produce      application/pdf
// @Success      200  {file}    binary
// @Failure      400  {object}  helper.HTTPError "ID inválido"
// @Failure      422  {object}  helper.HTTPError "Não encontrado"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /devolucao/pdf/{id} [get]
// @Security     BearerAuth
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
