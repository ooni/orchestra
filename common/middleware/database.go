package middleware

import (
	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/ooni/orchestra/common"
)

var ctx = log.WithFields(log.Fields{
	"pkg": "middleware",
})

// RunMigrations runs the database migrations
func RunMigrations(db *sqlx.DB) error {
	migrations := &migrate.AssetMigrationSource{
		Asset:    common.Asset,
		AssetDir: common.AssetDir,
		Dir:      "common/data/migrations",
	}
	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		return err
	}
	ctx.Infof("performed %d migrations", n)
	return nil
}

// GinDatabaseMiddleware a database aware middleware.
// It will set the DB property, that can be accessed via:
// db := c.MustGet("DB").(*sqlx.DB)
type GinDatabaseMiddleware struct {
	DB *sqlx.DB
}

// MiddlewareFunc this is what you register as the middleware
func (mw *GinDatabaseMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("DB", mw.DB)
		c.Next()
	}
}

// InitDatabaseMiddleware create the middleware that injects the database
func InitDatabaseMiddleware(dbType string, dbString string) (*GinDatabaseMiddleware, error) {
	db, err := sqlx.Open(dbType, dbString)
	if err != nil {
		ctx.Error("failed to open database")
		return nil, err
	}
	if err := RunMigrations(db); err != nil {
		return nil, err
	}

	return &GinDatabaseMiddleware{
		DB: db,
	}, nil
}
