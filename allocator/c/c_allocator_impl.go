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
	return unsafe.Pointer(ptr)
}

func (c *CAllocator) Free(ptr unsafe.Pointer) {
	C.free(ptr)
}
