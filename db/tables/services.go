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
+-----------------------------+---------------+------+-----+---------------------+----------------+
| Field                       | Type          | Null | Key | Default             | Extra          |
+-----------------------------+---------------+------+-----+---------------------+----------------+
| svc_hostid                  | varchar(30)   | YES  | MUL | NULL                |                |
| svcname                     | varchar(60)   | YES  | MUL | NULL                |                |
| svc_nodes                   | varchar(1000) | YES  |     | NULL                |                |
| svc_drpnode                 | varchar(30)   | YES  | MUL | NULL                |                |
| svc_drptype                 | varchar(7)    | YES  |     | NULL                |                |
| svc_autostart               | varchar(60)   | NO   |     |                     |                |
| svc_env                     | varchar(10)   | YES  |     | NULL                |                |
| svc_drpnodes                | varchar(1000) | YES  |     | NULL                |                |
| svc_comment                 | varchar(1000) | YES  |     | NULL                |                |
| svc_app                     | varchar(64)   | YES  | MUL | NULL                |                |
| svc_drnoaction              | varchar(1)    | YES  |     | F                   |                |
| svc_created                 | timestamp     | NO   |     | current_timestamp() |                |
| svc_config_updated          | datetime      | YES  |     | NULL                |                |
| svc_metrocluster            | varchar(10)   | YES  |     | NULL                |                |
| id                          | int(11)       | NO   | PRI | NULL                | auto_increment |
| svc_wave                    | varchar(10)   | NO   |     | 3                   |                |
| svc_config                  | mediumtext    | YES  |     | NULL                |                |
| updated                     | datetime      | NO   |     | NULL                |                |
| svc_topology                | varchar(20)   | YES  | MUL | failover            |                |
| svc_flex_min_nodes          | int(11)       | YES  |     | 1                   |                |
| svc_flex_max_nodes          | int(11)       | YES  |     | 0                   |                |
| svc_flex_cpu_low_threshold  | int(11)       | YES  |     | 0                   |                |
| svc_flex_cpu_high_threshold | int(11)       | YES  |     | 100                 |                |
| svc_status                  | varchar(10)   | YES  |     | undef               |                |
| svc_availstatus             | varchar(10)   | YES  |     | undef               |                |
| svc_ha                      | tinyint(1)    | YES  |     | 0                   |                |
| svc_status_updated          | datetime      | YES  |     | NULL                |                |
| svc_id                      | char(36)      | YES  | UNI |                     |                |
| svc_frozen                  | varchar(6)    | YES  |     | NULL                |                |
| svc_provisioned             | varchar(6)    | YES  |     | NULL                |                |
| svc_placement               | varchar(12)   | YES  |     | NULL                |                |
| svc_notifications           | varchar(1)    | YES  |     | T                   |                |
| svc_snooze_till             | datetime      | YES  |     | NULL                |                |
| cluster_id                  | char(36)      | YES  | MUL |                     |                |
| svc_flex_target             | int(11)       | YES  |     | NULL                |                |
+-----------------------------+---------------+------+-----+---------------------+----------------+
*/

type Service struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Svcname          string         `gorm:"column:svcname; index" json:"svcname"`
	SvcID            string         `gorm:"column:svc_id; size:36; uniqueIndex; size:36" json:"svc_id"`
	ClusterID        string         `gorm:"column:cluster_id; size:36; index" json:"cluster_id"`
	SvcConfigUpdated time.Time      `gorm:"column:svc_config_updated" json:"svc_config_updated"`
	SvcSnoozeTill    time.Time      `gorm:"column:svc_snooze_till" json:"svc_snooze_till"`
	SvcApp           string         `gorm:"column:svc_app; index" json:"svc_app"`
	SvcEnv           string         `gorm:"column:svc_env" json:"svc_env"`
	SvcTopology      string         `gorm:"column:svc_topology" json:"svc_topology"`
	SvcStatus        string         `gorm:"column:svc_status; size:10; index" json:"svc_status"`
	SvcAvailStatus   string         `gorm:"column:svc_avail_status; size:10; index" json:"svc_avail_status"`
	SvcHA            string         `gorm:"column:svc_ha; size:1; default:'F'" json:"svc_ha"`
	SvcFrozen        string         `gorm:"column:svc_frozen; size:6" json:"svc_frozen"`
	SvcProvisioned   string         `gorm:"column:svc_provisioned; size:6" json:"svc_provisioned"`
	SvcPlacement     string         `gorm:"column:svc_placement; size:12" json:"svc_placement"`
	SvcFlexTarget    int            `gorm:"column:svc_flex_target" json:"svc_flex_target"`
	SvcFlexMinNodes  int            `gorm:"column:svc_flex_min_nodes" json:"svc_flex_min_nodes"`
	SvcFlexMaxNodes  int            `gorm:"column:svc_flex_max_nodes" json:"svc_flex_max_nodes"`
	SvcWave          string         `gorm:"column:svc_wave; default:'3'" json:"svc_wave"`
	SvcConfig        string         `gorm:"column:svc_config; type:mediumtext" json:"svc_wave"`
	SvcComment       string         `gorm:"column:svc_config; type:mediumtext" json:"svc_wave"`
	Updated          time.Time      `gorm:"column:created" json:"created"`
	Created          time.Time      `gorm:"column:updated" json:"updated"`
}

func init() {
	db.Register(&db.Table{
		Name:  "services",
		Entry: Service{},
	})
}

func ServiceCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if reUUID.MatchString(id) {
			n, err := getServiceBySvcID(id)
			if err != nil {
				http.Error(w, http.StatusText(404), 404)
				return
			}
			ctx := context.WithValue(r.Context(), "service", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else if reID.MatchString(id) {
			n, err := getServiceByID(id)
			if err != nil {
				http.Error(w, http.StatusText(404), 404)
				return
			}
			ctx := context.WithValue(r.Context(), "service", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			n, err := getServiceByName(id)
			if err != nil {
				http.Error(w, http.StatusText(404), 404)
				return
			}
			ctx := context.WithValue(r.Context(), "service", n)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

func getServiceBySvcID(svcID string) (Service, error) {
	data := make([]Service, 0)
	result := db.DB().Where("svc_id = ?", svcID).Find(&data)
	if result.Error != nil {
		return Service{}, result.Error
	}
	if len(data) == 0 {
		return Service{}, fmt.Errorf("not found")
	}
	return data[0], nil
}

func getServiceByName(name string) (Service, error) {
	data := make([]Service, 0)
	result := db.DB().Where("svcname = ?", name).Find(&data)
	if result.Error != nil {
		return Service{}, result.Error
	}
	if len(data) == 0 {
		return Service{}, fmt.Errorf("not found")
	}
	return data[0], nil
}

func getServiceByID(id string) (Service, error) {
	data := make([]Service, 0)
	result := db.DB().Where("id = ?", id).Find(&data)
	if result.Error != nil {
		return Service{}, result.Error
	}
	if len(data) == 0 {
		return Service{}, fmt.Errorf("not found")
	}
	return data[0], nil
}
