// Command attachment-load exercises real handlers against disposable SQLite databases.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/http/controllers/api"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/attachments"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/attachmentservice"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type stage struct {
	SizeKiB                int            `json:"sizeKiB"`
	Workers                int            `json:"workers"`
	Requests               int            `json:"requests"`
	Success                int            `json:"success"`
	Seconds                float64        `json:"seconds"`
	RequestsPerSecond      float64        `json:"requestsPerSecond"`
	P50Ms                  float64        `json:"p50Ms"`
	P95Ms                  float64        `json:"p95Ms"`
	P99Ms                  float64        `json:"p99Ms"`
	AllocatedMiBPerRequest float64        `json:"allocatedMiBPerRequest"`
	PeakHeapMiB            float64        `json:"peakHeapMiB"`
	Errors                 map[string]int `json:"errors"`
}

type report struct {
	Started             string         `json:"started"`
	GoVersion           string         `json:"goVersion"`
	CPUs                int            `json:"cpus"`
	DatabaseConnections int            `json:"databaseConnections"`
	DatabaseMode        string         `json:"databaseMode"`
	Scope               string         `json:"scope"`
	Stages              []stage        `json:"stages"`
	ClaimRounds         int            `json:"claimRounds"`
	ClaimViolations     int            `json:"claimViolations"`
	CancelRounds        int            `json:"cancelRounds"`
	CancelViolations    int            `json:"cancelViolations"`
	QuotaRounds         int            `json:"quotaRounds"`
	QuotaExceededRounds int            `json:"quotaExceededRounds"`
	QuotaMaxAccepted    int            `json:"quotaMaxAccepted"`
	RaceErrors          map[string]int `json:"raceErrors"`
}

type response struct {
	Result      attachments.Info `json:"result"`
	MessageCode string           `json:"messageCode"`
}

type attempt struct {
	file     attachments.Info
	duration time.Duration
	err      string
}

type harness struct {
	db     *gorm.DB
	files  *gorm.DB
	client *http.Client
	url    string
	postNo atomic.Uint64
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func numbers(raw string) []int {
	var result []int
	for _, value := range strings.Split(raw, ",") {
		n, err := strconv.Atoi(value)
		must(err)
		if n < 1 {
			panic("load parameters must be positive")
		}
		result = append(result, n)
	}
	return result
}

func saveTrust(limit *int) {
	c := pageConfig.DefaultTrustSettings()
	for i := range c.LevelLimits {
		c.LevelLimits[i].MaxDailyUploads = limit
	}
	raw, err := json.Marshal(c)
	must(err)
	must(pageConfig.SaveConfig(pageConfig.TrustSettings, string(raw)))
	hotdataserve.ClearTrustSettingsCache()
}

func setup(connections int) (*harness, func()) {
	dir, err := os.MkdirTemp("", "gooseforum-attachment-load-")
	must(err)
	dir, err = filepath.EvalSymlinks(dir)
	must(err)
	ready := false
	defer func() {
		if !ready {
			_ = os.RemoveAll(dir)
		}
	}()
	for _, name := range []string{"default", "file"} {
		prefix := "db." + name + "."
		preferences.Set(prefix+"connection", "sqlite")
		preferences.Set(prefix+"path", filepath.Join(dir, name+".db"))
		preferences.Set(prefix+"maxOpenConnections", connections)
		preferences.Set(prefix+"maxIdleConnections", connections)
	}
	preferences.Set("badger.path", filepath.Join(dir, "badger"))
	db, files := dbconnect.Connect(), db4fileconnect.Connect()
	db.Logger = logger.Default.LogMode(logger.Silent)
	files.Logger = logger.Default.LogMode(logger.Silent)
	for _, conn := range []*gorm.DB{db, files} {
		var databases []struct{ File string }
		must(conn.Raw("PRAGMA database_list").Scan(&databases).Error)
		if len(databases) == 0 || !strings.HasPrefix(databases[0].File, dir+string(os.PathSeparator)) {
			panic("refusing load test outside temporary databases")
		}
	}
	must(db.AutoMigrate(&users.EntityComplete{}, &pageConfig.Entity{}, &attachments.Entity{}, &posts.Entity{}))
	must(files.AutoMigrate(&filedata.Entity{}))
	for id := uint64(1); id <= 2; id++ {
		u := users.EntityComplete{Id: id, Username: fmt.Sprintf("load-%d", id), UsernameLower: fmt.Sprintf("load-%d", id), TrustLevel: trustlevel.Member, CreatedAt: time.Now().Add(-72 * time.Hour)}
		must(db.Create(&u).Error)
	}
	saveTrust(new(100000))
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.POST("/attachments", func(c *gin.Context) {
		id := uint64(1)
		if c.GetHeader("X-Load-User") == "quota" {
			id = 2
		}
		c.Set("userId", id)
		api.UploadAttachment(c)
	})
	server := httptest.NewServer(router)
	transport := &http.Transport{MaxIdleConns: 128, MaxIdleConnsPerHost: 128}
	h := &harness{db: db, files: files, client: &http.Client{Transport: transport, Timeout: 30 * time.Second}, url: server.URL + "/attachments"}
	ready = true
	return h, func() {
		server.Close()
		transport.CloseIdleConnections()
		dbconnect.Close()
		db4fileconnect.Close()
		must(os.RemoveAll(dir))
	}
}

func form(size int) ([]byte, string, [32]byte) {
	data := bytes.Repeat([]byte("a"), size)
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "load.txt")
	must(err)
	_, err = part.Write(data)
	must(err)
	must(w.Close())
	return body.Bytes(), w.FormDataContentType(), sha256.Sum256(data)
}

