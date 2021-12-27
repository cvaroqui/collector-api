package routes

import (
	"net/http"

	"github.com/opensvc/collector-api/db/tables"
)

//
// GetService     godoc
// @Summary      Show a service
// @Description  Show a service by index, id or name
// @Security     BasicAuth
// @Security     BearerAuth
// @Tags         services
// @Accept       json
// @Produce      json
// @Success      200  {array}   tables.Service
// @Failure      404  {string}  string  "Not Found"
// @Failure      500  {string}  string  "Internal Server Error"
// @Param        id   path      string  true  "the index of the entry in database, or uuid, or name"
// @Router       /services/{id}  [get]
//
func GetService(w http.ResponseWriter, r *http.Request) {
	data := tables.ServiceFromCtx(r)
	if len(data) == 0 {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	jsonEncode(w, data)
}
