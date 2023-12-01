package generator

import (
	"hash/fnv"
	"strconv"
	"strings"
	"text/template"

	parser "github.com/mxmauro/gofile-parser"
)

// -----------------------------------------------------------------------------

func (sc *SaveContext) WriteStructs() error {
	for _, st := range sc.gen.structs {
		err := sc.WriteStructDeclaration(st)
		if err != nil {
			return err
		}
	}

	for _, st := range sc.gen.structs {
		err := sc.WriteStructAllocator(st)
		if err != nil {
			return err
		}
	}

	for _, st := range sc.gen.structs {
		err := sc.WriteStructFieldsSetters(st)
		if err != nil {
			return err
		}
	}

	// Done
	return nil
}

func (sc *SaveContext) WriteStructDeclaration(st *Struct) error {
	type StructFieldsDecl struct {
		Name              string
		TypeName          string
		TypeNamePrefixMod string
		Tag               string
	}
	type StructDecl struct {
		Name         string
		Fields       []StructFieldsDecl
		AllocatorPkg string
	}

	decl := StructDecl{
		Name:         st.name,
		Fields:       make([]StructFieldsDecl, 0),
		AllocatorPkg: sc.allocatorPkg,
	}

	for _, fld := range st.fields {
		fldDecl := StructFieldsDecl{
			Name:              strings.Join(fld.names, ", "),
			TypeName:          fld.typeName,
			TypeNamePrefixMod: fld.typeNamePrefixMod,
		}
		if len(fld.tags) > 0 {
			fldDecl.Tag = " " + fld.tags
		}

		decl.Fields = append(decl.Fields, fldDecl)
	}

	err := sc.WriteTemplate("StructDeclaration", `
type {{.Name}} struct {
{{- range $fldIdx, $fld := .Fields }}
	{{$fld.Name}} {{$fld.TypeNamePrefixMod}}{{$fld.TypeName}} {{$fld.Tag}}
{{- end }}

	__alloc {{.AllocatorPkg}}.Allocator
	__isInternal bool
	__freeing bool
}
`, nil, decl)
	if err != nil {
		return err
	}

	// Done
	return nil
}

