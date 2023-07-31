package console

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

func TestConsole(t *testing.T) {
	vm := goja.New()

	new(require.Registry).Enable(vm)
	Enable(vm)

	if c := vm.Get("console"); c == nil {
		t.Fatal("console not found")
	}

	if _, err := vm.RunString("console.log('')"); err != nil {
		t.Fatal("console.log() error", err)
	}

	if _, err := vm.RunString("console.error('')"); err != nil {
		t.Fatal("console.error() error", err)
	}

	if _, err := vm.RunString("console.warn('')"); err != nil {
		t.Fatal("console.warn() error", err)
	}

	if _, err := vm.RunString("console.info('')"); err != nil {
		t.Fatal("console.info() error", err)
	}

	if _, err := vm.RunString("console.debug('')"); err != nil {
		t.Fatal("console.debug() error", err)
	}
}

func TestConsoleWithPrinter(t *testing.T) {
	var stdoutStr, stderrStr string

	printer := StdPrinter{
		StdoutPrint: func(s string) { stdoutStr += s },
		StderrPrint: func(s string) { stderrStr += s },
	}

	vm := goja.New()

	registry := new(require.Registry)
	registry.Enable(vm)
	registry.RegisterNativeModule(ModuleName, RequireWithPrinter(printer))
	Enable(vm)

	if c := vm.Get("console"); c == nil {
		t.Fatal("console not found")
	}

	_, err := vm.RunString(`
		console.log('a')
		console.error('b')
		console.warn('c')
		console.debug('d')
		console.info('e')
	`)
	if err != nil {
		t.Fatal(err)
	}

	if want := "ade"; stdoutStr != want {
		t.Fatalf("Unexpected stdout output: got %q, want %q", stdoutStr, want)
	}

	if want := "bc"; stderrStr != want {
		t.Fatalf("Unexpected stderr output: got %q, want %q", stderrStr, want)
	}
}
