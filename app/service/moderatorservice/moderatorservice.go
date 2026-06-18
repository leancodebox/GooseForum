package moderatorservice

import (
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/moderators"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

const snapshotTTL = time.Minute

type Grant struct {
	UserID    uint64
	ScopeType string
	ScopeID   uint64
}

type Snapshot struct {
	Grants    []Grant
	ExpiresAt time.Time
}

var (
	snapshotValue atomic.Value
	refreshMu     sync.Mutex
)

func IsAdmin(userID uint64) bool {
	if userID == 0 {
		return false
	}
	roleID, ok := userservice.GetUserRoleId(userID)
	return ok && permission.CheckRole(roleID, permission.Admin)
}

func CanAccessModeration(userID uint64) bool {
	return IsAdmin(userID) || IsModerator(userID)
}

func IsModerator(userID uint64) bool {
	if userID == 0 {
		return false
	}
	for _, grant := range snapshot().Grants {
		if grant.UserID == userID {
			return true
		}
	}
	return false
}

func CanModerateAnyCategory(userID uint64, categoryIDs []uint64) bool {
	if IsAdmin(userID) {
		return true
	}
	if userID == 0 {
		return false
	}
	grants := snapshot().Grants
	for _, grant := range grants {
		if grant.UserID != userID {
			continue
		}
		if grant.ScopeType == moderators.ScopeGlobal {
			return true
		}
		if grant.ScopeType != moderators.ScopeCategory {
			continue
		}
		if slices.Contains(categoryIDs, grant.ScopeID) {
			return true
		}
	}
	return false
}

func ScopeForUser(userID uint64) (bool, []uint64) {
	if IsAdmin(userID) {
		return true, nil
	}
	categoryIDs := make([]uint64, 0)
	global := false
	for _, grant := range snapshot().Grants {
		if grant.UserID != userID {
			continue
		}
		if grant.ScopeType == moderators.ScopeGlobal {
			global = true
			continue
		}
		if grant.ScopeType == moderators.ScopeCategory {
			categoryIDs = append(categoryIDs, grant.ScopeID)
		}
	}
	return global, uniqueUint64(categoryIDs)
}

func Invalidate() {
	snapshotValue.Store(Snapshot{})
}

func snapshot() Snapshot {
	if value, ok := snapshotValue.Load().(Snapshot); ok && time.Now().Before(value.ExpiresAt) {
		return value
	}
	refreshMu.Lock()
	defer refreshMu.Unlock()
	if value, ok := snapshotValue.Load().(Snapshot); ok && time.Now().Before(value.ExpiresAt) {
		return value
	}
	next := loadSnapshot()
	snapshotValue.Store(next)
	return next
}

func loadSnapshot() Snapshot {
	rows := moderators.AllEnabled()
	grants := make([]Grant, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.UserId == 0 {
			continue
		}
		grants = append(grants, Grant{
			UserID:    row.UserId,
			ScopeType: row.ScopeType,
			ScopeID:   row.ScopeId,
		})
	}
	return Snapshot{
		Grants:    grants,
		ExpiresAt: time.Now().Add(snapshotTTL),
	}
}

func uniqueUint64(values []uint64) []uint64 {
	res := make([]uint64, 0, len(values))
	for _, value := range values {
		seen := slices.Contains(res, value)
		if !seen {
			res = append(res, value)
		}
	}
	return res
}
