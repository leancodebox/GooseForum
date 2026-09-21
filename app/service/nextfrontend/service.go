// Package nextfrontend manages the optional Next.js frontend process.
package nextfrontend

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	appresource "github.com/leancodebox/GooseForum/resource"
)

const (
	defaultNodeCommand   = "node"
	defaultEntry         = "./resource/apps/next/.next/standalone/apps/next/start.mjs"
	defaultHost          = "127.0.0.1"
	defaultStartupSecond = 15
	maxPortAttempts      = 3
	healthPath           = "/goose-internal/health"
	healthTokenHeader    = "X-Goose-Health-Token"
)

var nodeVersionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)`)

type processConfig struct {
	Enabled        bool
	NodeCommand    string
	Entry          string
	Host           string
	Port           int
	StartupTimeout time.Duration
	GoOrigin       string
	HealthToken    string
}

type process struct {
	cmd         *exec.Cmd
	done        chan struct{}
	waitErr     error
	runtimeDir  string
	stopOnce    sync.Once
	cleanupOnce sync.Once
	stopping    atomic.Bool
}

// Frontend starts the optional Next.js process and proxies public page requests
// while it is available.
type Frontend struct {
	config    processConfig
	proxy     *httputil.ReverseProxy
	available atomic.Bool
	process   *process
}

func configFromPreferences(goPort string) processConfig {
	return processConfig{
		Enabled:        preferences.GetBool("server.next", false),
		NodeCommand:    defaultNodeCommand,
		Entry:          defaultEntry,
		Host:           defaultHost,
		StartupTimeout: defaultStartupSecond * time.Second,
		GoOrigin:       "http://" + net.JoinHostPort(defaultHost, goPort),
	}
}

// New creates the optional frontend integration. It has no side effects until
// Start is called.
func New(goPort string) *Frontend {
	config := configFromPreferences(goPort)
	return &Frontend{config: config}
}

// Start enables the Next.js route after the child process is ready. Any error
// leaves the built-in Go frontend active.
func (f *Frontend) Start() {
	if !f.config.Enabled {
		return
	}
	process, port, err := startProcess(f.config)
	if err != nil {
		slog.Warn("next frontend unavailable; using Go frontend", "err", err)
		return
	}
	f.config.Port = port
	target := &url.URL{Scheme: "http", Host: net.JoinHostPort(f.config.Host, strconv.Itoa(port))}
	f.proxy = newFrontendProxy(target)
	f.process = process
	f.available.Store(true)
	closer.RegisterPriority(closer.PriorityProducer, f.Close)
	slog.Info("next frontend ready", "url", "http://"+net.JoinHostPort(f.config.Host, strconv.Itoa(f.config.Port)))
	go func() {
		<-process.done
		f.available.Store(false)
		process.cleanupRuntime()
		if process.stopping.Load() {
			return
		}
		if process.waitErr != nil {
			slog.Warn("next frontend stopped; using Go frontend", "err", process.waitErr)
		} else {
			slog.Warn("next frontend stopped; using Go frontend")
		}
	}()
}

func newFrontendProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
		slog.Warn("next frontend proxy request failed", "err", err)
		http.Error(writer, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
	}
	return proxy
}

// Proxy serves a request through Next.js when it is a public frontend route.
// It returns whether the response has been handled.
func (f *Frontend) Proxy(writer http.ResponseWriter, request *http.Request) bool {
	if !f.available.Load() || f.proxy == nil || !isNextRoute(request) {
		return false
	}
	f.proxy.ServeHTTP(writer, request)
	return true
}

func isNextRoute(request *http.Request) bool {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		return false
	}
	if strings.EqualFold(request.Header.Get("X-Goose-Page"), "true") {
		return false
	}
	path := request.URL.Path
	if strings.HasPrefix(path, "/_next/") || path == "/goose-page-data" {
		return true
	}
	for _, prefix := range []string{"/p/post/", "/u/", "/c/"} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	switch path {
	case "/", "/categories", "/members", "/links", "/sponsors", "/messages", "/drafts",
		"/moderation", "/settings", "/access-groups", "/theme-preview", "/notifications",
		"/publish", "/search", "/login", "/reset-password":
		return true
	default:
		return false
	}
}

// Close disables proxying before stopping the child process.
func (f *Frontend) Close() error {
	f.available.Store(false)
	if f.process == nil {
		return nil
	}
	return f.process.Close()
}

// Start validates the local runtime, starts Next.js, and waits until its TCP
// listener is available.
func startProcess(config processConfig) (*process, int, error) {
	runtimeDir, entryPath, err := prepareRuntime(config.Entry)
	if err != nil {
		return nil, 0, err
	}
	if runtimeDir != "" {
		config.Entry = entryPath
		defer func() {
			if runtimeDir != "" {
				_ = os.RemoveAll(runtimeDir)
			}
		}()
	}
	nodePath, entryPath, err := checkEnvironment(config)
	if err != nil {
		return nil, 0, err
	}

	var lastErr error
	for range maxPortAttempts {
		port, err := availableLoopbackPort(config.Host)
		if err != nil {
			return nil, 0, err
		}
		config.Port = port
		config.HealthToken, err = newHealthToken()
		if err != nil {
			return nil, 0, err
		}
		process, err := launchProcess(config, nodePath, entryPath)
		if err == nil {
			process.runtimeDir = runtimeDir
			runtimeDir = ""
			return process, port, nil
		}
		lastErr = err
		if !errors.Is(err, errProcessExitedBeforeReady) {
			break
		}
		slog.Warn("Next.js did not acquire dynamic port; retrying", "port", port, "err", err)
	}
	return nil, 0, lastErr
}

func newHealthToken() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf("create Next.js health token: %w", err)
	}
	return hex.EncodeToString(data), nil
}

func prepareRuntime(entry string) (string, string, error) {
	if entry != defaultEntry {
		return "", entry, nil
	}
	archive, embedded, err := appresource.GetNextStandaloneArchive()
	if err != nil {
		return "", "", fmt.Errorf("read embedded Next.js bundle: %w", err)
	}
	if !embedded {
		return "", entry, nil
	}
	directory, err := os.MkdirTemp("", "gooseforum-next-")
	if err != nil {
		return "", "", fmt.Errorf("create Next.js runtime directory: %w", err)
	}
	if err := extractArchive(archive, directory); err != nil {
		_ = os.RemoveAll(directory)
		return "", "", err
	}
	return directory, filepath.Join(directory, "apps", "next", "start.mjs"), nil
}

func extractArchive(data []byte, directory string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open embedded Next.js bundle: %w", err)
	}
	for _, item := range reader.File {
		name := filepath.Clean(filepath.FromSlash(item.Name))
		if name == "." || filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return fmt.Errorf("invalid embedded Next.js path %q", item.Name)
		}
		if !item.Mode().IsRegular() {
			return fmt.Errorf("unsupported embedded Next.js entry %q", item.Name)
		}
		target := filepath.Join(directory, name)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create Next.js runtime directory: %w", err)
		}
		input, err := item.Open()
		if err != nil {
			return fmt.Errorf("open embedded Next.js file %q: %w", item.Name, err)
		}
		output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, item.Mode().Perm())
		if err != nil {
			_ = input.Close()
			return fmt.Errorf("create Next.js runtime file %q: %w", item.Name, err)
		}
		_, copyErr := io.Copy(output, input)
		closeOutputErr := output.Close()
		closeInputErr := input.Close()
		if copyErr != nil {
			return fmt.Errorf("extract Next.js runtime file %q: %w", item.Name, copyErr)
		}
		if closeOutputErr != nil {
			return closeOutputErr
		}
		if closeInputErr != nil {
			return closeInputErr
		}
	}
	return nil
}

var errProcessExitedBeforeReady = errors.New("Next.js exited before becoming ready")

func launchProcess(config processConfig, nodePath, entryPath string) (*process, error) {
	cmd := exec.Command(
		nodePath,
		entryPath,
		"--goose-origin", config.GoOrigin,
		"--hostname", config.Host,
		"--port", strconv.Itoa(config.Port),
	)
	cmd.Dir = filepath.Dir(entryPath)
	cmd.Env = append(
		os.Environ(),
		"GOOSEFORUM_MANAGED=true",
		"GOOSEFORUM_HEALTH_TOKEN="+config.HealthToken,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start node: %w", err)
	}

	process := &process{cmd: cmd, done: make(chan struct{})}
	go func() {
		process.waitErr = cmd.Wait()
		close(process.done)
	}()

	if err := process.waitReady(config); err != nil {
		_ = process.Close()
		return nil, err
	}
	return process, nil
}

func availableLoopbackPort(host string) (int, error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		return 0, fmt.Errorf("allocate dynamic Next.js port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		return 0, fmt.Errorf("release dynamic Next.js port %d: %w", port, err)
	}
	return port, nil
}

func checkEnvironment(config processConfig) (string, string, error) {
	if config.NodeCommand == "" {
		return "", "", errors.New("Node.js command is empty")
	}
	if config.Entry == "" {
		return "", "", errors.New("Next.js entry is empty")
	}
	if !isLoopbackHost(config.Host) {
		return "", "", fmt.Errorf("Next.js host %q must be a loopback address", config.Host)
	}
	if config.StartupTimeout <= 0 {
		return "", "", errors.New("Next.js startup timeout must be positive")
	}
	if _, err := url.ParseRequestURI(config.GoOrigin); err != nil {
		return "", "", fmt.Errorf("invalid GooseForum origin: %w", err)
	}
	entryPath, err := filepath.Abs(config.Entry)
	if err != nil {
		return "", "", fmt.Errorf("resolve Next entry: %w", err)
	}
	info, err := os.Stat(entryPath)
	if err != nil {
		return "", "", fmt.Errorf("Next entry %q: %w", entryPath, err)
	}
	if !info.Mode().IsRegular() {
		return "", "", fmt.Errorf("Next entry %q is not a regular file", entryPath)
	}

	nodePath, err := exec.LookPath(config.NodeCommand)
	if err != nil {
		return "", "", fmt.Errorf("find Node.js executable %q: %w", config.NodeCommand, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, nodePath, "--version").CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("run %s --version: %w", nodePath, err)
	}
	if err := validateNodeVersion(strings.TrimSpace(string(output))); err != nil {
		return "", "", err
	}
	return nodePath, entryPath, nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(strings.TrimSpace(host), "localhost") {
		return true
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	return ip != nil && ip.IsLoopback()
}

func validateNodeVersion(version string) error {
	parts := nodeVersionPattern.FindStringSubmatch(version)
	if len(parts) != 4 {
		return fmt.Errorf("cannot parse Node.js version %q", version)
	}
	major, _ := strconv.Atoi(parts[1])
	minor, _ := strconv.Atoi(parts[2])
	if major < 20 || major == 20 && minor < 9 {
		return fmt.Errorf("Node.js %s is unsupported; Next.js requires >=20.9.0", version)
	}
	return nil
}

func (p *process) waitReady(config processConfig) error {
	deadline := time.NewTimer(config.StartupTimeout)
	defer deadline.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	healthURL := "http://" + address + healthPath
	client := &http.Client{
		Timeout: 500 * time.Millisecond,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for {
		select {
		case <-p.done:
			err := p.waitErr
			if err == nil {
				err = errors.New("process exited successfully before becoming ready")
			}
			return fmt.Errorf("%w on %s: %v", errProcessExitedBeforeReady, address, err)
		case <-deadline.C:
			return fmt.Errorf("Next.js did not listen on %s within %s", address, config.StartupTimeout)
		case <-ticker.C:
			if nextHealthReady(client, healthURL, config.HealthToken) {
				return nil
			}
		}
	}
}

func nextHealthReady(client *http.Client, healthURL, token string) bool {
	response, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	_ = response.Body.Close()
	actual := response.Header.Get(healthTokenHeader)
	return response.StatusCode == http.StatusNoContent &&
		len(actual) == len(token) &&
		subtle.ConstantTimeCompare([]byte(actual), []byte(token)) == 1
}

// Close stops the child process and waits briefly for it to exit.
func (p *process) Close() error {
	var closeErr error
	exited := false
	p.stopOnce.Do(func() {
		p.stopping.Store(true)
		if p.cmd == nil || p.cmd.Process == nil {
			exited = true
			return
		}
		var err error
		if runtime.GOOS == "windows" {
			err = p.cmd.Process.Kill()
		} else {
			err = p.cmd.Process.Signal(os.Interrupt)
		}
		if err != nil && !errors.Is(err, os.ErrProcessDone) {
			closeErr = err
		}
		exited = p.waitForExit(3 * time.Second)
		if !exited && runtime.GOOS != "windows" {
			if err := p.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) && closeErr == nil {
				closeErr = err
			}
			exited = p.waitForExit(3 * time.Second)
		}
		if !exited && closeErr == nil {
			closeErr = errors.New("Next.js did not exit after termination")
		}
	})
	if exited {
		p.cleanupRuntime()
	} else {
		go func() {
			<-p.done
			p.cleanupRuntime()
		}()
	}
	return closeErr
}

func (p *process) waitForExit(timeout time.Duration) bool {
	select {
	case <-p.done:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (p *process) cleanupRuntime() {
	p.cleanupOnce.Do(func() {
		if p.runtimeDir != "" {
			if err := os.RemoveAll(p.runtimeDir); err != nil {
				slog.Warn("remove Next.js runtime directory", "path", p.runtimeDir, "err", err)
			}
		}
	})
}
