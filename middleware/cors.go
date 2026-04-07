package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsConfig() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Remove barras no final para comparação segura
			origin = strings.TrimSuffix(origin, "/")

			if origin == "http://localhost:3000" ||
				origin == "https://sgepi-homologacao.radaptech.com.br" ||
				origin == "https://radaptech.com.br" ||
				origin == "https://www.radaptech.com.br" {
				return true
			}

			// Permitir qualquer subdominio
			if strings.HasSuffix(origin, ".radaptech.com.br") {
				return true
			}

			return false
		},

		AllowMethods: []string{"POST", "PUT", "GET", "PATCH", "DELETE", "OPTIONS"},

		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Tenant-ID", // Com T maiúsculo (Padrão)
			"X-tenant-id", // Todo minúsculo
			"X-tenant-ID", // Adicionei com t minúsculo por segurança
		},

		ExposeHeaders: []string{"Content-Length"},

		AllowCredentials: true,
		// O navegador armazena a permissão de CORS por esse tempo
		MaxAge: 12 * time.Hour,
	})
}
