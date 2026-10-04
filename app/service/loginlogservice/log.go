package loginlogservice

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
)

// UserID is zero until authentication has completed; attempted usernames are never stored.
type Event struct {
	UserID, SessionID                        uint64
	Action, Method, Provider, Result, Reason string
}

type window struct {
	started time.Time
	count   int
}
type failureBudget struct {
	sync.Mutex
	ips    map[string]window
	global window
}

var failures = failureBudget{ips: make(map[string]window)}
var digestKey = func() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return key
}()

func (b *failureBudget) allow(ip string, now time.Time, account ...string) bool {
	b.Lock()
	defer b.Unlock()
	if now.Sub(b.global.started) >= time.Minute {
		b.global = window{started: now}
		for key, value := range b.ips {
			if now.Sub(value.started) >= time.Minute {
				delete(b.ips, key)
			}
		}
	}
	if b.global.count >= 1000 {
		return false
	}
	keys := []string{"ip:" + ip}
	if len(account) > 0 && account[0] != "" {
		keys = append(keys, "account:"+account[0])
	}
	missing := 0
	for _, key := range keys {
		w, exists := b.ips[key]
		if !exists {
			missing++
		}
		if now.Sub(w.started) < time.Minute && w.count >= 10 {
			return false
		}
	}
	if len(b.ips)+missing > 4096 {
		return false
	}
	for _, key := range keys {
		w := b.ips[key]
		if now.Sub(w.started) >= time.Minute {
			w = window{started: now}
		}
		w.count++
		b.ips[key] = w
	}
	b.global.count++
	return true
}

func Record(c *gin.Context, event Event, accountKey ...string) {
	ip, ua := "", ""
	if c != nil && c.Request != nil {
		ip, ua = c.ClientIP(), c.Request.UserAgent()
	}
	if (event.Result == "failure" || event.Result == "challenge") && !failures.allow(ip, time.Now(), accountKey...) {
		return
	}
	row := authsessions.Log{UserId: event.UserID, UserAuthTokenId: event.SessionID,
		Action: clip(event.Action, 32), AuthMethod: clip(event.Method, 32), OAuthProvider: clip(event.Provider, 64),
		Result: clip(event.Result, 16), Reason: clip(event.Reason, 256), ClientIP: clip(ip, 45), UserAgent: clip(ua, 512), CreatedAt: time.Now()}
	if err := authsessions.CreateSecurityLog(&row); err != nil {
		slog.Error("Authentication audit write failed", "action", row.Action, "error", err)
	}
}

func Failure(c *gin.Context, method, provider, reason string, identity ...string) {
	key := ""
	if len(identity) > 0 && identity[0] != "" {
		digest := hmac.New(sha256.New, digestKey)
		digest.Write([]byte(strings.ToLower(strings.TrimSpace(identity[0]))))
		key = hex.EncodeToString(digest.Sum(nil))
	}
	Record(c, Event{Action: "login_failure", Method: method, Provider: provider, Result: "failure", Reason: reason}, key)
}

func clip(value string, max int) string {
	value = strings.ToValidUTF8(value, "")
	if len(value) <= max {
		return value
	}
	value = value[:max]
	return strings.ToValidUTF8(value, "")
}

type Filter = authsessions.LogFilter
type Row = authsessions.LogRow
type Page = authsessions.LogPage

func List(filter Filter, owner uint64, admin bool) (Page, error) {
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	var since, until *time.Time
	if filter.Since != "" {
		at, err := time.Parse(time.RFC3339, filter.Since)
		if err != nil {
			return Page{}, err
		}
		since = &at
	}
	if filter.Until != "" {
		at, err := time.Parse(time.RFC3339, filter.Until)
		if err != nil {
			return Page{}, err
		}
		until = &at
	}
	result, err := authsessions.ListLogs(filter, owner, admin, since, until)
	if err != nil {
		return Page{}, err
	}
	if !admin {
		for i := range result.List {
			result.List[i].Reason = ""
		}
	}
	return result, nil
}
