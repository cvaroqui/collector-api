package routes

import (
	"fmt"
	"net/http"

	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
)

//
// GetServiceTag     godoc
// @Summary   Show a tag attachment to a service
// @Security  BasicAuth
// @Security  BearerAuth
// @Tags      tags
// @Tags      services
// @Accept    json
// @Produce   json
// @Success   200  {array}   tables.ServiceTag
// @Failure   404  {string}  string  "Not Found"
// @Failure   500  {string}  string  "Internal Server Error"
// @Param     id   path      string  true  "the index of the entry in database, or uuid, or name"
// @Router    /services/{id}/tags/{id}  [get]
// @Router    /services/tags/{id}  [get]
//
func GetServiceTag(w http.ResponseWriter, r *http.Request) {
	data := make([]tables.ServiceTag, 0)
	tx := db.Tab("svc_tags").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	attachs := tables.ServiceTagFromCtx(r)
	if len(attachs) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	if err := tx.Where("svc_tags.id = ?", attachs[0].ID).Find(&data).Error; err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	jsonEncode(w, data)
}

//
// GetServicesTags     godoc
// @Summary   List tags attachments to services
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
// @Router    /services/tags  [get]
//
func GetServicesTags(w http.ResponseWriter, r *http.Request) {
	rq := db.Tab("svc_tags").Request()
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
// GetServiceTags     godoc
// @Summary   List tags attached to a service
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
// @Router    /services/{id}/tags  [get]
//
func GetServiceTags(w http.ResponseWriter, r *http.Request) {
	cs := tables.ServiceFromCtx(r)
	if len(cs) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	n := cs[0]
	rq := db.Tab("tags").Request()
	rq.AutoJoin("svc_tags")
	rq.Where("svc_tags.svc_id = ?", n.SvcID)
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
// GetServiceCandidateTags     godoc
// @Summary   List existing tags not already attached to a service
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
// @Router    /services/{id}/candidate_tags  [get]
//
func GetServiceCandidateTags(w http.ResponseWriter, r *http.Request) {
	cs := tables.ServiceFromCtx(r)
	if len(cs) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	n := cs[0]
	exclude := db.DB().Table("svc_tags").Where("svc_id = ?", n.SvcID).Select("tag_id")

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
// GetTagServices     godoc
// @Summary   List services having a specific tag
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
// @Router    /tags/{id}/services  [get]
//
func GetTagServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tag, _ := ctx.Value("tag").(tables.Tag)
	rq := db.Tab("services").Request()
	rq.AutoJoin("tags")
	rq.Where("svc_tags.tag_id = ?", tag.TagID)
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
