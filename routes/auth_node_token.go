package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/opensvc/collector-api/auth"
	"github.com/opensvc/collector-api/db/tables"
)

type TokenResponse struct {
	Token         string    `json:"token"`
	TokenExpireAt time.Time `json:"token_expire_at"`
}

//
// GetNodes     godoc
// @Summary      Get a node authentication token
// @Description  Get an authentication token from a node's credentials submitted with basic login.
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         auth
// @Produce      json
// @Success      200  {object}  TokenResponse
// @Failure      403  {string}  string
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /auth/node/token  [get]
//
func GetNodeToken(w http.ResponseWriter, r *http.Request) {
	exp := time.Minute * 10
	user := auth.User(r)
	nodes, err := tables.GetNodeByNodeID(user.GetID())
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	if len(nodes) == 0 {
		http.Error(w, fmt.Sprintf("%s (unknown node)", http.StatusText(403)), 403)
		return
	}
	n := nodes[0]
	tokenExpireAt := time.Now().Add(exp)
	claims := map[string]interface{}{
		"exp":        tokenExpireAt.Unix(),
		"authorized": true,
		"nodename":   n.Nodename,
		"node_id":    n.NodeID,
		"app":        n.App,
	}
	_, token, err := auth.TokenAuth.Encode(claims)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	auth.AppendToken(token, user)

	jsonEncode(w, TokenResponse{
		TokenExpireAt: tokenExpireAt,
		Token:         token,
	})
}
