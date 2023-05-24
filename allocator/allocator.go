package allocator

import (
	"unsafe"
)

// -----------------------------------------------------------------------------

type Allocator interface {
	Alloc(size uintptr) unsafe.Pointer
	Free(ptr unsafe.Pointer)

	ZeroMem(ptr unsafe.Pointer, size uintptr)
	CopyMem(dest, src unsafe.Pointer, size uintptr)
}
