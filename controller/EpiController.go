package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/service"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
)

type EpiService interface {
	Salvar(ctx context.Context, model model.EpiInserir, tenantID int32) error
	ListarEpis(ctx context.Context, f service.EpiFiltro, tenantId int32) (service.EpiPaginado, error)
	ListarEpi(ctx context.Context, id int, tenantid int32) (model.EpiDto, error)
	CancelarEpi(ctx context.Context, id int, tenantid int32) (int64, error)
	AtualizaEpi(ctx context.Context, model model.UpdateEpiInput, id, tenantId int32) error
	ListarEpiDashbord(ctx context.Context, tenantId int32) ([]model.EpiDashBord, error)
}

type EpiController struct {
	service EpiService
}

func NewEpiController(service EpiService) *EpiController {

	return &EpiController{
		service: service,
	}
}

func (e *EpiController) AdicionarEpi() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.EpiInserir

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		epi := model.EpiInserir{
			Nome:           input.Nome,
			Fabricante:     input.Fabricante,
			CA:             input.CA,
			Descricao:      input.Descricao,
			DataValidadeCa: input.DataValidadeCa,
			Idtamanho:      input.Idtamanho,
			IDprotecao:     input.IDprotecao,
			AlertaMinimo:   input.AlertaMinimo,
		}

		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		err := e.service.Salvar(ctx, epi, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrDadoDuplicado) {
				ctx.JSON(http.StatusConflict, gin.H{

					"error":   "CA ja registrado",
					"detalhe": err.Error(),
				})
				return
			}

			if errors.Is(err, helper.ErrDataMenor) {

				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error":    "data não pode ser menor que a atual",
					"detalhes": err.Error(),
				})
				return
			}

			if errors.Is(err, helper.ErrConflitoIntegridade) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error":    "tamanho ou protecao invalidos",
					"detalhes": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{

			"mensagem": "epi cadastrado",
		})
	}
}

func (e *EpiController) ListarEpis() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {

			ctx.JSON(500, gin.H{"error": "erro interno de tenant"})
			return
		}

		var filtro service.EpiFiltro

		if err := ctx.ShouldBindQuery(&filtro); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "parametros de paginacao invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		if filtro.Pagina <= 0 {
			filtro.Pagina = 1
		}
		if filtro.Quantidade <= 0 {
			filtro.Quantidade = 1000 // Padrão de 10 itens se não informar
		}

		epis, err := e.service.ListarEpis(ctx, filtro, tenantId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, epis)
	}
}

func (e *EpiController) ListarEpiPorId() gin.HandlerFunc {

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

		epi, err := e.service.ListarEpi(ctx, id, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error":   "epi nao encontrado",
					"detalhe": err.Error(),
				})

				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, epi)
	}
}

func (e *EpiController) DeletarEpi() gin.HandlerFunc {

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

		_, err = e.service.CancelarEpi(ctx, id, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": " epi nao encontrado",
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

func (e *EpiController) AtualizaEpi() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		idString := ctx.Param("id")

		id, err := strconv.Atoi(idString)
		if err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "id deve ser um numero",
			})
			return
		}

		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		var input model.UpdateEpiInput

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		err = e.service.AtualizaEpi(ctx, input, int32(id), tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error":    "epi nao encontrado",
					"detalhes": err.Error(),
				})

				return
			}

			if errors.Is(err, helper.ErrDadoDuplicado) {
				ctx.JSON(http.StatusConflict, gin.H{

					"error":    "CA ja cadastrado",
					"detalhes": err.Error(),
				})
				return
			}

			if errors.Is(err, helper.ErrDataMenor) {

				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error":    "data não pode ser menor que a atual",
					"detalhes": err.Error(),
				})
				return
			}

			if errors.Is(err, helper.ErrConflitoIntegridade) {

				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error":    "tamanho ou protecao nao encontrado",
					"detalhes": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    err.Error(),
				"detalhes": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"sucesso": "epi atualizado com sucesso"})
	}
}

func (e *EpiController) ListarEpiDashborController() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}



		epis, err:= e.service.ListarEpiDashbord(ctx, tenantId)
				if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, epis)

	}
}
