package permission

import (
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/userRoleRs"
	array "github.com/leancodebox/goose/collectionopt"
	"slices"
)

type Enum int

func (receiver Enum) Name() string {
	switch receiver {
	case Admin:
		return "管理员"
	case UserManager:
		return "用户管理"
	case ArticlesManager:
		return "文章管理"
	case PageManager:
		return "页面管理"
	case RoleManager:
		return "角色管理"
	}
	return ""
}
func (receiver Enum) Id() uint64 {
	return uint64(receiver)
}

const (
	Admin Enum = iota
	UserManager
	ArticlesManager
	PageManager
	RoleManager
)

func BuildOptions() []datastruct.Option[string, Enum] {
	var l []datastruct.Option[string, Enum]
	for i := Admin; i <= PageManager; i++ {
		l = append(l, datastruct.Option[string, Enum]{Name: i.Name(), Value: i})
	}
	return l
}

// CheckUser 检查某人是否有某权限
func CheckUser(userId uint64, permission Enum) bool {
	roleIds := userRoleRs.GetRoleIdsByUserId(userId)
	if roleIds == nil || len(roleIds) == 0 {
		return false
	}
	pList := GetPermission(roleIds)
	if pList == nil || len(pList) == 0 {
		return false
	}
	return slices.Contains(pList, permission) || slices.Contains(pList, Admin)
}

// GetPermission 获取权限
func GetPermission(roleIds []uint64) []Enum {
	return array.ArrayMap(func(t uint64) Enum {
		return Enum(t)
	}, rolePermissionRs.GetPermissionIdsByRoleIds(roleIds))
}
