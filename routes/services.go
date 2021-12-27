package routes

import (
	"fmt"
	"net/http"

	"github.com/opensvc/collector-api/db"
)

//
// GetNodesTags     godoc
// @Summary   List services
// @Security  BasicAuth
// @Security  BearerAuth
// @Tags      services
// @Accept    json
// @Produce   json
// @Success   200      {object}  db.TableResponse
// @Failure   500      {string}  string    "Internal Server Error"
// @Param     props    query     string    false  "properties to include, and optionally remap (comma separated)"
// @Param     groupby  query     string    false  "properties to group by (comma separated)"
// @Param     order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param     filters  query     []string  false  "property value filter (a, !a, a&b, a|b, (a,b),  a%,  a%&!ab%)"
// @Param     limit    query     int       false  "number of objets to include in response"
// @Param     offset   query     int       false  "offset of the first objet to include in response"
// @Param     meta     query     bool      false  "turn off metadata in response"
// @Router    /services  [get]
//
func GetServices(w http.ResponseWriter, r *http.Request) {
	rq := db.Tab("services").Request()
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
