package processor

import (
	"fmt"
	"strings"

	parser "github.com/mxmauro/gofile-parser"
	"github.com/mxmauro/unmanagedgen/generator"
)

// -----------------------------------------------------------------------------

func (proc *Processor) processStruct(psName string, ps *parser.ParsedStruct) error {
	var gs *generator.Struct

	for _, field := range ps.Fields {
		if tag, ok := field.Tags.GetTag("unmanaged"); ok {
			if tag.GetBoolProperty("omit") {
				continue
			}
		}

		if gs == nil {
			gs = proc.gen.AddStruct(generator.UnmanagedName(psName))
		}

		fieldNames := field.Names
		if len(fieldNames) == 0 {
			fieldNames = []string{field.ImplicitName}
		}

		fieldOpts := generator.FieldOptions{}

		switch fType := field.Type.(type) {
		case *parser.ParsedNativeType:
			fieldOpts.IsNative = true
			gs.AddField(fieldNames, fType.Name, field.Tags, fieldOpts)

		case *parser.ParsedNonNativeType:
			gs.AddField(fieldNames, fType.Name, field.Tags, fieldOpts)

		case *parser.ParsedStruct:
			return fmt.Errorf("[%v/%v] inline struct fields are not supported", psName, strings.Join(fieldNames, ","))

		case *parser.ParsedInterface:
			return fmt.Errorf("[%v/%v] inline interface fields are not supported", psName, strings.Join(fieldNames, ","))

		case *parser.ParsedMap:
			return fmt.Errorf("[%v/%v] map fields are not supported", psName, strings.Join(fieldNames, ","))

		case *parser.ParsedArray:
			if fType.Size == "..." {
				return fmt.Errorf("[%v/%v] ellipsis array fields are not supported", psName, strings.Join(fieldNames, ","))
			}
			fieldOpts.ArraySlice = &fType.Size

			switch fValueType := fType.ValueType.(type) {
			case *parser.ParsedNativeType:
				fieldOpts.IsNative = true
				gs.AddField(fieldNames, fValueType.Name, field.Tags, fieldOpts)

			case *parser.ParsedNonNativeType:
				gs.AddField(fieldNames, fValueType.Name, field.Tags, fieldOpts)

			case *parser.ParsedStruct:
				return fmt.Errorf("[%v/%v] arrays of inline structs are not supported", psName, strings.Join(fieldNames, ","))

			case *parser.ParsedInterface:
				return fmt.Errorf("[%v/%v] arrays of inline interface fields are not supported", psName, strings.Join(fieldNames, ","))

			case *parser.ParsedMap:
				return fmt.Errorf("[%v/%v] arrays of inline map fields are not supported", psName, strings.Join(fieldNames, ","))

			case *parser.ParsedArray:
				return fmt.Errorf("[%v/%v] multidimensional array fields are not supported", psName, strings.Join(fieldNames, ","))

			case *parser.ParsedPointer:
				fieldOpts.IsArraySliceOfPointers = true

				switch fToType := fValueType.ToType.(type) {
				case *parser.ParsedNativeType:
					fieldOpts.IsNative = true
					gs.AddField(fieldNames, fToType.Name, field.Tags, fieldOpts)

				case *parser.ParsedNonNativeType:
					gs.AddField(fieldNames, fToType.Name, field.Tags, fieldOpts)

				default:
					return fmt.Errorf("[%v/%v] unsupported array of pointers field type", psName, strings.Join(fieldNames, ","))
				}

			default:
				return fmt.Errorf("[%v/%v] unsupported array field type", psName, strings.Join(fieldNames, ","))
			}

		case *parser.ParsedPointer:
			fieldOpts.IsPointer = true
			switch fToType := fType.ToType.(type) {
			case *parser.ParsedNativeType:
				fieldOpts.IsNative = true
				gs.AddField(fieldNames, fToType.Name, field.Tags, fieldOpts)

			case *parser.ParsedNonNativeType:
				gs.AddField(fieldNames, fToType.Name, field.Tags, fieldOpts)

			case *parser.ParsedStruct:
				return fmt.Errorf("[%v/%v] pointers to inline struct fields are not supported", psName, strings.Join(fieldNames, ","))

			case *parser.ParsedInterface:
				return fmt.Errorf("[%v/%v] pointers to inline interface fields are not supported", psName, strings.Join(fieldNames, ","))

			case *parser.ParsedMap:
				return fmt.Errorf("[%v/%v] pointers to inline map fields are not supported", psName, strings.Join(fieldNames, ","))

			case *parser.ParsedArray:
				if fToType.Size == "..." {
					return fmt.Errorf("[%v/%v] pointers to ellipsis array fields are not supported", psName, strings.Join(fieldNames, ","))
				}
				fieldOpts.ArraySlice = &fToType.Size

				switch fValueType := fToType.ValueType.(type) {
				case *parser.ParsedNativeType:
					fieldOpts.IsNative = true
					gs.AddField(fieldNames, fValueType.Name, field.Tags, fieldOpts)

				case *parser.ParsedNonNativeType:
					gs.AddField(fieldNames, fValueType.Name, field.Tags, fieldOpts)

				case *parser.ParsedPointer:
					fieldOpts.IsArraySliceOfPointers = true

					switch fToType := fValueType.ToType.(type) {
					case *parser.ParsedNativeType:
						fieldOpts.IsNative = true
						gs.AddField(fieldNames, fToType.Name, field.Tags, fieldOpts)

					case *parser.ParsedNonNativeType:
						gs.AddField(fieldNames, fToType.Name, field.Tags, fieldOpts)

					default:
						return fmt.Errorf("[%v/%v] unsupported pointers to array of pointers field type", psName, strings.Join(fieldNames, ","))
					}

				default:
					return fmt.Errorf("[%v/%v] unsupported pointer to array field type", psName, strings.Join(fieldNames, ","))
				}

			case *parser.ParsedPointer:
				return fmt.Errorf("[%v/%v] double pointer fields not supported", psName, strings.Join(fieldNames, ","))

			default:
				return fmt.Errorf("[%v/%v] unsupported pointer field type", psName, strings.Join(fieldNames, ","))
			}

		default:
			return fmt.Errorf("[%v/%v] unsupported field type", psName, strings.Join(fieldNames, ","))
		}
	}

	return nil
}
