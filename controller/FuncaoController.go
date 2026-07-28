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

type FuncaoService interface {
	SalvarFuncao(ctx context.Context, model model.Funcao, tenantid int32) error
	ListasTodasFuncao(ctx context.Context, f service.FiltroFuncao, tenantId int32) (service.FuncaoPaginado, error)
	DeletarFuncao(ctx context.Context, id int, tenantId int32) error
	AtualizarFuncao(ctx context.Context, id int, funcao string, tenantId int32) error
	BuscaDepartamentosParaFuncao(ctx context.Context, tenantId int32) (map[string]int, error)
}

type FuncaoController struct {
	service FuncaoService
}

func NewFuncaoController(service FuncaoService) *FuncaoController {

	return &FuncaoController{service: service}
}

func (f *FuncaoController) ImportarFuncaoXLSX() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		fileHearder, err := ctx.FormFile("file")
		if err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{
				"erro":     "selecione um arquivo de planilha valido",
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

				"erro":     "o arquivo enviado não é uma planilha Excel  valida",
				"detalhes": err.Error(),
			})
			return
		}
		defer fi.Close()

		NomeDaPlanilha := fi.GetSheetName(0)
		linhas, err := fi.GetRows(NomeDaPlanilha)
		if err != nil || len(linhas) == 0 {

			ctx.JSON(http.StatusBadRequest, gin.H{
				"mensagem": "a planilha selecionada esta vazia",
				"detalhes": err.Error(),
			})
			return
		}

		tenantID, exists := middleware.GetTenantID(ctx)
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Sessão inválida ou expirada."})
			return
		}

		mapDep, err := f.service.BuscaDepartamentosParaFuncao(ctx, tenantID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message":  "Erro ao carregar departamentos para validação.",
				"detalhes": err.Error(),
			})
			return
		}

		cabecalho := false
		totalImportadas := 0
		totalIgnoradas := 0

		for indexLinha, linha := range linhas {
			if len(linha) < 2 {
				continue
			}

			colunaFuncao := strings.TrimSpace(linha[0])
			colunaDepartamento := strings.TrimSpace(linha[1])

			// Procura a linha do cabeçalho
			if strings.EqualFold(colunaFuncao, "Nome da Função") {
				cabecalho = true
				continue
			}

			if !cabecalho || colunaDepartamento == "" || colunaFuncao == "" {
				continue
			}

			// Busca o ID do departamento no mapa
			depId, exist := mapDep[strings.ToLower(colunaDepartamento)]
			if !exist {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"message": fmt.Sprintf("Erro na linha %d: O departamento '%s' não existe no sistema. Cadastre-o primeiro.", indexLinha+1, colunaDepartamento),
				})
				return
			}

			err = f.service.SalvarFuncao(ctx, model.Funcao{
				Funcao:         colunaFuncao,
				IdDepartamento: depId,
			}, tenantID)

			if err != nil {
				// 👉 OPÇÃO B: Se for duplicada ou conflito de integridade, apenas ignora e pula para a próxima!
				if errors.Is(err, helper.ErrDadoDuplicado) || errors.Is(err, helper.ErrConflitoIntegridade) {
					totalIgnoradas++
					continue // Pula para a próxima linha do loop
				}

				// Qualquer outro erro de banco (ex: conexão, query quebrada) aborta a requisição
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"message":  fmt.Sprintf("Erro interno ao salvar a função '%s' na linha %d.", colunaFuncao, indexLinha+1),
					"detalhes": err.Error(),
				})
				return
			}

			totalImportadas++
		}

		// Se nada foi importado porque tudo era duplicado ou inválido
		if totalImportadas == 0 && totalIgnoradas == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "Nenhuma função válida foi encontrada na planilha.",
			})
			return
		}

		// Resposta de sucesso informando quantos foram inseridos e quantos já existiam
		mensagemSucesso := fmt.Sprintf("%d função(ões) importada(s) com sucesso!", totalImportadas)
		if totalIgnoradas > 0 {
			mensagemSucesso = fmt.Sprintf("%d função(ões) importada(s) e %d ignorada(s) por já existirem.", totalImportadas, totalIgnoradas)
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message":    mensagemSucesso,
			"importados": totalImportadas,
			"ignorados":  totalIgnoradas,
		})

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Funções importadas com sucesso!",
		})
	}
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

		funcoes, err := f.service.ListasTodasFuncao(ctx, filtro, tenantID)
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
