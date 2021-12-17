package apiuser

import (
	"github.com/opensvc/collector-api/authuser"
	"github.com/opensvc/collector-api/db"
	"github.com/opensvc/collector-api/db/tables"
	"github.com/shaj13/go-guardian/v2/auth"
)

func DefaultApp(t auth.Info) string {
	var app string
	db.DB().
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

func PrimaryGroup(t auth.Info) string {
	var role string
	db.DB().
		Table("auth_group").
		Joins("JOIN auth_membership ON auth_group.id = auth_membership.group_id").
		Where("auth_membership.primary_group = 'T' AND auth_membership.user_id = ?", t.GetID()).
		Select("role").
		Take(&role)
	return role
}

func MakeNodeExtensions(node tables.Node) auth.Extensions {
	ext := make(auth.Extensions)
	ext[authuser.XNodeID] = []string{node.NodeID}
	ext[authuser.XApp] = []string{node.App}
	return ext
}

func MakeUserExtensions(user tables.User) auth.Extensions {
	ext := make(auth.Extensions)
	ext[authuser.XPrivileges] = getPrivileges(user.ID)
	return ext
}

func getPrivileges(id uint) []string {
	var roles []string
	db.DB().
		Table("auth_group").
		Joins("JOIN auth_membership ON auth_group.id = auth_membership.group_id").
		Joins("JOIN auth_user ON auth_membership.user_id = auth_user.id").
		Where("auth_group.privilege = ? AND auth_user.id = ?", "T", id).
		Pluck("role", &roles)
	return roles
}
