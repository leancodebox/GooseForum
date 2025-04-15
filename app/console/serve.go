package console

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
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
	var (
		debug = setting.IsDebug()
	)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	slog.Info("GooseForum:start")
	slog.Info(fmt.Sprintf("GooseForum:useMem %d KB", m.Alloc/1024/8))

	if debug {
		go func() {
			err := http.ListenAndServe("0.0.0.0:7071", nil)
			if err != nil {
				slog.Error("debug listen ", "err", err)
			}
		}()
	}

	//// 检查是否需要setup
	//if !setupservice.IsInitialized() {
	//	setupServe()
	//}

	// 启动主服务
	ginServe()
}

func ginServe() {
	go RunJob()
	defer StopJob()

	port := preferences.GetString("server.port", 8080)
	isDebug := setting.IsDebug()
	var engine *gin.Engine
	if !isDebug {
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
		engine = gin.New()
		engine.Use(gin.Recovery())
	} else {
		engine = gin.Default()
	}

	routes.RegisterByGin(engine)

	srv := &http.Server{
		Addr:           ":" + port,
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	slog.Info("GooseForum:listen " + port)
	slog.Info("use port:" + port)
	slog.Info("start use:" + cast.ToString(setting.GetUnitTime()))
	fmt.Println("if in local you can http://localhost:" + port)

	quit := make(chan os.Signal, 1)
	listenSignal(quit)
	data := <-quit
	slog.Info("Shutdown Server ...", "signal", data)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Info("Server Shutdown:", err)
	}
	slog.Info("Server exiting")
}
