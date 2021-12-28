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
+--------------+-------------+------+-----+---------------------+----------------+
| Field        | Type        | Null | Key | Default             | Extra          |
+--------------+-------------+------+-----+---------------------+----------------+
| id           | int(11)     | NO   | PRI | NULL                | auto_increment |
| app          | varchar(64) | YES  | UNI | NULL                |                |
| updated      | timestamp   | NO   |     | current_timestamp() |                |
| app_domain   | varchar(64) | YES  |     | NULL                |                |
| app_team_ops | varchar(64) | YES  | MUL | NULL                |                |
| description  | text        | NO   |     |                     |                |
+--------------+-------------+------+-----+---------------------+----------------+
*/

type App struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	App         string         `gorm:"column:app; index:i_app,unique; size:64" json:"app"`
	AppDomain   string         `gorm:"column:app_domain; size:64" json:"app_domain"`
	AppTeamOps  string         `gorm:"column:app_team_ops; index:idx_app_team_ops; size:64" json:"app_team_ops"`
	Description string         `gorm:"column:description; type:text" json:"description"`
}

func init() {
	db.Register(&db.Table{
		Name:  "apps",
		Entry: App{},
	})
}

func AppFromCtx(r *http.Request) []App {
	i := r.Context().Value("app")
	if i == nil {
		return []App{}
	}
	return i.([]App)
}

func AppCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if reID.MatchString(id) {
			if n, err := readableGetAppByID(r, id); err == nil {
				ctx := context.WithValue(r.Context(), "app", n)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
		} else {
			if n, err := readableGetAppByName(r, id); err == nil {
				ctx := context.WithValue(r.Context(), "app", n)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
		}
		http.Error(w, "unsupported app id format", 500)
	})
}

func readableGetAppByName(r *http.Request, id string) ([]App, error) {
	data := make([]App, 0)
	tx := db.Tab("apps").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	result := tx.Where("app = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func readableGetAppByID(r *http.Request, id string) ([]App, error) {
	data := make([]App, 0)
	tx := db.Tab("apps").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	result := tx.Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func GetAppByName(id string) ([]App, error) {
	data := make([]App, 0)
	tx := db.DB().Table("apps")
	result := tx.Where("app = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func GetAppByID(id string) ([]App, error) {
	data := make([]App, 0)
	tx := db.DB().Table("apps")
	result := tx.Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}
