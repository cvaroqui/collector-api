package routes

import (
	"fmt"
	"net/http"

	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
)

//
// GetNodeTag     godoc
// @Summary   Show a tag attachment to a node
// @Security  BasicAuth
// @Security  BearerAuth
// @Tags      tags
// @Tags      nodes
// @Accept    json
// @Produce   json
// @Success   200  {array}   tables.NodeTag
// @Failure   404  {string}  string  "Not Found"
// @Failure   500  {string}  string  "Internal Server Error"
// @Param     id   path      string  true  "the index of the entry in database, or uuid, or name"
// @Router    /nodes/{id}/tags/{id}  [get]
// @Router    /nodes/tags/{id}  [get]
//
func GetNodeTag(w http.ResponseWriter, r *http.Request) {
	data := make([]tables.NodeTag, 0)
	tx := db.Tab("node_tags").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)

	attachs := tables.NodeTagFromCtx(r)
	if len(attachs) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	if err := tx.Where("node_tags.id = ?", attachs[0].ID).Find(&data).Error; err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	jsonEncode(w, data)
}

//
// GetNodesTags     godoc
// @Summary   List tags attachments to nodes
// @Security  BasicAuth
// @Security  BearerAuth
// @Tags      tags
// @Accept    json
// @Produce   json
// @Success   200      {object}  db.TableResponse
// @Failure   500      {string}  string    "Internal Server Error"
// @Param     props    query     string    false  "properties to include, and optionally remap (comma separated)"
// @Param     groupby  query     string    false  "properties to group by (comma separated)"
// @Param     order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param     filters  query     []string  false  "property value filter (a, !a, a&b, a|b, (a,b),  a%,  a%&!ab%)"
// @Param     limit    query     int       false  "number of objets to include in response"
// @Param     offset   query     int       false  "offset of the first objet to include in response"
// @Param     meta     query     bool      false  "turn off metadata in response"
// @Router    /nodes/tags  [get]
//
func GetNodesTags(w http.ResponseWriter, r *http.Request) {
	rq := db.Tab("node_tags").Request()
	rq.AutoJoin("nodes")
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
// GetNodeTags     godoc
// @Summary   List tags attached to a node
// @Security  BasicAuth
// @Security  BearerAuth
// @Tags      tags
// @Accept    json
// @Produce   json
// @Success   200      {object}  db.TableResponse
// @Failure   500      {string}  string    "Internal Server Error"
// @Param     props    query     string    false  "properties to include, and optionally remap (comma separated)"
// @Param     groupby  query     string    false  "properties to group by (comma separated)"
// @Param     order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param     filters  query     []string  false  "property value filter (a, !a, a&b, a|b, (a,b),  a%,  a%&!ab%)"
// @Param     limit    query     int       false  "number of objets to include in response"
// @Param     offset   query     int       false  "offset of the first objet to include in response"
// @Param     meta     query     bool      false  "turn off metadata in response"
// @Router    /nodes/{id}/tags  [get]
//
func GetNodeTags(w http.ResponseWriter, r *http.Request) {
	data := tables.NodeFromCtx(r)
	if len(data) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	n := data[0]
	rq := db.Tab("node_tags").Request()
	rq.AutoJoin("nodes")
	rq.Where("nodes.id = ?", n.ID)
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
// GetNodeCandidateTags     godoc
// @Summary   List existing tags not already attached to a node
// @Security  BasicAuth
// @Security  BearerAuth
// @Tags      tags
// @Accept    json
// @Produce   json
// @Success   200      {object}  db.TableResponse
// @Failure   500      {string}  string    "Internal Server Error"
// @Param     props    query     string    false  "properties to include, and optionally remap (comma separated)"
// @Param     groupby  query     string    false  "properties to group by (comma separated)"
// @Param     order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param     filters  query     []string  false  "property value filter (a, !a, a&b, a|b, (a,b),  a%,  a%&!ab%)"
// @Param     limit    query     int       false  "number of objets to include in response"
// @Param     offset   query     int       false  "offset of the first objet to include in response"
// @Param     meta     query     bool      false  "turn off metadata in response"
// @Router    /nodes/{id}/candidate_tags  [get]
//
func GetNodeCandidateTags(w http.ResponseWriter, r *http.Request) {
	data := tables.NodeFromCtx(r)
	if len(data) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	n := data[0]
	exclude := db.DB().Table("node_tags").Where("node_id = ?", n.NodeID).Select("tag_id")

	rq := db.Tab("tags").Request()
	rq.Where("tags.tag_id NOT IN (?)", exclude)
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
// GetTagNodes     godoc
// @Summary   List nodes having a specific tag
// @Security  BasicAuth
// @Security  BearerAuth
// @Tags      tags
// @Accept    json
// @Produce   json
// @Success   200      {object}  db.TableResponse
// @Failure   500      {string}  string    "Internal Server Error"
// @Param     props    query     string    false  "properties to include, and optionally remap (comma separated)"
// @Param     groupby  query     string    false  "properties to group by (comma separated)"
// @Param     order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param     filters  query     []string  false  "property value filter (a, !a, a&b, a|b, (a,b),  a%,  a%&!ab%)"
// @Param     limit    query     int       false  "number of objets to include in response"
// @Param     offset   query     int       false  "offset of the first objet to include in response"
// @Param     meta     query     bool      false  "turn off metadata in response"
// @Router    /tags/{id}/nodes  [get]
//
func GetTagNodes(w http.ResponseWriter, r *http.Request) {
	data := tables.TagFromCtx(r)
	if len(data) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	tag := data[0]
	rq := db.Tab("nodes").Request()
	rq.AutoJoin("tags")
	rq.Where("node_tags.tag_id = ?", tag.TagID)
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
