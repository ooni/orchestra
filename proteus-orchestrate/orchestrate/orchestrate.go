package orchestrate

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/apex/log"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/rubenv/sql-migrate"
	"github.com/spf13/viper"
	"gopkg.in/gin-contrib/cors.v1"

	"github.com/thetorproject/proteus/proteus-common/middleware"

	"github.com/thetorproject/proteus/proteus-orchestrate/orchestrate/api/v1"
	"github.com/thetorproject/proteus/proteus-orchestrate/orchestrate/handler"
	"github.com/thetorproject/proteus/proteus-orchestrate/orchestrate/sched"
)

var ctx = log.WithFields(log.Fields{
	"cmd": "events",
})

func initDatabase() (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", viper.GetString("database.url"))
	if err != nil {
		ctx.Error("failed to open database")
		return nil, err
	}
	return db, err
}

func runMigrations(db *sqlx.DB) error {
	migrations := &migrate.AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "proteus-common/data/migrations",
	}
	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		return err
	}
	ctx.Infof("performed %d migrations", n)
	return nil
}

func loadTemplates(list ...string) multitemplate.Render {
	r := multitemplate.New()
	for _, x := range list {
		templateString, err := Asset("data/templates/" + x)
		if err != nil {
			ctx.WithError(err).Error("failed to load template")
		}

		tmplMessage, err := template.New(x).Parse(string(templateString))
		if err != nil {
			ctx.WithError(err).Error("failed to parse template")
		}
		r.Add(x, tmplMessage)
	}
	return r
}

// Start starts the events backend including the web handlers
func Start() {
	var (
		err error
		db  *sqlx.DB
	)
	db, err = initDatabase()
	if err != nil {
		ctx.WithError(err).Error("failed to connect to DB")
		return
	}
	defer db.Close()

	err = runMigrations(db)
	if err != nil {
		ctx.WithError(err).Error("failed to run DB migration")
		return
	}

	authMiddleware, err := middleware.InitAuthMiddleware(db)
	if err != nil {
		ctx.WithError(err).Error("failed to initialise authMiddlewareDevice")
		return
	}

	scheduler := sched.NewScheduler(db)
	handler.InitHandlers(scheduler, db)

	router := gin.Default()
	router.Use(cors.New(middleware.CorsConfig()))
	router.HTMLRender = loadTemplates("home.tmpl")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.tmpl", gin.H{
			"title":                "proteus-events",
			"componentName":        "proteus-events",
			"componentDescription": LongDescription,
		})
	})
	err = apiv1.BindAPI(router)
	if err != nil {
		ctx.WithError(err).Error("failed to BinAPI")
		return
	}

	Addr := fmt.Sprintf("%s:%d", viper.GetString("api.address"),
		viper.GetInt("api.port"))
	ctx.Infof("starting on %s", Addr)

	scheduler.Start()
	s := &http.Server{
		Addr:    Addr,
		Handler: router,
	}
	gracehttp.Serve(s)
}
