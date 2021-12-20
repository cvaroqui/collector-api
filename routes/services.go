package routes

import (
	"fmt"
	"net/http"

	"github.com/opensvc/collector-api/db"
)

func GetServices(w http.ResponseWriter, r *http.Request) {
	rq := db.Tab("services").Request()
	td, err := rq.MakeTableResponse(r)
	if err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
	if err := jsonEncode(w, td); err != nil {
		http.Error(w, fmt.Sprint(err), 500)
	}
}
