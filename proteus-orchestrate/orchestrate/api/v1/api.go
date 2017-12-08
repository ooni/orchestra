package apiv1

import (
	"github.com/thetorproject/proteus/proteus-common/middleware"
	"github.com/thetorproject/proteus/proteus-orchestrate/orchestrate/handler"

	"github.com/apex/log"
	"github.com/gin-gonic/gin"
)

var ctx = log.WithFields(log.Fields{
	"pkg": "apiv1",
})

// MakeRouter bind all the request handlers and middleware
func BindAPI(router *gin.Engine) error {
	v1 := router.Group("/api/v1")

	admin := v1.Group("/admin")
	admin.Use(authMiddleware.MiddlewareFunc(middleware.AdminAuthorizor))
	{
		admin.GET("/jobs", handler.HandleListJobs)
		admin.POST("/job", handler.HandleAddJob)
		admin.DELETE("/job/:job_id", handler.HandleDeleteJob)
	}

	device := v1.Group("/")
	device.Use(authMiddleware.MiddlewareFunc(middleware.DeviceAuthorizor))
	{
		device.GET("/rendezvous", handler.HandleRendezvous)

		device.GET("/tasks", handler.HandleListTasks)

		device.GET("/task/:task_id", handler.HandleGetTask)
		device.POST("/task/:task_id/accept", handler.HandleAcceptTask)
		device.POST("/task/:task_id/reject", handler.HandleRejectTask)
		device.POST("/task/:task_id/done", handler.HandleDoneTask)
	}
	return nil
}
