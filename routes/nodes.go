package routes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/opensvc/collector-api/apiuser"
	"github.com/opensvc/collector-api/auth"
	"github.com/opensvc/collector-api/authuser"
	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//
// GetNodes	godoc
// @Summary      List nodes
// @Description  List nodes
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         nodes
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
// @Router       /nodes  [get]
//
func GetNodes(w http.ResponseWriter, r *http.Request) {
	rq := db.Tab("nodes").Request()
	td, err := rq.MakeReadTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}

//
// PostNodes	godoc
// @Summary      Create or update nodes
// @Description  The app code of the nodes is forced to one the user is responsible of.
// @Description  The team responsible of the nodes defaults to the user's primary group.
// @Description  The user must be in the NodeManager privilege group.
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         nodes
// @Accept       json
// @Produce      json
// @Param        nodes  body      []Node  true  "list of nodes to create or update"
// @Success      200    {array}   Node
// @Failure      500      {string}  string    "Internal Server Error"
// @Router       /nodes  [post]
//
func PostNodes(w http.ResponseWriter, r *http.Request) {
	user := auth.User(r)
	if !authuser.HasPrivilege(user, "NodeManager") {
		authuser.PrivError(w, "NodeManager")
		return
	}
	nodes := make([]tables.Node, 0)
	node := tables.Node{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("read request body: %s", err), 500)
		return
	}
	if err := json.Unmarshal(body, &node); err == nil {
		// single entry
		nodes = append(nodes, node)
	} else if err := json.Unmarshal(body, &nodes); err != nil {
		// list of entry
		http.Error(w, fmt.Sprintf("unmarshal json: %s", err), 500)
		return
	}
	userPrimaryGroup := apiuser.PrimaryGroup(user)
	userDefaultApp := apiuser.DefaultApp(user)
	var myNodes *gorm.DB
	if !authuser.HasPrivilege(user, "Manager") {
		myNodes = db.DB().
			Joins("JOIN apps ON apps.app = nodes.app").
			Joins("JOIN apps_responsibles ON apps_responsibles.app_id = apps.id").
			Joins("JOIN auth_membership ON auth_membership.group_id = apps_responsibles.group_id AND auth_membership.user_id = ?", user.GetID()).
			Select("nodes.id")
		temp := []tables.Node{}
		if err := myNodes.Find(&temp).Error; err != nil {
			http.Error(w, fmt.Sprintf("nodes under user responsabilty: %s", err), 500)
			return
		}
	}
	for i, n := range nodes {
		if n.ID != 0 {
			db.DB().Where("id = ?", n.ID).Take(&n)
		} else if n.NodeID != "" {
			db.DB().Where("node_id = ?", n.NodeID).Take(&n)
		} else if n.Nodename != "" && n.App != "" {
			db.DB().Where("nodename = ? AND app = ?", n.Nodename, n.App).Take(&n)
		}
		if n.ID == 0 {
			// new entry ... populate required field we have defaults for

			if n.App == "" {
				// set a default team responsible
				if userDefaultApp == "" {
					http.Error(w, "insert or update: user has no default app, an app must be set", 500)
					return
				}
				n.App = userDefaultApp

				// new chance to find an existing node
				if n.Nodename != "" && n.App != "" {
					db.DB().Where("nodename = ? AND app = ?", n.Nodename, n.App).Take(&n)
				}
			}
		}
		if n.ID == 0 {
			if n.TeamResponsible == "" {
				// set a default team responsible
				if userPrimaryGroup == "" {
					http.Error(w, "insert or update: user has no primary group, a team responsible must be set", 500)
					return
				}
				n.TeamResponsible = userPrimaryGroup
			}
		} else if myNodes != nil {
			var i int64
			if err := myNodes.Where("nodes.id = ?", n.ID).Count(&i).Error; err != nil {
				http.Error(w, fmt.Sprintf("insert or update: %s", err), 500)
				return
			}
			if i == 0 {
				http.Error(w, fmt.Sprintf("insert or update: user is not responsible for node %s in app %s", n.Nodename, n.App), 500)
				return
			}
		}
		nodes[i] = n
	}

	tx := db.DB().Clauses(clause.OnConflict{UpdateAll: true})
	if err := tx.Create(&nodes).Error; err != nil {
		http.Error(w, fmt.Sprintf("insert or update: %s", err), 500)
		return
	}
	if err := jsonEncode(w, nodes); err != nil {
		http.Error(w, fmt.Sprintf("json encode: %s", err), 500)
		return
	}
}
