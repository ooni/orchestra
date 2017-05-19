package proteus_mw

import (
	"time"
	"gopkg.in/gin-contrib/cors.v1"
)

func CorsConfig() cors.Config {
	return cors.Config{
		AllowMethods:		[]string{"GET", "POST", "PUT", "HEAD", "DELETE"},
		AllowHeaders:		[]string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowAllOrigins:	true,
		AllowCredentials:	false,
		MaxAge:				12 * time.Hour,
	}
}

