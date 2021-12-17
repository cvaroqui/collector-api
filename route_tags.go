package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/opensvc/collector-api/apiuser"
	"github.com/shaj13/go-guardian/v2/auth"
	"gorm.io/gorm/clause"
)

//
// GetTags     godoc
// @Summary      List tags
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         tags
// @Accept       json
// @Produce      json
// @Success      200      {object}  TableResponse
// @Failure      500      {string}  string  "Internal Server Error"
// @Param        props    query     string    false  "properties to include, and optionally remap (comma separated)"
// @Param        groupby  query     string    false  "properties to group by (comma separated)"
// @Param        order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param        filters  query     []string  false  "property value filter (a, !a, a&b, a|b, (a,b),  a%,  a%&!ab%)"
// @Param        limit    query     int       false  "number of objets to include in response"
// @Param        offset   query     int       false  "offset of the first objet to include in response"
// @Param        meta     query     bool      false  "turn off metadata in response"
// @Router       /tags  [get]
//
func getTags(w http.ResponseWriter, r *http.Request) {
	rq := tables["tags"].Request(TableRequestWithACL(false))
	td, err := rq.MakeReadTableResponse(r)
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
// PostTags    godoc
// @Summary      Create or update tags
// @Description  Creating tags requires no privilege.
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        tags  body      []Tag  true  "list of tags to create or update"
// @Success      200    {array}   Tag
// @Failure      500      {string}  string    "Internal Server Error"
// @Router       /tags  [post]
//
func postTags(w http.ResponseWriter, r *http.Request) {
	tags := make([]Tag, 0)
	tag := Tag{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("read request body: %s", err), 500)
		return
	}
	if err := json.Unmarshal(body, &tag); err == nil {
		// single entry
		tags = append(tags, tag)
	} else if err := json.Unmarshal(body, &tags); err != nil {
		// list of entry
		http.Error(w, fmt.Sprintf("unmarshal json: %s", err), 500)
	}
	tx := db.Clauses(clause.OnConflict{UpdateAll: true})
	if err := tx.Create(&tags).Error; err != nil {
		http.Error(w, fmt.Sprintf("insert or update: %s", err), 500)
		return
	}
	if err := jsonEncode(w, tags); err != nil {
		http.Error(w, fmt.Sprintf("json encode: %s", err), 500)
		return
	}
}

//
// DelTags     godoc
// @Summary      Delete tags
// @Description  Deleting tags requires the TagManager privilege.
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         tags
// @Accept       json
// @Produce      json
// @Success      200      {object}  []Tag
// @Failure      403      {string}  string  "Forbidden"
// @Failure      500      {string}  string  "Internal Server Error"
// @Param        filters  query     []string  false  "property value filter (a, !a, a&b, a|b, (a,b),  a%,  a%&!ab%)"
// @Param        order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param        limit    query     int       false  "number of objets to include in response"
// @Param        offset   query     int       false  "offset of the first objet to include in response"
// @Router       /tags  [delete]
//
func delTags(w http.ResponseWriter, r *http.Request) {
	user := auth.User(r)
	if !apiuser.HasPrivilege(user, "TagManager") {
		apiuser.PrivError(w, "TageManager")
		return
	}
	tags := make([]Tag, 0)
	tag := Tag{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("read request body: %s", err), 500)
		return
	}
	if err := json.Unmarshal(body, &tag); err == nil {
		// single entry
		tags = append(tags, tag)
	} else if err := json.Unmarshal(body, &tags); err != nil {
		// list of entry
	}
	rq := tables["tags"].Request(
		TableRequestWithACL(false),
		TableRequestWithWriteIntent(true),
	)
	tx := rq.TX(r)
	if len(tags) == 0 && !rq.HasValidFilters() {
		http.Error(w, "a valid json body or a valid filter is required, to prevent deleting all entries", 403)
		return
	}
	if err := tx.Find(&tags).Error; err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	if len(tags) == 0 {
		http.Error(w, http.StatusText(204), 204)
		return
	}
	if err := db.Delete(&tags).Error; err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	if err := jsonEncode(w, tags); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
}
