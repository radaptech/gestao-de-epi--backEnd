package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/service"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"

	storage_go "github.com/supabase-community/storage-go"
)

type EntregasService interface {
	Salvar(ctx context.Context, model model.EntregaParaInserir, tenantid int32, token string) error
	ListaEntregas(ctx context.Context, f service.FiltroEntregas, tenantId int32) (service.EntregaPaginada, error)
	CancelarEntrega(ctx context.Context, tenantId, id, iduser int) error
	GerarDadosPdfService(ctx context.Context, matricula string, tenantId int32) (helper.DadosPdf, error)
	BuscaEntregaDash(ctx context.Context, tenantId int32) ([]model.EntregaDashbord, error)
	BuscaItemDash(ctx context.Context, tenantID int32) ([]model.EntregaItensDashBord, error)
	TokenEntrega(ctx context.Context, tenantId, idFuncionario int32) (string, error)
}

type EntregaController struct {
	Service EntregasService
}

func NewEntregaController(service EntregasService) *EntregaController {

	return &EntregaController{
		Service: service,
	}
}

func (e *EntregaController) UploadAssinaturaSupabase(base64str string, token string) (string, error) {

	if strings.Contains(base64str, ",") {
		base64str = strings.Split(base64str, ",")[1]
	}

	decodificador, err := base64.StdEncoding.DecodeString(base64str)
	if err != nil {

		return "", fmt.Errorf("erro ao decodificar string. %w", err)
	}

	supabaseUrl := os.Getenv("SUPABASE_URL")
	secretKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	bucket := os.Getenv("SUPABASE_BUCKET")

	cliente := storage_go.NewClient(supabaseUrl+"/storage/v1", secretKey, nil)

	contentType := "image/png"
	arquivo := fmt.Sprintf("%s_%d.png", token, time.Now().Unix())
	opts := storage_go.FileOptions{
		ContentType: &contentType, // 🌟 Força o formato PNG
	}
	_, errS := cliente.UploadFile(bucket, arquivo, bytes.NewReader(decodificador), opts)
	if errS != nil {

		return "", fmt.Errorf("erro ao enviar o arquivo para o supabase, %v", errS)
	}

	urlPublic := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseUrl, bucket, arquivo)

	return urlPublic, err
}

func (e *EntregaController) Adicionar() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input model.EntregaParaInserir

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
				"erro": "erro au setar usuario",
			})
			return
		}

		// 1. Gera o Token de Auditoria
		token, errToken := e.Service.TokenEntrega(ctx, tenantId, input.ID_funcionario)
		if errToken != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao gerar token de auditoria"})
			return
		}

		// 2. Faz o Upload da Assinatura para o Bucket
		urlAssinatura, errA := e.UploadAssinaturaSupabase(input.Assinatura_Digital, token)
		if errA != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":    "falha ao salvar assinatura digital",
				"detalhes": errA.Error(),
			})
			return
		}

		// 3. Atualiza o input com a URL do bucket antes de mandar para o Service
		// Assim você economiza memória não criando uma struct nova do zero
		input.Assinatura_Digital = urlAssinatura
		input.Id_user = userId

		// 4. Salva no Banco de Dados (Transação)
		err := e.Service.Salvar(ctx, input, tenantId, token)
		if err != nil {
			// Tratamento de erros específicos
			if errors.Is(err, helper.ErrNaoEncontrado) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "funcionario ou registro não encontrado"})
				return
			}
			if strings.Contains(err.Error(), "estoque insuficiente") {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao salvar entrega", "detalhes": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"mensagem": "entrega cadastrada com sucesso",
			 // Opcional: retornar o token para o front
		})
	}
}

func (e *EntregaController) ListarEntregas() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var filtro service.FiltroEntregas

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

		entregas, err := e.Service.ListaEntregas(ctx.Request.Context(), filtro, tenantId)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": "erro ao realizar buscar das entregas de epi",
			})
			return
		}

		ctx.JSON(http.StatusOK, entregas)
	}
}

func (e *EntregaController) CancelarEntrega() gin.HandlerFunc {

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

		idUser, existe := ctx.Get("userId")
		if !existe {
			ctx.JSON(http.StatusUnauthorized, gin.H{

				"error": "Token inválido ou sem id",
			})

			return
		}

		err = e.Service.CancelarEntrega(ctx, int(tenantId), id, int(idUser.(uint)))
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {

				ctx.JSON(http.StatusNotFound, gin.H{

					"error":    "entrega não encontrada",
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

func (e *EntregaController) GerarFichaEpiPDF() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		matricula := ctx.Param("matricula")
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

		fmt.Printf("DEBUG: Matricula do Param: '%s' | Tenant do Middleware: %d\n", matricula, tenantId)
		entregaDadosPdf, err := e.Service.GerarDadosPdfService(ctx.Request.Context(), matricula, tenantId)
		if err != nil {

			if errors.Is(err, helper.ErrNaoEncontrado) {
				ctx.JSON(http.StatusUnprocessableEntity, gin.H{
					"error":    "dados nao encontrados",
					"detalhes": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    err.Error(),
				"detalhes": "dados não obtidos para gerar o pdf",
			})
			return
		}

		documento, err := helper.CreatePdf(entregaDadosPdf, auditoria, responsavel)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    err.Error(),
				"detalhes": "erro na geração do pdf",
			})
			return

		}

		// Baixa o arquivo direto no navegador
		ctx.Header("Content-Disposition", "attachment; filename=Ficha_EPI_"+matricula+".pdf")
		ctx.Data(http.StatusOK, "application/pdf", documento.GetBytes()) // Extrai os bytes limpos!
	}
}

func (e *EntregaController) BuscarEntregaDashbord() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno de tenant",
			})
			return
		}

		entregas, err := e.Service.BuscaEntregaDash(ctx, tenantId)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, entregas)

	}
}

func (e *EntregaController) BuscarEntregaItenDashbord() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantId, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno de tenant",
			})
			return
		}

		itens, err := e.Service.BuscaItemDash(ctx, tenantId)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, itens)

	}
}
