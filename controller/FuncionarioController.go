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

type FuncionarioService interface {
	SalvarFuncionario(ctx context.Context, model model.FuncionarioINserir, tenantId int32) error
	ListarFuncionario(ctx context.Context, matricula string, tenantId int32) (model.Funcionario_Dto, error)
	ListaTodosFuncionarios(ctx context.Context, f service.FiltroFuncionario, tenantId int32) (service.FuncionarioPaginado, error)
	DeletarFuncionario(ctx context.Context, id int, tenantId int32) error
	AtualizarFuncionarioCompleto(ctx context.Context, id int, req model.UpdateFuncionarioRequest, tenantId int) error
}

type FuncionarioController struct {
	Service FuncionarioService
}

func NewFuncionarioController(service FuncionarioService) *FuncionarioController {

	return &FuncionarioController{
		Service: service,
	}
}

// RegistraDepartamento godoc
// @Summary      Cadastrar novo funcionarios
// @Description  Cadastra um novo funcionario no sistema
// @Tags         funcionarios
// @Accept       json
// @Produce      json
// @Param        funcionario body model.FuncionarioINserir true "Dados do funcionario"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  helper.HTTPError "Dados inválidos"
// @Failure      409  {object}  helper.HTTPError "Departamento já existe"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /cadastro-funcionario [post]
// @Security     BearerAuth
func (f *FuncionarioController) Adicionar() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.FuncionarioINserir

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}


		novoFunc := model.FuncionarioINserir{
			Nome:            input.Nome,
			ID_departamento: input.ID_departamento,
			ID_funcao:       input.ID_funcao,
		}
		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		err := f.Service.SalvarFuncionario(ctx, novoFunc, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrDadoDuplicado) {
				ctx.JSON(http.StatusConflict, gin.H{

					"error": err.Error(),
				})
				return
			}

			if errors.Is(err, helper.ErrConflitoIntegridade) {

				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error": "departamento ou funcao nao encontrado",
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{

			"mensagem": "funcionario cadastrado",
		})
	}

}

// ListarFuncionarios godoc
// @Summary      Listar todos
// @Description  Retorna uma lista com todos os funcionarios
// @Tags         funcionarios
// @Produce      json
// @Success      200  {array}   model.Funcionario_Dto
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /funcionarios [get]
// @Security     BearerAuth
func (f *FuncionarioController) ListarFuncionarios() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var filtro service.FiltroFuncionario
		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		if err := ctx.ShouldBindQuery(&filtro); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error":    "parametros de busca invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		if filtro.Pagina <= 0 {
			filtro.Pagina = 1
		}
		if filtro.Quantidade <= 0 {
			filtro.Quantidade = 10 // Padrão de 10 itens se não informar
		}

		funcs, err := f.Service.ListaTodosFuncionarios(ctx, filtro, tenantID)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro interno ao listar departamentos",
			})
			return
		}

		ctx.JSON(http.StatusOK, funcs)
	}
}

// ListarFuncionarioPorMatricula godoc
// @Summary      Buscar por matricula
// @Description  Retorna os detalhes de um único funcionario
// @Tags         funcionarios
// @Produce      json
// @Param        id   path      int  true  "matricula do funcionario"
// @Success      200  {object}  model.Funcionario_Dto
// @Failure      400  {object}  helper.HTTPError "ID inválido"
// @Failure      404  {object}  helper.HTTPError "Não encontrado"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /funcionario/{matricula} [get]
// @Security     BearerAuth
func (f *FuncionarioController) ListarFuncionarioPorMatricula() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		matricula := ctx.Param("matricula")

		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		funcionario, err := f.Service.ListarFuncionario(ctx, matricula, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": "funcionario nao encontrado",
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, funcionario)
	}
}

// DeletarFuncionaioI godoc
// @Summary      Deletar funcionario
// @Description  Remove (ou inativa) um funcionario pelo ID
// @Tags         funcionarios
// @Param        id   path      int  true  "ID do funcionario"
// @Success      204  "Sem Conteúdo (Sucesso)"
// @Failure      400  {object}  helper.HTTPError "ID inválido"
// @Failure      404  {object}  helper.HTTPError "Não encontrado"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /funcionario/{id} [delete]
// @Security     BearerAuth
func (f *FuncionarioController) DeletarFuncionaioId() gin.HandlerFunc {

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

		err = f.Service.DeletarFuncionario(ctx, id, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": "funcionario nao encontrado",
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

// AtualizaFuncionario godoc
// @Summary      Atualizar funcionario
// @Description  Atualiza os dados de um funcionario existente
// @Tags         funcionarios
// @Accept       json
// @Produce      json
// @Param        id   path      int                      true  "ID do funcionario"
// @Param        body body      model.UpdateFuncionarioRequest true  "funcionario novos dados"
// @Success      200  {object}  map[string]string "Sucesso"
// @Failure      400  {object}  helper.HTTPError "Erro de validação (ID ou Nome curto)"
// @Failure      404  {object}  helper.HTTPError "Não encontrado"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /funcionario/{id} [patch]
func (f *FuncionarioController) AtualizaFuncionario() gin.HandlerFunc {

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

		var input model.UpdateFuncionarioRequest

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		err = f.Service.AtualizarFuncionarioCompleto(ctx, id, input, int(tenantID))
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": "funcionario nao encontrado",
				})

				return
			}

			if errors.Is(err, helper.ErrDadoDuplicado) {
				ctx.JSON(http.StatusConflict, gin.H{

					"error": "matricula ja cadastrada",
				})
				return
			}

			if errors.Is(err, helper.ErrConflitoIntegridade) {

				ctx.JSON(http.StatusUnprocessableEntity, gin.H{

					"error": "departamento ou funcao nao encontrado",
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return

		}

		ctx.JSON(http.StatusOK, gin.H{"sucesso": "funcionario atualizado"})

	}
}
