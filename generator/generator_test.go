package generator_test

import (
	"bufio"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/mxmauro/unmanagedgen/processor"
)

// -----------------------------------------------------------------------------

func TestSample1(t *testing.T) {
	// Compile the sample1 code
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "../testdata/sample1/structs.go")
	t.Logf("Gnerating code: %v...", path)
	err := processor.ProcessFile(path)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Running Sample1 test")
	cmd := exec.Command("go", "test", "github.com/mxmauro/unmanagedgen/testdata/sample1")
	cmd.Dir = filepath.Join(filepath.Dir(filename), "..")
	err = runCmd(t, cmd)
	if err != nil {
		t.Fatal(err)
	}
}

func runCmd(t *testing.T, cmd *exec.Cmd) error {
	var cmdStdErr io.ReadCloser

	cmdStdOut, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmdStdErr, err = cmd.StderrPipe()
	if err != nil {
		_ = cmdStdOut.Close()
		return err
	}

	wg := sync.WaitGroup{}
	processOutput := func(reader io.ReadCloser) {
		r := bufio.NewReader(reader)
		defer func() {
			_ = reader.Close()
		}()

		for {
			line, _, err := r.ReadLine()
			if err != nil {
				break
			}
			if len(line) > 0 {
				t.Log(string(line))
			}
		}
		wg.Done()
	}

	wg.Add(2)
	go processOutput(cmdStdOut)
	go processOutput(cmdStdErr)

	err = cmd.Run()
	wg.Wait()
	return err
}
