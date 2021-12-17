package tables

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/opensvc/collector-api/db"
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
	db.Register(&db.Table{
		Name:  "node_tags",
		Entry: NodeTag{},
	})
}

func NodeTagCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		n, err := getNodeTagByID(r, id)
		if err != nil { // TODO: return list and let users decide about errs
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "nodeTag", n)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getNodeTagByID(r *http.Request, id string) (NodeTag, error) {
	data := make([]NodeTag, 0)
	result := db.Tab("node_tags").Request().TX(r).Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return NodeTag{}, result.Error
	}
	if len(data) == 0 {
		return NodeTag{}, fmt.Errorf("not found")
	}
	return data[0], nil
}
