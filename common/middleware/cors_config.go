package middleware

import (
	"gopkg.in/gin-contrib/cors.v1"
	"time"
)

// CorsConfig stores the Cross Origin Resource Sharing configuration for orchestra
func CorsConfig() cors.Config {
	return cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "HEAD", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowAllOrigins:  true,
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
}
