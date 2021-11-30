package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"gorm.io/gorm"
)

type NodeTag struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	NodeID        string         `gorm:"column:node_id; size:36" json:"node_id"`
	TagID         string         `gorm:"column:tag_id; size:40; index" json:"tag_id"`
	TagAttachData string         `gorm:"column:tag_attach_data; type:text" json:"tag_attach_data"`
	Created       time.Time      `gorm:"column:created; autoCreateTime" json:"created"`
}

func init() {
	tables["node_tags"] = newTable("node_tags").
		SetEntry(NodeTag{}).
		SetJoin("nodes", "left join nodes on nodes.node_id=node_tags.node_id").
		SetJoin("svcmon", "left join svcmon on node_tags.node_id=svcmon.node_id").
		SetJoin("services", "left join svcmon on node_tags.node_id=svcmon.node_id left join services on svcmon.svc_id=services.svc_id")
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

func getNodeTags(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	tx := tables["node_tags"].DBTable().Joins("left join nodes on node_tags.node_id = nodes.node_id")
	if app, ok := claims["app"]; ok && app != "" {
		tx = tx.Where("nodes.app = ?", app)
	}
	data := make([]NodeTag, 0)
	td, err := tables["node_tags"].MakeResponse(r, tx, &data)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}
