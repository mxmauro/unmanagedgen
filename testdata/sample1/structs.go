package sample1

import (
	"go/ast"
)

type Sample struct {
	SomeInt                     int
	SomeString                  string
	SomeSubsample               SubSample
	ArrayOfInts                 [4]int
	ArrayOfStrings              [4]string
	ArrayOfSubsamples           [4]SubSample
	SliceOfInts                 []int
	SliceOfStrings              []string
	SliceOfSubsamples           []SubSample
	ArrayOfPtrToInts            [4]*int
	ArrayOfPtrToStrings         [4]*string
	ArrayOfPtrToSubsamples      [4]*SubSample
	SliceOfPtrToInts            []*int
	SliceOfPtrToStrings         []*string
	SliceOfPtrToSubsamples      []*SubSample
	PtrToInt                    *int
	PtrToString                 *string
	PtrToSomeSubsample          *SubSample
	PtrToArrayOfInts            *[4]int
	PtrToArrayOfStrings         *[4]string
	PtrToArrayOfSubsamples      *[4]SubSample
	PtrToSliceOfInts            *[]int
	PtrToSliceOfStrings         *[]string
	PtrToSliceOfSubsamples      *[]SubSample
	PtrToArrayOfPtrToInts       *[4]*int
	PtrToArrayOfPtrToStrings    *[4]*string
	PtrToArrayOfPtrToSubsamples *[4]*SubSample
	PtrToSliceOfPtrToInts       *[]*int
	PtrToSliceOfPtrToStrings    *[]*string
	PtrToSliceOfPtrToSubsamples *[]*SubSample

	AT ast.ArrayType `unmanaged:"omit"`
}

type SubSample struct {
	SomeInt    int
	SomeString string
}
