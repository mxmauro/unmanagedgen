package generator

import (
	"fmt"

	parser "github.com/mxmauro/gofile-parser"
)

// -----------------------------------------------------------------------------

func (sc *SaveContext) WriteImports() error {
	sc.allocatorPkg = sc.gen.NextId() + "_alloc"

	sc.WriteLine("import (")
	sc.WriteLine("\"unsafe\"")
	sc.WriteLine("")
	sc.WriteLine("%v \"github.com/mxmauro/unmanagedgen/allocator\"", sc.allocatorPkg)

	// Create a list of used package names
	pkgNames := make([]string, 0)
	for _, st := range sc.gen.structs {
		for _, fld := range st.fields {
			pkgName, _ := parser.GetIdentifierParts(fld.typeName)
			if len(pkgName) > 0 {
				found := false
				for _, name := range pkgNames {
					if name == pkgName {
						found = true
						break
					}
				}
				if !found {
					pkgNames = append(pkgNames, pkgName)
				}
			}
		}
	}

	first := true
	for _, pkgName := range pkgNames {
		pkgImpIndex := -1
		for impIdx, imp := range sc.gen.imports {
			if pkgName == imp.PackageName() {
				pkgImpIndex = impIdx
				break
			}
		}
		if pkgImpIndex < 0 {
			return fmt.Errorf("unable to find import of package '%v'", pkgName)
		}

		imp := &sc.gen.imports[pkgImpIndex]

		impName := ""
		if len(imp.Name) > 0 {
			impName = imp.Name + " "
		}

		if first {
			sc.WriteLine("")
			first = false
		}
		sc.WriteLine("%v\"%v\"", impName, imp.Path)
	}

	sc.WriteLine(")")

	// Done
	return nil
}
