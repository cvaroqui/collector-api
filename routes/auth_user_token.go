package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/opensvc/collector-api/auth"
	"github.com/opensvc/collector-api/authuser"
)

//
// GetNodes     godoc
// @Summary      Get a user authentication token
// @Description  Get an authentication token from a user's credentials submitted with basic login.
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         auth
// @Produce      json
// @Success      200  {object}  TokenResponse
// @Failure      403  {string}  string
// @Failure      500  {string}  string  "Internal Server Error"
// @Router       /auth/user/token  [get]
//
func GetUserToken(w http.ResponseWriter, r *http.Request) {
	exp := time.Minute * 10
	user := auth.User(r)
	tokenExpireAt := time.Now().Add(exp)
	privileges := strings.Join(authuser.Privileges(user), ",")
	claims := map[string]interface{}{
		"exp":        tokenExpireAt.Unix(),
		"authorized": true,
		"user_id":    user.GetID(),
		"privileges": privileges,
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
