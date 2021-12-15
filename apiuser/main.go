package apiuser

import (
	"net/http"

	"github.com/shaj13/go-guardian/v2/auth"
	"gorm.io/gorm"
)

const (
	XPrivileges string = "privileges"
)

func DefaultApp(db *gorm.DB, t auth.Info) string {
	var app string
	db.
		Table("apps").
		Joins("JOIN apps_responsibles ON apps.id = apps_responsibles.app_id").
		Joins("JOIN auth_membership ON apps_responsibles.group_id = auth_membership.group_id").
		Where("auth_membership.user_id = ?", t.GetID()).
		Select("apps.app").
		Order("apps.app").
		Limit(1).
		Find(&app)
	return app
}

func PrimaryGroup(db *gorm.DB, t auth.Info) string {
	var role string
	db.
		Table("auth_group").
		Joins("JOIN auth_membership ON auth_group.id = auth_membership.group_id").
		Where("auth_membership.primary_group = 'T' AND auth_membership.user_id = ?", t.GetID()).
		Select("role").
		Take(&role)
	return role
}

func MakeExtensions(db *gorm.DB, id uint) auth.Extensions {
	ext := make(auth.Extensions)
	ext[XPrivileges] = getPrivileges(db, id)
	return ext
}

func getPrivileges(db *gorm.DB, id uint) []string {
	var roles []string
	db.
		Table("auth_group").
		Joins("JOIN auth_membership ON auth_group.id = auth_membership.group_id").
		Joins("JOIN auth_user ON auth_membership.user_id = auth_user.id").
		Where("auth_group.privilege = ? AND auth_user.id = ?", "T", id).
		Pluck("role", &roles)
	return roles
}

func IsManager(t auth.Info) bool {
	return HasPrivilege(t, "Manager")
}

func HasPrivilege(t auth.Info, priv string) bool {
	privs := t.GetExtensions().Values(XPrivileges)
	for _, s := range privs {
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
	http.Error(w, priv, 401)
}
