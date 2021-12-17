package main

import (
	"fmt"
	"net/http"

	"github.com/opensvc/collector-api/apiuser"
	"github.com/shaj13/go-guardian/v2/auth"
)

//
// GetNode     godoc
// @Summary      Show a node
// @Description  Show a node by index, id or name
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         nodes
// @Accept       json
// @Produce      json
// @Success      200  {array}   Node
// @Failure      404  {string}  string  "Not Found"
// @Failure      500  {string}  string  "Internal Server Error"
// @Param        id   path      string  true  "the index of the entry in database, or uuid, or name"
// @Router       /nodes/{id}  [get]
//
func getNode(w http.ResponseWriter, r *http.Request) {
	nodes := nodeFromCtx(r)
	if len(nodes) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	jsonEncode(w, nodes)
}

//
// DelNode     godoc
// @Summary      Delete a node
// @Description  Delete a node by index, id or name.
// @Description  The user must have the NodeManager privilege.
// @Description  The user must be responsible for the node, via app responsibles.
// @Description  Cascade delete on services instances, dashboard entries, checks, packages and patches.
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         nodes
// @Accept       json
// @Produce      json
// @Success      200  {array}   Node
// @Success      204  {array}   string  "No Content"
// @Failure      500  {string}  string  "Internal Server Error"
// @Param        id   path      string  true  "the index of the entry in database, or uuid, or name"
// @Router       /nodes/{id}  [delete]
//
func delNode(w http.ResponseWriter, r *http.Request) {
	user := auth.User(r)
	if !apiuser.HasPrivilege(user, "NodeManager") {
		apiuser.PrivError(w, "NodeManager")
		return
	}
	nodes := nodeFromCtx(r)
	if len(nodes) == 0 {
		http.Error(w, http.StatusText(202), 202)
		return
	}
	if err := db.Delete(&nodes).Error; err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	jsonEncode(w, nodes)
}
