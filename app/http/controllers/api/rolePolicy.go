package api

import (
	"errors"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/rolemanagementservice"
	"gorm.io/gorm"
)

func roleMutationError(err error) component.Response {
	if errors.Is(err, rolemanagementservice.ErrProtected) {
		return component.FailResponseCode(component.MessageAdminRestrictionProtected, nil)
	}
	if errors.Is(err, rolemanagementservice.ErrInvalid) {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return component.FailResponseCode(component.MessageAdminRoleNotFound, nil)
	}
	return component.FailResponseCode(component.MessageOperationFailed, nil)
}
