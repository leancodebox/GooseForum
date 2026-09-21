package nextfrontend

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	appresource "github.com/leancodebox/GooseForum/resource"
)

func TestValidateNodeVersion(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{version: "v20.9.0", valid: true},
		{version: "v22.0.0", valid: true},
		{version: "20.8.1", valid: false},
		{version: "v18.20.0", valid: false},
		{version: "unknown", valid: false},
	}
	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			err := validateNodeVersion(test.version)
			if test.valid && err != nil {
				t.Fatalf("validateNodeVersion(%q) error = %v", test.version, err)
			}
			if !test.valid && err == nil {
				t.Fatalf("validateNodeVersion(%q) unexpectedly succeeded", test.version)
			}
		})
	}
}

func TestIsLoopbackHost(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1", "localhost"} {
		if !isLoopbackHost(host) {
			t.Fatalf("isLoopbackHost(%q) = false", host)
		}
	}
	for _, host := range []string{"", "0.0.0.0", "192.0.2.1", "example.com"} {
		if isLoopbackHost(host) {
			t.Fatalf("isLoopbackHost(%q) = true", host)
		}
	}
}

func TestIsNextRoute(t *testing.T) {
	tests := []struct {
		method string
		path   string
		header bool
		want   bool
	}{
		{method: http.MethodGet, path: "/", want: true},
		{method: http.MethodGet, path: "/p/post/1", want: true},
		{method: http.MethodGet, path: "/_next/static/app.js", want: true},
		{method: http.MethodGet, path: healthPath, want: false},
		{method: http.MethodGet, path: "/goose-page-data", want: true},
		{method: http.MethodGet, path: "/api/forum/topics", want: false},
		{method: http.MethodGet, path: "/file/img/a.webp", want: false},
		{method: http.MethodGet, path: "/oauth2/authorize", want: false},
		{method: http.MethodGet, path: "/admin/users", want: false},
		{method: http.MethodGet, path: "/activate", want: false},
		{method: http.MethodGet, path: "/unknown", want: false},
		{method: http.MethodGet, path: "/robots.txt", want: false},
		{method: http.MethodPost, path: "/", want: false},
		{method: http.MethodGet, path: "/", header: true, want: false},
	}
	for _, test := range tests {
		request, err := http.NewRequest(test.method, "http://example.com"+test.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if test.header {
			request.Header.Set("X-Goose-Page", "true")
		}
		if got := isNextRoute(request); got != test.want {
			t.Errorf("isNextRoute(%s %s, header=%v) = %v, want %v", test.method, test.path, test.header, got, test.want)
		}
	}
}

func TestFrontendProxy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Frontend", "next")
		_, _ = writer.Write([]byte(request.URL.RequestURI()))
	}))
	defer upstream.Close()
	target, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	frontend := &Frontend{proxy: newTestProxy(target)}
	frontend.available.Store(true)
	request := httptest.NewRequest(http.MethodGet, "http://forum.example/p/post/1?page=2", nil)
	response := httptest.NewRecorder()
	if !frontend.Proxy(response, request) {
		t.Fatal("Proxy() did not handle a public page")
	}
	if response.Code != http.StatusOK || response.Header().Get("X-Frontend") != "next" || response.Body.String() != "/p/post/1?page=2" {
		t.Fatalf("unexpected proxy response: code=%d header=%q body=%q", response.Code, response.Header().Get("X-Frontend"), response.Body.String())
	}
}

func TestFrontendProxyErrorDoesNotDisableNext(t *testing.T) {
	target, err := url.Parse("http://127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	proxy := newFrontendProxy(target)
	proxy.Transport = roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, context.Canceled
	})
	frontend := &Frontend{proxy: proxy}
	frontend.available.Store(true)
	request := httptest.NewRequest(http.MethodGet, "http://forum.example/", nil)
	response := httptest.NewRecorder()

	if !frontend.Proxy(response, request) {
		t.Fatal("Proxy() did not handle a public page")
	}
	if response.Code != http.StatusBadGateway {
		t.Fatalf("proxy status = %d, want %d", response.Code, http.StatusBadGateway)
	}
	if !frontend.available.Load() {
		t.Fatal("one canceled request disabled Next globally")
	}
}

func TestNextHealthReadyRequiresLaunchToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set(healthTokenHeader, "other-process")
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client := server.Client()

	if nextHealthReady(client, server.URL, "launch-token") {
		t.Fatal("accepted a listener with the wrong launch token")
	}
	if !nextHealthReady(client, server.URL, "other-process") {
		t.Fatal("rejected the listener with the matching launch token")
	}
}

func TestNewHealthToken(t *testing.T) {
	first, err := newHealthToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newHealthToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 64 || len(second) != 64 || first == second {
		t.Fatalf("unexpected health tokens: first=%q second=%q", first, second)
	}
}

