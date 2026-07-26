package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/service"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type FornecedorService interface {
	Adicionar(ctx context.Context, model model.FornecedorInserir, tenantId int32) error
	ListarFornecedor(ctx context.Context, filt service.FiltroFornecedores, tenatId int32) (service.FornecedoresPaginados, error)
	CancelarFornecedor(ctx context.Context, id, tenantId int32) error
	AtualizaFornecedor(ctx context.Context, model model.FornecedorUpdate, id, tenantId int64) error
}

type FornecedorController struct {
	service FornecedorService
}

func NewFornecedorController(service FornecedorService) *FornecedorController {

	return &FornecedorController{
		service: service,
	}
}
func (f *FornecedorController) ImportFornecedor() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		fileHearder, err := ctx.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{

				"erro":     "selecione um arquivo de planilas valido",
				"detalhes": err.Error(),
			})
			return
		}

		file, err := fileHearder.Open()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{

				"erro":     "erro ao ler a planilha",
				"detalhes": err.Error(),
			})
			return
		}

		defer file.Close()

		fi, err := excelize.OpenReader(file)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"erro":     "O arquivo enviado não é uma planilha Excel (.xlsx) válida.",
				"detalhes": err.Error(),
			})
			return
		}
		defer fi.Close()

		NomeDaPlanilha := fi.GetSheetName(0)
		linhas, err := fi.GetRows(NomeDaPlanilha)
		if err != nil || len(linhas) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message":  "A planilha selecionada está vazia.",
				"detalhes": err.Error(),
			})
			return
		}

		tenantID, exists := middleware.GetTenantID(ctx)
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Sessão inválida ou expirada."})
			return
		}
		var fornecedores []string
		cabecalho := false

		for _, linha := range linhas {

			if len(linha) < 3 {
				continue
			}

			razaoSocial := strings.TrimSpace(linha[0])
			NomeFantasia := strings.TrimSpace(linha[1])
			cnpj := strings.TrimSpace(linha[2])

			inscricaoEstadual := ""
			if len(linha) >= 4 {

				inscricaoEstadual = strings.TrimSpace(linha[3])

			}

			if strings.EqualFold(razaoSocial, "Razão Social") {

				cabecalho = true
				continue
			}

			if !cabecalho || razaoSocial == "" || cnpj == "" {
				continue
			}

			fornecedor := model.FornecedorInserir{
				RazaoSocial:       razaoSocial,
				NomeFantasia:      NomeFantasia,
				CNPJ:              cnpj,
				InscricaoEstadual: inscricaoEstadual,
			}

			err = f.service.Adicionar(ctx, fornecedor, tenantID)
			if err != nil {
				// Ignora duplicados e continua o loop (Opção B)
				if errors.Is(err, helper.ErrDadoDuplicado) || errors.Is(err, helper.ErrConflitoIntegridade) {

					continue
				}

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"message":  fmt.Sprintf("Erro ao salvar o fornecedor '%s'.", razaoSocial),
					"detalhes": err.Error(),
				})
				return
			}
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Importação concluída com sucesso!",
			"total":   len(fornecedores),
		})
	}

}
func (f *FornecedorController) Adicionar() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.FornecedorInserir

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error":    "dados invalidos, ",
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

		err := f.service.Adicionar(ctx, input, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrDadoDuplicado) {
				ctx.JSON(http.StatusConflict, gin.H{
					"error":    "CNPJ ja existe no sistema",
					"detalhes": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"detalhes": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{

			"mensagem": "fornecedo cadastrado",
		})
	}
}

func (f *FornecedorController) ListarFornecedores() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var filtro service.FiltroFornecedores

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

		fornecedores, err := f.service.ListarFornecedor(ctx, filtro, tenantId)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":    "erro ao realizar busca em fornecedores",
				"detalhes": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, fornecedores)
	}
}

func (f *FornecedorController) CancelarFornecedor() gin.HandlerFunc {

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

		err = f.service.CancelarFornecedor(ctx, int32(id), tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error":    "fornecedor não encontrado",
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

func (f *FornecedorController) AtualizaFornecedor() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		idParam := ctx.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{
				"erro": err.Error(),
			})
			return
		}

		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		var input model.FornecedorUpdate

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		err = f.service.AtualizaFornecedor(ctx, input, int64(id), int64(tenantID))
		if err != nil {

			if errors.Is(err, helper.ErrDadoDuplicado) {

				ctx.JSON(http.StatusConflict, gin.H{

					"error":    "cnpj ja cadastrado",
					"detalhes": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    "erro interno no servidor",
				"detalhes": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"sucesso": "fornecedor atualizado com sucesso"})
	}
}
