package routes

import (
	"fmt"
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

//
// GetUserGroups     godoc
// @Summary      List groups the user is a member of
// @Description  Managers and UserManager are allowed to see all users' information.
// @Description  Others can only see information for users in their organization groups.
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200      {object}  db.TableResponse
// @Failure      500    {string}  string  "Internal Server Error"
// @Param        props    query     string    false  "properties to include, and optionally remap (comma separated)"
// @Param        groupby  query     string    false  "properties to group by (comma separated)"
// @Param        order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param        filters  query     []string  false  "property value filter (a, !a, a&b, a|b, (a,b),  a%,  a%&!ab%)"
// @Param        limit    query     int       false  "number of objets to include in response"
// @Param        offset   query     int       false  "offset of the first objet to include in response"
// @Param        meta     query     bool      false  "turn off metadata in response"
// @Router       /users/{id}/groups  [get]
//
func GetUserGroups(w http.ResponseWriter, r *http.Request) {
	users := tables.UserFromCtx(r)
	if len(users) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	u := users[0]
	rq := db.Tab("auth_group").Request(
		db.TableRequestWithACL(false),
	)
	rq.AutoJoin("auth_membership")
	rq.Where("auth_membership.user_id = ?", u.ID)
	td, err := rq.MakeTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
}

//
// GetUserAppsPublication     godoc
// @Summary      List apps the user can read
// @Description  Managers and UserManager are allowed to see all users' information.
// @Description  Others can only see information for users in their organization groups.
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200      {object}  db.TableResponse
// @Failure      500    {string}  string  "Internal Server Error"
// @Param        props    query     string    false  "properties to include, and optionally remap (comma separated)"
// @Param        groupby  query     string    false  "properties to group by (comma separated)"
// @Param        order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param        filters  query     []string  false  "property value filter (a, !a, a&b, a|b, (a,b),  a%,  a%&!ab%)"
// @Param        limit    query     int       false  "number of objets to include in response"
// @Param        offset   query     int       false  "offset of the first objet to include in response"
// @Param        meta     query     bool      false  "turn off metadata in response"
// @Router       /users/{id}/apps/publication  [get]
//
func GetUserAppsPublication(w http.ResponseWriter, r *http.Request) {
	users := tables.UserFromCtx(r)
	if len(users) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	u := users[0]
	rq := db.Tab("apps").Request(
		db.TableRequestWithACL(false),
	)
	if !u.IsManager() {
		rq.AutoJoin("apps_publications")
		rq.TX(r).Joins("JOIN auth_membership ON apps_publications.group_id = auth_membership.group_id")
		rq.Where("auth_membership.user_id = ?", u.ID)
	}
	td, err := rq.MakeTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
}

//
// GetUserAppsResponsible     godoc
// @Summary      List apps the user is a responsible of
// @Description  Managers and UserManager are allowed to see all users' information.
// @Description  Others can only see information for users in their organization groups.
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200      {object}  db.TableResponse
// @Failure      500    {string}  string  "Internal Server Error"
// @Param        props    query     string    false  "properties to include, and optionally remap (comma separated)"
// @Param        groupby  query     string    false  "properties to group by (comma separated)"
// @Param        order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param        filters  query     []string  false  "property value filter (a, !a, a&b, a|b, (a,b),  a%,  a%&!ab%)"
// @Param        limit    query     int       false  "number of objets to include in response"
// @Param        offset   query     int       false  "offset of the first objet to include in response"
// @Param        meta     query     bool      false  "turn off metadata in response"
// @Router       /users/{id}/apps/responsible  [get]
//
func GetUserAppsResponsible(w http.ResponseWriter, r *http.Request) {
	users := tables.UserFromCtx(r)
	if len(users) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	u := users[0]

	rq := db.Tab("apps").Request(
		db.TableRequestWithACL(false),
	)
	if !u.IsManager() {
		rq.AutoJoin("apps_responsibles")
		rq.TX(r).Joins("JOIN auth_membership ON apps_responsibles.group_id = auth_membership.group_id")
		rq.Where("auth_membership.user_id = ?", u.ID)
	}
	td, err := rq.MakeTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
}
