package httpnotifyservice

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

const (
	EventTopicPublished   = "topic.published"
	EventTopicUpdated     = "topic.updated"
	EventCommentCreated   = "comment.created"
	EventUserSignup       = "user.signup"
	EventReportCreated    = "moderation.report.created"
	defaultTimeoutSeconds = 2
	maxTimeoutSeconds     = 15
	contentTypeJSON       = "application/json"
	disableAfterFailures  = 3
	maxConfigEndpoints    = 20
	maxStatusUpdateTries  = 3
)

var updateMu sync.Mutex

var supportedEvents = map[string]struct{}{
	EventTopicPublished: {},
	EventTopicUpdated:   {},
	EventCommentCreated: {},
	EventUserSignup:     {},
	EventReportCreated:  {},
}

var sendRequest = func(req *http.Request, timeout time.Duration) (*http.Response, error) {
	return (&http.Client{Timeout: timeout}).Do(req)
}

type Envelope struct {
	Event     string `json:"event"`
	Timestamp int64  `json:"timestamp"`
	Data      any    `json:"data"`
}

func Notify(ctx context.Context, eventName string, data any) {
	config := hotdataserve.GetHttpNotifyConfigCache()
	if !shouldNotify(config, eventName) {
		return
	}
	now := time.Now().Unix()
	body, err := json.Marshal(Envelope{
		Event:     eventName,
		Timestamp: now,
		Data:      data,
	})
	if err != nil {
		slog.Error("httpnotify: marshal payload failed", "event", eventName, "err", err)
		return
	}
	for _, endpoint := range config.Endpoints {
		if ctx.Err() != nil {
			return
		}
		if !endpointAccepts(endpoint, eventName) {
			continue
		}
		deliver(ctx, endpoint, eventName, now, body)
	}
}

func ShouldNotify(eventName string) bool {
	return shouldNotify(hotdataserve.GetHttpNotifyConfigCache(), eventName)
}

func shouldNotify(config pageConfig.HttpNotifyConfig, eventName string) bool {
	if !config.Enabled {
		return false
	}
	for _, endpoint := range config.Endpoints {
		if endpointAccepts(endpoint, eventName) {
			return true
		}
	}
	return false
}

func endpointAccepts(endpoint pageConfig.HttpNotifyEndpoint, eventName string) bool {
	if !endpoint.Enabled || endpoint.AbnormalTerminated || strings.TrimSpace(endpoint.URL) == "" {
		return false
	}
	return slices.Contains(endpoint.Events, eventName)
}

func deliver(ctx context.Context, endpoint pageConfig.HttpNotifyEndpoint, eventName string, timestamp int64, body []byte) {
	req, err := buildRequest(endpoint, eventName, deliveryID(), timestamp, body)
	if err != nil {
		slog.Error("httpnotify: build request failed", "endpoint", endpoint.Name, "event", eventName, "err", err)
		recordDeliveryResult(endpoint, false, err.Error())
		return
	}
	req = req.WithContext(ctx)
	resp, err := sendRequest(req, endpointTimeout(endpoint))
	if err != nil {
		if ctx.Err() != nil {
			slog.Debug("httpnotify: delivery canceled with event context", "endpoint", endpoint.Name, "event", eventName)
			return
		}
		slog.Error("httpnotify: request failed", "endpoint", endpoint.Name, "event", eventName, "err", err)
		recordDeliveryResult(endpoint, false, err.Error())
		return
	}
	statusCode, status := resp.StatusCode, resp.Status
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
	_ = resp.Body.Close()
	if statusCode < 200 || statusCode >= 300 {
		slog.Warn("httpnotify: non-2xx response", "endpoint", endpoint.Name, "event", eventName, "status", statusCode)
		recordDeliveryResult(endpoint, false, status)
		return
	}
	recordDeliveryResult(endpoint, true, "")
}

func buildRequest(endpoint pageConfig.HttpNotifyEndpoint, eventName string, deliveryID string, timestamp int64, body []byte) (*http.Request, error) {
	targetURL, err := url.Parse(strings.TrimSpace(endpoint.URL))
	if err != nil {
		return nil, err
	}
	if targetURL.Scheme != "http" && targetURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported url scheme: %s", targetURL.Scheme)
	}
	req, err := http.NewRequest(http.MethodPost, targetURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("X-Goose-Event", eventName)
	req.Header.Set("X-Goose-Delivery", deliveryID)
	req.Header.Set("X-Goose-Timestamp", strconv.FormatInt(timestamp, 10))
	if endpoint.Secret != "" {
		req.Header.Set("X-Goose-Signature", sign(endpoint.Secret, timestamp, body))
	}
	return req, nil
}

