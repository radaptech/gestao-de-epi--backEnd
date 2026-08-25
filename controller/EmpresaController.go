package controller

import (
	"context"
	"strconv"

	"log"
	"net/http"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/gin-gonic/gin"
)

type EmpresaService interface {
	Salvar(ctx context.Context, model model.EmpresaInserir) error
	EmpresaDashboard(ctx context.Context) (model.ResumoDashboard, error)
	EmpresaRecentes(ctx context.Context) ([]model.EmpresaRecente, error)
	DadosEmpresas(ctx context.Context) ([]model.Empresa, error)
	EditarEmpresa(ctx context.Context, id int32, model model.EditarEmpresaRequest) error
}

type EmpresaController struct {
	service EmpresaService
}

func NewEmpresaController(serv EmpresaService) *EmpresaController {

	return &EmpresaController{
		service: serv,
	}
}

// Salvar godoc
// @Summary      Criar empresa
// @Description  Cadastra uma nova empresa no sistema
// @Tags         Empresas
// @Accept       json
// @Produce      json
// @Param        empresa body model.EmpresaInserir true "Dados da empresa"
// @Success      201  {object}  map[string]string "Empresa criada"
// @Failure      400  {object}  helper.HTTPError "Dados inválidos"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /empresa [post]
// @Security     BearerAuth
func (e *EmpresaController) Salvar() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input model.EmpresaInserir

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "Dados inválidos ou formato incorreto.",
				"detalhes": err.Error(),
			})
			return
		}

		err := e.service.Salvar(ctx, input)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":    "Erro ao salvar empresa no banco de dados.",
				"detalhes": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{
			"sucesso": "Empresa cadastrada com sucesso",
		})
	}
}


// ResumoDashboard godoc
// @Summary      Dashboard
// @Description  Retorna o resumo para o dashboard da empresa
// @Tags         Empresas
// @Produce      json
// @Success      200  {object}  model.ResumoDashboard
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /dashboard [get]
// @Security     BearerAuth
func (e *EmpresaController) ResumoDashboard() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		resumo, err := e.service.EmpresaDashboard(ctx)
		if err != nil {
			log.Printf("erro dashbord: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    "erro ao realizar buscar dos dados do Dashbord",
				"detalhes": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, resumo)
	}
}


// EmpresaRecentes godoc
// @Summary      Empresas recentes
// @Description  Lista as empresas adicionadas recentemente
// @Tags         Empresas
// @Produce      json
// @Success      200  {array}   model.EmpresaRecente
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /empresas/recentes [get]
// @Security     BearerAuth
func (e *EmpresaController) EmpresaRecentes() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		empresa, err := e.service.EmpresaRecentes(ctx)
		if err != nil {
			log.Printf("erro dashbord: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    "erro ao realizar buscar dos dados das empresas",
				"detalhes": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, empresa)

	}
}

// DadosEmpresas godoc
// @Summary      Listar empresas
// @Description  Retorna uma lista com todas as empresas
// @Tags         Empresas
// @Produce      json
// @Success      200  {array}   model.Empresa
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /empresas [get]
// @Security     BearerAuth
func (e *EmpresaController) DadosEmpresas() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		empresa, err := e.service.DadosEmpresas(ctx)
		if err != nil {
			log.Printf("erro em pegar os dados das empresas: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    "erro ao realizar buscar dos dados das empresas",
				"detalhes": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, empresa)

	}
}

// EditarEmpresa godoc
// @Summary      Editar empresa
// @Description  Atualiza os dados de uma empresa existente
// @Tags         Empresas
// @Accept       json
// @Produce      json
// @Param        id   path      int                    true  "ID da empresa"
// @Param        body body      model.EditarEmpresaRequest true  "Dados para atualização"
// @Success      204  "Sem conteúdo (Sucesso)"
// @Failure      400  {object}  helper.HTTPError "ID ou dados inválidos"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /empresa/{id} [put]
// @Security     BearerAut
func (e *EmpresaController) EditarEmpresa() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		idparam := ctx.Param("id")
		id, err := strconv.Atoi(idparam)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "ID do plano inválido"})
			return
		}
		var input model.EditarEmpresaRequest

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"erro": err.Error(),
			})
			return
		}

		err = e.service.EditarEmpresa(ctx, int32(id), input)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    "erro ao editar empresa",
				"detalhes": err.Error(),
			})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}
