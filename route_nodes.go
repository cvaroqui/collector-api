package main

import (
	"fmt"
	"net/http"
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
// @Success      200      {object}  TableResponse
// @Failure      500      {string}  string    "ok"
// @Param        props    query     string    false  "properties to include, and optionally remap (comma separated)"
// @Param        groupby  query     string    false  "properties to group by (comma separated)"
// @Param        order    query     string    false  "properties to order by (comma separated, prefix with '~' to reverse)"
// @Param        filters  query     []string  false  "property value filter (a&b:a AND b, a|b:a OR b, (a,b):IN  (a,b),  !a:NOT  a,  %a:LIKE  a%)"
// @Param        limit    query     int       false  "number of objets to include in response"
// @Param        offset   query     int       false  "offset of the first objet to include in response"
// @Param        meta     query     bool      false  "turn off metadata in response"
// @Router       /nodes  [get]
//
func getNodes(w http.ResponseWriter, r *http.Request) {
	req := tables["nodes"].Request()
	td, err := req.MakeResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}
