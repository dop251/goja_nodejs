package require

import (
	"testing"

	js "github.com/dop251/goja"
)

func TestRequire(t *testing.T) {
	const SCRIPT = `
	var m = require("./testdata/m.js");
	m.test();
	`

	vm := js.New()

	registry := new(Registry)
	registry.Enable(vm)

	v, err := vm.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	if !v.StrictEquals(vm.ToValue("passed")) {
		t.Fatalf("Unexpected result: %v", v)
	}
}
