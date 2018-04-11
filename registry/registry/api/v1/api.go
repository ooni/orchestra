package apiv1

import (
	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/ooni/orchestra/common/middleware"
	"github.com/ooni/orchestra/registry/registry/handler"
)

var ctx = log.WithFields(log.Fields{
	"pkg": "registry",
	"cmd": "ooni-registry",
})

// BindAPI bind all the request handlers and middleware
func BindAPI(router *gin.Engine, authMiddleware *middleware.GinJWTMiddleware) error {
	v1 := router.Group("/api/v1")

	v1.POST("/login", authMiddleware.LoginHandler)
	v1.POST("/register", handler.RegisterHandler)

	admin := v1.Group("/admin")
	admin.Use(authMiddleware.MiddlewareFunc(middleware.AdminAuthorizor))
	{
		admin.GET("/clients", handler.ListClientsHandler)
	}

	device := v1.Group("/")
	device.Use(authMiddleware.MiddlewareFunc(middleware.DeviceAuthorizor))
	{
		// XXX do we also want to support a PATCH method?
		device.PUT("/update/:client_id", handler.UpdateHandler)
	}

	return nil
}
