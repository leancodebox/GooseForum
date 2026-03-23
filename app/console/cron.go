package console

import (
	"errors"
	"log/slog"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/logging"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
	"github.com/robfig/cron/v3"
)

var c = cron.New(
	cron.WithLogger(cron.VerbosePrintfLogger(logging.CronLogging{})),
)
var runCron = false

func RunJob() {
	closer.Register(StopJob)
	slog.Info("start cron")
	backupSpec := preferences.Get("db.spec", "0 3 * * *")
	entryID, err := c.AddFunc(backupSpec, upCmd(func() {
		dbconnect.BackupSQLiteHandle()
		db4fileconnect.BackupSQLiteHandle()
	}))
	slog.Info("reg cron", "entryID", entryID, "spec", backupSpec, "err", err)
	entryID, err = c.AddFunc("3 3 * * *", upCmd(func() {
		// 实现未来7天的创建。检查除了今天以外6天的是否创建，如果没有创建则进行创建
		now := time.Now()
		keys := []dailyStats.StatType{
			dailyStats.StatTypeRegCount,
			dailyStats.StatTypeArticleCount,
			dailyStats.StatTypeReplyCount,
		}
		for i := range 7 {
			date := now.AddDate(0, 0, i)
			for _, key := range keys {
				_ = dailyStats.InitStats(date, key)
			}
		}
	}))
	slog.Info("reg cron", "entryID", entryID, "spec", backupSpec, "err", err)
	runCron = true
	c.Start()
}

func StopJob() error {
	if !runCron {
		return nil
	}
	ctx := c.Stop()
	select {
	case <-ctx.Done():
		runCron = false
		return nil
	case <-time.After(10 * time.Second):
		slog.Error("timed out waiting for job to stop")
		return errors.New("timed out waiting for job to stop")
	}
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
