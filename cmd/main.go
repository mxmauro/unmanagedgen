package main

import (
	"fmt"
	"os"

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
		Use:   "unmanagedgen --file filemask",
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
	err := processor.ProcessFolder(settings.FileMask, func(filename string) {
		fmt.Printf("Processing: %v...", filename)
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		return 1
	}
	return 0
}
