package apiv1

import (
	"github.com/thetorproject/proteus/proteus-common/middleware"
	"github.com/thetorproject/proteus/proteus-orchestrate/orchestrate/handler"

	"github.com/apex/log"
	"github.com/gin-gonic/gin"
)

var ctx = log.WithFields(log.Fields{
	"pkg": "apiv1",
	"cmd": "proteus-orchestrate",
})

// BindAPI bind all the request handlers and middleware
func BindAPI(router *gin.Engine, authMiddleware *middleware.GinJWTMiddleware) error {
	v1 := router.Group("/api/v1")

	admin := v1.Group("/admin")
	admin.Use(authMiddleware.MiddlewareFunc(middleware.AdminAuthorizor))
	{
		admin.GET("/jobs", handler.ListJobsHandler)
		admin.POST("/job", handler.AddJobHandler)
		admin.DELETE("/job/:job_id", handler.DeleteJobHandler)
	}

	rendezvous := v1.Group("/")
	// XXX we should make authentication here optional
	{
		rendezvous.GET("/urls", handler.URLsHandler)
		rendezvous.GET("/collectors", handler.CollectorsHandler)
		rendezvous.GET("/test-helpers", handler.TestHelpersHandler)
	}

	device := v1.Group("/")
	device.Use(authMiddleware.MiddlewareFunc(middleware.DeviceAuthorizor))
	{
		device.GET("/tasks", handler.ListTasksHandler)

		device.GET("/task/:task_id", handler.GetTaskHandler)
		device.POST("/task/:task_id/accept", handler.AcceptTaskHandler)
		device.POST("/task/:task_id/reject", handler.RejectTaskHandler)
		device.POST("/task/:task_id/done", handler.DoneTaskHandler)
	}
	return nil
}
