package sched

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// GinSchedMiddleware a database aware middleware.
// It will set the DB property, that can be accessed via:
// db := c.MustGet("DB").(*sqlx.DB)
type GinSchedMiddleware struct {
	db        *sqlx.DB
	scheduler *Scheduler
}

// MiddlewareFunc this is what you register as the middleware
func (mw *GinSchedMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("Scheduler", mw.scheduler)
		c.Next()
	}
}

// InitSchedMiddleware create the middleware that injects the database
func InitSchedMiddleware(db *sqlx.DB) (*GinSchedMiddleware, error) {
	scheduler := NewScheduler(db)
	scheduler.Start()
	return &GinSchedMiddleware{
		db:        db,
		scheduler: scheduler,
	}, nil
}
