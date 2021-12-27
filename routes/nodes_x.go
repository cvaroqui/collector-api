package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/opensvc/collector-api/authuser"
	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
	"github.com/opensvc/collector-api/xmap"
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
// @Success      200  {array}   tables.Node
// @Failure      404  {string}  string  "Not Found"
// @Failure      500  {string}  string  "Internal Server Error"
// @Param        id   path      string  true  "the index of the entry in database, or uuid, or name"
// @Router       /nodes/{id}  [get]
//
func GetNode(w http.ResponseWriter, r *http.Request) {
	nodes := tables.NodeFromCtx(r)
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
// @Success      200  {array}   tables.Node
// @Success      204  {array}   string  "No Content"
// @Failure      500  {string}  string  "Internal Server Error"
// @Param        id   path      string  true  "the index of the entry in database, or uuid, or name"
// @Router       /nodes/{id}  [delete]
//
func DelNode(w http.ResponseWriter, r *http.Request) {
	user := auth.User(r)
	if !authuser.HasPrivilege(user, "NodeManager") {
		authuser.PrivError(w, "NodeManager")
		return
	}
	nodes := tables.NodeFromCtx(r)
	if len(nodes) == 0 {
		http.Error(w, http.StatusText(202), 202)
		return
	}
	if err := db.DB().Delete(&nodes).Error; err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	jsonEncode(w, nodes)
}

//
// PostNode	godoc
// @Summary      Update a node
// @Description  The user must be in the NodeManager privilege group.
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         nodes
// @Accept       json
// @Produce      json
// @Param        id     path      string       true  "the index of the entry in database, or uuid, or name"
// @Param        nodes  body      tables.Node  true  "node properties to create or update"
// @Success      200    {array}   tables.Node
// @Failure      401    {string}  string  "missing NodeManager privilege"
// @Failure      404    {string}  string  "the entry to update does not exist"
// @Failure      500    {string}  string
// @Router       /nodes/{id}  [post]
//
func PostNode(w http.ResponseWriter, r *http.Request) {
	user := auth.User(r)
	if !authuser.HasPrivilege(user, "NodeManager") {
		authuser.PrivError(w, "NodeManager")
		return
	}
	data := make(map[string]interface{})
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("read request body: %s", err), 500)
		return
	}
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, fmt.Sprintf("unmarshal json: %s", err), 500)
		return
	}
	if _, ok := data["id"]; ok {
		delete(data, "id")
	}
	currents := tables.NodeFromCtx(r)
	if len(currents) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	current := currents[0]
	var i int64
	rq := db.Tab("nodes").Request(db.TableRequestWithWriteIntent(true))
	if err := rq.TX(r).Where("nodes.id = ?", current.ID).Count(&i).Error; err != nil {
		http.Error(w, fmt.Sprintf("select from write: %s", err), 500)
		return
	}
	if i == 0 {
		http.Error(w, fmt.Sprintf("user is not responsible for node %s in app %s", current.Nodename, current.App), 500)
		return
	}
	props := xmap.Keys(data)
	if err := db.DB().Table("nodes").Select(props).Where("id = ?", current.ID).Updates(data).Error; err != nil {
		http.Error(w, fmt.Sprintf("update: %s", err), 500)
		return
	}
	if err := db.DB().Take(&current).Error; err != nil {
		http.Error(w, fmt.Sprintf("select after update: %s", err), 500)
		return
	}
	if err := jsonEncode(w, []tables.Node{current}); err != nil {
		http.Error(w, fmt.Sprintf("json encode: %s", err), 500)
		return
	}
	// enqueue dashboard alerts refresh
}
