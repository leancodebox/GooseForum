package permission

import (
	"fmt"
	"slices"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/samber/lo"
)

var rolePermissionCache = localcache.Cache[[]Enum]{MaxEntries: 256}

type Enum int

// i18nKey returns the stable translation key for the permission label.
func (receiver Enum) i18nKey() string {
	switch receiver {
	case Admin:
		return "permission.admin"
	case UserManager:
		return "permission.userManager"
	case TopicsManager:
		return "permission.articlesManager"
	case PageManager:
		return "permission.pageManager"
	case RoleManager:
		return "permission.roleManager"
	case SiteManager:
		return "permission.siteManager"
	}
	return ""
}

// Name returns the permission label in the fallback locale (zh). Prefer
// LocalizedName when a request locale is available.
func (receiver Enum) Name() string {
	return receiver.LocalizedName(i18n.Fallback)
}

// LocalizedName returns the permission label translated into lang.
func (receiver Enum) LocalizedName(lang string) string {
	key := receiver.i18nKey()
	if key == "" {
		return ""
	}
	return i18n.T(lang, key)
}

func (receiver Enum) Id() uint64 {
	return uint64(receiver)
}

const (
	Admin Enum = iota
	UserManager
	TopicsManager
	PageManager
	RoleManager
	SiteManager
)

func BuildOptions(lang string) []datastruct.Option[string, Enum] {
	return lo.Map(lo.RangeFrom(int(Admin), int(SiteManager-Admin+1)), func(i int, _ int) datastruct.Option[string, Enum] {
		item := Enum(i)
		name := item.LocalizedName(lang)
		return datastruct.Option[string, Enum]{Name: name, Label: name, Value: item}
	})
}

func All() []Enum {
	return lo.Map(lo.RangeFrom(int(Admin), int(SiteManager-Admin+1)), func(i int, _ int) Enum {
		return Enum(i)
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

func InvalidateRole(roleId uint64) {
	rolePermissionCache.Delete(fmt.Sprintf("role_permission:%d", roleId))
}

// CheckRole 检查某人是否有某权限
func CheckRole(roleId uint64, permission Enum) bool {
	pList := GetPermissionByRoleId(roleId)
	if len(pList) == 0 {
		return false
	}
	return slices.Contains(pList, permission) || slices.Contains(pList, Admin)
}

func CheckAnyRole(roleId uint64) bool {
	all := All()
	for _, item := range GetPermissionByRoleId(roleId) {
		if slices.Contains(all, item) {
			return true
		}
	}
	return false
}
