//go:build windows

package gui

import (
	"syscall"
	"unsafe"
)

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceExW = kernel32.NewProc("GetDiskFreeSpaceExW")
)

// getDiskFreeGB returns free disk space in gigabytes on Windows via GetDiskFreeSpaceExW.
func getDiskFreeGB() float64 {
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes int64
	dir, err := syscall.UTF16PtrFromString(".")
	if err != nil {
		return 0.0
	}
	r1, _, _ := getDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(dir)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)
	if r1 == 0 {
		return 0.0
	}
	return float64(freeBytesAvailable) / 1024 / 1024 / 1024
}
