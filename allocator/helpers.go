package allocator

// #include <memory.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

// -----------------------------------------------------------------------------

var sizeOfPtr = 4 << (^uintptr(0) >> 63)
var safeMul = uintptr(1) << (4 * unsafe.Sizeof(sizeOfPtr))

// -----------------------------------------------------------------------------

func ZeroMem(ptr unsafe.Pointer, size uintptr) {
	C.memset(ptr, 0, C.size_t(size))
}

func CopyMem(dest, src unsafe.Pointer, size uintptr) {
	C.memcpy(dest, src, C.size_t(size))
}

func AddUintptr(a, b uintptr) (uintptr, bool) {
	overflow := a+b < a
	return a + b, overflow
}

func MulUintptr(a, b uintptr) (uintptr, bool) {
	if a|b < safeMul || a == 0 {
		return a * b, false
	}
	overflow := b > ^uintptr(0)/a
	return a * b, overflow
}