func (h *harness) upload(body []byte, contentType string, quota bool) attempt {
	started := time.Now()
	req, err := http.NewRequest(http.MethodPost, h.url, bytes.NewReader(body))
	must(err)
	req.Header.Set("Content-Type", contentType)
	if quota {
		req.Header.Set("X-Load-User", "quota")
	}
	res, err := h.client.Do(req)
	if err != nil {
		return attempt{duration: time.Since(started), err: err.Error()}
	}
	defer res.Body.Close()
	var payload response
	err = json.NewDecoder(res.Body).Decode(&payload)
	if err != nil {
		return attempt{duration: time.Since(started), err: err.Error()}
	}
	result := attempt{file: payload.Result, duration: time.Since(started)}
	if res.StatusCode != 200 || result.file.ID == 0 {
		result.err = fmt.Sprintf("HTTP %d %s", res.StatusCode, payload.MessageCode)
	}
	return result
}

func parallel(n, workers int, run func(int)) {
	var next atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range workers {
		wg.Go(func() {
			<-start
			for {
				i := int(next.Add(1)) - 1
				if i >= n {
					return
				}
				run(i)
			}
		})
	}
	close(start)
	wg.Wait()
}

func (h *harness) measure(sizeKiB, workers, count int) stage {
	body, contentType, hash := form(sizeKiB * 1024)
	results := make([]attempt, count)
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	var peak atomic.Uint64
	peak.Store(before.HeapAlloc)
	done, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(stopped)
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				var current runtime.MemStats
				runtime.ReadMemStats(&current)
				if current.HeapAlloc > peak.Load() {
					peak.Store(current.HeapAlloc)
				}
			}
		}
	}()
	started := time.Now()
	parallel(count, workers, func(i int) { results[i] = h.upload(body, contentType, false) })
	elapsed := time.Since(started)
	close(done)
	<-stopped
	runtime.ReadMemStats(&after)
	s := stage{SizeKiB: sizeKiB, Workers: workers, Requests: count, Seconds: elapsed.Seconds(), Errors: map[string]int{}, PeakHeapMiB: float64(max(peak.Load(), after.HeapAlloc)) / (1 << 20), AllocatedMiBPerRequest: float64(after.TotalAlloc-before.TotalAlloc) / float64(count) / (1 << 20)}
	latencies := make([]float64, 0, count)
	for _, result := range results {
		latencies = append(latencies, float64(result.duration.Microseconds())/1000)
		if result.err != "" {
			s.Errors[result.err]++
			continue
		}
		s.Success++
		_, object, err := attachmentservice.Open(context.Background(), 1, result.file.ID)
		must(err)
		digest := sha256.New()
		n, err := io.Copy(digest, object.Body)
		must(err)
		must(object.Body.Close())
		if n != int64(sizeKiB*1024) || !bytes.Equal(digest.Sum(nil), hash[:]) {
			panic("uploaded content corrupted")
		}
		must(attachmentservice.Cancel(context.Background(), 1, result.file.ID))
	}
	slices.Sort(latencies)
	percentile := func(p int) float64 { return latencies[max(0, (count*p+99)/100-1)] }
	s.P50Ms, s.P95Ms, s.P99Ms = percentile(50), percentile(95), percentile(99)
	s.RequestsPerSecond = float64(s.Success) / elapsed.Seconds()
	fmt.Fprintf(os.Stderr, "size=%dKiB workers=%d success=%d/%d rps=%.1f p95=%.1fms heap=%.1fMiB errors=%v\n", sizeKiB, workers, s.Success, count, s.RequestsPerSecond, s.P95Ms, s.PeakHeapMiB, s.Errors)
	return s
}

func (h *harness) bind(file attachments.Info) error {
	post := posts.Entity{TopicId: 1, PostNo: h.postNo.Add(1), UserId: 1, Content: "load", Attachments: []attachments.Info{file}}
	return h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&post).Error; err != nil {
			return err
		}
		return attachments.BindWithDB(tx, 1, post.Id, post.Attachments)
	})
}

