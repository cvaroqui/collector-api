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

func NodeTagFromCtx(r *http.Request) []NodeTag {
	i := r.Context().Value("nodeTag")
	if i == nil {
		return []NodeTag{}
	}
	return i.([]NodeTag)
}

func NodeTagCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nodes := NodeFromCtx(r)
		tags := TagFromCtx(r)
		if len(nodes) == 1 && len(tags) == 1 {
			n, err := getNodeTagByNodeIDAndTagID(r, nodes[0].NodeID, tags[0].TagID)
			if err != nil {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
			ctx := context.WithValue(r.Context(), "nodeTag", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			id := chi.URLParam(r, "id")
			n, err := getNodeTagByID(r, id)
			if err != nil {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
			ctx := context.WithValue(r.Context(), "nodeTag", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

	})
}

func getNodeTagByNodeIDAndTagID(r *http.Request, nodeID, tagID string) ([]NodeTag, error) {
	data := make([]NodeTag, 0)
	tx := db.Tab("node_tags").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	if err := tx.Where("node_tags.node_id = ? AND node_tags.tag_id = ?", nodeID, tagID).Find(&data).Error; err != nil {
		return data, err
	}
	return data, nil
}

func getNodeTagByID(r *http.Request, id string) ([]NodeTag, error) {
	data := make([]NodeTag, 0)
	tx := db.Tab("node_tags").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	if err := tx.Where("id = ?", id).Find(&data).Error; err != nil {
		return data, err
	}
	return data, nil
}
