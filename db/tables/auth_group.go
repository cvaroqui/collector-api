package tables

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/opensvc/collector-api/db"
	"gorm.io/gorm"
)

/*
+-------------+--------------+------+-----+---------+----------------+
| Field       | Type         | Null | Key | Default | Extra          |
+-------------+--------------+------+-----+---------+----------------+
| id          | int(11)      | NO   | PRI | NULL    | auto_increment |
| role        | varchar(255) | YES  | UNI | NULL    |                |
| description | longtext     | YES  |     | NULL    |                |
| privilege   | varchar(1)   | YES  | MUL | F       |                |
+-------------+--------------+------+-----+---------+----------------+
*/

type Group struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Role        string         `gorm:"column:role; index:idx2,unique; size:255" json:"role"`
	Description string         `gorm:"column:description; type:longtext" json:"description"`
	Privilege   bool           `gorm:"column:privilege; index:idx1" json:"privilege"`
}

func init() {
	db.Register(&db.Table{
		Name:  "auth_group",
		Entry: Group{},
	})
}

func GroupFromCtx(r *http.Request) []Group {
	i := r.Context().Value("group")
	if i == nil {
		return []Group{}
	}
	return i.([]Group)
}

func GroupCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if reID.MatchString(id) {
			if n, err := readableGetGroupByID(r, id); err == nil {
				ctx := context.WithValue(r.Context(), "group", n)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
		} else {
			if n, err := readableGetGroupByName(r, id); err == nil {
				ctx := context.WithValue(r.Context(), "group", n)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
		}
		http.Error(w, "unsupported group id format", 500)
	})
}

func readableGetGroupByName(r *http.Request, id string) ([]Group, error) {
	data := make([]Group, 0)
	tx := db.Tab("auth_group").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	result := tx.Where("role = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func readableGetGroupByID(r *http.Request, id string) ([]Group, error) {
	data := make([]Group, 0)
	tx := db.Tab("auth_group").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	result := tx.Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func GetGroupByName(id string) ([]Group, error) {
	data := make([]Group, 0)
	tx := db.DB().Table("auth_group")
	result := tx.Where("role = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func GetGroupByID(id string) ([]Group, error) {
	data := make([]Group, 0)
	tx := db.DB().Table("auth_group")
	result := tx.Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}
