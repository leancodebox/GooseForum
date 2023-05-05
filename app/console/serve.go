package console

import (
	"context"
	"fmt"
	"github.com/leancodebox/GooseForum/bundles/app"
	"github.com/leancodebox/GooseForum/bundles/logging"
	"github.com/leancodebox/GooseForum/routes"
	"github.com/leancodebox/goose/preferences"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

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
		debug = preferences.GetBool("app.debug", true)
	)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	info("GooseForum:start")
	info(fmt.Sprintf("GooseForum:useMem %d KB", m.Alloc/1024/8))

	if debug {
		go func() {
			err := http.ListenAndServe("0.0.0.0:7071", nil)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	ginServe()
}

const (
	ENV      = "env"
	EnvProd  = "production"
	EnvLocal = "local"
)

var (
	port   = preferences.GetString("app.port", 8080)
	isProd = preferences.Get("app.env") == EnvProd
)

func ginServe() {
	var engine *gin.Engine
	if isProd {
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
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	info("GooseForum:listen " + port)
	fmt.Printf("use port:%s\n", port)
	fmt.Println("start use:" + cast.ToString(app.GetUnitTime()))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logging.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logging.Println("Server Shutdown:", err)
	}
	logging.Println("Server exiting")
}
