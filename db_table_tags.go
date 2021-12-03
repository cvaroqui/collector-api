package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	tables["tags"] = newTable("tags").SetEntry(Tag{})
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

func delTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tag, ok := ctx.Value("tag").(Tag)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	// TODO: priv check
	db.Where("tags.id = ?", tag.ID).Delete(&Tag{})
	jsonEncode(w, []Tag{tag})
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
	req := tables["tags"].Request()
	td, err := req.MakeResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func postTags(w http.ResponseWriter, r *http.Request) {
	tags := make([]Tag, 0)
	tag := Tag{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := json.Unmarshal(body, &tag); err == nil {
		// single entry
		tags = append(tags, tag)
	} else if err := json.Unmarshal(body, &tags); err != nil {
		// list of entry
		http.Error(w, fmt.Sprint(err), 500)
	}
	for _, tag := range tags {
		tx := db.Clauses(clause.OnConflict{UpdateAll: true})
		if err := tx.Create(&tag).Error; err != nil {
			http.Error(w, fmt.Sprint(err), 500)
		}
	}
}