func sign(secret string, timestamp int64, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte("."))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func endpointTimeout(endpoint pageConfig.HttpNotifyEndpoint) time.Duration {
	seconds := endpoint.TimeoutSeconds
	if seconds <= 0 {
		seconds = defaultTimeoutSeconds
	}
	if seconds > maxTimeoutSeconds {
		seconds = maxTimeoutSeconds
	}
	return time.Duration(seconds) * time.Second
}

func ValidateConfig(config pageConfig.HttpNotifyConfig) error {
	if len(config.Endpoints) > maxConfigEndpoints {
		return fmt.Errorf("HTTP notification endpoints exceed maximum limit of %d", maxConfigEndpoints)
	}
	ids := make(map[string]struct{}, len(config.Endpoints))
	active := 0
	for index, endpoint := range config.Endpoints {
		id := strings.TrimSpace(endpoint.Id)
		if id == "" || len(id) > 128 {
			return fmt.Errorf("HTTP notification endpoint %d has an invalid id", index+1)
		}
		if _, exists := ids[id]; exists {
			return fmt.Errorf("HTTP notification endpoint id %q is duplicated", id)
		}
		ids[id] = struct{}{}
		if len(endpoint.Name) > 128 || len(endpoint.Secret) > 4096 || len(endpoint.URL) > 2048 {
			return fmt.Errorf("HTTP notification endpoint %q exceeds field length limits", id)
		}
		if !config.Enabled || !endpoint.Enabled {
			continue
		}
		active++
		target, err := url.Parse(strings.TrimSpace(endpoint.URL))
		if err != nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") {
			return fmt.Errorf("HTTP notification endpoint %q has an invalid URL", id)
		}
		if endpoint.TimeoutSeconds < 1 || endpoint.TimeoutSeconds > maxTimeoutSeconds {
			return fmt.Errorf("HTTP notification endpoint %q has an invalid timeout", id)
		}
		if len(endpoint.Events) == 0 {
			return fmt.Errorf("HTTP notification endpoint %q has no events", id)
		}
		seenEvents := make(map[string]struct{}, len(endpoint.Events))
		for _, eventName := range endpoint.Events {
			if _, ok := supportedEvents[eventName]; !ok {
				return fmt.Errorf("HTTP notification endpoint %q has unsupported event %q", id, eventName)
			}
			if _, exists := seenEvents[eventName]; exists {
				return fmt.Errorf("HTTP notification endpoint %q repeats event %q", id, eventName)
			}
			seenEvents[eventName] = struct{}{}
		}
	}
	if config.Enabled && active == 0 {
		return errors.New("HTTP notification requires at least one enabled endpoint")
	}
	return nil
}

func deliveryID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b[:])
}

func recordDeliveryResult(endpoint pageConfig.HttpNotifyEndpoint, success bool, message string) {
	updateMu.Lock()
	defer updateMu.Unlock()

	for range maxStatusUpdateTries {
		entity := pageConfig.GetByPageType(pageConfig.HttpNotify)
		if entity.Id == 0 {
			return
		}
		config := jsonopt.Decode[pageConfig.HttpNotifyConfig](entity.Config)
		config, changed := applyDeliveryResult(config, endpoint.Id, endpoint.URL, success, message)
		if !changed {
			return
		}
		saved, err := pageConfig.CompareAndSwapConfig(entity, jsonopt.Encode(config))
		if err != nil {
			slog.Error("httpnotify: record delivery result failed", "endpoint", endpoint.Name, "err", err)
			return
		}
		if saved {
			hotdataserve.ClearHttpNotifyConfigCache()
			return
		}
	}
	slog.Warn("httpnotify: delivery result dropped after concurrent configuration updates", "endpoint", endpoint.Name)
}

func applyDeliveryResult(config pageConfig.HttpNotifyConfig, endpointId string, endpointURL string, success bool, message string) (pageConfig.HttpNotifyConfig, bool) {
	for i := range config.Endpoints {
		endpoint := &config.Endpoints[i]
		if endpoint.Id != "" && endpointId != "" {
			if endpoint.Id != endpointId {
				continue
			}
		} else if endpoint.URL != endpointURL {
			continue
		}
		if success {
			if endpoint.FailureCount == 0 && endpoint.LastError == "" && !endpoint.AbnormalTerminated {
				return config, false
			}
			endpoint.FailureCount = 0
			endpoint.LastError = ""
			endpoint.AbnormalTerminated = false
			return config, true
		}
		endpoint.FailureCount++
		endpoint.LastError = message
		if endpoint.FailureCount >= disableAfterFailures {
			endpoint.Enabled = false
			endpoint.AbnormalTerminated = true
		}
		return config, true
	}
	return config, false
}
