package sample1

import (
	"go/ast"
)

type Sample struct {
	SomeInt                   int
	SomeString                string
	ArrayOfInts               [4]int
	SlicesOfBytes             []byte
	ArrayOfStrings            [4]string
	SlicesOfStrings           []string
	ArrayOfPtrToInts          [4]*int
	SlicesOfPtrToBytes        []*byte
	ArrayOfPtrToStrings       [4]*string
	SlicesOfPtrToStrings      []*string
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

	AT ast.ArrayType `unmanaged:"omit"`
}
