//go:build windows

package signalwatch

import (
	"os"
	"os/signal"
	"syscall"
)

func ListenSignal(quit chan os.Signal) {
	signal.Notify(quit,
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGQUIT)
}