func TestFrontendStartFallsBackWhenEntryIsMissing(t *testing.T) {
	frontend := &Frontend{config: processConfig{
		Enabled:        true,
		NodeCommand:    defaultNodeCommand,
		Entry:          t.TempDir() + "/missing.mjs",
		Host:           defaultHost,
		StartupTimeout: time.Second,
		GoOrigin:       "http://127.0.0.1:5234",
	}}
	frontend.Start()
	if frontend.available.Load() || frontend.process != nil {
		t.Fatal("Next.js should remain unavailable when the standalone entry is missing")
	}
}

func TestExtractArchive(t *testing.T) {
	var data bytes.Buffer
	archive := zip.NewWriter(&data)
	header := &zip.FileHeader{Name: "apps/next/start.mjs", Method: zip.Deflate}
	header.SetMode(0o644)
	file, err := archive.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("export {};")); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err := extractArchive(data.Bytes(), directory); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(directory, "apps", "next", "start.mjs"))
	if err != nil || string(content) != "export {};" {
		t.Fatalf("extracted content %q, err=%v", content, err)
	}
}

func TestExtractArchiveRejectsTraversal(t *testing.T) {
	var data bytes.Buffer
	archive := zip.NewWriter(&data)
	file, err := archive.Create("../escape.mjs")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = file.Write([]byte("bad"))
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := extractArchive(data.Bytes(), t.TempDir()); err == nil {
		t.Fatal("path traversal archive was accepted")
	}
}

func TestPrepareEmbeddedRuntime(t *testing.T) {
	if _, embedded, err := appresource.GetNextStandaloneArchive(); err != nil {
		t.Fatal(err)
	} else if !embedded {
		t.Skip("Next.js bundle is not embedded in this build")
	}
	directory, entry, err := prepareRuntime(defaultEntry)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(directory)
	if directory == "" || !strings.HasPrefix(entry, directory+string(filepath.Separator)) {
		t.Fatalf("runtime directory=%q entry=%q", directory, entry)
	}
	if _, err := os.Stat(entry); err != nil {
		t.Fatalf("embedded entry missing: %v", err)
	}
}

func TestStartEmbeddedRuntimeAndCleanup(t *testing.T) {
	if _, embedded, err := appresource.GetNextStandaloneArchive(); err != nil {
		t.Fatal(err)
	} else if !embedded {
		t.Skip("Next.js bundle is not embedded in this build")
	}
	if _, err := exec.LookPath(defaultNodeCommand); err != nil {
		t.Skip("Node.js is not available")
	}
	backend := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Goose-Page") != "true" {
			t.Errorf("missing internal page header")
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"component":"home.index","props":{"sort":"latest","tabs":[],"topics":[],"pagination":{"page":1,"nextPage":2,"hasNext":false,"nextUrl":""},"announcement":{"enabled":true,"html":"<p>Embedded Next integration</p>"}},"layout":{"site":{"name":"GooseForum","description":"","logo":"","favicon":"","brandType":"default","brandText":"","brandImage":""},"viewer":{"id":0,"username":"","email":"","avatarUrl":"","isAuthenticated":false,"canAccessAdmin":false,"isModerator":false,"requiresEmailVerification":false,"adminPermissions":[]},"header":[],"sidebar":{"activeKey":"topics","categories":[]},"footer":{"links":[],"primary":[]},"unread":{"notifications":false,"messages":false},"theme":{"enabled":false,"current":"gf-light","themeColor":"#fbfdff"}},"meta":{"title":"Embedded Next integration"},"url":"/","version":"1.0"}`))
	}))
	defer backend.Close()
	process, port, err := startProcess(processConfig{
		Enabled:        true,
		NodeCommand:    defaultNodeCommand,
		Entry:          defaultEntry,
		Host:           defaultHost,
		StartupTimeout: 15 * time.Second,
		GoOrigin:       backend.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	directory := process.runtimeDir
	if directory == "" || port == 0 {
		t.Fatalf("runtime directory=%q port=%d", directory, port)
	}
	response, err := http.Get("http://" + net.JoinHostPort(defaultHost, strconv.Itoa(port)) + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, readErr := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if readErr != nil || response.StatusCode != http.StatusOK || !strings.Contains(string(body), "Embedded Next integration") {
		t.Fatalf("embedded Next response status=%d body=%q err=%v", response.StatusCode, body, readErr)
	}
	if err := process.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(directory); !os.IsNotExist(err) {
		t.Fatalf("runtime directory was not removed: %v", err)
	}
}

func TestAvailableLoopbackPort(t *testing.T) {
	occupied, err := net.Listen("tcp", net.JoinHostPort(defaultHost, "0"))
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	occupiedPort := occupied.Addr().(*net.TCPAddr).Port

	port, err := availableLoopbackPort(defaultHost)
	if err != nil {
		t.Fatal(err)
	}
	if port < 1024 || port > 65535 {
		t.Fatalf("availableLoopbackPort() = %d, want a dynamic high port", port)
	}
	if port == occupiedPort {
		t.Fatalf("availableLoopbackPort() reused occupied port %d", port)
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(defaultHost, strconv.Itoa(port)))
	if err != nil {
		t.Fatalf("allocated port %d cannot be reused: %v", port, err)
	}
	_ = listener.Close()
}

func newTestProxy(target *url.URL) *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(target)
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (function roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
