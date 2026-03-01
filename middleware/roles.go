package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


func VerificaRole(rolePermitida string) gin.HandlerFunc{

	return  func(ctx *gin.Context) {

		roleAtual := ctx.GetString("user_role")

		if rolePermitida != roleAtual{

			ctx.JSON(http.StatusForbidden, gin.H{

				"erro": "Acesso negado,apenas " + rolePermitida + "s podem fazer isso",
			})
			ctx.Abort()
			return 
		}

		ctx.Next()

	}
}