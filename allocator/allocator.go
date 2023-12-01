package allocator

import (
	"unsafe"
)

// -----------------------------------------------------------------------------

type Allocator interface {
	Alloc(size uintptr) unsafe.Pointer
	Free(ptr unsafe.Pointer)
}
