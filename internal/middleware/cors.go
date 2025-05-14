package middleware

import "github.com/gin-gonic/gin"

// CORSHeaders is a map of CORS headers
var CORSHeaders = map[string]bool{
	"Access-Control-Allow-Origin":      true,
	"Access-Control-Allow-Methods":     true,
	"Access-Control-Allow-Headers":     true,
	"Access-Control-Allow-Credentials": true,
	"Access-Control-Expose-Headers":    true,
}

// CORSMiddleware returns a middleware that adds CORS headers to all responses
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

		// Handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}