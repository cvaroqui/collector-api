package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
)

var (
	addr             string = ":8080"
	tokenAuth        *jwtauth.JWTAuth
	verifyBytes      []byte
	verifyKey        *rsa.PublicKey
	signKey          *rsa.PrivateKey
	jwtSignKeyPath   string
	jwtVerifyKeyPath string
)

func fatal(err interface{}) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if err := initJWT(); err != nil {
		fatal(err)
	}
	if err := initDB(); err != nil {
		fatal(err)
	}
	if listen := os.Getenv("LISTEN"); listen != "" {
		addr = listen
	}
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
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator)

		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["user_id"])))
		})

		r.Route("/nodes", func(r chi.Router) {
			r.With(pager).Get("/", getNodes)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(nodeCtx)
				r.Get("/", getNode)
			})
			r.Route("/tags", func(r chi.Router) {
				r.With(pager).Get("/", getNodeTags)
				r.Route("/{id}", func(r chi.Router) {
					r.Use(nodeTagCtx)
					r.Get("/", getNodeTag)
				})
			})
		})
		r.Route("/tags", func(r chi.Router) {
			r.With(pager).Get("/", getTags)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(tagCtx)
				r.Get("/", getTag)
			})
		})
	})

	// NodeAuth routes
	r.Group(func(r chi.Router) {
		r.Use(nodeAuth(time.Minute * 30))

		r.Route("/nodes/login", func(r chi.Router) {
			r.Get("/", getNodeToken)
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
