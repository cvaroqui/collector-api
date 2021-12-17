package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/shaj13/go-guardian/v2/auth"
)

func getNodeToken(w http.ResponseWriter, r *http.Request) {
	exp := time.Minute * 10
	user := auth.User(r)
	nodes, err := getNodeByNodeID(user.GetID())
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	if len(nodes) == 0 {
		http.Error(w, fmt.Sprintf("%s (unknown node)", http.StatusText(403)), 403)
		return
	}
	n := nodes[0]
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
