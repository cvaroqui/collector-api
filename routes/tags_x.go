package routes

import (
	"net/http"

	"github.com/opensvc/collector-api/auth"
	"github.com/opensvc/collector-api/authuser"
	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
)

//
// DelTag     godoc
// @Summary      Delete a tag
// @Description  Delete a tag by index, uuid or name.
// @Description  Requires the TagManager privilege.
// @Description  Cascade deletes the tag attachements to nodes and services.
// @Security  BasicAuth
// @Security  BearerAuth
// @Tags      tags
// @Accept    json
// @Produce   json
// @Success   200  {array}   tables.Tag
// @Success      204  {string}  string  "No Content"
// @Failure      403  {string}  string  "Forbidden"
// @Failure   500  {string}  string  "Internal Server Error"
// @Param     id   path      string  true  "the index of the entry in database, or uuid, or name"
// @Router       /tags/{id}  [delete]
//
func DelTag(w http.ResponseWriter, r *http.Request) {
	user := auth.User(r)
	if !authuser.HasPrivilege(user, "TagManager") {
		authuser.PrivError(w, "TagManager")
		return
	}
	tags := tables.TagFromCtx(r)
	if len(tags) == 0 {
		http.Error(w, http.StatusText(204), 204)
		return
	}
	db.DB().Where("tags.id = ?", tags[0].ID).Delete(&[]tables.Tag{})
	jsonEncode(w, tags)
}

//
// GetTag     godoc
// @Summary   Show a tag
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         tags
// @Accept       json
// @Produce      json
// @Success      200  {array}   tables.Tag
// @Failure   404  {string}  string  "Not Found"
// @Failure      500  {string}  string  "Internal Server Error"
// @Param        id   path      string  true  "the index of the entry in database, or uuid, or name"
// @Router    /tags/{id}  [get]
//
func GetTag(w http.ResponseWriter, r *http.Request) {
	tags := tables.TagFromCtx(r)
	if len(tags) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	jsonEncode(w, tags)
}
