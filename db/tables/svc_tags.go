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

func ServiceTagCtx(next http.Handler) http.Handler {
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

func getServiceTagByID(id string) (ServiceTag, error) {
	data := make([]ServiceTag, 0)
	result := db.DB().Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return ServiceTag{}, result.Error
	}
	if len(data) == 0 {
		return ServiceTag{}, fmt.Errorf("not found")
	}
	return data[0], nil
}
