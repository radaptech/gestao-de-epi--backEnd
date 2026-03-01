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

	return func(ctx *gin.Context) {


		const portador = "Bearer "               //usado para fazer a verificação do Header
		header := ctx.GetHeader("Authorization") //busca a key

		if header == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token nao encontrado"})
			return
		}

		//verifica se a meu headaer NAO contem a palavra "Bearer", se nao tiver, da erro
		if !strings.HasPrefix(header, portador) {

			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "formato do token invalido"})
			return
		}

		//retirando a palavra "Bearer " do meu header, deixando apenas o token
		tokenString := header[len(portador):]

		//lendo o token e chamando a funcao anonima para retornar a chave secreta que esta no .env
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {

			// Verifica se o algoritmo é HMAC (o mesmo usado para criar a secret key)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inesperado: %v", t.Header["alg"])
			}

			//retornando a chave secreta
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		//verificando a validade do token
		if err != nil || !token.Valid {

			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token invalido ou expirado"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {

			// "sub"  guarda o ID do usuário
			if userId, ok := claims["sub"].(float64); ok {

				ctx.Set("userId", uint(userId)) //anotando o id do usuario no contexto
			}

			if role, ok:= claims["role"].(string); ok {

				ctx.Set("user_role", role)
			}

			if TenantId, ok:= claims["tenantId"].(float64); ok{

				ctx.Set("user_TenantId", int32(TenantId))
			}
		}else {
			
			ctx.JSON(http.StatusUnauthorized, gin.H{

				"error": "falha ao extrair os dados do token",
			})

			ctx.Abort()
			return 
		}

		ctx.Next()
	}
}
