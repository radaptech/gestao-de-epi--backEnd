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
            origin = strings.TrimSuffix(origin, "/")

            // 👇 ADICIONE O LOCALHOST:5173 (ou a porta exata que seu Front está rodando)
            if origin == "http://localhost:3000" ||
                origin == "https://sgepi-homologacao.radaptech.com.br" ||
                origin == "https://radaptech.com.br" ||
                origin == "https://www.radaptech.com.br" {
                return true
            }

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
            "X-Tenant-ID", 
            "X-tenant-id", 
            "X-tenant-ID",
            
        },

        // 👇 ADICIONE Content-Disposition PARA DOWNLOADS DE ARQUIVO
        ExposeHeaders: []string{"Content-Length", "Content-Disposition"},

        AllowCredentials: true,
        MaxAge: 12 * time.Hour,
    })
}
