package tables

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/opensvc/collector-api/db"
	"gorm.io/gorm"
)

/*
+---------------------------+---------------------------------------------------+------+-----+---------+----------------+
| Field                     | Type                                              | Null | Key | Default | Extra          |
+---------------------------+---------------------------------------------------+------+-----+---------+----------------+
| id                        | int(11)                                           | NO   | PRI | NULL    | auto_increment |
| first_name                | varchar(128)                                      | YES  |     | NULL    |                |
| last_name                 | varchar(128)                                      | YES  |     | NULL    |                |
| email                     | varchar(512)                                      | YES  |     | NULL    |                |
| password                  | varchar(512)                                      | YES  |     | NULL    |                |
| registration_key          | varchar(512)                                      | YES  |     | NULL    |                |
| reset_password_key        | varchar(512)                                      | YES  |     |         |                |
| email_notifications       | varchar(1)                                        | YES  |     | T       |                |
| im_notifications          | varchar(1)                                        | YES  |     | T       |                |
| im_type                   | varchar(16)                                       | YES  |     | NULL    |                |
| im_username               | varchar(100)                                      | YES  |     | NULL    |                |
| email_log_level           | enum('debug','info','warning','error','critical') | YES  |     | warning |                |
| im_log_level              | enum('debug','info','warning','error','critical') | YES  |     | warning |                |
| lock_filter               | varchar(1)                                        | YES  |     | F       |                |
| phone_work                | varchar(16)                                       | YES  |     | NULL    |                |
| registration_id           | varchar(512)                                      | YES  |     |         |                |
| quota_app                 | int(11)                                           | YES  |     | NULL    |                |
| quota_org_group           | int(11)                                           | YES  |     | NULL    |                |
| username                  | varchar(128)                                      | YES  |     | NULL    |                |
| quota_docker_registries   | int(11)                                           | YES  |     | NULL    |                |
| im_notifications_delay    | int(11)                                           | YES  |     | 0       |                |
| email_notifications_delay | int(11)                                           | YES  |     | 0       |                |
+---------------------------+---------------------------------------------------+------+-----+---------+----------------+
*/

type User struct {
	ID                    uint           `gorm:"primarykey" json:"id"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Username              string         `gorm:"column:username; size:128" json:"username"`
	FirstName             string         `gorm:"column:first_name; size:128" json:"first_name"`
	LastName              string         `gorm:"column:last_name; size:128" json:"last_name"`
	Email                 string         `gorm:"column:email; size:512" json:"email"`
	Password              string         `gorm:"column:password; size:512" json:"password"`
	ResetPasswordKey      string         `gorm:"column:reset_password_key; size:512" json:"reset_password_key"`
	RegistrationKey       string         `gorm:"column:registration_key; size:512" json:"registration_key"`
	RegistrationID        string         `gorm:"column:registration_id; size:512" json:"registration_id"`
	EmailNotifications    string         `gorm:"column:email_notifications; size:1" json:"email_notifications"`
	EmailLogLevel         string         `gorm:"column:email_log_level; type:enum('debug','info','warning','error','critical'); default:warning" json:"email_log_level"`
	IMNotifications       string         `gorm:"column:im_notifications; size:1" json:"im_notifications"`
	IMNotificationsDelay  int            `gorm:"column:im_notifications_delay" json:"im_notifications_delay"`
	IMType                string         `gorm:"column:im_type; size:16" json:"im_type"`
	IMUsername            string         `gorm:"column:im_username; size:100" json:"im_username"`
	IMLogLevel            string         `gorm:"column:im_log_level; type:enum('debug','info','warning','error','critical'); default:warning" json:"im_log_level"`
	LockFilter            string         `gorm:"column:lock_filter; size:1" json:"lock_filter"`
	PhoneWork             string         `gorm:"column:phone_work; size:16" json:"phone_work"`
	QuotaApp              int            `gorm:"column:quota_app" json:"quota_app"`
	QuotaOrgGroup         int            `gorm:"column:quota_org_group" json:"quota_org_group"`
	QuotaDockerRegistries int            `gorm:"column:quota_docker_registries" json:"quota_docker_registries"`
}

func init() {
	db.Register(&db.Table{
		Name:  "auth_user",
		Entry: User{},
	})
}

func UserFromCtx(r *http.Request) []User {
	i := r.Context().Value("user")
	if i == nil {
		return []User{}
	}
	return i.([]User)
}

func UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if reID.MatchString(id) {
			if n, err := readableGetUserByID(r, id); err == nil {
				ctx := context.WithValue(r.Context(), "user", n)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
		} else if _, err := mail.ParseAddress(id); err == nil {
			if n, err := readableGetUserByEmail(r, id); err == nil {
				ctx := context.WithValue(r.Context(), "user", n)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
		} else {
			if n, err := readableGetUserByUsername(r, id); err == nil {
				ctx := context.WithValue(r.Context(), "user", n)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				http.Error(w, fmt.Sprint(err), 500)
				return
			}
		}
		http.Error(w, "unsupported node id format", 500)
	})
}

func readableGetUserByUsername(r *http.Request, id string) ([]User, error) {
	data := make([]User, 0)
	tx := db.Tab("auth_user").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	result := tx.Where("username = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func readableGetUserByEmail(r *http.Request, id string) ([]User, error) {
	data := make([]User, 0)
	tx := db.Tab("auth_user").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	result := tx.Where("email = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func readableGetUserByID(r *http.Request, id string) ([]User, error) {
	data := make([]User, 0)
	tx := db.Tab("auth_user").Request(
		db.TableRequestWithFilters(false),
		db.TableRequestWithPaging(false),
	).TX(r)
	result := tx.Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func GetUserByUsername(id string) ([]User, error) {
	data := make([]User, 0)
	tx := db.DB().Table("auth_user")
	result := tx.Where("username = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func GetUserByEmail(id string) ([]User, error) {
	data := make([]User, 0)
	tx := db.DB().Table("auth_user")
	result := tx.Where("email = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}

func GetUserByID(id string) ([]User, error) {
	data := make([]User, 0)
	tx := db.DB().Table("auth_user")
	result := tx.Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return data, result.Error
	}
	return data, nil
}
