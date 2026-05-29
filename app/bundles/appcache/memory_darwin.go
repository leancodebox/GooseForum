//go:build darwin

package appcache

import "golang.org/x/sys/unix"

func totalMemoryBytes() (uint64, bool) {
	total, err := unix.SysctlUint64("hw.memsize")
	return total, err == nil && total > 0
}
