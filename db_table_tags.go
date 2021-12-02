package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Tag struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	TagID      string         `gorm:"column:tag_id; size:40; index" json:"tag_id"`
	TagName    string         `gorm:"->;column:tag_name; unique; size:128; type:GENERATED ALWAYS AS (sha(tag_name)) STORED" json:"tag_name"`
	TagExclude string         `gorm:"column:tag_exclude; size:128" json:"tag_exclude"`
	TagData    string         `gorm:"column:tag_data; type:text" json:"tag_data"`
	TagCreated time.Time      `gorm:"column:tag_created; autoCreateTime" json:"tag_created"`
}

var (
	reTagID, _ = regexp.Compile("[0-9a-f]{40}")
)

func init() {
	tables["tags"] = newTable("tags").
		SetEntry(Tag{}).
		SetJoin("node_tags", "left join node_tags on tags.tag_id=node_tags.tag_id").
		SetJoin("nodes", "left join node_tags on tags.tag_id=node_tags.tag_id left join nodes on node_tags.node_id=nodes.node_id").
		SetJoin("svc_tags", "left join svc_tags on tags.tag_id=svc_tags.tag_id").
		SetJoin("svcmon", "left join svc_tags on tags.tag_id=svc_tags.tag_id left join svcmon on svc_tags.svc_id=svcmon.svc_id").
		SetJoin("services", "left join svc_tags on tags.tag_id=svc_tags.tag_id left join services on svc_tags.svc_id=services.svc_id")
}

func tagCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if reTagID.MatchString(id) {
			n, err := getTagByTagID(id)
			if err != nil {
				http.Error(w, http.StatusText(404), 404)
				return
			}
			ctx := context.WithValue(r.Context(), "tag", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			n, err := getTagByID(id)
			if err != nil {
				http.Error(w, http.StatusText(404), 404)
				return
			}
			ctx := context.WithValue(r.Context(), "tag", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func getTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, ok := ctx.Value("tag").(Tag)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	jsonEncode(w, []Tag{n})
}

func getTagByTagID(id string) (Tag, error) {
	data := make([]Tag, 0)
	result := db.Where("tag_id = ?", id).Find(&data)
	if result.Error != nil {
		return Tag{}, result.Error
	}
	if len(data) == 0 {
		return Tag{}, fmt.Errorf("not found")
	}
	return data[0], nil
}

func getTagByID(id string) (Tag, error) {
	data := make([]Tag, 0)
	result := db.Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return Tag{}, result.Error
	}
	if len(data) == 0 {
		return Tag{}, fmt.Errorf("not found")
	}
	return data[0], nil
}

func getTags(w http.ResponseWriter, r *http.Request) {
	//_, claims, _ := jwtauth.FromContext(r.Context())
	tx := tables["tags"].DBTable()
	data := make([]Tag, 0)
	td, err := tables["tags"].MakeResponse(r, tx, &data)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}
