package registry

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-contrib/multitemplate"
	"github.com/ooni/orchestra/common/middleware"
	"github.com/ooni/orchestra/registry/registry/api/v1"

	"github.com/apex/log"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gopkg.in/gin-contrib/cors.v1"
)

var ctx = log.WithFields(log.Fields{
	"pkg": "registry",
	"cmd": "ooni-registry",
})

func loadTemplates(list ...string) multitemplate.Render {
	r := multitemplate.New()
	for _, x := range list {
		templateString, err := Asset("registry/data/templates/" + x)
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

// Start the registry server
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

	router := gin.Default()
	router.Use(dbMiddleware.MiddlewareFunc())
	router.Use(cors.New(middleware.CorsConfig()))
	router.HTMLRender = loadTemplates("home.tmpl")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.tmpl", gin.H{
			"title":                "ooni-registry",
			"componentName":        "ooni-registry",
			"componentDescription": LongDescription,
		})
	})
	err = apiv1.BindAPI(router, authMiddleware)
	if err != nil {
		ctx.WithError(err).Error("failed to BindAPI")
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
