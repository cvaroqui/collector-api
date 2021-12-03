package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type NodeTag struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	NodeID        string         `gorm:"column:node_id; size:36" json:"node_id"`
	TagID         string         `gorm:"column:tag_id; size:40; index" json:"tag_id"`
	TagAttachData datatypes.JSON `gorm:"column:tag_attach_data; type:text" json:"tag_attach_data"`
	Created       time.Time      `gorm:"column:created; autoCreateTime" json:"created"`
}

func init() {
	tables["node_tags"] = newTable("node_tags").SetEntry(NodeTag{})
}

func nodeTagCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		n, err := getNodeTagByID(id)
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "nodeTag", n)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getNodeTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, ok := ctx.Value("nodeTag").(NodeTag)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	jsonEncode(w, []NodeTag{n})
}

func getNodeTagByID(id string) (NodeTag, error) {
	data := make([]NodeTag, 0)
	result := db.Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return NodeTag{}, result.Error
	}
	if len(data) == 0 {
		return NodeTag{}, fmt.Errorf("not found")
	}
	return data[0], nil
}

func getNodesTags(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	rq := tables["node_tags"].Request()
	rq.AutoJoin("nodes")
	if app, ok := claims["app"]; ok && app != "" {
		rq.Where("nodes.app = ?", app)
	}
	td, err := rq.MakeResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func getNodeTags(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	ctx := r.Context()
	n, _ := ctx.Value("node").(Node)
	rq := tables["node_tags"].Request()
	rq.AutoJoin("nodes")
	rq.Where("nodes.id = ?", n.ID)
	if app, ok := claims["app"]; ok && app != "" {
		rq.Where("nodes.app = ?", app)
	}
	td, err := rq.MakeResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func getNodeCandidateTags(w http.ResponseWriter, r *http.Request) {
	//_, claims, _ := jwtauth.FromContext(r.Context())
	ctx := r.Context()
	n, _ := ctx.Value("node").(Node)
	exclude := db.Table("node_tags").Where("node_id = ?", n.NodeID).Select("tag_id")

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

func getTagNodes(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	ctx := r.Context()
	tag, _ := ctx.Value("tag").(Tag)
	fmt.Println("xx", tag)
	rq := tables["nodes"].Request()
	rq.AutoJoin("tags")
	if app, ok := claims["app"]; ok && app != "" {
		rq.Where("nodes.app = ?", app)
	}
	rq.Where("node_tags.tag_id = ?", tag.TagID)
	if app, ok := claims["app"]; ok && app != "" {
		rq.Where("nodes.app = ?", app)
	}
	td, err := rq.MakeResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}