func (sc *SaveContext) WriteStructAllocator(st *Struct) error {
	type AllocNewFreeField struct {
		Name     string
		TypeName string
		Opts     intFieldOptions
	}
	type AllocNewFree struct {
		NewFuncName       string
		StructName        string
		ManagedStructName string
		AllocatorPkg      string
		MustFreeStrings   bool
		HaveArrays        bool
		Fields            []AllocNewFreeField
	}

	allocNF := AllocNewFree{
		StructName:        st.name,
		ManagedStructName: st.managedName,
		AllocatorPkg:      sc.allocatorPkg,
		Fields:            make([]AllocNewFreeField, 0),
	}

	// New method
	if parser.IsPublic(st.name) {
		allocNF.NewFuncName = "New" + st.name
	} else {
		allocNF.NewFuncName = "new" + capitalizeFirstLetter(st.name)
	}

	for _, fld := range st.fields {
		if fld.opts.IsString {
			allocNF.MustFreeStrings = true
		}
		if fld.opts.ArraySlice != nil {
			allocNF.HaveArrays = true
		}

		for _, name := range fld.names {
			ff := AllocNewFreeField{
				Name:     name,
				TypeName: fld.typeName,
				Opts:     fld.opts,
			}
			allocNF.Fields = append(allocNF.Fields, ff)
		}
	}

	funcMap := template.FuncMap{
		"counter": templateCounter(),
		"isSlice": func(s *string) bool {
			return s != nil && len(*s) == 0
		},
		"isArray": func(s *string) bool {
			return s != nil && len(*s) > 0
		},
		"isArrayOrSlice": func(s *string) bool {
			return s != nil
		},
	}

	err := sc.WriteTemplate("StructAllocator", `
// {{.NewFuncName}} creates a new {{.StructName}} object and returns a pointer to it
func {{.NewFuncName}}(alloc {{.AllocatorPkg}}.Allocator) *{{.StructName}} {

	ptr := alloc.Alloc(unsafe.Sizeof({{.StructName}}{}))
	if ptr == nil {
		panic("cannot allocate memory for {{$.StructName}}")
	}
	allocator.ZeroMem(ptr, unsafe.Sizeof({{.StructName}}{}))

	v := (*{{.StructName}})(ptr)
	v.__alloc = alloc
	v.initNonPointerNonNativeFields()
	return v
}

// InitAllocator sets the allocator used for fields
func (v *{{.StructName}}) InitAllocator(alloc {{.AllocatorPkg}}.Allocator) {
	v.__alloc = alloc
	v.__isInternal = true
	v.initNonPointerNonNativeFields()
}

// Free deletes the object and frees memory
func (v *{{.StructName}}) Free() {
{{- if .MustFreeStrings }} 
	var bytePtr *byte
{{- end }}
{{- if .HaveArrays }} 
	var arrLen int
{{- end }}

	if v.__freeing {
		return
	}
	v.__freeing = true

{{range $fldIdx, $fld := .Fields}}
	{{- if $fld.Opts.IsPointer }}
		if v.{{$fld.Name}} != nil {
			{{- if not (isArrayOrSlice $fld.Opts.ArraySlice) }}
				// {{$fld.Name}} is a simple pointer
				{{- if $fld.Opts.IsNative }}
					v.__alloc.Free(unsafe.Pointer(v.{{$fld.Name}}))
				{{- else }}
					v.{{$fld.Name}}.Free()
				{{- end }}
			{{- else }}
				{{- if $fld.Opts.IsArraySliceOfPointers }}
					// {{$fld.Name}} is a pointer to an array/slice of pointers
					// Free each non-nil element of the array (they are supposed to be unmanaged too)
					{{- $c := counter}}
					vv{{$c}} := *v.{{$fld.Name}}
					arrLen = len(vv{{$c}})
					for idx := 0; idx < arrLen; idx += 1 {
						if vv{{$c}}[idx] != nil {
							{{- if $fld.Opts.IsNative }}
								v.__alloc.Free(unsafe.Pointer(vv{{$c}}[idx]))
							{{- else }}
								vv{{$c}}[idx].Free()
							{{- end }}
						}
					}
				{{- else if $fld.Opts.IsNative }}
					{{- if $fld.Opts.IsString }}
						// {{$fld.Name}} is a pointer to an array/slice of strings
						// Free each string data in the array (we don't own the string headers)
						{{- $c := counter}}
						vv{{$c}} := *v.{{$fld.Name}}
						arrLen = len(vv{{$c}})
						for idx := 0; idx < arrLen; idx += 1 {
							bytePtr = unsafe.StringData(vv{{$c}}[idx])
							if bytePtr != nil {
								v.__alloc.Free(unsafe.Pointer(bytePtr))
							}
						}
					{{- /* else it is an array of things we don't need to free */ -}}
					{{- end }}
				{{- else }}
					// {{$fld.Name}} is a pointer to an array/slice of non-native objects (they are supposed to be unmanaged too)
					{{- $c := counter}}
					vv{{$c}} := *v.{{$fld.Name}}
					arrLen = len(vv{{$c}})
					for idx := 0; idx < arrLen; idx += 1 {
						vv{{$c}}[idx].Free()
					}
				{{- end }}
				// Free the array/slice
				v.__alloc.Free(unsafe.Pointer(v.{{$fld.Name}}))
			{{- end }}
		}
	{{- else if isArrayOrSlice $fld.Opts.ArraySlice }}
		{{- if $fld.Opts.IsArraySliceOfPointers }}
			// {{$fld.Name}} is an array/slice of pointers
			// Free each non-nil element of the array (they are supposed to be unmanaged too)
			arrLen = len(v.{{$fld.Name}})
			for idx := 0; idx < arrLen; idx += 1 {
				if v.{{$fld.Name}}[idx] != nil {
					{{- if $fld.Opts.IsNative }}
						v.__alloc.Free(unsafe.Pointer(v.{{$fld.Name}}[idx]))
					{{- else }}
						v.{{$fld.Name}}[idx].Free()
					{{- end }}
				}
			}
		{{- else if $fld.Opts.IsNative }}
			{{- if $fld.Opts.IsString }}
				// {{$fld.Name}} is an array/slice of strings
				// Free each string data (we don't own the string headers)
				{{- $c := counter}}
				arrLen = len(v.{{$fld.Name}})
				for idx := 0; idx < arrLen; idx += 1 {
					bytePtr = unsafe.StringData(v.{{$fld.Name}}[idx])
					if bytePtr != nil {
						v.__alloc.Free(unsafe.Pointer(bytePtr))
					}
				}
			{{- /* else it is an array of things we don't need to free */ -}}
			{{- end }}
		{{- else }}
			// {{$fld.Name}} is an array/slice of non-native objects (they are supposed to be unmanaged too)
			{{- $c := counter}}
			arrLen = len(v.{{$fld.Name}})
			for idx := 0; idx < arrLen; idx += 1 {
				v.{{$fld.Name}}[idx].Free()
			}
		{{- end }}
		{{- if isSlice $fld.Opts.ArraySlice }}
			// Free the slice data
			{{- $c := counter}}
			slicePtr{{$c}} := unsafe.SliceData(v.{{$fld.Name}})
			if slicePtr{{$c}} != nil {
				v.__alloc.Free(unsafe.Pointer(slicePtr{{$c}}))
			}
		{{- end }}
	{{- else if $fld.Opts.IsNative }}
		{{- if $fld.Opts.IsString }}
			// {{$fld.Name}} is a string
			// Free string data (we don't own the string header)
			bytePtr = unsafe.StringData(v.{{$fld.Name}})
			if bytePtr != nil {
				v.__alloc.Free(unsafe.Pointer(bytePtr))
			}
		{{- end }}
	{{- else }}
		// {{$fld.Name}} is a non-native objects (it is supposed to be unmanaged too)
		v.{{$fld.Name}}.Free()
	{{- end }}
{{end }}

	if !v.__isInternal {
		v.__alloc.Free(unsafe.Pointer(v))
	} else {
		v.__freeing = false
	}
}

func (v *{{.StructName}}) Allocator() {{.AllocatorPkg}}.Allocator {
	return v.__alloc
}

func (v *{{.StructName}}) initNonPointerNonNativeFields() {
{{- range $fldIdx, $fld := .Fields}}
	{{- if and (not $fld.Opts.IsPointer) (not $fld.Opts.IsNative) }}
		{{- if isArrayOrSlice $fld.Opts.ArraySlice }}
			{{- if isArray $fld.Opts.ArraySlice }}
				{{- if not $fld.Opts.IsArraySliceOfPointers }}
					{{- $c := counter}}
					arrLen{{$c}} := len(v.{{$fld.Name}})
					for idx := 0; idx < arrLen{{$c}}; idx += 1 {
						v.{{$fld.Name}}[idx].InitAllocator(v.__alloc)
					}
				{{- end }}
			{{- end }}
		{{- else }}
			v.{{$fld.Name}}.InitAllocator(v.__alloc)
		{{- end }}
	{{- end }}
{{- end }}
}

func (v *{{$.StructName}}) zeroAlloc(size uintptr) unsafe.Pointer {
	ptr := v.__alloc.Alloc(size)
	if ptr == nil {
		panic("cannot allocate memory for {{$.StructName}}")
	}
	allocator.ZeroMem(ptr, size)
	return ptr
}
`, funcMap, allocNF)
	if err != nil {
		return err
	}

	// Done
	return nil
}

