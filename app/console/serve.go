package console

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/signalwatch"
	"github.com/leancodebox/GooseForum/app/service/oauthservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"

	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/routes"
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

	if setting.IsDebug() {
		go pprofServe()
	}

	// 启动主服务
	ginServe()
}

func pprofServe() {
	// go tool pprof http://localhost:19070/debug/pprof/profile
	// go tool pprof -http=:9001 http://localhost:19070/debug/pprof/heap
	// http://127.0.0.1:19070/debug/pprof/
	err := http.ListenAndServe("127.0.0.1:19070", nil)
	if err != nil {
		slog.Error("debug listen ", "err", err)
	}
}

func ginServe() {
	// 初始化OAuth配置
	oauthservice.InitOAuth()
	defer userservice.CloseUpdateUserLastActiveTime()
	RunJob()
	defer StopJob()

	port := preferences.GetString("server.port", 8080)
	var engine *gin.Engine
	if !setting.IsDebug() {
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
		engine = gin.New()
		engine.Use(gin.Recovery())
	} else {
		engine = gin.Default()
	}

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Info("Server Shutdown:", err)
	}
	slog.Info("Server exiting")
}
