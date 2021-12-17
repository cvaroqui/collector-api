package routes

import (
	"net/http"

	"github.com/opensvc/collector-api/db/tables"
)

func GetService(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, ok := ctx.Value("service").(tables.Service)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	jsonEncode(w, []tables.Service{n})
}
