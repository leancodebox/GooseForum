package api

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/oidcproviderservice"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
)

type OIDCAdminGrantReq struct {
	ClientID string `form:"clientId" json:"clientId"`
	UserID   string `form:"userId" json:"userId"`
	After    string `form:"after" json:"after"`
}

type OIDCAdminGrantView struct {
	oidcProviderStore.AdminGrant
	Username string `json:"username"`
}

func ListOIDCClientGrants(req component.BetterRequest[OIDCAdminGrantReq]) component.Response {
	store, err := newOIDCClientStore()
	if err != nil {
		return oidcClientFailure("initialize store", err)
	}
	if req.Params.ClientID == "" {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	rows, more, err := store.ListClientGrants(requestContext(req.GinContext), req.Params.ClientID, req.Params.After, req.Params.UserID)
	if err != nil {
		return oidcClientFailure("list client grants", err)
	}
	ids := make([]uint64, 0, len(rows))
	for _, row := range rows {
		if id, err := strconv.ParseUint(row.UserID, 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	identities, err := users.MentionIdentities(ids, nil)
	if err != nil {
		return oidcClientFailure("resolve grant users", err)
	}
	names := make(map[string]string, len(identities))
	for _, identity := range identities {
		names[strconv.FormatUint(identity.Id, 10)] = identity.Username
	}
	views := make([]OIDCAdminGrantView, 0, len(rows))
	for _, row := range rows {
		views = append(views, OIDCAdminGrantView{AdminGrant: row, Username: names[row.UserID]})
	}
	return component.SuccessResponse(map[string]any{"items": views, "hasMore": more})
}

func RevokeOIDCClientGrant(req component.BetterRequest[OIDCAdminGrantReq]) component.Response {
	store, err := newOIDCClientStore()
	if err != nil {
		return oidcClientFailure("initialize store", err)
	}
	if req.Params.ClientID == "" {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	if req.Params.UserID == "" {
		err = store.RevokeClientGrants(requestContext(req.GinContext), req.Params.ClientID, time.Now())
	} else {
		if _, e := strconv.ParseUint(req.Params.UserID, 10, 64); e != nil {
			return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
		}
		err = store.RevokeGrant(requestContext(req.GinContext), req.Params.UserID, req.Params.ClientID, time.Now())
	}
	if err != nil {
		return oidcClientFailure("revoke client grant", err)
	}
	slog.Info("admin OIDC grant revoked", "administrator", req.UserId, "client", req.Params.ClientID, "user", req.Params.UserID)
	optlogger.UserOpt(req.UserId, optlogger.ManageOIDC, req.Params.ClientID, "Revoke OIDC grant: user="+req.Params.UserID)
	return component.SuccessResponse(true)
}

func DeleteOIDCClient(req component.BetterRequest[RotateOIDCClientSecretReq]) component.Response {
	if req.Params.ClientID == "" {
		return component.FailResponseCode(component.MessageRequestInvalidParams, nil)
	}
	store, err := newOIDCClientStore()
	if err != nil {
		return oidcClientFailure("initialize store", err)
	}
	if err := store.DeleteClient(requestContext(req.GinContext), req.Params.ClientID, time.Now()); err != nil {
		return oidcClientFailure("delete client", err)
	}
	slog.Info("admin OIDC client deleted", "administrator", req.UserId, "client", req.Params.ClientID)
	optlogger.UserOpt(req.UserId, optlogger.ManageOIDC, req.Params.ClientID, "Delete OIDC client")
	return component.SuccessResponse(true)
}

func ResetOIDCSigningKey(req component.BetterRequest[component.Null]) component.Response {
	if err := oidcproviderservice.ResetDefaultSigningKey(requestContext(req.GinContext)); err != nil {
		return oidcClientFailure("reset signing key", err)
	}
	slog.Info("admin OIDC signing key reset", "administrator", req.UserId)
	optlogger.UserOpt(req.UserId, optlogger.ManageOIDC, "oidc", "Reset unrecoverable OIDC signing key")
	return component.SuccessResponse(oidcproviderservice.Status())
}
