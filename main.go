package main

import (
	"crypto/rsa"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

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
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.Timeout(60 * time.Second))

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
			r.With(pager).Get("/", getNodes)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(nodeCtx)
				r.Get("/", getNode)
				r.With(pager).Get("/candidate_tags", getNodeCandidateTags)
				r.With(pager).Get("/tags", getNodeTags)
			})
			r.Route("/tags", func(r chi.Router) {
				r.With(pager).Get("/", getNodesTags)
				r.Route("/{id}", func(r chi.Router) {
					r.Use(nodeTagCtx)
					r.Get("/", getNodeTag)
				})
			})
		})
		r.Route("/services", func(r chi.Router) {
			r.With(pager).Get("/", getServices)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(serviceCtx)
				r.Get("/", getService)
				//r.With(pager).Get("/candidate_tags", getServiceCandidateTags)
				//r.With(pager).Get("/tags", getServiceTags)
			})
			/*
				r.Route("/tags", func(r chi.Router) {
					r.With(pager).Get("/", getServicesTags)
					r.Route("/{id}", func(r chi.Router) {
						r.Use(serviceTagCtx)
						r.Get("/", getServiceTag)
					})
				})
			*/
		})
		r.Route("/tags", func(r chi.Router) {
			r.With(pager).Get("/", getTags)
			r.Post("/", postTags)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(tagCtx)
				r.Get("/", getTag)
				r.Delete("/", delTag)
				r.Route("/nodes", func(r chi.Router) {
					r.With(pager).Get("/", getTagNodes)
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
