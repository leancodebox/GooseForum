package rolemanagementservice

import (
	"errors"
	"maps"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/adminpolicyservice"
	"github.com/leancodebox/GooseForum/app/service/authsessionservice"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

var ErrProtected = errors.New("protected role")
var ErrInvalid = errors.New("invalid role")

type Change struct {
	Id          uint64
	Name        string
	Permissions []uint64
	Delete      bool
}

func Save(actorID uint64, id uint64, name string, permissions []uint64) error {
	return change(actorID, Change{Id: id, Name: name, Permissions: permissions})
}
func Delete(actorID, id uint64) error { return change(actorID, Change{Id: id, Delete: true}) }

func change(actorID uint64, change Change) error {
	id, members, err := apply(actorID, change, time.Now())
	permission.InvalidateRole(id)
	for _, member := range members {
		authsessionservice.InvalidateUser(member)
		userservice.InvalidateUserCaches(member)
	}
	return err
}

func apply(actorID uint64, change Change, now time.Time) (uint64, []uint64, error) {
	change.Name = strings.TrimSpace(change.Name)
	next := map[uint64]bool{}
	if change.Delete {
		if change.Id == 0 {
			return 0, nil, ErrInvalid
		}
	} else {
		if change.Name == "" || utf8.RuneCountInString(change.Name) > 255 || len(change.Permissions) == 0 || len(change.Permissions) > 100 {
			return 0, nil, ErrInvalid
		}
		for _, id := range change.Permissions {
			if id > permission.SiteManager.Id() {
				return 0, nil, ErrInvalid
			}
			next[id] = true
		}
	}
	var memberIDs []uint64
	roleID := change.Id
	apply := func() error {
		snapshot, err := adminpolicyservice.ReadSnapshot()
		if err != nil {
			return err
		}
		actor, err := users.GetAccountState(actorID)
		if err != nil {
			return ErrProtected
		}
		permissions, err := adminpolicyservice.Permissions(actor.RoleId, change.Id)
		if err != nil {
			return err
		}
		owned := permissions[actor.RoleId]
		actorAdmin := owned[permission.Admin.Id()]
		if actor.EffectiveRestriction(now) != users.RestrictionNormal || !actorAdmin && !owned[permission.RoleManager.Id()] {
			return ErrProtected
		}
		current := permissions[change.Id]
		if !actorAdmin {
			for _, set := range []map[uint64]bool{current, next} {
				for id := range set {
					if !owned[id] {
						return ErrProtected
					}
				}
			}
		}
		var entity role.Entity
		if change.Id != 0 {
			entity, err = role.GetForPolicy(change.Id)
			if err != nil {
				return err
			}
		} else {
			entity.Effective = 1
		}
		if current[permission.Admin.Id()] && !next[permission.Admin.Id()] {
			remaining, err := users.HasNormalAdministrator(snapshot.AdminRoleIDs, 0, change.Id, now, snapshot.EmailVerificationEnabled)
			if err != nil {
				return err
			}
			if !remaining {
				return ErrProtected
			}
		}
		if change.Id != 0 && (change.Delete || !maps.Equal(current, next)) {
			memberIDs, err = users.RoleMemberIDs(change.Id)
			if err != nil {
				return err
			}
		}
		if change.Delete {
			if err := rolePermissionRs.DeletePolicyGrants(change.Id); err != nil {
				return err
			}
			if err := role.DeleteForPolicy(&entity); err != nil {
				return err
			}
		} else {
			entity.RoleName = change.Name
			if err := role.SaveForPolicy(&entity); err != nil {
				return err
			}
			roleID = entity.Id
			if !maps.Equal(current, next) {
				ids := make([]uint64, 0, len(next))
				for id := range next {
					ids = append(ids, id)
				}
				if err := rolePermissionRs.ReplacePolicyGrants(entity.Id, ids); err != nil {
					return err
				}
			}
		}
		if change.Delete || !maps.Equal(current, next) {
			if len(memberIDs) > 0 {
				if err := users.RevokeRoleMembers(change.Id, change.Delete); err != nil {
					return err
				}
			}
		}
		return nil
	}
	err := apply()
	return roleID, memberIDs, err
}
