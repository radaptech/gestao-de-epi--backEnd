package controller

import (
	"context"
	"errors"
	"log"

	"net/http"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
)

type LoginService interface {
	Registrar(ctx context.Context, model model.Usuario, tenantId int32) error
	FazerLogin(ctx context.Context, email, senha string, tenantId int32) (string, repository.BuscarUsuarioPorEmailRow, error)
	BuscarPorId(ctx context.Context, id uint, tenantId int32) (model.RecuperaUser, error)
	ListarUsuario(ctx context.Context, tenantId int32) ([]model.UsuarioResponse, error)
	RecuperacaoSenha(ctx context.Context, rl model.RecuperaLogin) error
	RedefinirSenha(ctx context.Context, rs model.RedefinirSenha) error
}

type LoginController struct {
	service LoginService
}

func NewLoginController(service LoginService) *LoginController {

	return &LoginController{
		service: service,
	}
}

func (l *LoginController) Registrar() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.Usuario

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		novoUsuario := model.Usuario{

			Nome:  input.Nome,
			Email: input.Email,
			Senha: input.Senha,
			Role:  input.Role,
		}

		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}
		err := l.service.Registrar(ctx, novoUsuario, tenantID)
		if err != nil {

			if errors.Is(err, helper.ErrDadoDuplicado) {

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

		ctx.JSON(http.StatusCreated, gin.H{

			"mensagem": "usuario cadastrado",
		})
	}
}

func (l *LoginController) Login() gin.HandlerFunc {

	return func(c *gin.Context) {

		var input model.LoginInput

		if err := c.ShouldBindJSON(&input); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "Dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		tenantID, ok := middleware.GetTenantID(c)
		if !ok {
			c.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}
		token, user, err := l.service.FazerLogin(c, input.Email, input.Senha, tenantID)
		if err != nil {

			if err.Error() == "email ou senha inválidos" {

				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "email ou senha incorretos",
				})

				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro interno ao realizar login",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{

			"token": token,
			"usuario": gin.H{
				"id":    user.ID,
				"nome":  user.Nome,
				"email": user.Email,
				"role":  user.Role.String,
			},
		})
	}
}

func (l *LoginController) VerPerfil() gin.HandlerFunc {

	return func(c *gin.Context) {

		id, existe := c.Get("userId")
		if !existe {
			c.JSON(http.StatusUnauthorized, gin.H{

				"error": "Token inválido ou sem id",
			})

			return
		}
		tenantID, ok := middleware.GetTenantID(c)
		if !ok {
			c.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		usuario, err := l.service.BuscarPorId(c, id.(uint), tenantID)
		if err != nil {

			c.JSON(404, gin.H{"error": "Usuário não encontrado"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":    usuario.Id,
			"nome":  usuario.Nome,
			"email": usuario.Email,
			"role":  usuario.Role,
		})
	}
}

func (l *LoginController) ListarUsuario() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		users, err := l.service.ListarUsuario(ctx, tenantID)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    "erro interno do servidor",
				"detalhes": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, users)

	}
}

func (l *LoginController) SalvarToken() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.RecuperaLogin

		if err := ctx.ShouldBindJSON(&input); err != nil {

			log.Printf("erro: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Não foi possível processar a solicitação no momento. Tente novamente mais tarde.",
			})
			return
		}

		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}
 
		input.TenantId = int(tenantID)

		err := l.service.RecuperacaoSenha(ctx, input)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    "erro interno do servidor",
				"detalhes": err.Error(),
			})
			return
		}

		ctx.Status(http.StatusOK)
	}
}

func (l *LoginController) RedefinirSenha() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		var input model.RedefinirSenha

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		tenantID, ok := middleware.GetTenantID(ctx)
		if !ok {
			ctx.JSON(500, gin.H{"error": "Erro interno de tenant"})
			return
		}

		input.TenantId = int(tenantID)

		err := l.service.RedefinirSenha(ctx, input)
		if err != nil {
			if err.Error() == "link inválido, expirado ou não pertence a esta empresa" {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao redefinir senha."})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"mensagem": "Senha redefinida com sucesso! Você já pode fazer login.",
		})
	}
}
