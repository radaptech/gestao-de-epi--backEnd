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

type FuncaoService interface {
	SalvarFuncao(ctx context.Context, model model.Funcao, tenantid int32) error
	ListasTodasFuncao(ctx context.Context, f service.FiltroFuncao, tenantId int32) (service.FuncaoPaginado, error)
	DeletarFuncao(ctx context.Context, id int, tenantId int32) error
	AtualizarFuncao(ctx context.Context, id int, funcao string, tenantId int32) error
}

type FuncaoController struct {
	service FuncaoService
}

func NewFuncaoController(service FuncaoService) *FuncaoController {

	return &FuncaoController{service: service}
}

// RegistraFuncao godoc
// @Summary      Criar uma funcao
// @Description  Cadastra uma nova funcao no sistema
// @Tags         funcao
// @Accept       json
// @Produce      json
// @Param        funcao body model.Funcao true "Dados da funcao"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  helper.HTTPError "Dados inválidos"
// @Failure      409  {object}  helper.HTTPError "funcao já existe"
// @Failure      409  {object}  helper.HTTPError "id de departamento nao existe no sistema"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /cadastro-funcao [post]
// @Security     BearerAuth
func (f *FuncaoController) RegistraFuncao() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.Funcao

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error": err.Error(),
			})
			return
		}

		novaFuncao := model.Funcao{
			Funcao:         input.Funcao,
			IdDepartamento: input.IdDepartamento,
		}
		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		err := f.service.SalvarFuncao(ctx, novaFuncao, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrDadoDuplicado) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error": err.Error(),
				})
				return
			}

			if errors.Is(err, helper.ErrConflitoIntegridade) {
				ctx.JSON(http.StatusConflict, gin.H{
					"error": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{

			"mensagem": "função cadastrada",
		})

	}
}

// ListarFuncoes godoc
// @Summary      Listar todos
// @Description  Retorna uma lista com todos os funcoes
// @Tags         funcao
// @Produce      json
// @Success      200  {array}   model.FuncaoDto
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /funcoes [get]
// @Security     BearerAuth
func (f *FuncaoController) ListarFuncoes() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var filtro service.FiltroFuncao

		if err := ctx.ShouldBindQuery(&filtro); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error":    "parametros de busca invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		if filtro.Pagina <= 0 {
			filtro.Pagina = 1
		}
		if filtro.Quantidade <= 0 {
			filtro.Quantidade = 1000 // Padrão de 10 itens se não informar
		}

		funcoes, err := f.service.ListasTodasFuncao(ctx,filtro ,tenantID)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro interno ao listar funcoes",
			})
			return
		}

		ctx.JSON(http.StatusOK, funcoes)
	}
}


// DeletarFuncao godoc
// @Summary      Deletar funcao
// @Description  Remove (ou inativa) uma funcao pelo ID
// @Tags         funcao
// @Param        id   path      int  true  "ID da funcao"
// @Success      204  "Sem Conteúdo (Sucesso)"
// @Failure      400  {object}  helper.HTTPError "ID inválido"
// @Failure      404  {object}  helper.HTTPError "Não encontrado"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /funcao/{id} [delete]
// @Security     BearerAuth
func (f *FuncaoController) DeletarFuncao() gin.HandlerFunc {

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

		err = f.service.DeletarFuncao(ctx, id, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error": "funcao nao encontrada",
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

// UpdateFuncao godoc
// @Summary      Atualizar funcao
// @Description  Atualiza o nome de uma funcao e seu departamento existente
// @Tags         funcao
// @Accept       json
// @Produce      json
// @Param        id   path      int                      true  "ID da funcao"
// @Param        body body      model.Funcao true  "Novo nome"
// @Success      200  {object}  map[string]string "Sucesso"
// @Failure      400  {object}  helper.HTTPError "Erro de validação (ID ou Nome curto)"
// @Failure      404  {object}  helper.HTTPError "Não encontrado"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /funcao/{id} [put]
func (f *FuncaoController) AtualizarFuncao() gin.HandlerFunc {

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

		var input model.Funcao

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error": err.Error(),
			})
			return
		}

		err = f.service.AtualizarFuncao(ctx, id, input.Funcao, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrNomeCurto) {

				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error": "nome da funcao tem que possui 2 ou mais letras",
				})
				return
			}
			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": "funcao nao encontrado para atualizar",
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"erro": err.Error(),
			})

			return
		}

		ctx.JSON(http.StatusOK, gin.H{"sucesso": "funcao atualizado"})

	}
}
