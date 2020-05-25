// +build windows

package goridge

import (
	"fmt"
	"syscall"
	"unsafe"
)

func getAllocSize() uint {
	var mod = syscall.NewLazyDLL("kernel32.dll")
	var proc = mod.NewProc("GetPhysicallyInstalledSystemMemory")
	var mem uint64

	ret, _, err := proc.Call(uintptr(unsafe.Pointer(&mem)))
	fmt.Printf("Ret: %d, err: %v, Physical memory: %d\n", ret, err, mem)
	return 0
}