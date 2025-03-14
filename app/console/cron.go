package console

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/logging"
	"github.com/robfig/cron/v3"
	"log/slog"
)

var c = cron.New(
	cron.WithLogger(cron.VerbosePrintfLogger(logging.CronLogging{})),
)
var runCron = false

func RunJob() {
	slog.Info("start cron")
	backupSpec := preferences.Get("db.spec", "0 3 * * *")
	spec, err := c.AddFunc(backupSpec, upCmd(func() {
		dbconnect.BackupSQLiteHandle()
		db4fileconnect.BackupSQLiteHandle()
	}))
	slog.Info("reg cron", "spec", spec, "err", err)
	runCron = true
	c.Run()
}

func StopJob() {
	if !runCron {
		return
	}
	c.Stop()
}

func upCmd(cmd func()) func() {
	return func() {
		defer func() {
			if p := recover(); p != nil {
				slog.Error("cron panic ", "p", p)
			}
		}()
		cmd()
	}
}
