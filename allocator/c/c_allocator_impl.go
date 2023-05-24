//go:build cgo

package c

// #include <memory.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

// -----------------------------------------------------------------------------

type CAllocator struct {
}

func New() *CAllocator {
	return &CAllocator{}
}

func (c *CAllocator) Alloc(size uintptr) unsafe.Pointer {
	ptr := C.malloc(C.size_t(size))
	c.ZeroMem(ptr, size)
	return unsafe.Pointer(ptr)
}

func (c *CAllocator) Free(ptr unsafe.Pointer) {
	C.free(ptr)
}

func (c *CAllocator) ZeroMem(ptr unsafe.Pointer, size uintptr) {
	C.memset(ptr, 0, C.size_t(size))
}

func (c *CAllocator) CopyMem(dest, src unsafe.Pointer, size uintptr) {
	C.memcpy(dest, src, C.size_t(size))
}
