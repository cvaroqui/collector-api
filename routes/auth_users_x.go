package routes

import (
	"net/http"

	"github.com/opensvc/collector-api/authuser"
	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
	"github.com/shaj13/go-guardian/v2/auth"
)

//
// DelUser     godoc
// @Summary      Delete a user
// @Description  Delete a user by index, email or login
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {array}   tables.User
// @Success      204  {string}  string  "No Content"
// @Failure      403  {string}  string  "Forbidden"
// @Failure      500  {string}  string  "Internal Server Error"
// @Param        id   path      string  true  "the index of the entry in database, or email, or login name"
// @Router       /users/{id}  [delete]
//
func DelUser(w http.ResponseWriter, r *http.Request) {
	user := auth.User(r)
	if authuser.HasPrivilege(user, "UserManager") {
		http.Error(w, http.StatusText(403), 403)
		return
	}
	users := tables.UserFromCtx(r)
	if len(users) == 0 {
		http.Error(w, http.StatusText(204), 204)
		return
	}
	usr := users[0]
	db.DB().Table("auth_user").Where("auth_user.id = ?", usr.ID).Delete(&tables.User{})
	jsonEncode(w, []tables.User{usr})
}

//
// GetUser     godoc
// @Summary      Show a user
// @Description  Show a user by index, email or login
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {array}   tables.User
// @Failure      404  {string}  string  "Not Found"
// @Failure      500  {string}  string  "Internal Server Error"
// @Param        id   path      string  true  "the index of the entry in database, or uuid, or name"
// @Router       /users/{id}  [get]
//
func GetUser(w http.ResponseWriter, r *http.Request) {
	users := tables.UserFromCtx(r)
	if len(users) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	jsonEncode(w, users)
}
