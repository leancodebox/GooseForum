//go:build darwin || (openbsd && !mips64) || linux

package signalwatch

import (
	"os"
	"os/signal"
	"syscall"
)

func ListenSignal(quit chan os.Signal) {
	signal.Notify(quit,
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
}
