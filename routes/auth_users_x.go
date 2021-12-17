package routes

import (
	"net/http"

	"github.com/opensvc/collector-api/authuser"
	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
	"github.com/shaj13/go-guardian/v2/auth"
)

func DelUser(w http.ResponseWriter, r *http.Request) {
	user := auth.User(r)
	if authuser.HasPrivilege(user, "UserManager") {
		http.Error(w, http.StatusText(403), 403)
		return
	}
	ctx := r.Context()
	usr, ok := ctx.Value("user").(tables.User)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	db.DB().Table("auth_user").Where("auth_user.id = ?", usr.ID).Delete(&tables.User{})
	jsonEncode(w, []tables.User{usr})
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, ok := ctx.Value("user").(tables.User)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	jsonEncode(w, []tables.User{n})
}
