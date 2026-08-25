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

type EpiService interface {
	Salvar(ctx context.Context, model model.EpiInserir, tenantID int32) error
	ListarEpis(ctx context.Context, f service.EpiFiltro, tenantId int32) (service.EpiPaginado, error)
	ListarEpi(ctx context.Context, id int, tenantid int32) (model.EpiDto, error)
	CancelarEpi(ctx context.Context, id int, tenantid int32) (int64, error)
	AtualizaEpi(ctx context.Context, model model.UpdateEpiInput, id, tenantId int32) error
	ListarEpiDashbord(ctx context.Context, tenantId int32) ([]model.EpiDashBord, error)
	BuscarEpiDoFuncionario(ctx context.Context, tenantId, IdFuncionario int32) ([]model.EpiDtoDevolucao, error)
}

type EpiController struct {
	service EpiService
}

func NewEpiController(service EpiService) *EpiController {

	return &EpiController{
		service: service,
	}
}


// AdicionarEpi godoc
// @Summary      Cadastrar novo EPI
// @Description  Adiciona um novo EPI ao estoque da empresa
// @Tags         EPIs
// @Accept       json
// @Produce      json
// @Param        epi body model.EpiInserir true "Dados do EPI"
// @Success      200  {object}  map[string]string "EPI cadastrado"
// @Failure      409  {object}  helper.HTTPError "CA já registrado"
// @Failure      422  {object}  helper.HTTPError "Erro de validação (Data, Integridade)"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /epis [post]
// @Security     BearerAuth
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
			IdTamanho:      input.IdTamanho,
			IDProtecao:     input.IDProtecao,
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

// ListarEpis godoc
// @Summary      Listar EPIs
// @Description  Retorna uma lista paginada de EPIs
// @Tags         EPIs
// @Produce      json
// @Param        pagina     query    int     false  "Página"
// @Param        quantidade query    int     false  "Quantidade por página"
// @Success      200  {object}  service.EpiPaginado
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /epis [get]
// @Security     BearerAuth
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

// ListarEpiPorId godoc
// @Summary      Buscar EPI por ID
// @Description  Retorna os detalhes de um EPI específico
// @Tags         EPIs
// @Param        id   path      int  true  "ID do EPI"
// @Produce      json
// @Success      200  {object}  model.EpiDto
// @Failure      422  {object}  helper.HTTPError "EPI não encontrado"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /epis/{id} [get]
// @Security     BearerAut
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

// DeletarEpi godoc
// @Summary      Deletar EPI
// @Description  Remove um EPI do cadastro
// @Tags         EPIs
// @Param        id   path      int  true  "ID do EPI"
// @Success      204  "Sem conteúdo"
// @Failure      404  {object}  helper.HTTPError "EPI não encontrado"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /epis/{id} [delete]
// @Security     BearerAuth
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

// AtualizaEpi godoc
// @Summary      Atualizar EPI
// @Description  Atualiza os dados de um EPI existente
// @Tags         EPIs
// @Accept       json
// @Produce      json
// @Param        id   path      int                   true  "ID do EPI"
// @Param        body body      model.UpdateEpiInput  true  "Dados para atualização"
// @Success      200  {object}  map[string]string "Sucesso"
// @Failure      409  {object}  helper.HTTPError "CA já cadastrado"
// @Failure      422  {object}  helper.HTTPError "Erro de validação"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /epis/{id} [put]
// @Security     BearerAuth
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

		fmt.Printf("INPUT RECEBIDO: %+v\n", input)
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

// ListarEpiDashborController godoc
// @Summary      Dashboard de EPIs
// @Description  Retorna dados resumidos para o dashboard de EPIs
// @Tags         EPIs
// @Produce      json
// @Success      200  {array}   model.EpiDashBord
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /epis/dashboard [get]
// @Security     BearerAuth
func (e *EpiController) ListarEpiDashborController() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		epis, err := e.service.ListarEpiDashbord(ctx, tenantId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, epis)

	}
}

// ListarEpiFuncionario godoc
// @Summary      EPIs por funcionário
// @Description  Retorna a lista de EPIs vinculados a um funcionário específico
// @Tags         EPIs
// @Param        id   path      int  true  "ID do Funcionário"
// @Produce      json
// @Success      200  {array}   model.EpiDtoDevolucao
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /epis/funcionario/{id} [get]
// @Security     BearerAuth
func (e *EpiController) ListarEpiFuncionario() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		idString := ctx.Param("id")
		IdFuncionario, err := strconv.Atoi(idString)
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

		epis, err := e.service.BuscarEpiDoFuncionario(ctx, tenantId, int32(IdFuncionario))
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, epis)
	}

	

}
