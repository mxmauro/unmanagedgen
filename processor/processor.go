package processor

import (
	parser "github.com/mxmauro/gofile-parser"
	"github.com/mxmauro/unmanagedgen/allocator"
	"github.com/mxmauro/unmanagedgen/generator"
)

// -----------------------------------------------------------------------------

type UnmanagedSample struct {
	N       int
	B       string
	X       []int
	XB      []string
	__alloc allocator.Allocator
}

type Processor struct {
	pf  *parser.ParsedFile
	gen *generator.Generator
}

// -----------------------------------------------------------------------------

func ProcessFile(filename string) error {
	var err error

	proc := Processor{}

	proc.pf, err = parser.ParseFile(parser.ParseFileOptions{
		Filename: filename,
		// ResolveModule: false,
	})
	if err != nil {
		return err
	}

	proc.gen = generator.New(proc.pf)

	for _, decl := range proc.pf.Declarations {
		switch tDecl := decl.Type.(type) {
		case *parser.ParsedStruct:
			if tag, ok := decl.Tags.GetTag("unmanaged"); ok {
				if tag.GetBoolProperty("omit") {
					break
				}
			}

			err = proc.processStruct(decl.Name, tDecl)
			if err != nil {
				return err
			}
		}
	}

	err = proc.gen.Save()
	if err != nil {
		return err
	}

	// Done
	return nil
}
