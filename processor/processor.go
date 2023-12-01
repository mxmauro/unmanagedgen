package processor

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mxmauro/FastGlobbing"
	parser "github.com/mxmauro/gofile-parser"
	"github.com/mxmauro/unmanagedgen/generator"
)

// -----------------------------------------------------------------------------

type Processor struct {
	pf  *parser.ParsedFile
	gen *generator.Generator
}

// -----------------------------------------------------------------------------

// ProcessFolder process all go files that matches the provided file-mask
func ProcessFolder(fileMask string, cb func(filename string)) error {
	var baseDir string
	var err error

	baseDir, fileMask, err = getBaseDirAndMask(fileMask)
	if err != nil {
		return err
	}

	vendorMask := string(os.PathSeparator) + "vendor" + string(os.PathSeparator)
	matcher := FastGlobbing.NewGitWildcardMatcher(baseDir + fileMask)
	return filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("unable to access path '%v' [err=%v]", path, err.Error())
		}
		if info.IsDir() || strings.Contains(path, vendorMask) {
			return nil
		}
		if (!strings.HasSuffix(path, ".go")) || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_unmanaged.go") {
			return nil
		}
		if matcher.Test(path) {
			if cb != nil {
				cb(path)
			}
			err = ProcessFile(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// ProcessFile process the specified file
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

// -----------------------------------------------------------------------------

func getBaseDirAndMask(fileMask string) (string, string, error) {
	if !filepath.IsAbs(fileMask) {
		d, err := os.Getwd()
		if err != nil {
			return "", "", fmt.Errorf("unable to query current directory [err=%v]", err.Error())
		}
		if !strings.HasSuffix(d, string(os.PathSeparator)) {
			d += string(os.PathSeparator)
		}
		fileMask = d + fileMask
	}

	fileMask = filepath.Clean(fileMask)

	lastSlash := -1
	fileMaskRunes := []rune(fileMask)
runeScan:
	for idx, r := range fileMaskRunes {
		switch r {
		case os.PathSeparator:
			lastSlash = idx

		case '[':
			fallthrough
		case '*':
			fallthrough
		case '?':
			break runeScan
		}
	}
	if lastSlash < 0 {
		return "", "", errors.New("unable to determine base directory")
	}
	// Done
	return string(fileMaskRunes[0:(lastSlash + 1)]), string(fileMaskRunes[(lastSlash + 1):]), nil
}