func (h *harness) races(out *report, rounds int) {
	for range rounds {
		file, err := attachmentservice.Upload(context.Background(), 1, "race.txt", bytes.NewBufferString("race"))
		must(err)
		results := make([]error, 16)
		parallel(16, 16, func(i int) { results[i] = h.bind(file) })
		winners := 0
		for _, err := range results {
			if err == nil {
				winners++
			} else if !errors.Is(err, attachments.ErrInvalid) {
				out.RaceErrors[err.Error()]++
			}
		}
		if winners != 1 {
			out.ClaimViolations++
		}
		out.ClaimRounds++
		file, err = attachmentservice.Upload(context.Background(), 1, "cancel-race.txt", bytes.NewBufferString("race"))
		must(err)
		original, err := attachments.Get(file.ID)
		must(err)
		var bound, canceled error
		parallel(2, 2, func(i int) {
			if i == 0 {
				bound = h.bind(file)
			} else {
				canceled = attachmentservice.Cancel(context.Background(), 1, file.ID)
			}
		})
		var row attachments.Entity
		found := h.db.First(&row, file.ID).Error
		var objects int64
		must(h.files.Model(&filedata.Entity{}).Where("name = ?", original.ObjectName).Count(&objects).Error)
		valid := (bound == nil && canceled != nil && found == nil && row.PostID != 0 && objects == 1) || (bound != nil && canceled == nil && errors.Is(found, gorm.ErrRecordNotFound) && objects == 0)
		if !valid {
			out.CancelViolations++
		}
		for _, err := range []error{bound, canceled} {
			if err != nil && !errors.Is(err, attachments.ErrInvalid) {
				out.RaceErrors[err.Error()]++
			}
		}
		out.CancelRounds++
	}
}

func (h *harness) quota(out *report, rounds int) {
	one := 1
	saveTrust(&one)
	defer saveTrust(new(100000))
	body, contentType, _ := form(1024)
	for range rounds {
		results := make([]attempt, 32)
		parallel(32, 32, func(i int) { results[i] = h.upload(body, contentType, true) })
		accepted := 0
		for _, result := range results {
			if result.err == "" {
				accepted++
				must(attachmentservice.Cancel(context.Background(), 2, result.file.ID))
			} else if result.err != "HTTP 400 upload.dailyLimit" {
				out.RaceErrors[result.err]++
			}
		}
		if accepted > 1 {
			out.QuotaExceededRounds++
		}
		out.QuotaMaxAccepted = max(out.QuotaMaxAccepted, accepted)
		out.QuotaRounds++
	}
}

func main() {
	workers := flag.String("workers", "1,4,16,32", "concurrent upload workers")
	sizes := flag.String("sizes-kib", "64,5120", "payload sizes in KiB")
	requests := flag.Int("requests", 128, "requests per small-file stage; files >=1MiB use at most 32")
	rounds := flag.Int("rounds", 50, "rounds of competing bind and cancel operations")
	connections := flag.Int("db-connections", 5, "connections per disposable SQLite database")
	output := flag.String("out", "", "optional JSON report path")
	allocProfile := flag.String("alloc-profile", "", "optional Go allocation profile path")
	flag.Parse()
	if *connections < 1 || *requests < 1 || *rounds < 1 {
		panic("load parameters must be positive")
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	h, cleanup := setup(*connections)
	defer cleanup()
	var mode string
	must(h.db.Raw("PRAGMA journal_mode").Scan(&mode).Error)
	out := report{Started: time.Now().UTC().Format(time.RFC3339), GoVersion: runtime.Version(), CPUs: runtime.GOMAXPROCS(0), DatabaseConnections: *connections, DatabaseMode: mode, Scope: "Temporary file-backed SQLite; loopback HTTP real upload handler with fixed test identity; normal uploads retain the quota-count query with a 100000/day test limit; Go heap/allocation metrics include client and server; post binding uses real transactions.", RaceErrors: map[string]int{}}
	for _, size := range numbers(*sizes) {
		if size > attachmentservice.HardMaxSize/1024 {
			panic("payload exceeds hard attachment limit")
		}
		for _, count := range numbers(*workers) {
			n := *requests
			if size >= 1024 {
				n = min(n, 32)
			}
			out.Stages = append(out.Stages, h.measure(size, count, n))
		}
	}
	h.races(&out, *rounds)
	h.quota(&out, min(*rounds, 20))
	if *allocProfile != "" {
		runtime.GC()
		profile, err := os.Create(*allocProfile)
		must(err)
		must(pprof.Lookup("allocs").WriteTo(profile, 0))
		must(profile.Close())
	}
	raw, err := json.MarshalIndent(out, "", "  ")
	must(err)
	fmt.Println(string(raw))
	if *output != "" {
		must(os.WriteFile(*output, append(raw, '\n'), 0600))
	}
}
