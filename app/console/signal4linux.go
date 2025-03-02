//go:build darwin || (openbsd && !mips64) || linux

package console

import (
	"os"
	"os/signal"
	"syscall"
)

func listenSignal(quit chan os.Signal) {
	signal.Notify(quit,
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
}
