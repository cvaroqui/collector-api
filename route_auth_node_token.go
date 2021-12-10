package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shaj13/go-guardian/v2/auth"
)

func getNodeToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	exp := time.Minute * 10
	user, _ := ctx.Value("user").(auth.Info)
	n, _ := getNodeByNodeID(user.GetID())
	expireAt := time.Now().Add(exp)
	claims := map[string]interface{}{
		"exp":        expireAt.Unix(),
		"authorized": true,
		"nodename":   n.Nodename,
		"node_id":    n.NodeID,
		"app":        n.App,
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
