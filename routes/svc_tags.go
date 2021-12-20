package routes

import (
	"fmt"
	"net/http"

	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
)

func GetServiceTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, ok := ctx.Value("serviceTag").(tables.ServiceTag)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	jsonEncode(w, []tables.ServiceTag{n})
}

func GetServicesTags(w http.ResponseWriter, r *http.Request) {
	rq := db.Tab("svc_tags").Request()
	td, err := rq.MakeTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func GetServiceTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, _ := ctx.Value("service").(tables.Service)
	rq := db.Tab("tags").Request()
	rq.AutoJoin("svc_tags")
	rq.Where("svc_tags.svc_id = ?", n.SvcID)
	td, err := rq.MakeTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func GetServiceCandidateTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, _ := ctx.Value("service").(tables.Service)
	exclude := db.DB().Table("svc_tags").Where("svc_id = ?", n.SvcID).Select("tag_id")

	rq := db.Tab("tags").Request()
	rq.Where("tags.tag_id NOT IN (?)", exclude)
	td, err := rq.MakeTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func GetTagServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tag, _ := ctx.Value("tag").(tables.Tag)
	rq := db.Tab("services").Request()
	rq.AutoJoin("tags")
	rq.Where("svc_tags.tag_id = ?", tag.TagID)
	td, err := rq.MakeTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}
