package controller

import (
	"context"
	"errors"
	"log"
	"strconv"

	"net/http"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/helper"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/internal/model"
	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/middleware"
	"github.com/gin-gonic/gin"
)

type LoginService interface {
	Registrar(ctx context.Context, model model.Usuario) error
	FazerLogin(ctx context.Context, email, senha string, tenantId int32) (string, repository.BuscarUsuarioPorEmailRow, error)
	BuscarPorId(ctx context.Context, id uint, tenantId int32) (model.RecuperaUser, error)
	ListarUsuario(ctx context.Context, tenantId int32) ([]model.UsuarioResponse, error)
	RecuperacaoSenha(ctx context.Context, rl model.RecuperaLogin) error
	RedefinirSenha(ctx context.Context, rs model.RedefinirSenha) error
	UltimoAcesso(ctx context.Context, id, tenantId int32) error
	MostrarUsuariosPainel(ctx context.Context) ([]model.UsuarioResponsePainel, error)
	EditarUsuario(ctx context.Context, id int32, model model.EditarUsuarioRequest) error
	EditarStatusUsuario(ctx context.Context, id int32, model model.AlterarStatusRequest) error
	
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

			Nome:      input.Nome,
			Email:     input.Email,
			Senha:     input.Senha,
			Role:      input.Role,
			EmpresaID: input.EmpresaID,
		}

		err := l.service.Registrar(ctx, novoUsuario)
		if err != nil {

			if errors.Is(err, helper.ErrDadoDuplicado) {

				ctx.JSON(http.StatusConflict, gin.H{

					"error": err.Error(),
				})

				return
			}

			if errors.Is(err, helper.ErrLimiteExcedido) {
				ctx.JSON(http.StatusForbidden, gin.H{

					"error":"limite de usuarios excedidos",
					"detalhes":err.Error(),
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

//utilizando HTTP only
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

			log.Printf("erro ao realizar login: %v", err)
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

		err = l.service.UltimoAcesso(c, user.ID, tenantID)
		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{

				"error":    "Erro interno ao realizar login",
				"detalhes": err.Error(),
			})
			return

		}

		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie(
			"token",
			token,
			86400,
			"/",
			"",
			true, // botar para true depois
			true,
		)

		c.JSON(http.StatusOK, gin.H{
			"usuario": gin.H{
				"id":    user.ID,
				"nome":  user.Nome,
				"email": user.Email,
				"role":  user.Role.String,
			},
		})
	}
}

func (l *LoginController) Logout() gin.HandlerFunc {

	return  func(ctx *gin.Context) {

		ctx.SetCookie(
			"token",
			"",
			-1,
			"/",
			"",
			false,
			true,
		)

		ctx.JSON(http.StatusOK, gin.H{
            "message": "Logout realizado com sucesso",
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

		idConvertid:=uint(id.(int32)) 
		usuario, err := l.service.BuscarPorId(c, idConvertid, tenantID)
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

			log.Println("erro ao enviar email de recuperaçao: %w", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error": "erro interno do servidor",
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

func (l *LoginController) MostrarUsuariosPainel() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		users, err := l.service.MostrarUsuariosPainel(ctx)
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

func (l *LoginController) EditarUsuario() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		idParam := ctx.Param("id")
		idUsuario, err := strconv.Atoi(idParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "ID inválido",
			})
			return
		}
		log.Printf("[DEBUG] ID recebido para edição: %d", idUsuario)
		var input model.EditarUsuarioRequest

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		err = l.service.EditarUsuario(ctx, int32(idUsuario), input)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    "erro ao atualizar usuario",
				"detalhes": err.Error(),
			})
			return
		}

		ctx.Status(http.StatusNoContent)
	}
}

func (l *LoginController) EditarStatusUsuario() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		idParam := ctx.Param("id")
		idUsuario, err := strconv.Atoi(idParam)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "ID inválido",
			})
			return
		}
		
		var input model.AlterarStatusRequest

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":    "dados invalidos",
				"detalhes": err.Error(),
			})
			return
		}

		err = l.service.EditarStatusUsuario(ctx, int32(idUsuario), input)
		if err != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{

				"error":    "erro ao atualizar usuario",
				"detalhes": err.Error(),
			})
			return 
		}

		ctx.Status(http.StatusNoContent)
	}
}
