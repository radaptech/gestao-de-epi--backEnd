package controller

import (
	"context"
	"errors"
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

type DepartamentoService interface {
	SalvarDepartamento(ctx context.Context, tenantId int32, m model.Departamento) (model.DepartamentoDto, error)
	ListarTodosDepartamentos(ctx context.Context, f service.FiltroDepartamento, tenantId int32) (service.DepartamentoPaginado, error)
	DeletarDepartamento(ctx context.Context, id int, tenantId int32) error
	AtualizarDepartamento(ctx context.Context, id int32, novoNome string, tenantId int32) error
}

type DepartamentoController struct {
	service DepartamentoService
}

func NewDepartamentoController(service DepartamentoService) *DepartamentoController {

	return &DepartamentoController{service: service}
}

func (d *DepartamentoController) ImportDepartamentoXLSX() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		fileHearder, err := ctx.FormFile("file")
		if err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"erro:":     "selecione um arquivo de planilhas valido",
				"detalhes:": err.Error(),
			})
			return
		}

		//Abre o arquivo diretamente do stream da requisição
		filer, err := fileHearder.Open()
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"erro":     "erro ao ler a planilha",
				"detalhes": err.Error(),
			})
			return
		}
		defer filer.Close()

		// Inicializa o leitor do Excelize a partir do buffer na memória
		f, err := excelize.OpenReader(filer)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{

				"erro":     "O arquivo enviado não é uma planilha Excel (.xlsx) válida.",
				"detalhes": err.Error(),
			})
			return
		}
		defer f.Close()

		NomeDaPlanilha := f.GetSheetName(0)
		linhas, err := f.GetRows(NomeDaPlanilha)
		if err != nil || len(linhas) == 0 {

			ctx.JSON(http.StatusBadRequest, gin.H{
				"message":  "A planilha selecionada está vazia.",
				"detalhes": err.Error(),
			})
			return
		}

		var deps []string
		cabecalho := false

		for _, linha := range linhas {
			if len(linha) == 0 {
				continue
			}

			valorColuna := strings.TrimSpace(linha[0])
			if valorColuna == "" {
				continue
			}

			if strings.EqualFold(valorColuna, "Nome do Departamento") {
				cabecalho = true
				continue // Pula a linha do próprio cabeçalho
			}
			if !cabecalho {
				continue
			}

			deps = append(deps, valorColuna)

		}

		if len(deps) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "Nenhum departamento válido foi encontrado na planilha.",
			})

			return
		}

		tenantID, exists := middleware.GetTenantID(ctx)
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Sessão inválida ou expirada."})
			return
		}
		for _, dep := range deps {

			var departamento model.Departamento

			departamento.Departamento = dep

			_, err := d.service.SalvarDepartamento(ctx, tenantID, departamento)
			if err != nil {

				if errors.Is(err, helper.ErrDadoDuplicado) {
				ctx.JSON(http.StatusConflict, gin.H{

					"error":    "departamento ja existe no sistema",
					"detalhes": err.Error(),
				})
				return 
				}
				ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Erro ao registrar os departamentos no banco de dados."})
				return
				
			}
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Importação concluída com sucesso!",
			"total":   len(deps),
		})
	}
}

// RegistraDepartamento godoc
// @Summary      Criar novo departamento
// @Description  Cadastra um novo departamento no sistema
// @Tags         Departamentos
// @Accept       json
// @Produce      json
// @Param        departamento body model.Departamento true "Dados do departamento"
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  helper.HTTPError "Dados inválidos"
// @Failure      409  {object}  helper.HTTPError "Departamento já existe"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /cadastro-departamento [post]
// @Security     BearerAuth
func (d *DepartamentoController) RegistraDepartamento() gin.HandlerFunc {

	return func(c *gin.Context) {

		var input model.Departamento

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		novoDep := model.Departamento{
			Departamento: input.Departamento,
		}
		tenantID, ok := middleware.GetTenantID(c)
		if !ok {
			c.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		depStts, err := d.service.SalvarDepartamento(c, tenantID, novoDep)
		if err != nil {

			if errors.Is(err, helper.ErrDadoDuplicado) {
				c.JSON(http.StatusConflict, gin.H{

					"error":    "departamento ja existe no sistema",
					"detalhes": err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{

			"mensagem":     "departamento cadastrado",
			"departamento": depStts,
		})
	}
}

// ListarDepartamentos godoc
// @Summary      Listar todos
// @Description  Retorna uma lista com todos os departamentos
// @Tags         Departamentos
// @Produce      json
// @Success      200  {array}   model.DepartamentoDto
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /departamentos [get]
// @Security     BearerAuth
func (d *DepartamentoController) ListarDepartamentos() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var filtro service.FiltroDepartamento

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

		deps, err := d.service.ListarTodosDepartamentos(ctx, filtro, tenantID)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    "erro ao realizar buscar dos departamentos",
				"detalhes": err.Error(),
			})
			return

		}

		ctx.JSON(http.StatusOK, deps)

	}
}

// DeletarDepartamento godoc
// @Summary      Deletar departamento
// @Description  Remove (ou inativa) um departamento pelo ID
// @Tags         Departamentos
// @Param        id   path      int  true  "ID do Departamento"
// @Success      204  "Sem Conteúdo (Sucesso)"
// @Failure      400  {object}  helper.HTTPError "ID inválido"
// @Failure      404  {object}  helper.HTTPError "Não encontrado"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /departamento/{id} [delete]
// @Security     BearerAuth
func (d *DepartamentoController) DeletarDepartamento() gin.HandlerFunc {

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
		err = d.service.DeletarDepartamento(ctx, id, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": "departamento nao encontrado",
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

// UpdateDepartamento godoc
// @Summary      Atualizar departamento
// @Description  Atualiza o nome de um departamento existente
// @Tags         Departamentos
// @Accept       json
// @Produce      json
// @Param        id   path      int                      true  "ID do Departamento"
// @Param        body body      model.Departamento true  "Novo nome"
// @Success      200  {object}  map[string]string "Sucesso"
// @Failure      400  {object}  helper.HTTPError "Erro de validação (ID ou Nome curto)"
// @Failure      404  {object}  helper.HTTPError "Não encontrado"
// @Failure      500  {object}  helper.HTTPError "Erro interno"
// @Router       /departamentos/{id} [put]
func (d *DepartamentoController) AtualizarDepartamento() gin.HandlerFunc {

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
		var input model.Departamento

		if err := ctx.ShouldBindJSON(&input); err != nil {

			ctx.JSON(http.StatusBadRequest, gin.H{

				"error": err.Error(),
			})
			return
		}

		err = d.service.AtualizarDepartamento(ctx, int32(id), input.Departamento, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrNomeCurto) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": "nome do departamento tem que possui 2 ou mais letras",
				})
				return
			}
			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error": "departamento nao encontrado para atualizar",
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"erro": err.Error(),
			})

			return
		}

		ctx.JSON(http.StatusOK, gin.H{"sucesso": "departamento atualizado"})
	}
}
