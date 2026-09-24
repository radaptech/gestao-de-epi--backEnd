package middleware

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/radaptech/ginmw"
)

func LoggerComUsuario() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Processa a requisição
		c.Next()

		// Depois que terminou, pega o status
		status := c.Writer.Status()

		// Tenta pegar o ID do usuário (se estiver logado)
		usuarioID, existe := ginmw.UserID(c)
		userLog := "Anonimo"
		if existe {
			userLog = fmt.Sprintf("User:%v", usuarioID)
		}

		// Se deu erro (400 ou 500), imprime um log mais detalhado
		if status >= 400 {
			log.Printf("⚠️  ERRO | %d | %s | %s | %s",
				status,
				c.Request.Method,
				c.Request.URL.Path,
				userLog,
			)
		}
	}
}
