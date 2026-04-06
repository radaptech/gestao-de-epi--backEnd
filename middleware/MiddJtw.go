package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	
)

func AutenticacaoJWT() gin.HandlerFunc {
    // Carregue a secret uma vez aqui (fora do loop de requisição)
    jwtSecret := []byte(os.Getenv("JWT_SECRET"))
    
    // Verificação de segurança: Se a secret estiver vazia, o servidor nem deve subir direito
    if len(jwtSecret) == 0 {
        panic("JWT_SECRET não configurada no ambiente!")
    }

    return func(ctx *gin.Context) {
        const portador = "Bearer "
        header := ctx.GetHeader("Authorization")

        if header == "" || !strings.HasPrefix(header, portador) {
            ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Acesso negado: Token ausente ou formato inválido"})
            return
        }

        tokenString := strings.TrimPrefix(header, portador)

        token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("método de assinatura inesperado: %v", t.Header["alg"])
            }
            return jwtSecret, nil
        })

        if err != nil || !token.Valid {
            ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token inválido ou expirado"})
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Falha ao processar permissões"})
            return
        }

        // Exemplo de validação crítica: Se não tiver o ID do usuário, o token é inútil
        userId, ok := claims["sub"].(float64)
        if !ok {
            ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token não contém identificação do usuário"})
            return
        }

        // Setando os dados no contexto
        ctx.Set("userId", uint(userId))
        
        if role, ok := claims["role"].(string); ok {
            ctx.Set("user_role", role)
        }
        
        if tenantId, ok := claims["tenantId"].(float64); ok {
            ctx.Set("user_TenantId", int32(tenantId))
        }

        if nome, ok := claims["nome"].(string); ok {
            ctx.Set("user_nome", nome)
        }

        ctx.Next()
    }
}
