package sample1

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/mxmauro/unmanagedgen/allocator/c"
)

// -----------------------------------------------------------------------------

const SamplesCount = 1

// -----------------------------------------------------------------------------

func TestSample1(t *testing.T) {
	alloc := c.NewWithDebug()

	t.Logf("Initializing %v elements", SamplesCount)
	arr := make([]*UnmanagedSample, SamplesCount)
	for idx := 0; idx < len(arr); idx++ {
		arr[idx] = NewUnmanagedSample(alloc)
	}

	t.Log("Making changes")
	for idx := 0; idx < len(arr)*200; idx++ {
		makeChange(arr[rand.Intn(len(arr))])
	}

	t.Log("Freeing elements")
	for idx := 0; idx < len(arr); idx++ {
		arr[idx].Free()
	}

	if alloc.Usage() != 0 {
		t.Fatalf("Usage is not zero! [%v]", alloc.Usage())
	}
}

func makeChange(v *UnmanagedSample) {
	switch rand.Intn(10) {
	case 0:
		v.SomeInt = rand.Int()

	case 1:
		v.SetSomeString(strings.Repeat("*", 16+rand.Intn(128)))

	case 2:
		v.ArrayOfInts[rand.Intn(4)] = rand.Int()

	case 3:
		if v.SlicesOfBytes == nil {
			v.SetSlicesOfBytesCapacity(rand.Intn(16), true)
		}
		if len(v.SlicesOfBytes) > 0 {
			v.SlicesOfBytes[rand.Intn(len(v.SlicesOfBytes))] = byte(rand.Intn(255))
		}

	case 4:
		v.SetArrayOfStrings(rand.Intn(4), strings.Repeat("*", 16+rand.Intn(128)))

	case 5:
		if v.SlicesOfStrings == nil {
			v.SetSlicesOfStringsCapacity(rand.Intn(16), true)
		}
		if len(v.SlicesOfStrings) > 0 {
			v.SetSlicesOfStrings(rand.Intn(len(v.SlicesOfStrings)), strings.Repeat("*", 16+rand.Intn(128)))
		}

	case 6:
		var intPtr *int

		intVal := rand.Intn(32)
		if intVal > 0 {
			intPtr = &intVal
		}
		v.SetArrayOfPtrToInts(rand.Intn(4), intPtr)

	case 7:
		if v.SlicesOfPtrToBytes == nil {
			v.SetSlicesOfPtrToBytesCapacity(rand.Intn(16), true)
		}
		if len(v.SlicesOfPtrToBytes) > 0 {
			var bytePtr *byte

			byteVal := byte(rand.Intn(32))
			if byteVal > 0 {
				bytePtr = &byteVal
			}
			v.SetSlicesOfPtrToBytes(rand.Intn(len(v.SlicesOfPtrToBytes)), bytePtr)
		}

	case 8:
		var strPtr *string

		intVal := rand.Intn(128)
		if intVal > 0 {
			s := strings.Repeat("*", 16+intVal)
			strPtr = &s
		}
		v.SetArrayOfPtrToStrings(rand.Intn(4), strPtr)

	case 9:
		if v.SlicesOfPtrToStrings == nil {
			v.SetSlicesOfPtrToStringsCapacity(rand.Intn(16), true)
		}
		if len(v.SlicesOfPtrToStrings) > 0 {
			var strPtr *string

			intVal := rand.Intn(128)
			if intVal > 0 {
				s := strings.Repeat("*", 16+intVal)
				strPtr = &s
			}
			v.SetSlicesOfPtrToStrings(rand.Intn(len(v.SlicesOfPtrToStrings)), strPtr)
		}

		/*
			PtrToInt                  *int
			PtrToString               *string
			PtrToArrayOfInts          *[4]int
			PtrToSlicesOfBytes        *[]byte
			PtrToArrayOfStrings       *[4]string
			PtrToSlicesOfStrings      *[]string
			PtrToArrayOfPtrToInts     *[4]*int
			PtrToSlicesOfPtrToBytes   *[]*byte
			PtrToArrayOfPtrToStrings  *[4]*string
			PtrToSlicesOfPtrToStrings *[]*string
		*/
	}
}
