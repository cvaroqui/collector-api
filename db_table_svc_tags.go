package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ServiceTag struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	SvcID         string         `gorm:"column:svc_id; size:36" json:"svc_id"`
	TagID         string         `gorm:"column:tag_id; size:40; index" json:"tag_id"`
	TagAttachData datatypes.JSON `gorm:"column:tag_attach_data; type:text" json:"tag_attach_data"`
	Created       time.Time      `gorm:"column:created; autoCreateTime" json:"created"`
}

func init() {
	tables["svc_tags"] = newTable("svc_tags").SetEntry(ServiceTag{})
}

func serviceTagCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		n, err := getServiceTagByID(id)
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "serviceTag", n)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getServiceTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, ok := ctx.Value("serviceTag").(ServiceTag)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	jsonEncode(w, []ServiceTag{n})
}

func getServiceTagByID(id string) (ServiceTag, error) {
	data := make([]ServiceTag, 0)
	result := db.Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return ServiceTag{}, result.Error
	}
	if len(data) == 0 {
		return ServiceTag{}, fmt.Errorf("not found")
	}
	return data[0], nil
}

func getServicesTags(w http.ResponseWriter, r *http.Request) {
	rq := tables["svc_tags"].Request()
	td, err := rq.MakeResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func getServiceTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, _ := ctx.Value("service").(Service)
	rq := tables["tags"].Request()
	rq.AutoJoin("svc_tags")
	rq.Where("svc_tags.svc_id = ?", n.SvcID)
	td, err := rq.MakeResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func getServiceCandidateTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, _ := ctx.Value("service").(Service)
	exclude := db.Table("svc_tags").Where("svc_id = ?", n.SvcID).Select("tag_id")

	rq := tables["tags"].Request()
	rq.Where("tags.tag_id NOT IN (?)", exclude)
	td, err := rq.MakeResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func getTagServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tag, _ := ctx.Value("tag").(Tag)
	rq := tables["services"].Request()
	rq.AutoJoin("tags")
	rq.Where("svc_tags.tag_id = ?", tag.TagID)
	td, err := rq.MakeResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}
