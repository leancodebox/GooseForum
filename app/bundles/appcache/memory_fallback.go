//go:build !darwin && !linux

package appcache

func totalMemoryBytes() (uint64, bool) {
	return 0, false
}
