package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mxmauro/FastGlobbing"
	"github.com/mxmauro/unmanagedgen/processor"
	"github.com/spf13/cobra"
)

// -----------------------------------------------------------------------------

var settings struct {
	FileMask string
}

// -----------------------------------------------------------------------------

func main() {
	// Create command line parser/executor
	rootCmd := &cobra.Command{
		Use:   "unmanagedgen -f filemask",
		Short: "Generate allocator helpers for defined structs using an external allocator such as C.malloc/C.free",
		Run: func(cmd *cobra.Command, args []string) {
			exitCode := runGenerator()
			if exitCode != 0 {
				os.Exit(exitCode)
			}
		},
	}

	rootCmd.SilenceUsage = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.Flags().StringVarP(&settings.FileMask, "file", "F", "", "File specification containing definitions to process")
	_ = rootCmd.MarkFlagRequired("file")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runGenerator() int {
	rootDir, fileMask, b := getRootDir()
	if !b {
		return 1
	}

	vendorMask := string(os.PathSeparator) + "vendor" + string(os.PathSeparator)
	matcher := FastGlobbing.NewGitWildcardMatcher(rootDir + fileMask)
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error: Unable to access path '%v' [err=%v]\n", path, err.Error())
			return err
		}
		if info.IsDir() || strings.Contains(path, vendorMask) {
			return nil
		}
		if (!strings.HasSuffix(path, ".go")) || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_unmanaged.go") {
			return nil
		}
		if matcher.Test(path) {
			fmt.Printf("Processing: %v...", path)
			err = processor.ProcessFile(path)
			if err != nil {
				fmt.Printf(" Error: %v\n", err.Error())
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 1
	}

	// Done
	return 0
}

func getRootDir() (string, string, bool) {
	fileMask := settings.FileMask
	if !filepath.IsAbs(settings.FileMask) {
		d, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error: Unable to query current directory. [err=%v]", err.Error())
			return "", "", false
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
		fmt.Printf("Internal error: Unable to determine root dir.")
		return "", "", false
	}
	// Done
	return string(fileMaskRunes[0:(lastSlash + 1)]), string(fileMaskRunes[(lastSlash + 1):]), true
}
