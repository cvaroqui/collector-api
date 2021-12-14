package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/opensvc/collector-api/apiuser"
	"github.com/shaj13/go-guardian/v2/auth"
)

func getUserToken(w http.ResponseWriter, r *http.Request) {
	exp := time.Minute * 10
	user := auth.User(r)
	expireAt := time.Now().Add(exp)
	privileges := strings.Join(apiuser.Privileges(user), ",")
	claims := map[string]interface{}{
		"exp":        expireAt.Unix(),
		"authorized": true,
		"user_id":    user.GetID(),
		"privileges": privileges,
	}
	_, token, err := tokenAuth.Encode(claims)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	auth.Append(tokenStrategy, token, user)

	jsonEncode(w, map[string]interface{}{
		"expire_at": expireAt,
		"token":     token,
	})
}
