package httpnotifyservice

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func TestEndpointAcceptsSelectedEvent(t *testing.T) {
	endpoint := pageConfig.HttpNotifyEndpoint{Enabled: true, URL: "http://example.com/hook", Events: []string{"topic.published"}}

	if !endpointAccepts(endpoint, "topic.published") {
		t.Fatal("expected endpoint to accept selected event")
	}
	if endpointAccepts(endpoint, "comment.created") {
		t.Fatal("expected endpoint to reject unselected event")
	}
}

func TestShouldNotifyConfig(t *testing.T) {
	config := pageConfig.HttpNotifyConfig{Enabled: true, Endpoints: []pageConfig.HttpNotifyEndpoint{{
		Enabled: true,
		URL:     "http://example.com/hook",
		Events:  []string{"topic.published"},
	}}}

	if !shouldNotify(config, "topic.published") {
		t.Fatal("expected enabled matching endpoint to notify")
	}
	if shouldNotify(config, "comment.created") {
		t.Fatal("expected unmatched event to skip notification")
	}
	config.Enabled = false
	if shouldNotify(config, "topic.published") {
		t.Fatal("expected disabled config to skip notification")
	}
}

func TestBuildSignedRequest(t *testing.T) {
	body := []byte(`{"event":"topic.published"}`)
	req, err := buildRequest(pageConfig.HttpNotifyEndpoint{
		URL:    "http://example.com/hook",
		Secret: "secret",
	}, "topic.published", "delivery-1", 1710000000, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte("1710000000."))
	mac.Write(body)
	wantSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if req.Header.Get("X-Goose-Event") != "topic.published" {
		t.Fatalf("event header = %q", req.Header.Get("X-Goose-Event"))
	}
	if req.Header.Get("X-Goose-Delivery") != "delivery-1" {
		t.Fatalf("delivery header = %q", req.Header.Get("X-Goose-Delivery"))
	}
	if req.Header.Get("X-Goose-Signature") != wantSignature {
		t.Fatalf("signature = %q, want %q", req.Header.Get("X-Goose-Signature"), wantSignature)
	}
}

func TestApplyDeliveryFailureDisablesEndpointAfterThreeFailures(t *testing.T) {
	config := pageConfig.HttpNotifyConfig{Enabled: true, Endpoints: []pageConfig.HttpNotifyEndpoint{{
		Id:      "endpoint-1",
		Enabled: true,
		URL:     "http://example.com/hook",
	}}}

	config, changed := applyDeliveryResult(config, "endpoint-1", "http://example.com/hook", false, "timeout")
	if !changed || config.Endpoints[0].FailureCount != 1 || !config.Endpoints[0].Enabled {
		t.Fatalf("first failure not recorded correctly: %+v", config.Endpoints[0])
	}
	config, _ = applyDeliveryResult(config, "endpoint-1", "http://example.com/hook", false, "timeout")
	config, _ = applyDeliveryResult(config, "endpoint-1", "http://example.com/hook", false, "timeout")

	endpoint := config.Endpoints[0]
	if endpoint.Enabled {
		t.Fatal("expected endpoint disabled after three failures")
	}
	if !endpoint.AbnormalTerminated {
		t.Fatal("expected endpoint marked abnormal terminated")
	}
	if endpoint.LastError != "timeout" {
		t.Fatalf("last error = %q", endpoint.LastError)
	}
}

func TestApplyDeliverySuccessResetsFailureCount(t *testing.T) {
	config := pageConfig.HttpNotifyConfig{Endpoints: []pageConfig.HttpNotifyEndpoint{{
		Id:           "endpoint-1",
		Enabled:      true,
		URL:          "http://example.com/hook",
		FailureCount: 2,
		LastError:    "timeout",
	}}}

	config, changed := applyDeliveryResult(config, "endpoint-1", "http://example.com/hook", true, "")
	if !changed {
		t.Fatal("expected successful delivery to update endpoint")
	}
	endpoint := config.Endpoints[0]
	if endpoint.FailureCount != 0 || endpoint.LastError != "" || endpoint.AbnormalTerminated {
		t.Fatalf("success did not reset endpoint failure state: %+v", endpoint)
	}
}

func TestValidateConfigAllowsLocalAndPrivateEndpoints(t *testing.T) {
	for index, endpointURL := range []string{
		"http://localhost:8080/hook",
		"http://127.0.0.1/hook",
		"http://192.168.1.20/hook",
		"https://hooks.example.com/events",
	} {
		config := pageConfig.HttpNotifyConfig{Enabled: true, Endpoints: []pageConfig.HttpNotifyEndpoint{{
			Id: fmt.Sprintf("endpoint-%d", index), Enabled: true, URL: endpointURL,
			Events: []string{EventTopicPublished}, TimeoutSeconds: 2,
		}}}
		if err := ValidateConfig(config); err != nil {
			t.Fatalf("ValidateConfig(%q): %v", endpointURL, err)
		}
	}
}

func TestValidateConfigRejectsInvalidActiveEndpoints(t *testing.T) {
	valid := pageConfig.HttpNotifyEndpoint{
		Id: "endpoint-1", Enabled: true, URL: "https://hooks.example.com/events",
		Events: []string{EventTopicPublished}, TimeoutSeconds: 2,
	}
	tests := []struct {
		name      string
		endpoints []pageConfig.HttpNotifyEndpoint
	}{
		{name: "missing id", endpoints: []pageConfig.HttpNotifyEndpoint{{Enabled: true, URL: valid.URL, Events: valid.Events, TimeoutSeconds: 2}}},
		{name: "duplicate id", endpoints: []pageConfig.HttpNotifyEndpoint{valid, valid}},
		{name: "missing host", endpoints: []pageConfig.HttpNotifyEndpoint{{Id: "endpoint-1", Enabled: true, URL: "http:///hook", Events: valid.Events, TimeoutSeconds: 2}}},
		{name: "unsupported event", endpoints: []pageConfig.HttpNotifyEndpoint{{Id: "endpoint-1", Enabled: true, URL: valid.URL, Events: []string{"unknown"}, TimeoutSeconds: 2}}},
		{name: "invalid timeout", endpoints: []pageConfig.HttpNotifyEndpoint{{Id: "endpoint-1", Enabled: true, URL: valid.URL, Events: valid.Events, TimeoutSeconds: 30}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateConfig(pageConfig.HttpNotifyConfig{Enabled: true, Endpoints: test.endpoints}); err == nil {
				t.Fatal("invalid configuration was accepted")
			}
		})
	}
}

func TestValidateConfigAllowsDisablingBrokenEndpoint(t *testing.T) {
	config := pageConfig.HttpNotifyConfig{Enabled: false, Endpoints: []pageConfig.HttpNotifyEndpoint{{
		Id: "endpoint-1", Enabled: true, URL: "broken", Events: []string{"unknown"}, TimeoutSeconds: 30,
	}}}
	if err := ValidateConfig(config); err != nil {
		t.Fatalf("disabled configuration should remain saveable: %v", err)
	}
}
