package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
	"gorm.io/gorm/clause"
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	rq := db.Tab("auth_user").Request()
	td, err := rq.MakeTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

func PostUsers(w http.ResponseWriter, r *http.Request) {
	users := make([]tables.User, 0)
	user := tables.User{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := json.Unmarshal(body, &user); err == nil {
		// single entry
		users = append(users, user)
	} else if err := json.Unmarshal(body, &users); err != nil {
		// list of entry
		http.Error(w, fmt.Sprint(err), 500)
	}
	for _, user := range users {
		tx := db.DB().Clauses(clause.OnConflict{UpdateAll: true})
		if err := tx.Create(&user).Error; err != nil {
			http.Error(w, fmt.Sprint(err), 500)
		}
	}
}
