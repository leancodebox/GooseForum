package console

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/captchaOpt"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	paniclog "github.com/leancodebox/GooseForum/app/bundles/recovery"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/bundles/signalwatch"
	"github.com/leancodebox/GooseForum/app/console/job"
	"github.com/leancodebox/GooseForum/app/http/routes"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/oauthservice"
	"github.com/spf13/cast"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// CmdServe represents the available web sub-command.
var CmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Start web server",
	Run:   runWeb,
	Args:  cobra.NoArgs,
}

func runWeb(_ *cobra.Command, _ []string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	slog.Info("GooseForum:start")
	slog.Info(fmt.Sprintf("GooseForum:useMem %d KB", m.Alloc/1024/8))

	startDebugServices()
	ginServe()
}

func startDebugServices() {
	if !setting.IsDebug() {
		return
	}
	go servePprof()
}

func servePprof() {
	defer paniclog.Recover("pprof_server")
	// go tool pprof http://localhost:19070/debug/pprof/profile
	// go tool pprof -http=:9001 http://localhost:19070/debug/pprof/heap
	// http://127.0.0.1:19070/debug/pprof/
	const addr = "127.0.0.1:19070"
	srv := &http.Server{
		Addr:              addr,
		Handler:           pprofMux(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("debug listen ", "err", err)
	}
}

func pprofMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return mux
}

func ginServe() {
	// 初始化OAuth配置
	oauthservice.InitOAuth()
	captchaOpt.StartCleanup()
	mailservice.StartEmailProcessor()
	job.Run()

	port := preferences.GetString("server.port", 8080)
	engine := newGinEngine()
	routes.RegisterByGin(engine)
	host := ``
	if setting.IsLocal() {
		host = `127.0.0.1`
	}
	address := fmt.Sprintf("%v:%v", host, port)
	srv := &http.Server{
		Addr:           address,
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	quit := make(chan os.Signal, 1)
	signalwatch.ListenSignal(quit)
	go func() {
		defer paniclog.Recover("http_server")
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http serve ", "err", err)
			fmt.Println("http serve ", "err", err)
			quit <- os.Interrupt
		}
	}()

	slog.Info("GooseForum:listen " + port)
	slog.Info("use port:" + port)
	slog.Info("start use:" + cast.ToString(setting.GetUnitTime()))
	fmt.Println("if in local you can http://localhost:" + port)

	data := <-quit
	slog.Info("Shutdown Server ...", "signal", data)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Info("Server Shutdown", "err", err)
	}

	slog.Info("Server exiting")
}

func newGinEngine() *gin.Engine {
	if setting.IsDebug() {
		gin.SetMode(gin.DebugMode)
		return gin.Default()
	} else {
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
	}
	return gin.New()
}
