package main

import (
	"crypto/rsa"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/opensvc/collector-api/docs"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/spf13/viper"
)

var (
	tokenAuth        *jwtauth.JWTAuth
	verifyBytes      []byte
	verifyKey        *rsa.PublicKey
	signKey          *rsa.PrivateKey
	jwtSignKeyFile   string
	jwtVerifyKeyFile string
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
	if err := initJWT(); err != nil {
		fatal(err)
	}
	if err := initAuth(); err != nil {
		fatal(err)
	}
	if err := initDB(); err != nil {
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
		r.Use(authMiddleware)
		r.Route("/auth/node/token", func(r chi.Router) {
			r.Get("/", getNodeToken)
		})
		r.Route("/auth/user/token", func(r chi.Router) {
			r.Get("/", getUserToken)
		})
		r.Route("/nodes", func(r chi.Router) {
			r.Get("/", getNodes)
			r.Post("/", postNodes)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(nodeCtx)
				r.Get("/", getNode)
				r.Delete("/", delNode)
				r.Get("/candidate_tags", getNodeCandidateTags)
				r.Get("/tags", getNodeTags)
			})
			r.Route("/tags", func(r chi.Router) {
				r.Get("/", getNodesTags)
				r.Route("/{id}", func(r chi.Router) {
					r.Use(nodeTagCtx)
					r.Get("/", getNodeTag)
				})
			})
		})
		r.Route("/services", func(r chi.Router) {
			r.Get("/", getServices)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(serviceCtx)
				r.Get("/", getService)
				r.Get("/candidate_tags", getServiceCandidateTags)
				r.Get("/tags", getServiceTags)
			})
			r.Route("/tags", func(r chi.Router) {
				r.Get("/", getServicesTags)
				r.Route("/{id}", func(r chi.Router) {
					r.Use(serviceTagCtx)
					r.Get("/", getServiceTag)
				})
			})
		})
		r.Route("/tags", func(r chi.Router) {
			r.Get("/", getTags)
			r.Post("/", postTags)
			r.Delete("/", delTags)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(tagCtx)
				r.Get("/", getTag)
				r.Delete("/", delTag)
				r.Route("/nodes", func(r chi.Router) {
					r.Get("/", getTagNodes)
				})
				r.Route("/services", func(r chi.Router) {
					r.Get("/", getTagServices)
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

func jsonEncode(w io.Writer, data interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(data)
}
