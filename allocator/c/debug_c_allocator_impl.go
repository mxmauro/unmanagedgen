//go:build cgo

package c

// #include <memory.h>
// #include <stdlib.h>
import "C"
import (
	"sync/atomic"
	"unsafe"
)

// -----------------------------------------------------------------------------

const sizeOfUintptr = 32 << uintptr(^uintptr(0)>>63)

// -----------------------------------------------------------------------------

type DebugCAllocator struct {
	usage int64
}

func NewWithDebug() *DebugCAllocator {
	return &DebugCAllocator{}
}

func (c *DebugCAllocator) Alloc(size uintptr) unsafe.Pointer {
	ptr := C.malloc(C.size_t(size) + sizeOfUintptr)
	*((*uintptr)(ptr)) = size
	atomic.AddInt64(&c.usage, int64(size))
	ptr = unsafe.Add(ptr, sizeOfUintptr)
	c.ZeroMem(ptr, size)
	return unsafe.Pointer(ptr)
}

func (c *DebugCAllocator) Free(ptr unsafe.Pointer) {
	if ptr != nil {
		ptr = unsafe.Add(ptr, -sizeOfUintptr)
		size := *((*uintptr)(ptr))
		newUsage := atomic.AddInt64(&c.usage, -int64(size))
		if int(newUsage) < 0 {
			panic("DebugCAllocator usage below 0")
		}
		C.free(ptr)
	}
}

func (c *DebugCAllocator) ZeroMem(ptr unsafe.Pointer, size uintptr) {
	C.memset(ptr, 0, C.size_t(size))
}

func (c *DebugCAllocator) CopyMem(dest, src unsafe.Pointer, size uintptr) {
	C.memcpy(dest, src, C.size_t(size))
}

func (c *DebugCAllocator) Usage() int64 {
	return atomic.LoadInt64(&c.usage)
}
