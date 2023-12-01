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
const guardSize = 16

var guard [guardSize]byte

// -----------------------------------------------------------------------------

func init() {
	for idx := 0; idx < guardSize; idx++ {
		guard[idx] = byte(0x01 + idx)
	}
}

type DebugCAllocator struct {
	usage int64
}

func NewWithDebug() *DebugCAllocator {
	return &DebugCAllocator{}
}

func (c *DebugCAllocator) Alloc(size uintptr) unsafe.Pointer {
	ptr := C.malloc(C.size_t(size) + sizeOfUintptr + guardSize*2)

	C.memcpy(ptr, unsafe.Pointer(&guard), C.size_t(guardSize))
	ptr = unsafe.Add(ptr, guardSize)

	*((*uintptr)(ptr)) = size
	atomic.AddInt64(&c.usage, int64(size))
	ptr = unsafe.Add(ptr, sizeOfUintptr)

	ptr2 := unsafe.Add(ptr, size)
	C.memcpy(ptr2, unsafe.Pointer(&guard), C.size_t(guardSize))

	return unsafe.Pointer(ptr)
}

func (c *DebugCAllocator) Free(ptr unsafe.Pointer) {
	if ptr != nil {
		realPtr := unsafe.Add(ptr, -(sizeOfUintptr + guardSize))
		if C.memcmp(realPtr, unsafe.Pointer(&guard), C.size_t(guardSize)) != 0 {
			panic("DebugCAllocator::bufferoverflow/pre detected")
		}

		sizePtr := unsafe.Add(ptr, -sizeOfUintptr)
		size := *((*uintptr)(sizePtr))
		newUsage := atomic.AddInt64(&c.usage, -int64(size))
		if int(newUsage) < 0 {
			panic("DebugCAllocator usage below 0")
		}

		ptr = unsafe.Add(ptr, size)
		if C.memcmp(ptr, unsafe.Pointer(&guard), C.size_t(guardSize)) != 0 {
			panic("DebugCAllocator::bufferoverflow/post detected")
		}

		C.free(realPtr)
	}
}

func (c *DebugCAllocator) Usage() int64 {
	return atomic.LoadInt64(&c.usage)
}
