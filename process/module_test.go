package process

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

func TestProcessEnvStructure(t *testing.T) {
	vm := goja.New()

	new(require.Registry).Enable(vm)
	Enable(vm)

	if c := vm.Get("process"); c == nil {
		t.Fatal("process not found")
	}

	if c, err := vm.RunString("process.env"); c == nil || err != nil {
		t.Fatal("error accessing process.env")
	}
}

func TestProcessEnvValuesArtificial(t *testing.T) {
	os.Setenv("GOJA_IS_AWESOME", "true")
	defer os.Unsetenv("GOJA_IS_AWESOME")

	vm := goja.New()

	new(require.Registry).Enable(vm)
	Enable(vm)

	jsRes, err := vm.RunString("process.env['GOJA_IS_AWESOME']")

	if err != nil {
		t.Fatal(fmt.Sprintf("Error executing: %s", err))
	}

	if jsRes.String() != "true" {
		t.Fatal(fmt.Sprintf("Error executing: got %s but expected %s", jsRes, "true"))
	}
}

func TestProcessEnvValuesBrackets(t *testing.T) {
	vm := goja.New()

	new(require.Registry).Enable(vm)
	Enable(vm)

	for _, e := range os.Environ() {
		envKeyValue := strings.SplitN(e, "=", 2)
		jsExpr := fmt.Sprintf("process.env['%s']", envKeyValue[0])

		jsRes, err := vm.RunString(jsExpr)

		if err != nil {
			t.Fatal(fmt.Sprintf("Error executing %s: %s", jsExpr, err))
		}

		if jsRes.String() != envKeyValue[1] {
			t.Fatal(fmt.Sprintf("Error executing %s: got %s but expected %s", jsExpr, jsRes, envKeyValue[1]))
		}
	}
}
