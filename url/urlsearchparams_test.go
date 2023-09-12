package url

import (
	_ "embed"
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
)

func createVM() *goja.Runtime {
	vm := goja.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	Enable(vm)
	return vm
}

func TestURLSearchParams(t *testing.T) {
	vm := createVM()

	if c := vm.Get("URLSearchParams"); c == nil {
		t.Fatal("URLSearchParams not found")
	}

	script := `const params = new URLSearchParams();`

	if _, err := vm.RunString(script); err != nil {
		t.Fatal("Failed to process url script.", err)
	}
}

//go:embed testdata/url_search_params.js
var url_search_params string

func TestURLSearchParameters(t *testing.T) {
	vm := createVM()

	if c := vm.Get("URLSearchParams"); c == nil {
		t.Fatal("URLSearchParams not found")
	}

	// Script will throw an error on failed validation

	_, err := vm.RunScript("testdata/url_search_params.js", url_search_params)
	if err != nil {
		if ex, ok := err.(*goja.Exception); ok {
			t.Fatal(ex.String())
		}
		t.Fatal("Failed to process url script.", err)
	}
}
