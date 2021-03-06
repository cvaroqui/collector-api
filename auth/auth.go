package auth

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/fifo"
	"github.com/spf13/viper"

	"github.com/shaj13/go-guardian/v2/auth"
	"github.com/shaj13/go-guardian/v2/auth/strategies/basic"
	"github.com/shaj13/go-guardian/v2/auth/strategies/ldap"
	"github.com/shaj13/go-guardian/v2/auth/strategies/token"
	"github.com/shaj13/go-guardian/v2/auth/strategies/union"

	"github.com/opensvc/collector-api/apiuser"
	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
	"github.com/opensvc/collector-api/w2pcrypt"
)

var (
	strategy      union.Union
	TokenStrategy auth.Strategy
	cache         libcache.Cache
	w2pCryptObj   *w2pcrypt.Crypt
)

func AppendToken(key interface{}, info auth.Info) error {
	return auth.Append(TokenStrategy, key, info)
}

func User(r *http.Request) auth.Info {
	return auth.User(r)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, user, err := strategy.AuthenticateRequest(r)
		if err != nil {
			log.Println(err)
			code := http.StatusUnauthorized
			http.Error(w, http.StatusText(code), code)
			return
		}
		log.Printf("User %s (%s) authenticated\n", user.GetUserName(), user.GetID())
		r = auth.RequestWithUser(user, r)
		next.ServeHTTP(w, r)
	})
}

func validateNode(ctx context.Context, r *http.Request, username, password string) (auth.Info, error) {
	data := make([]tables.Node, 0)
	passwordMD5B := md5.Sum([]byte(password))
	passwordMD5 := hex.EncodeToString(passwordMD5B[:])
	result := db.DB().Table("auth_node").Joins("JOIN nodes ON nodes.node_id = auth_node.node_id").Where("auth_node.nodename = ? and md5(auth_node.uuid) = ?", username, passwordMD5).Select("nodes.*").Find(&data)
	if result.Error != nil {
		return nil, fmt.Errorf("node auth: %s", result.Error)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("node auth: Invalid credentials")
	}
	extensions := apiuser.MakeNodeExtensions(data[0])
	return auth.NewDefaultUser(username, data[0].NodeID, nil, extensions), nil
}

func validateUser(ctx context.Context, r *http.Request, username, password string) (auth.Info, error) {
	var (
		users []tables.User
		err   error
	)
	if strings.Contains(username, "@") {
		users, err = tables.GetUserByEmail(username)
	} else {
		users, err = tables.GetUserByUsername(username)
	}
	if err != nil {
		return nil, err
	}
	switch len(users) {
	case 0:
		return nil, fmt.Errorf("user not found")
	case 1:
		// ok
	default:
		return nil, fmt.Errorf("too many users found")
	}
	user := users[0]
	if user.ID == 0 {
		return nil, fmt.Errorf("user id zero")
	}
	if ok, err := w2pCryptObj.IsEqual(password, user.Password); err != nil {
		return nil, fmt.Errorf("user auth: %s", err)
	} else if ok {
		extensions := apiuser.MakeUserExtensions(user)
		return auth.NewDefaultUser(username, fmt.Sprint(user.ID), nil, extensions), nil
	}
	return nil, fmt.Errorf("user auth: invalid credentials")
}

func initCache() error {
	cache = libcache.FIFO.New(0)
	cache.SetTTL(time.Minute * 5)
	cache.RegisterOnExpired(func(key, _ interface{}) {
		cache.Peek(key)
	})
	return nil
}

func initLDAP() []auth.Strategy {
	log.Println("init ldap auth strategy")
	strategies := make([]auth.Strategy, 0)
	data := viper.Sub("auth.ldap")
	if data == nil {
		return strategies
	}
	for k, _ := range data.AllSettings() {
		prefix := fmt.Sprintf("auth.ldap.%s.", k)
		host := viper.GetString(prefix + "host")
		port := viper.GetString(prefix + "port")
		baseDN := viper.GetString(prefix + "base_dn")
		log.Printf("  %s %s:%s %s", k, host, port, baseDN)
		tlsCfg := tls.Config{}
		cfg := &ldap.Config{
			BaseDN:       baseDN,
			BindDN:       viper.GetString(prefix + "bind_dn"),
			Port:         port,
			Host:         host,
			BindPassword: viper.GetString(prefix + "bind_password"),
			Filter:       viper.GetString(prefix + "filter"),
			TLS:          &tlsCfg,
		}
		ldapStrategy := ldap.NewCached(cfg, cache)
		strategies = append(strategies, ldapStrategy)
	}
	return strategies
}

func initBasicNode() []auth.Strategy {
	log.Println("init basic node auth strategy")
	basicNodeStrategy := basic.NewCached(validateNode, cache)
	return []auth.Strategy{basicNodeStrategy}
}

func initBasicUser() []auth.Strategy {
	log.Println("init basic user auth strategy")
	hmacAlg := viper.GetString("auth.web2py.hmac.alg")
	if hmacAlg != "" {
		log.Println("  using hmac algo:", hmacAlg)
	}
	hmacKey := viper.GetString("auth.web2py.hmac.key")
	if hmacKey != "" {
		log.Println("  using hmac key")
	}
	basicUserStrategy := basic.NewCached(validateUser, cache)
	w2pCryptObj = w2pcrypt.NewCrypt(hmacKey, hmacAlg)
	return []auth.Strategy{basicUserStrategy}
}

func initToken() []auth.Strategy {
	log.Println("init token auth strategy")
	TokenStrategy = token.New(token.NoOpAuthenticate, cache)
	return []auth.Strategy{TokenStrategy}
}

func Init() error {
	if err := initJWT(); err != nil {
		return err
	}
	if err := initCache(); err != nil {
		return err
	}
	strategies := make([]auth.Strategy, 0)
	strategies = append(strategies, initToken()...)
	strategies = append(strategies, initBasicNode()...)
	strategies = append(strategies, initBasicUser()...)
	strategies = append(strategies, initLDAP()...)
	strategy = union.New(strategies...)
	return nil
}
