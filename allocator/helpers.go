package allocator

import (
	"unsafe"
)

// -----------------------------------------------------------------------------

var sizeOfPtr = 4 << (^uintptr(0) >> 63)
var safeMul = uintptr(1) << (4 * unsafe.Sizeof(sizeOfPtr))

// -----------------------------------------------------------------------------

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
