package permission

import "slices"

type Enum int

func (receiver Enum) name() string {
	switch receiver {
	case Admin:
		return "管理员"
	case UserManager:
		return "用户管理"
	case ArticlesManager:
		return "文章管理"
	case PageManager:
		return "页面管理"
	}
	return ""
}

const (
	Admin Enum = iota
	UserManager
	ArticlesManager
	PageManager
)

// CheckUser 检查某人是否有某权限
func CheckUser(userId uint64, permission Enum) bool {
	// todo  getRoleByUserId userId 通过用户获取角色
	// todo  getPermission by Role 通过角色获取权限点
	pList := []Enum{}
	return slices.Contains(pList, permission) || slices.Contains(pList, Admin) || true
}
