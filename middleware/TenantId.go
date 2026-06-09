package middleware

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/davi-fernandesx/sistema-de-gestao-de-epi/database/repository"
	"github.com/gin-gonic/gin"
)

const TenantId = "tenantId"

func TenantMiddleware(querie *repository.Queries) gin.HandlerFunc {

	return func(c *gin.Context) {

		subdominio := strings.TrimSpace(c.GetHeader("X-tenant-ID"))

		if subdominio == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Tenant não informado (X-Tenant-ID ausente)"})
			return
		}

		// 3. Ignora o'www' ou 'api' se vierem no header
		if subdominio == "www" || subdominio == "api" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Subdomínio reservado"})
			return
		}

		// Busca no banco (usando o Context da request para cancelamento/timeout)
		empresa, err := querie.GetTenantBySubdomain(c.Request.Context(), subdominio)
		if err != nil {

			log.Printf("subdominio: %s", subdominio)
			log.Printf("erro: %v", err)
			if err == sql.ErrNoRows {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Empresa não encontrada"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao validar empresa",

				"detalhes": err.Error()})
			return
		}
		// SUCESSO: Salva o ID dentro do contexto do GIN
		// O Gin tem um mapa chave/valor interno otimizado (c.Set/c.Get)
		c.Set(TenantId, empresa.ID)

		// Passa para o próximo handler
		c.Next()
	}
}

// Helper para pegar o ID dentro dos Controllers de forma tipada
func GetTenantID(c *gin.Context) (int32, bool) {
	val, exists := c.Get(TenantId)
	if !exists {
		return 0, false
	}
	// Faz o cast para int32 (o tipo que o sqlc geralmente usa para IDs)
	id, ok := val.(int32)
	return id, ok
}
