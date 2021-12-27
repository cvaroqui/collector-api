package tables

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/opensvc/collector-api/db"
	"gorm.io/gorm"
)

type Tag struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	TagID      string         `gorm:"->;column:tag_id; size:40; index; type:GENERATED ALWAYS AS (sha(tag_name)) STORED" json:"tag_id"`
	TagName    string         `gorm:"column:tag_name; unique; size:128" json:"tag_name"`
	TagExclude string         `gorm:"column:tag_exclude; size:128" json:"tag_exclude"`
	TagData    string         `gorm:"column:tag_data; type:text" json:"tag_data"`
	TagCreated time.Time      `gorm:"column:tag_created; autoCreateTime" json:"tag_created"`
}

var (
	reTagID, _ = regexp.Compile("[0-9a-f]{40}")
)

func init() {
	db.Register(&db.Table{
		Name:  "tags",
		Entry: Tag{},
	})
}

func TagFromCtx(r *http.Request) []Tag {
	i := r.Context().Value("tag")
	if i == nil {
		return []Tag{}
	}
	return i.([]Tag)
}

func TagCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if reTagID.MatchString(id) {
			tags, err := getTagByTagID(id)
			if err != nil {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
			ctx := context.WithValue(r.Context(), "tag", tags)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			tags, err := getTagByID(id)
			if err != nil {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
			ctx := context.WithValue(r.Context(), "tag", tags)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func getTagByTagID(id string) ([]Tag, error) {
	data := make([]Tag, 0)
	result := db.DB().Where("tag_id = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func getTagByID(id string) ([]Tag, error) {
	data := make([]Tag, 0)
	result := db.DB().Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}
