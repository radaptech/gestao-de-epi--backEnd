package middleware

import "github.com/gin-gonic/gin"

const userId = "userId"

func GetUserID(c *gin.Context) (int32, bool) {
	val, exists := c.Get(userId)
	if !exists {
		return 0, false
	}
	// Faz o cast para int32 (o tipo que o sqlc geralmente usa para IDs)
	id, ok := val.(int32)
	return id, ok
}
