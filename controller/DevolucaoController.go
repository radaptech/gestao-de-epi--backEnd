package controller

import (
	"context"
	"errors"
	"net/http"
	

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/service"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
)

type DevolucaoService interface {
	SalvarDevolucao(ctx context.Context, modelDevolucao model.DevolucaoInserir, tenantId int32, token string) error
	CancelarDevolucao(ctx context.Context, id, iduser, tenantId int) error
	ListarDevolucoes(ctx context.Context, f service.FiltroDevolucao, tenantId int32) (service.DevolucaoPaginada, error)
	TokenDevolucao(ctx context.Context, tenantId, Idfuncionario int32) (string, error)
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
