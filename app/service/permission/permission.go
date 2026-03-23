package permission

import (
	"fmt"
	"slices"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/datacache"
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/samber/lo"
)

var rolePermissionCache = datacache.Cache[[]Enum]{}

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
	case SiteManager:
		return "站点管理"
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
	SiteManager
)

func BuildOptions() []datastruct.Option[string, Enum] {
	return lo.Map(lo.RangeFrom(int(Admin), int(SiteManager-Admin+1)), func(i int, _ int) datastruct.Option[string, Enum] {
		item := Enum(i)
		return datastruct.Option[string, Enum]{Name: item.Name(), Label: item.Name(), Value: item}
	})
}

// GetPermissionByRoleId 获取单个角色的权限（带缓存）
func GetPermissionByRoleId(roleId uint64) []Enum {
	key := fmt.Sprintf("role_permission:%d", roleId)
	return rolePermissionCache.GetOrLoad(key, func() ([]Enum, error) {
		// 从数据库加载
		pIds := rolePermissionRs.GetPermissionIdsByRoleIds([]uint64{roleId})
		return lo.Map(pIds, func(t uint64, _ int) Enum {
			return Enum(t)
		}), nil
	}, 10*time.Minute) // 缓存 10 分钟
}

// CheckRole 检查某人是否有某权限
func CheckRole(roleId uint64, permission Enum) bool {
	pList := GetPermissionByRoleId(roleId)
	if len(pList) == 0 {
		return false
	}
	return slices.Contains(pList, permission) || slices.Contains(pList, Admin)
}

// GetPermission 获取权限 (兼容旧接口)
func GetPermission(roleIds []uint64) []Enum {
	var result []Enum
	for _, rid := range roleIds {
		result = append(result, GetPermissionByRoleId(rid)...)
	}
	return lo.Uniq(result)
}
