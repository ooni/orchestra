package apiv1

import (
	"github.com/ooni/orchestra/common/middleware"
	"github.com/ooni/orchestra/orchestrate/orchestrate/handler"

	"github.com/apex/log"
	"github.com/gin-gonic/gin"
)

var ctx = log.WithFields(log.Fields{
	"pkg": "apiv1",
	"cmd": "ooni-orchestrate",
})

// BindAPI bind all the request handlers and middleware
func BindAPI(router *gin.Engine, authMiddleware *middleware.GinJWTMiddleware) error {
	v1 := router.Group("/api/v1")

	admin := v1.Group("/admin")
	admin.Use(authMiddleware.MiddlewareFunc(middleware.AdminAuthorizor))
	{
		admin.GET("/alerts", handler.ListAlertsHandler)
		admin.POST("/alert", handler.AddAlertHandler)
		admin.DELETE("/alert/:alert_id", handler.DeleteAlertHandler)

		admin.POST("/experiment", handler.AddExperimentHandler)
	}

	rendezvous := v1.Group("/")
	// This means that authentication is optional
	rendezvous.Use(authMiddleware.MiddlewareFunc(middleware.NullAuthorizor))
	{
		rendezvous.GET("/urls", handler.URLsHandler)
		rendezvous.GET("/test-list/urls", handler.URLsHandler)
		rendezvous.GET("/collectors", handler.CollectorsHandler)
		rendezvous.GET("/test-helpers", handler.TestHelpersHandler)
	}

	device := v1.Group("/")
	device.Use(authMiddleware.MiddlewareFunc(middleware.DeviceAuthorizor))
	{
		/*
			experiments.GET("/experiments", handler.ListExperimentsHandler)
			experiments.GET("/experiment/:exp_id", handler.GetExperimentHandler)
			experiments.GET("/experiment/:exp_id/accept", handler.AcceptExperimentHandler)
			experiments.GET("/experiment/:exp_id/reject", handler.RejectExperimentHandler)
			experiments.GET("/experiment/:exp_id/done", handler.DoneExperimentHandler)
		*/

		device.GET("/tasks", handler.ListTasksHandler)
		device.GET("/task/:task_id", handler.GetTaskHandler)
		device.POST("/task/:task_id/accept", handler.AcceptTaskHandler)
		device.POST("/task/:task_id/reject", handler.RejectTaskHandler)
		device.POST("/task/:task_id/done", handler.DoneTaskHandler)
	}
	return nil
}
