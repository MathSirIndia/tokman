//go:build !windows

package gui

import "syscall"

// getDiskFreeGB returns free disk space in gigabytes on Unix/Linux/macOS.
func getDiskFreeGB() float64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(".", &stat); err == nil {
		return float64(stat.Bavail * uint64(stat.Bsize)) / 1024 / 1024 / 1024
	}
	return 0.0
}
