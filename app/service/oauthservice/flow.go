package oauthservice

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/redirectopt"
	"github.com/leancodebox/GooseForum/app/bundles/sessionstore"
)

const oauthFlowSessionName = "gooseforum_oauth_flow"

type Flow struct {
	Provider string `json:"provider"`
	Mode     string `json:"mode"`
	UserID   uint64 `json:"userId,omitempty"`
	Redirect string `json:"redirect,omitempty"`
	IssuedAt int64  `json:"issuedAt"`
}

func StartFlow(res http.ResponseWriter, req *http.Request, provider string, userID uint64, requestedMode, redirect string) error {
	mode := "login"
	if requestedMode == "bind" {
		if userID == 0 {
			return errors.New("login is required to bind an OAuth account")
		}
		mode = "bind"
	}
	flow := Flow{
		Provider: provider,
		Mode:     mode,
		UserID:   userID,
		Redirect: safeRedirect(redirect),
		IssuedAt: time.Now().Unix(),
	}
	encoded, err := json.Marshal(flow)
	if err != nil {
		return err
	}
	session, err := sessionstore.GetSession().Get(req, oauthFlowSessionName)
	if err != nil {
		return err
	}
	session.Values[provider] = string(encoded)
	return session.Save(req, res)
}

func ConsumeFlow(res http.ResponseWriter, req *http.Request, provider string) (Flow, error) {
	session, err := sessionstore.GetSession().Get(req, oauthFlowSessionName)
	if err != nil {
		return Flow{}, err
	}
	raw, ok := session.Values[provider].(string)
	if !ok || raw == "" {
		return Flow{}, errors.New("OAuth flow context is missing or expired")
	}
	delete(session.Values, provider)
	if err := session.Save(req, res); err != nil {
		return Flow{}, err
	}
	var flow Flow
	if err := json.Unmarshal([]byte(raw), &flow); err != nil {
		return Flow{}, err
	}
	if flow.Provider != provider || time.Since(time.Unix(flow.IssuedAt, 0)) > 10*time.Minute {
		return Flow{}, errors.New("OAuth flow context is invalid or expired")
	}
	return flow, nil
}

func safeRedirect(value string) string {
	return redirectopt.Local(value)
}