func (sc *SaveContext) WriteStructFieldsSetters(st *Struct) error {
	type SetterField struct {
		FuncName              string
		SetFuncPrefix         string
		Name                  string
		TypeName              string
		TypeNamePrefixMod     string
		Opts                  intFieldOptions
		FriendlyArraySize     string
		FriendlyArrayTypeName string
	}

	type SetterNeedAllocArrayItem map[string]string

	type Setter struct {
		StructName         string
		AllocatorPkg       string
		SetterFields       []SetterField
		NeedAllocSlice     map[string]string
		NeedAllocSlicePtr  map[string]string
		NeedAllocString    bool
		NeedAllocStringPtr bool
		NeedAllocArrayPtr  map[string]SetterNeedAllocArrayItem
	}

	setter := Setter{
		StructName:        st.name,
		AllocatorPkg:      sc.allocatorPkg,
		SetterFields:      make([]SetterField, 0),
		NeedAllocSlice:    make(map[string]string),
		NeedAllocSlicePtr: make(map[string]string),
		NeedAllocArrayPtr: make(map[string]SetterNeedAllocArrayItem),
	}

	for _, fld := range st.fields {
		if fld.opts.IsPointer {
			if fld.opts.ArraySlice == nil {
				if fld.opts.IsString {
					setter.NeedAllocStringPtr = true
				}
			} else {
				fieldType := fld.typeName
				if fld.opts.IsArraySliceOfPointers {
					fieldType = "*" + fieldType
				}
				if fld.opts.IsString {
					setter.NeedAllocStringPtr = true
				}

				if len(*fld.opts.ArraySlice) > 0 {
					siz := friendlyArraySize(*fld.opts.ArraySlice) + friendlyArrayTypeName(fieldType)
					if m, ok := setter.NeedAllocArrayPtr[siz]; ok {
						m[*fld.opts.ArraySlice] = fieldType
					} else {
						m2 := make(SetterNeedAllocArrayItem)
						m2[*fld.opts.ArraySlice] = fieldType
						setter.NeedAllocArrayPtr[siz] = m2
					}
				} else {
					setter.NeedAllocSlicePtr[friendlyArrayTypeName(fieldType)] = fieldType
				}
			}
		} else if fld.opts.ArraySlice != nil {
			fieldType := fld.typeName
			if fld.opts.IsArraySliceOfPointers {
				fieldType = "*" + fieldType
			}
			if fld.opts.IsString {
				setter.NeedAllocStringPtr = true
			}

			if len(*fld.opts.ArraySlice) == 0 {
				setter.NeedAllocSlice[friendlyArrayTypeName(fieldType)] = fieldType
			}
		} else if fld.opts.IsString {
			setter.NeedAllocString = true
		}

		for _, name := range fld.names {
			arrayTypeName := fld.typeName
			if fld.opts.ArraySlice != nil && fld.opts.IsArraySliceOfPointers {
				arrayTypeName = "*" + arrayTypeName
			}

			arraySize := ""
			if fld.opts.ArraySlice != nil && len(*fld.opts.ArraySlice) > 0 {
				arraySize = *fld.opts.ArraySlice
			}

			setterField := SetterField{
				Name:                  name,
				TypeName:              fld.typeName,
				TypeNamePrefixMod:     fld.typeNamePrefixMod,
				Opts:                  fld.opts,
				FriendlyArraySize:     friendlyArraySize(arraySize),
				FriendlyArrayTypeName: friendlyArrayTypeName(arrayTypeName),
			}

			// Set method
			if parser.IsPublic(name) {
				setterField.FuncName = name
				setterField.SetFuncPrefix = "Set"
			} else {
				setterField.FuncName = capitalizeFirstLetter(name)
				setterField.SetFuncPrefix = "set"
			}

			setter.SetterFields = append(setter.SetterFields, setterField)
		}
	}

	funcMap := template.FuncMap{
		"counter": templateCounter(),
		"derefStr": func(s *string) string {
			return *s
		},
		"isSlice": func(s *string) bool {
			return s != nil && len(*s) == 0
		},
		"isArray": func(s *string) bool {
			return s != nil && len(*s) > 0
		},
		"isArrayOrSlice": func(s *string) bool {
			return s != nil
		},
	}

	err := sc.WriteTemplate("StructFieldsSetters", `
{{range $fldIdx, $fld := .SetterFields}}
	{{- if $fld.Opts.IsPointer }}
		{{if not (isArrayOrSlice $fld.Opts.ArraySlice) }}
			{{- /* a simple pointer */ -}}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}(value *{{$fld.TypeName}}) {
			{{- if $fld.Opts.IsNative }}
				{{- if $fld.Opts.IsString }}
					{{- /* a pointer to a string */ -}}
					if v.{{$fld.Name}} != nil {
						v.__alloc.Free(unsafe.Pointer(v.{{$fld.Name}}))
					}
					if value != nil {
						v.{{$fld.Name}} = v.dupStringPtr(*value)
					} else {
						v.{{$fld.Name}} = nil
					}
				{{- else }}
					{{- /* a pointer to a native type */ -}}
					if value != nil {
						valueSize := unsafe.Sizeof(*value)
						if v.{{$fld.Name}} == nil {
							v.{{$fld.Name}} = (*{{$fld.TypeName}})(v.zeroAlloc(valueSize))
						}
						{{$.AllocatorPkg}}.CopyMem(unsafe.Pointer(v.{{$fld.Name}}), unsafe.Pointer(value), valueSize)
					} else if v.{{$fld.Name}} != nil {
						v.__alloc.Free(unsafe.Pointer(v.{{$fld.Name}}))
						v.{{$fld.Name}} = nil
					}
				{{- end }}
			{{- else }}
				{{- /* a pointer to a non-native object (it is supposed to be unmanaged too) */ -}}
				if v.{{$fld.Name}} != nil {
					v.{{$fld.Name}}.Free()
				}
				v.{{$fld.Name}} = value
			{{- end }}
}
		{{- else }}
			{{if isSlice $fld.Opts.ArraySlice }}
				{{- /* a pointer to a slice of something */ -}}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}Capacity(sliceLen int, preserve bool) {
				var newSlice {{$fld.TypeNamePrefixMod}}{{$fld.TypeName}}

				// Create the new slice
				if sliceLen > 0 {
					newSlice = v.allocSlicePtr_{{$fld.FriendlyArrayTypeName}}(sliceLen)
				}

				toPreserve := 0

				// Copy original items if preserve up to the size of the new slice and free the rest
				if v.{{$fld.Name}} != nil {
					vv := *v.{{$fld.Name}}
					oldSliceLen := len(vv)

					// Calculate the number of entries to preserve
					if preserve {
						toPreserve = oldSliceLen
						if toPreserve > sliceLen {
							toPreserve = sliceLen
						}
						for idx := 0; idx < toPreserve; idx++ {
							(*newSlice)[idx] = vv[idx]
						}
					}

					{{- if or $fld.Opts.IsArraySliceOfPointers (or (not $fld.Opts.IsNative) $fld.Opts.IsString) }}
						// Free unused entries
						for idx := toPreserve; idx < oldSliceLen; idx++ {
							{{- if $fld.Opts.IsArraySliceOfPointers }}
								v.{{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx, nil)
							{{- else if $fld.Opts.IsNative }}
								{{- if $fld.Opts.IsString }}
									v.{{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx, unsafe.String(nil, 0))
								{{- /* else it is pointer to a slice of things we don't need to free */ -}}
								{{- end }}
							{{- else }}
								{{- /* a pointer to a slice of non-native objects (they are supposed to be unmanaged too) */ -}}
								vv[idx].Free()
							{{- end }}
						}
					{{- end }}

					// Free old slice
					v.__alloc.Free(unsafe.Pointer(v.{{$fld.Name}}))
				}

				{{- if and (not $fld.Opts.IsArraySliceOfPointers) (not $fld.Opts.IsNative) }}
					// Initialize added non-native structs
					for idx := toPreserve; idx < sliceLen; idx++ {
						(*newSlice)[idx].InitAllocator(v.__alloc)
					}
				{{- end }}

				// Replace
				v.{{$fld.Name}} = newSlice
}
			{{else }}
				{{- /* a pointer to an array of something */ -}}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}CreateArray() {
				v.{{$fld.SetFuncPrefix}}{{$fld.FuncName}}DestroyArray()
				v.{{$fld.Name}} = v.allocArrayPtr_{{$fld.FriendlyArraySize}}{{$fld.FriendlyArrayTypeName}}()

				{{- if and (not $fld.Opts.IsArraySliceOfPointers) (not $fld.Opts.IsNative) }}
					arrLen := len(v.{{$fld.Name}})
					// Initialize added non-native structs
					for idx := 0; idx < arrLen; idx++ {
						v.{{$fld.Name}}[idx].InitAllocator(v.__alloc)
					}
				{{- end }}
}

func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}DestroyArray() {
				if v.{{$fld.Name}} != nil {
					{{- if or $fld.Opts.IsArraySliceOfPointers (or (not $fld.Opts.IsNative) $fld.Opts.IsString) }}
						// Free all entries
						arrLen := len(v.{{$fld.Name}})
						for idx := 0; idx < arrLen; idx++ {
							{{- if $fld.Opts.IsArraySliceOfPointers }}
								v.{{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx, nil)
							{{- else if $fld.Opts.IsNative }}
								{{- if $fld.Opts.IsString }}
									v.{{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx, unsafe.String(nil, 0))
								{{- /* else it is an array of things we don't need to free */ -}}
								{{- end }}
							{{- else }}
								{{- /* array/slice of non-native objects (they are supposed to be unmanaged too) */ -}}
								v.{{$fld.Name}}[idx].Free()
							{{- end }}
						}
					{{- end }}

					// Free array
					v.__alloc.Free(unsafe.Pointer(v.{{$fld.Name}}))
					v.{{$fld.Name}} = nil
				}
}
			{{- end }}
			{{if $fld.Opts.IsArraySliceOfPointers }}
				{{- /* a pointer to an array/slice of pointers */ -}}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx int, value *{{$fld.TypeName}}) {
				// assert v.{{$fld.Name}} != nil && idx >= 0 && idx < len(*v.{{$fld.Name}})
				vv := &((*v.{{$fld.Name}})[idx])
				{{- if $fld.Opts.IsNative }}
					{{- if $fld.Opts.IsString }}
						if *vv != nil {
							v.__alloc.Free(unsafe.Pointer(*vv))
						}
						if value != nil {
							*vv = v.dupStringPtr(*value)
						} else {
							*vv = nil
						}
					{{- else }}
						if value != nil {
							valueSize := unsafe.Sizeof(*value)
							if *vv == nil {
								*vv = (*{{$fld.TypeName}})(v.zeroAlloc(valueSize))
							}
							{{$.AllocatorPkg}}.CopyMem(unsafe.Pointer(*vv), unsafe.Pointer(value), valueSize)
						} else if *vv != nil {
							v.__alloc.Free(unsafe.Pointer(*vv))
							*vv = nil
						}
					{{- end }}
				{{- else }}
					if *vv != nil {
						(*vv).Free()
					}
					*vv = value
				{{- end }}
}
			{{else }}
				{{- if $fld.Opts.IsNative }}
					{{- if $fld.Opts.IsString }}
						{{- /* a pointer to an array/slice of strings (we don't own the string headers) */ -}}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx int, value {{$fld.TypeName}}) {
						// assert v.{{$fld.Name}} != nil && idx >= 0 && idx < len(*v.{{$fld.Name}})
						vv := &((*v.{{$fld.Name}})[idx])
						bytePtr := unsafe.StringData(*vv)
						if bytePtr != nil {
							v.__alloc.Free(unsafe.Pointer(bytePtr))
						}
						*vv = v.dupString(value)
}
					{{- /* else it is an array of things we don't need to handle */ -}}
					{{- end }}
				{{- else }}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx int, value {{$fld.TypeName}}) {
					{{- /* a pointer to an array/slice of non-native objects (they are supposed to be unmanaged too) */ -}}
					// assert v.{{$fld.Name}} != nil && idx >= 0 && idx < len(*v.{{$fld.Name}})
					vv := &((*v.{{$fld.Name}})[idx])
					vv.Free()
					*vv = value
}
				{{- end }}
			{{end }}
		{{- end }}
	{{- else if isArrayOrSlice $fld.Opts.ArraySlice }}
		{{if isSlice $fld.Opts.ArraySlice }}
			{{- /* a slice of something */ -}}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}Capacity(sliceLen int, preserve bool) {
			var newSlice {{$fld.TypeNamePrefixMod}}{{$fld.TypeName}}

			// Create the new slice
			if sliceLen > 0 {
				newSlice = v.allocSlice_{{$fld.FriendlyArrayTypeName}}(sliceLen)
			}

			toPreserve := 0

			// Copy original items if preserve up to the size of the new slice and free the rest
			if v.{{$fld.Name}} != nil {
				oldSliceLen := len(v.{{$fld.Name}})

				// Calculate the number of entries to preserve
				if preserve {
					toPreserve = oldSliceLen
					if toPreserve > sliceLen {
						toPreserve = sliceLen
					}
					for idx := 0; idx < toPreserve; idx++ {
						newSlice[idx] = v.{{$fld.Name}}[idx]
					}
				}

				{{- if or $fld.Opts.IsArraySliceOfPointers (or (not $fld.Opts.IsNative) $fld.Opts.IsString) }}
					// Free unused entries
					for idx := toPreserve; idx < oldSliceLen; idx++ {
						{{- if $fld.Opts.IsArraySliceOfPointers }}
							v.{{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx, nil)
						{{- else if $fld.Opts.IsNative }}
							{{- if $fld.Opts.IsString }}
								v.{{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx, unsafe.String(nil, 0))
							{{- /* else it is pointer to a slice of things we don't need to free */ -}}
							{{- end }}
						{{- else }}
							{{- /* a pointer to a slice of non-native objects (they are supposed to be unmanaged too) */ -}}
							v.{{$fld.Name}}[idx].Free()
						{{- end }}
					}
				{{- end }}

				// Free old slice data
				{{- $c := counter}}
				slicePtr{{$c}} := unsafe.SliceData(v.{{$fld.Name}})
				if slicePtr{{$c}} != nil {
					v.__alloc.Free(unsafe.Pointer(slicePtr{{$c}}))
				}
			}

			{{- if and (not $fld.Opts.IsArraySliceOfPointers) (not $fld.Opts.IsNative) }}
				// Initialize added non-native structs
				for idx := toPreserve; idx < sliceLen; idx++ {
					newSlice[idx].InitAllocator(v.__alloc)
				}
			{{- end }}

			// Replace
			v.{{$fld.Name}} = newSlice
}
		{{- end }}
		{{if $fld.Opts.IsArraySliceOfPointers }}
			{{- /* an array/slice of pointers */ -}}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx int, value *{{$fld.TypeName}}) {
			// assert idx >= 0 && idx < len(v.{{$fld.Name}})
			vv := &(v.{{$fld.Name}}[idx])
			{{- if $fld.Opts.IsNative }}
				{{- if $fld.Opts.IsString }}
					if *vv != nil {
						v.__alloc.Free(unsafe.Pointer(*vv))
					}
					if value != nil {
						*vv = v.dupStringPtr(*value)
					} else {
						*vv = nil
					}
				{{- else }}
					if value != nil {
						valueSize := unsafe.Sizeof(*value)
						if *vv == nil {
							*vv = (*{{$fld.TypeName}})(v.zeroAlloc(valueSize))
						}
						{{$.AllocatorPkg}}.CopyMem(unsafe.Pointer(*vv), unsafe.Pointer(value), valueSize)
					} else if *vv != nil {
						v.__alloc.Free(unsafe.Pointer(*vv))
						*vv = nil
					}
				{{- end }}
			{{- else }}
				if *vv != nil {
					(*vv).Free()
				}
				*vv = value
			{{- end }}
}
		{{- else }}
			{{- /* an array/slice of something */ -}}
			{{- if $fld.Opts.IsNative }}
				{{if $fld.Opts.IsString }}
					{{- /* an array/slice of strings (we don't own the string header) */ -}}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx int, value {{$fld.TypeName}}) {
					// assert idx >= 0 && idx < len(v.{{$fld.Name}})
					vv := &(v.{{$fld.Name}}[idx])
					bytePtr := unsafe.StringData(*vv)
					if bytePtr != nil {
						v.__alloc.Free(unsafe.Pointer(bytePtr))
					}
					*vv = v.dupString(value)
}
				{{- /* else it is an array of things we don't need to handle */ -}}
				{{- end }}
			{{else }}
				{{- /* an array/slice of non-native objects (they are supposed to be unmanaged too) */ -}}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}(idx int, value {{$fld.TypeName}}) {
				// assert idx >= 0 && idx < len(v.{{$fld.Name}})
				vv := &(v.{{$fld.Name}}[idx])
				vv.Free()
				*vv = value
}
			{{- end }}
		{{- end }}
	{{- else if $fld.Opts.IsNative }}
		{{if $fld.Opts.IsString }}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}(value {{$fld.TypeName}}) {
			bytePtr := unsafe.StringData(v.{{$fld.Name}})
			if bytePtr != nil {
				v.__alloc.Free(unsafe.Pointer(bytePtr))
			}
			v.{{$fld.Name}} = v.dupString(value)
}
		{{- end }}
	{{else }}
		{{- /* a non-native objects (it is supposed to be unmanaged too) */ -}}
func (v *{{$.StructName}}) {{$fld.SetFuncPrefix}}{{$fld.FuncName}}(value {{$fld.TypeName}}) {
		v.{{$fld.Name}}.Free()
		v.{{$fld.Name}} = value
}
	{{- end }}
{{- end }}

{{range $key, $value := .NeedAllocSlice }}
func (v *{{$.StructName}}) allocSlice_{{$key}}(sliceLen int) []{{$value}} {
	var tempT {{$value}}

	memSize, overflow := {{$.AllocatorPkg}}.MulUintptr(unsafe.Sizeof(tempT), uintptr(sliceLen))
	if overflow || sliceLen < 0 {
		panic("{{$.StructName}}::allocSlice[{{$value}}]: size out of range")
	}

	data := v.zeroAlloc(memSize)
	destSlice := unsafe.Slice((*{{$value}})(data), sliceLen)
	return destSlice
}

func (v *{{$.StructName}}) dupSlice_{{$key}}(src []{{$value}}) []{{$value}} {
	var tempT {{$value}}

	arrLen := len(src)
	destSlice := v.allocSlice_{{$key}}(arrLen)
	memSize, _ := {{$.AllocatorPkg}}.MulUintptr(unsafe.Sizeof(tempT), uintptr(arrLen))
	{{$.AllocatorPkg}}.CopyMem(unsafe.Pointer(unsafe.SliceData(destSlice)), unsafe.Pointer(unsafe.SliceData(src)), memSize)
	return destSlice
}
{{- end }}

{{range $key, $value := .NeedAllocSlicePtr }}
func (v *{{$.StructName}}) allocSlicePtr_{{$key}}(sliceLen int) *[]{{$value}} {
	var tempT {{$value}}
	var memSize uintptr

	dataSize, overflow := {{$.AllocatorPkg}}.MulUintptr(unsafe.Sizeof(tempT), uintptr(sliceLen))
	if overflow || sliceLen < 0 {
		panic("{{$.StructName}}::allocSlicePtr[{{$value}}]: size out of range")
	}

	hdrSize := unsafe.Sizeof([]{{$value}}{})
	memSize, overflow = {{$.AllocatorPkg}}.AddUintptr(dataSize, hdrSize)
	if overflow {
		panic("{{$.StructName}}::allocSlicePtr[{{$value}}]: size out of range")
	}
	ptr := v.zeroAlloc(memSize)
	data := unsafe.Add(ptr, hdrSize)
	tmpSlice := unsafe.Slice((*{{$value}})(data), sliceLen)
	{{$.AllocatorPkg}}.CopyMem(ptr, unsafe.Pointer(&tmpSlice), hdrSize)
	return (*[]{{$value}})(ptr)
}

func (v *{{$.StructName}}) dupSlicePtr_{{$key}}(src []{{$value}}) *[]{{$value}} {
	var tempT {{$value}}

	sliceLen := len(src)
	destSlice := v.allocSlicePtr_{{$key}}(sliceLen)
	memSize, _ := {{$.AllocatorPkg}}.MulUintptr(unsafe.Sizeof(tempT), uintptr(sliceLen))
	{{$.AllocatorPkg}}.CopyMem(unsafe.Pointer(unsafe.SliceData(*destSlice)), unsafe.Pointer(unsafe.SliceData(src)), memSize)
	return destSlice
}
{{- end }}

{{- range $siz, $item := .NeedAllocArrayPtr }}
{{range $key, $value := $item }}
func (v *{{$.StructName}}) allocArrayPtr_{{$siz}}() *[{{$key}}]{{$value}} {
	var tempT {{$value}}

	memSize, overflow := {{$.AllocatorPkg}}.MulUintptr(unsafe.Sizeof(tempT), uintptr({{$key}}))
	if overflow {
		panic("{{$.StructName}}::allocArrayPtr[[{{$key}}]{{$value}}]: size out of range")
	}
	ptr := v.zeroAlloc(memSize)
	return (*[{{$key}}]{{$value}})(ptr)
}
{{- end }}
{{- end }}

{{if .NeedAllocString }}
func (v *{{$.StructName}}) allocString(strLen int) string {
	if strLen == 0 {
		return unsafe.String(nil, 0)
	}
	data := v.zeroAlloc(uintptr(strLen))
	destStr := unsafe.String((*byte)(data), strLen)
	return destStr
}

func (v *{{$.StructName}}) dupString(s string) string {
	strLen := len(s)
	destStr := v.allocString(strLen)
	{{$.AllocatorPkg}}.CopyMem(unsafe.Pointer(unsafe.StringData(destStr)), unsafe.Pointer(unsafe.StringData(s)), uintptr(strLen))
	return destStr
}
{{- end }}

{{if .NeedAllocStringPtr }}
func (v *{{$.StructName}}) allocStringPtr(strLen int) *string {
	hdrSize := unsafe.Sizeof("")
	memSize, overflow := {{.AllocatorPkg}}.AddUintptr(uintptr(strLen), hdrSize)
	if overflow {
		panic("{{$.StructName}}::allocStringPtr: size out of range")
	}
	ptr := v.zeroAlloc(memSize)
	data := unsafe.Add(ptr, hdrSize)
	*((*string)(ptr)) = unsafe.String((*byte)(data), strLen)
	return (*string)(ptr)
}

func (v *{{$.StructName}}) dupStringPtr(s string) *string {
	strLen := len(s)
	destStr := v.allocStringPtr(strLen)
	{{$.AllocatorPkg}}.CopyMem(unsafe.Pointer(unsafe.StringData(*destStr)), unsafe.Pointer(unsafe.StringData(s)), uintptr(strLen))
	return destStr
}
{{- end }}
`, funcMap, setter)
	if err != nil {
		return err
	}

	// Done
	return nil
}

func friendlyArraySize(arraySliceSize string) string {
	if _, err := strconv.ParseInt(arraySliceSize, 10, 64); err != nil {
		h := fnv.New32a()
		h.Sum([]byte(arraySliceSize))
		return "_" + strings.ToUpper(strconv.FormatInt(int64(h.Sum32()), 16))
	}
	return arraySliceSize
}

func friendlyArrayTypeName(typeName string) string {
	isPointer := false
	if strings.HasPrefix(typeName, "*") {
		isPointer = true
		typeName = typeName[1:]
	}
	l := len(typeName)
	for idx := 0; idx < l; idx++ {
		if (typeName[idx] < 'A' || typeName[idx] > 'Z') && (typeName[idx] < 'a' || typeName[idx] > 'z') && (typeName[idx] < '0' || typeName[idx] > '9') && typeName[idx] != '_' {
			h := fnv.New32a()
			if isPointer {
				h.Sum([]byte("*"))
			}
			h.Sum([]byte(typeName))
			return "_" + strings.ToUpper(strconv.FormatInt(int64(h.Sum32()), 16))
		}
	}
	if isPointer {
		return "PtrTo" + capitalizeFirstLetter(typeName)
	}
	return typeName
}
