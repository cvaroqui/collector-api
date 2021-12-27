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
	db.Register(&db.Table{
		Name:  "svc_tags",
		Entry: ServiceTag{},
	})
}

func ServiceTagFromCtx(r *http.Request) []ServiceTag {
	i := r.Context().Value("serviceTag")
	if i == nil {
		return []ServiceTag{}
	}
	return i.([]ServiceTag)
}

func ServiceTagCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		services := ServiceFromCtx(r)
		tags := TagFromCtx(r)
		if len(services) == 1 && len(tags) == 1 {
			n, err := getServiceTagBySvcIDAndTagID(r, services[0].SvcID, tags[0].TagID)
			if err != nil {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
			ctx := context.WithValue(r.Context(), "serviceTag", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			id := chi.URLParam(r, "id")
			n, err := getServiceTagByID(r, id)
			if err != nil {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
			ctx := context.WithValue(r.Context(), "serviceTag", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func getServiceTagBySvcIDAndTagID(r *http.Request, svcID, tagID string) ([]ServiceTag, error) {
	data := make([]ServiceTag, 0)
	tx := db.Tab("svc_tags").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	if err := tx.Where("svc_tags.node_id = ? AND svc_tags.tag_id = ?", svcID, tagID).Find(&data).Error; err != nil {
		return data, err
	}
	return data, nil
}

func getServiceTagByID(r *http.Request, id string) ([]ServiceTag, error) {
	data := make([]ServiceTag, 0)
	tx := db.Tab("svc_tags").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	if err := tx.Where("id = ?", id).Find(&data).Error; err != nil {
		return data, err
	}
	return data, nil
}
