package authuser

import (
	"fmt"
	"net/http"

	"github.com/shaj13/go-guardian/v2/auth"
)

const (
	XNodeID     string = "node_id"
	XApp        string = "app"
	XPrivileges string = "privileges"
)

func IsManager(t auth.Info) bool {
	return HasPrivilege(t, "Manager")
}

func HasPrivilege(t auth.Info, priv string) bool {
	privs := t.GetExtensions().Values(XPrivileges)
	for _, s := range privs {
		if s == "Manager" {
			return true
		}
		if s == priv {
			return true
		}
	}
	return false
}

func Privileges(t auth.Info) []string {
	return t.GetExtensions().Values(XPrivileges)
}

func PrivError(w http.ResponseWriter, priv string) {
	http.Error(w, fmt.Sprintf("%s: requires %s", http.StatusText(403), priv), 401)
}

func IsNode(t auth.Info) bool {
	return t.GetExtensions().Has(XNodeID)
}
