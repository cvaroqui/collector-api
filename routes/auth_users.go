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

//
// GetNodes     godoc
// @Summary      List users
// @Description  List users
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200      {object}  db.TableResponse
// @Failure      500    {string}  string  "Internal Server Error"
// @Param        props    query     string    false  "properties to include, and optionally remap (comma separated)"
// @Param        groupby  query     string    false  "properties to group by (comma separated)"
// @Param        order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param        filters  query     []string  false  "property value filter (a, !a, a&b, a|b, (a,b),  a%,  a%&!ab%)"
// @Param        limit    query     int       false  "number of objets to include in response"
// @Param        offset   query     int       false  "offset of the first objet to include in response"
// @Param        meta     query     bool      false  "turn off metadata in response"
// @Router       /users  [get]
//
func GetUsers(w http.ResponseWriter, r *http.Request) {
	rq := db.Tab("auth_user").Request()
	td, err := rq.MakeTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
}

//
// PostUsers	godoc
// @Summary      Create or update users
// @Description  The user must be in the UserManager privilege group to modify tiers users properties.
// @Description  The user must be in the SelfManager privilege group to modify its user properties.
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        users  body      []tables.User  true  "list of users to create or update"
// @Success      200    {array}   tables.User
// @Failure      500      {string}  string    "Internal Server Error"
// @Router       /users  [post]
//
func PostUsers(w http.ResponseWriter, r *http.Request) {
	users := make([]tables.User, 0)
	user := tables.User{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	if err := json.Unmarshal(body, &user); err == nil {
		// single entry
		users = append(users, user)
	} else if err := json.Unmarshal(body, &users); err != nil {
		// list of entry
		http.Error(w, fmt.Sprint(err), 500)
		return
	}
	for _, user := range users {
		tx := db.DB().Clauses(clause.OnConflict{UpdateAll: true})
		if err := tx.Create(&user).Error; err != nil {
			http.Error(w, fmt.Sprint(err), 500)
			return
		}
	}
}
