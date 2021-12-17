package main

import (
	"log"
	"net/http"
	"time"

	"github.com/opensvc/collector-api/auth"
	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
	_ "github.com/opensvc/collector-api/docs"
	"github.com/opensvc/collector-api/routes"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/spf13/viper"
)

func fatal(err interface{}) {
	if err != nil {
		log.Fatal(err)
	}
}

//
// @title                       OpenSVC collector API
// @version                     1.0
// @description                 Organization clusters, nodes, services and more.
// @contact.name                OpenSVC SAS
// @contact.url                 https://www.opensvc.com
// @contact.email               collector-api-contact@opensvc.com
// @license.name                Apache License 2.0
// @license.url                 https://www.apache.org/licenses/LICENSE-2.0
// @BasePath                    /
// @securityDefinitions.basic   BasicAuth
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
//
func main() {
	if err := initConf(); err != nil {
		fatal(err)
	}
	if err := auth.Init(); err != nil {
		fatal(err)
	}
	if err := db.Init(); err != nil {
		fatal(err)
	}
	addr := viper.GetString("Listen")
	log.Printf("Starting server on %v\n", addr)
	if err := http.ListenAndServe(addr, router()); err != nil {
		fatal(err)
	}
}

func router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.Timeout(60 * time.Second))
	r.Mount("/swagger", httpSwagger.WrapHandler)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Route("/auth/node/token", func(r chi.Router) {
			r.Get("/", routes.GetNodeToken)
		})
		r.Route("/auth/user/token", func(r chi.Router) {
			r.Get("/", routes.GetUserToken)
		})
		r.Route("/nodes", func(r chi.Router) {
			r.Get("/", routes.GetNodes)
			r.Post("/", routes.PostNodes)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(tables.NodeCtx)
				r.Get("/", routes.GetNode)
				r.Delete("/", routes.DelNode)
				r.Post("/", routes.PostNode)
				r.Get("/candidate_tags", routes.GetNodeCandidateTags)
				r.Get("/tags", routes.GetNodeTags)
			})
			r.Route("/tags", func(r chi.Router) {
				r.Get("/", routes.GetNodesTags)
				r.Route("/{id}", func(r chi.Router) {
					r.Use(tables.NodeTagCtx)
					r.Get("/", routes.GetNodeTag)
				})
			})
		})
		r.Route("/services", func(r chi.Router) {
			r.Get("/", routes.GetServices)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(tables.ServiceCtx)
				r.Get("/", routes.GetService)
				r.Get("/candidate_tags", routes.GetServiceCandidateTags)
				r.Get("/tags", routes.GetServiceTags)
			})
			r.Route("/tags", func(r chi.Router) {
				r.Get("/", routes.GetServicesTags)
				r.Route("/{id}", func(r chi.Router) {
					r.Use(tables.ServiceTagCtx)
					r.Get("/", routes.GetServiceTag)
				})
			})
		})
		r.Route("/tags", func(r chi.Router) {
			r.Get("/", routes.GetTags)
			r.Post("/", routes.PostTags)
			r.Delete("/", routes.DelTags)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(tables.TagCtx)
				r.Get("/", routes.GetTag)
				r.Delete("/", routes.DelTag)
				r.Route("/nodes", func(r chi.Router) {
					r.Get("/", routes.GetTagNodes)
				})
				r.Route("/services", func(r chi.Router) {
					r.Get("/", routes.GetTagServices)
				})
			})
		})
	})

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome anonymous"))
		})
	})

	return r
}
