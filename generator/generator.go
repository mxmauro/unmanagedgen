package generator

import (
	"hash/fnv"
	"strconv"
	"strings"

	parser "github.com/mxmauro/gofile-parser"
)

// -----------------------------------------------------------------------------

type Generator struct {
	filename    string
	packageName string
	imports     []parser.ParsedImport
	structs     []*Struct
	idPrefix    string
	idCounter   uint
}

type Struct struct {
	name   string
	fields []Field
}

type Field struct {
	names             []string
	typeName          string
	typeNamePrefixMod string
	tags              string
	opts              intFieldOptions
}

type FieldOptions struct {
	IsNative               bool
	isString               bool
	IsPointer              bool
	ArraySlice             *string
	IsArraySliceOfPointers bool
}

type intFieldOptions struct {
	IsNative               bool
	IsString               bool
	IsPointer              bool
	ArraySlice             *string
	IsArraySliceOfPointers bool
}

// -----------------------------------------------------------------------------

func UnmanagedName(name string) string {
	dotIdx := strings.Index(name, ".")
	namespace := ""
	if dotIdx >= 0 {
		namespace = name[0:dotIdx]
		name = name[(dotIdx + 1):]
	}
	if parser.IsPublic(name) {
		name = "Unmanaged" + name
	} else {
		runePsName := []rune(name)
		name = "unmanaged" + strings.ToUpper(string(runePsName[0])) + string(runePsName[1:])
	}
	return namespace + name
}

func New(pf *parser.ParsedFile) *Generator {
	gen := &Generator{
		filename:    pf.Filename[0:len(pf.Filename)-3] + "_unmanaged.go",
		packageName: pf.Package,
		imports:     pf.Imports,
		structs:     make([]*Struct, 0),
		idCounter:   0,
	}

	h := fnv.New32a()
	h.Sum([]byte(gen.filename))
	gen.idPrefix = "unmgd" + strings.ToUpper(strconv.FormatInt(int64(h.Sum32()), 16)) + "_"

	return gen
}

func (gen *Generator) NextId() string {
	gen.idCounter += 1
	return gen.idPrefix + strconv.FormatUint(uint64(gen.idCounter), 10)
}

func (gen *Generator) AddStruct(name string) *Struct {
	gs := &Struct{
		name:   name,
		fields: make([]Field, 0),
	}
	gen.structs = append(gen.structs, gs)
	return gs
}

func (gen *Generator) Save() error {
	var err error

	sc := newSaveContext(gen)

	sc.WriteLine("package " + sc.gen.packageName)
	sc.WriteLine("")

	err = sc.WriteImports()
	if err != nil {
		return err
	}

	err = sc.WriteStructs()
	if err != nil {
		return err
	}

	err = sc.Save()
	if err != nil {
		return err
	}

	// Done
	return nil
}

func (gs *Struct) AddField(names []string, typeName string, tags parser.ParsedTags, opts FieldOptions) {
	finalTags := ""
	for k, v := range tags {
		if k != "unmanaged" {
			if len(finalTags) > 0 {
				finalTags += " "
			}
			finalTags += k + ":" + strconv.Quote(string(v))
		}
	}
	if len(finalTags) > 0 {
		finalTags = "`" + finalTags + "`"
	}

	iOpts := intFieldOptions{
		IsNative:               opts.IsNative,
		IsString:               opts.IsNative && typeName == "string",
		IsPointer:              opts.IsPointer,
		ArraySlice:             opts.ArraySlice,
		IsArraySliceOfPointers: opts.IsArraySliceOfPointers,
	}
	if !opts.IsNative {
		typeName = UnmanagedName(typeName)
	}

	filteredNames := make([]string, 0)
	for _, s := range names {
		if len(s) > 0 {
			filteredNames = append(filteredNames, s)
		}
	}

	typeNamePrefixMod := ""
	if opts.IsPointer {
		typeNamePrefixMod = "*"
	}
	if opts.ArraySlice != nil {
		typeNamePrefixMod += "[" + (*opts.ArraySlice) + "]"
		if opts.IsArraySliceOfPointers {
			typeNamePrefixMod += "*"
		}
	}

	gs.fields = append(gs.fields, Field{
		names:             filteredNames,
		typeName:          typeName,
		typeNamePrefixMod: typeNamePrefixMod,
		tags:              finalTags,
		opts:              iOpts,
	})
}
