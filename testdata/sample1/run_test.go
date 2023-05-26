package sample1

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/mxmauro/unmanagedgen/allocator/c"
)

// -----------------------------------------------------------------------------

const SamplesCount = 1000000

// -----------------------------------------------------------------------------

func TestSample1(t *testing.T) {
	alloc := c.NewWithDebug()

	t.Logf("Initializing %v elements", SamplesCount)
	arr := make([]*UnmanagedSample, SamplesCount)
	for idx := 0; idx < len(arr); idx++ {
		arr[idx] = NewUnmanagedSample(alloc)
	}

	t.Log("Making changes")
	total := len(arr) * 200
	oldPct := -1
	for idx := 0; idx < total; idx++ {
		makeSampleChange(arr[rand.Intn(len(arr))])
		pct := ((idx + 1) * 100) / total
		if pct != oldPct {
			oldPct = pct
			if pct%10 == 0 {
				t.Logf("  -> %v%%", pct)
			}
		}
	}

	t.Log("Freeing elements")
	for idx := 0; idx < len(arr); idx++ {
		arr[idx].Free()
	}

	if alloc.Usage() != 0 {
		t.Fatalf("Usage is not zero! [%v]", alloc.Usage())
	}
}

func makeSampleChange(v *UnmanagedSample) {
	switch rand.Intn(24) {
	case 0:
		v.SomeInt = rand.Int()

	case 1:
		v.SetSomeString(strings.Repeat("*", 16+rand.Intn(128)))

	case 2:
		makeSubsampleChange(&v.SomeSubsample)

	case 3:
		v.ArrayOfInts[rand.Intn(len(v.ArrayOfInts))] = rand.Int()

	case 4:
		v.SetArrayOfStrings(rand.Intn(len(v.ArrayOfStrings)), strings.Repeat("*", 16+rand.Intn(128)))

	case 5:
		makeSubsampleChange(&v.ArrayOfSubsamples[rand.Intn(len(v.ArrayOfSubsamples))])

	case 6:
		if v.SliceOfInts == nil {
			v.SetSliceOfIntsCapacity(rand.Intn(16), true)
		}
		if len(v.SliceOfInts) > 0 {
			v.SliceOfInts[rand.Intn(len(v.SliceOfInts))] = rand.Int()
		}

	case 7:
		if v.SliceOfStrings == nil {
			v.SetSliceOfStringsCapacity(rand.Intn(16), true)
		}
		if len(v.SliceOfStrings) > 0 {
			v.SetSliceOfStrings(rand.Intn(len(v.SliceOfStrings)), strings.Repeat("*", 16+rand.Intn(128)))
		}

	case 8:
		if v.SliceOfSubsamples == nil {
			v.SetSliceOfSubsamplesCapacity(rand.Intn(16), true)
		}
		if len(v.SliceOfSubsamples) > 0 {
			makeSubsampleChange(&v.SliceOfSubsamples[rand.Intn(len(v.SliceOfSubsamples))])
		}

	case 9:
		v.SetArrayOfPtrToInts(rand.Intn(len(v.ArrayOfPtrToInts)), getRandomPtrToInt())

	case 10:
		v.SetArrayOfPtrToStrings(rand.Intn(len(v.ArrayOfPtrToStrings)), getRandomPtrToString())

	case 11:
		ss := NewUnmanagedSubSample(v.Allocator())
		v.SetArrayOfPtrToSubsamples(rand.Intn(len(v.ArrayOfPtrToSubsamples)), ss)
		makeSubsampleChange(ss)

	case 12:
		if v.SliceOfPtrToInts == nil {
			v.SetSliceOfPtrToIntsCapacity(rand.Intn(16), true)
		}
		if len(v.SliceOfPtrToInts) > 0 {
			v.SetSliceOfPtrToInts(rand.Intn(len(v.SliceOfPtrToInts)), getRandomPtrToInt())
		}

	case 13:
		if v.SliceOfPtrToStrings == nil {
			v.SetSliceOfPtrToIntsCapacity(rand.Intn(16), true)
		}
		if len(v.SliceOfPtrToStrings) > 0 {
			v.SetSliceOfPtrToStrings(rand.Intn(len(v.SliceOfPtrToStrings)), getRandomPtrToString())
		}

	case 14:
		if v.SliceOfPtrToSubsamples == nil {
			v.SetSliceOfPtrToSubsamplesCapacity(rand.Intn(16), true)
		}
		if len(v.SliceOfPtrToSubsamples) > 0 {
			ss := NewUnmanagedSubSample(v.Allocator())
			v.SetSliceOfPtrToSubsamples(rand.Intn(len(v.SliceOfPtrToSubsamples)), ss)
			makeSubsampleChange(ss)
		}

	case 15:
		v.SetPtrToInt(getRandomPtrToInt())

	case 16:
		v.SetPtrToString(getRandomPtrToString())

	case 17:
		if rand.Intn(8) == 0 {
			ss := NewUnmanagedSubSample(v.Allocator())
			v.SetPtrToSomeSubsample(ss)
			makeSubsampleChange(ss)
		} else {
			v.SetPtrToSomeSubsample(nil)
		}

	case 18:
		if v.PtrToArrayOfInts == nil {
			v.SetPtrToArrayOfIntsCreateArray()
		} else if rand.Intn(16) == 0 {
			v.SetPtrToArrayOfIntsDestroyArray()
		}
		if v.PtrToArrayOfInts != nil {
			(*v.PtrToArrayOfInts)[rand.Intn(len(*v.PtrToArrayOfInts))] = rand.Int()
		}

	case 19:
		if v.PtrToArrayOfStrings == nil {
			v.SetPtrToArrayOfStringsCreateArray()
		} else if rand.Intn(16) == 0 {
			v.SetPtrToArrayOfStringsDestroyArray()
		}
		if v.PtrToArrayOfStrings != nil {
			v.SetPtrToArrayOfStrings(rand.Intn(len(*v.PtrToArrayOfStrings)), strings.Repeat("*", 16+rand.Intn(128)))
		}

	case 20:
		if v.PtrToArrayOfSubsamples == nil {
			v.SetPtrToArrayOfSubsamplesCreateArray()
		} else if rand.Intn(16) == 0 {
			v.SetPtrToArrayOfSubsamplesDestroyArray()
		}
		if v.PtrToArrayOfSubsamples != nil {
			makeSubsampleChange(&((*v.PtrToArrayOfSubsamples)[rand.Intn(len(*v.PtrToArrayOfSubsamples))]))
		}

	case 21:
		if v.PtrToSliceOfInts == nil {
			v.SetPtrToSliceOfIntsCapacity(rand.Intn(16), true)
		}
		if v.PtrToSliceOfInts != nil && len(*v.PtrToSliceOfInts) > 0 {
			(*v.PtrToSliceOfInts)[rand.Intn(len(*v.PtrToSliceOfInts))] = rand.Int()
		}

	case 22:
		if v.PtrToSliceOfStrings == nil {
			v.SetPtrToSliceOfStringsCapacity(rand.Intn(16), true)
		}
		if v.PtrToSliceOfStrings != nil && len(*v.PtrToSliceOfStrings) > 0 {
			v.SetPtrToSliceOfStrings(rand.Intn(len(*v.PtrToSliceOfStrings)), strings.Repeat("*", 16+rand.Intn(128)))
		}

	case 23:
		if v.PtrToSliceOfSubsamples == nil {
			v.SetPtrToSliceOfSubsamplesCapacity(rand.Intn(16), true)
		}
		if v.PtrToSliceOfSubsamples != nil && len(*v.PtrToSliceOfSubsamples) > 0 {
			makeSubsampleChange(&((*v.PtrToSliceOfSubsamples)[rand.Intn(len(*v.PtrToSliceOfSubsamples))]))
		}

		/*

			PtrToArrayOfPtrToInts       *[4]*int
			PtrToArrayOfPtrToStrings    *[4]*string
			PtrToArrayOfPtrToSubSamples *[4]*UnmanagedSubSample
			PtrToSliceOfPtrToInts       *[]*int
			PtrToSliceOfPtrToStrings    *[]*string
			PtrToSliceOfPtrToSubSamples *[]*UnmanagedSubSample
		*/
	}
}

func makeSubsampleChange(v *UnmanagedSubSample) {
	switch rand.Intn(2) {
	case 0:
		v.SomeInt = rand.Int()

	case 1:
		v.SetSomeString(strings.Repeat("*", 16+rand.Intn(128)))
	}
}

func getRandomPtrToInt() *int {
	intVal := rand.Intn(32)
	if intVal == 0 {
		return nil
	}
	return &intVal
}

func getRandomPtrToString() *string {
	intVal := rand.Intn(128)
	if intVal == 0 {
		return nil
	}
	s := strings.Repeat("*", 16+intVal)
	return &s
}
