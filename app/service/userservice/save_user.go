package userservice

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/http/controllers/transform"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/usercardservice"
)

func SaveUser(userEntity *users.EntityComplete) error {
	if err := users.Save(userEntity); err != nil {
		return err
	}
	InvalidateUserCaches(userEntity)
	return nil
}

func InvalidateUserCaches(userEntity *users.EntityComplete) {
	if userEntity == nil || userEntity.Id == 0 {
		return
	}
	users.InvalidateRoleCache(userEntity.Id)
	usercardservice.Invalidate(userEntity.Id)
	if cacheErr := hotdataserve.Reload(hotdataserve.UserShowCacheKey(userEntity.Id), transform.User2userShow(*userEntity)); cacheErr != nil {
		slog.Error("reload user cache failed", "userId", userEntity.Id, "err", cacheErr)
	}
}
