// +build linux

package goridge

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

func getAllocSize() uint {
	// ARCH detection
	// 64bit = 1 on 64-bit systems, 0 on 32-bit systems
	_64bit := 1 << (^uintptr(0) >> 63) / 2

	// maxAlloc is the maximum size of an allocation. On 64-bit,
	// it's theoretically possible to allocate 1<<heapAddrBits (Addr bits) bytes. On
	// 32-bit, however, this is one less than 1<<32 because the
	// number of bytes in the address space doesn't actually fit
	// in a uintptr.

	//       Platform  Addr bits  Arena size  L1 entries   L2 entries
	// --------------  ---------  ----------  ----------  -----------
	//       */64-bit         48        64MB           1    4M (32MB)
	// windows/64-bit         48         4MB          64    1M  (8MB)
	//       */32-bit         32         4MB           1  1024  (4KB)
	//     */mips(le)         31         4MB           1   512  (2KB)

	// But to be in safe zone, we are limiting allocation to 17.1 GB in x64 and 2.14 on x86
	// theoretically we can allow allocate up to 128Gb on x64
	// https://github.com/golang/go/blob/release-branch.go1.4/src/runtime/malloc.h#L152
	maxAlloc := uint(1)
	if _64bit == 1 {
		maxAlloc = maxAlloc << 34 // approx 17179869184 bytes or 17 Gb in 64 bit system
	} else {
		maxAlloc = maxAlloc << 31 // approx 2147483648 bytes or 2.14 Gb in 32 bit system
	}
	return maxAlloc
}

func MemAvail() uint64 {
	fname := "/proc/meminfo"
	FileBytes, err := ioutil.ReadFile(fname)
	if err != nil {
		return uint64(1) << 31 //2.14 Gb
	}
	bufr := bytes.NewBuffer(FileBytes)
	for {
		line, err := bufr.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

		}
		ndx := strings.Index(line, "MemFree:")
		if ndx >= 0 {
			line = strings.TrimSpace(line[9:])
			fmt.Printf("%q\n", line)
			line = line[:len(line)-3]
			fmt.Printf("%q\n", line)
			mem, err := strconv.ParseUint(line, 10, 64)
			if err == nil {
				return mem
			}
			return uint64(1) << 31 //2.14 Gb

		}
	}

	return uint64(1) << 31 //2.14 Gb
}
