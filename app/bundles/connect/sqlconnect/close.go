package sqlconnect

import "log/slog"

func (itself *Connect) Close() {
	dbIns := itself.Connect
	if dbIns == nil {
		return
	}
	db, err := dbIns.DB()
	if err != nil {
		return
	}
	if db == nil {
		return
	}
	if err = db.Close(); err != nil {
		slog.Error("dbClose", "err", err)
	} else {
		slog.Info("dbCloseSuccess")
	}
}
