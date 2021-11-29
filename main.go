package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
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

func init() {
}

func initJWT() error {
	jwtSignKeyPath = os.Getenv("JWT_SIGN_KEY")
	jwtVerifyKeyPath = os.Getenv("JWT_VERIFY_KEY")

	if jwtSignKeyPath == "" && jwtVerifyKeyPath == "" {
		return fmt.Errorf("JWT_SIGN_KEY must be to either a secret string, or the path of a RSA key. In the later case, JWT_VERIFY_KEY must also be set to the path of the RSA public key.")
	} else if strings.HasPrefix(jwtSignKeyPath, "/") {
		signBytes, err := ioutil.ReadFile(jwtSignKeyPath)
		if err != nil {
			return err
		}
		if signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes); err != nil {
			return err
		}
		if jwtVerifyKeyPath == "" {
			return fmt.Errorf("JWT_SIGN_KEY is set to the path of a RSA key. In this case, JWT_VERIFY_KEY must also be set to the path of the RSA public key.")
		}
		if verifyBytes, err = ioutil.ReadFile(jwtVerifyKeyPath); err != nil {
			return err
		} else {
			log.Printf("Verify key:\n%s", string(verifyBytes))
		}

		if verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes); err != nil {
			return err
		}

		tokenAuth = jwtauth.New("RS256", signKey, verifyKey)
	} else {
		log.Printf("Using JWT HMAC signature. This is less secure than RS256 signature and verification. Set both JWT_SIGN_KEY and JWT_VERIFY_KEY to paths of a RSA key-pair.")
		tokenAuth = jwtauth.New("HMAC", signKey, nil)
	}
	return nil
}

func main() {
	if err := initJWT(); err != nil {
		fatal(err)
	}
	if listen := os.Getenv("LISTEN"); listen != "" {
		addr = listen
	}
	log.Printf("Starting server on %v\n", addr)
	http.ListenAndServe(addr, router())
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
			r.Route("/{nodeID}", func(r chi.Router) {
				r.Use(nodeCtx)
				r.Get("/", getNode)
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

func jsonEncode(w io.Writer, data interface{}) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(data)
}
