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
	"github.com/spf13/viper"
	"gopkg.in/gin-contrib/cors.v1"

	"github.com/thetorproject/proteus/proteus-common/middleware"

	"github.com/thetorproject/proteus/proteus-orchestrate/orchestrate/api/v1"
	"github.com/thetorproject/proteus/proteus-orchestrate/orchestrate/sched"
)

var ctx = log.WithFields(log.Fields{
	"pkg": "orchestrate",
	"cmd": "proteus-orchestrate",
})

func initDatabase() (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", viper.GetString("database.url"))
	if err != nil {
		ctx.Error("failed to open database")
		return nil, err
	}
	return db, err
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
	)

	dbMiddleware, err := middleware.InitDatabaseMiddleware("postgres", viper.GetString("database.url"))
	if err != nil {
		ctx.WithError(err).Error("failed to init database middleware")
		return
	}
	defer dbMiddleware.DB.Close()

	authMiddleware, err := middleware.InitAuthMiddleware(dbMiddleware.DB)
	if err != nil {
		ctx.WithError(err).Error("failed to initialise authMiddlewareDevice")
		return
	}
	schedMiddleware, err := sched.InitSchedMiddleware(dbMiddleware.DB)
	if err != nil {
		ctx.WithError(err).Error("failed to initialise schedMiddleware")
		return
	}

	router := gin.Default()
	router.Use(schedMiddleware.MiddlewareFunc())
	router.Use(dbMiddleware.MiddlewareFunc())
	router.Use(cors.New(middleware.CorsConfig()))
	router.HTMLRender = loadTemplates("home.tmpl")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.tmpl", gin.H{
			"title":                "proteus-events",
			"componentName":        "proteus-events",
			"componentDescription": LongDescription,
		})
	})
	err = apiv1.BindAPI(router, authMiddleware)
	if err != nil {
		ctx.WithError(err).Error("failed to BinAPI")
		return
	}

	Addr := fmt.Sprintf("%s:%d", viper.GetString("api.address"),
		viper.GetInt("api.port"))
	ctx.Infof("starting on %s", Addr)

	s := &http.Server{
		Addr:    Addr,
		Handler: router,
	}
	gracehttp.Serve(s)
}
