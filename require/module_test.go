package require

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"testing"

	js "github.com/dop251/goja"
)

func mapFileSystemSourceLoader(files map[string]string) SourceLoader {
	return func(path string) ([]byte, error) {
		s, ok := files[path]
		if !ok {
			return nil, ModuleFileDoesNotExistError
		}
		return []byte(s), nil
	}
}

func TestRequireNativeModule(t *testing.T) {
	const SCRIPT = `
	var m = require("test/m");
	m.test();
	`

	vm := js.New()

	registry := new(Registry)
	registry.Enable(vm)

	RegisterNativeModule("test/m", func(runtime *js.Runtime, module *js.Object) {
		o := module.Get("exports").(*js.Object)
		o.Set("test", func(call js.FunctionCall) js.Value {
			return runtime.ToValue("passed")
		})
	})

	v, err := vm.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	if !v.StrictEquals(vm.ToValue("passed")) {
		t.Fatalf("Unexpected result: %v", v)
	}
}

func TestRequireRegistryNativeModule(t *testing.T) {
	const SCRIPT = `
	var log = require("test/log");
	log.print('passed');
	`

	logWithOutput := func(w io.Writer, prefix string) ModuleLoader {
		return func(vm *js.Runtime, module *js.Object) {
			o := module.Get("exports").(*js.Object)
			o.Set("print", func(call js.FunctionCall) js.Value {
				fmt.Fprint(w, prefix, call.Argument(0).String())
				return js.Undefined()
			})
		}
	}

	vm1 := js.New()
	buf1 := &bytes.Buffer{}

	registry1 := new(Registry)
	registry1.Enable(vm1)

	registry1.RegisterNativeModule("test/log", logWithOutput(buf1, "vm1 "))

	vm2 := js.New()
	buf2 := &bytes.Buffer{}

	registry2 := new(Registry)
	registry2.Enable(vm2)

	registry2.RegisterNativeModule("test/log", logWithOutput(buf2, "vm2 "))

	_, err := vm1.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	s := buf1.String()
	if s != "vm1 passed" {
		t.Fatalf("vm1: Unexpected result: %q", s)
	}

	_, err = vm2.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	s = buf2.String()
	if s != "vm2 passed" {
		t.Fatalf("vm2: Unexpected result: %q", s)
	}
}

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

func TestSourceLoader(t *testing.T) {
	const SCRIPT = `
	var m = require("m.js");
	m.test();
	`

	const MODULE = `
	function test() {
		return "passed1";
	}

	exports.test = test;
	`

	vm := js.New()

	registry := NewRegistry(WithGlobalFolders("."), WithLoader(func(name string) ([]byte, error) {
		if name == "m.js" {
			return []byte(MODULE), nil
		}
		return nil, errors.New("Module does not exist")
	}))
	registry.Enable(vm)

	v, err := vm.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	if !v.StrictEquals(vm.ToValue("passed1")) {
		t.Fatalf("Unexpected result: %v", v)
	}
}

func TestStrictModule(t *testing.T) {
	const SCRIPT = `
	var m = require("m.js");
	m.test();
	`

	const MODULE = `
	"use strict";

	function test() {
		var a = "passed1";
		eval("var a = 'not passed'");
		return a;
	}

	exports.test = test;
	`

	vm := js.New()

	registry := NewRegistry(WithGlobalFolders("."), WithLoader(func(name string) ([]byte, error) {
		if name == "m.js" {
			return []byte(MODULE), nil
		}
		return nil, errors.New("Module does not exist")
	}))
	registry.Enable(vm)

	v, err := vm.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	if !v.StrictEquals(vm.ToValue("passed1")) {
		t.Fatalf("Unexpected result: %v", v)
	}
}

func TestResolve(t *testing.T) {
	testRequire := func(src, fpath string, globalFolders []string, fs map[string]string) (*js.Runtime, js.Value, error) {
		vm := js.New()
		r := NewRegistry(WithGlobalFolders(globalFolders...), WithLoader(mapFileSystemSourceLoader(fs)))
		r.Enable(vm)
		t.Logf("Require(%s)", fpath)
		ret, err := vm.RunScript(path.Join(src, "test.js"), fmt.Sprintf("require('%s')", fpath))
		if err != nil {
			return nil, nil, err
		}
		return vm, ret, nil
	}

	globalFolders := []string{
		"/usr/lib/node_modules",
		"/home/src/.node_modules",
	}

	fs := map[string]string{
		"/home/src/app/app.js":                   `exports.name = "app"`,
		"/home/src/app2/app2.json":               `{"name": "app2"}`,
		"/home/src/app3/index.js":                `exports.name = "app3"`,
		"/home/src/app4/index.json":              `{"name": "app4"}`,
		"/home/src/app5/package.json":            `{"main": "app5.js"}`,
		"/home/src/app5/app5.js":                 `exports.name = "app5"`,
		"/home/src/app6/package.json":            `{"main": "."}`,
		"/home/src/app6/index.js":                `exports.name = "app6"`,
		"/home/src/app7/package.json":            `{"main": "./a/b/c/file.js"}`,
		"/home/src/app7/a/b/c/file.js":           `exports.name = "app7"`,
		"/usr/lib/node_modules/app8":             `exports.name = "app8"`,
		"/home/src/app9/app9.js":                 `exports.name = require('./a/file.js').name`,
		"/home/src/app9/a/file.js":               `exports.name = require('./b/file.js').name`,
		"/home/src/app9/a/b/file.js":             `exports.name = require('./c/file.js').name`,
		"/home/src/app9/a/b/c/file.js":           `exports.name = "app9"`,
		"/home/src/.node_modules/app10":          `exports.name = "app10"`,
		"/home/src/app11/app11.js":               `exports.name = require('d/file.js').name`,
		"/home/src/app11/a/b/c/app11.js":         `exports.name = require('d/file.js').name`,
		"/home/src/app11/node_modules/d/file.js": `exports.name = "app11"`,
		"/app12.js":                              `exports.name = require('a/file.js').name`,
		"/node_modules/a/file.js":                `exports.name = "app12"`,
		"/app13/app13.js":                        `exports.name = require('b/file.js').name`,
		"/node_modules/b/file.js":                `exports.name = "app13"`,
		"node_modules/app14/index.js":            `exports.name = "app14"`,
		"../node_modules/app15/index.js":         `exports.name = "app15"`,
	}

	for i, tc := range []struct {
		src   string
		path  string
		ok    bool
		field string
		value string
	}{
		{"/home/src", "./app/app", true, "name", "app"},
		{"/home/src", "./app/app.js", true, "name", "app"},
		{"/home/src", "./app/bad.js", false, "", ""},
		{"/home/src", "./app2/app2", true, "name", "app2"},
		{"/home/src", "./app2/app2.json", true, "name", "app2"},
		{"/home/src", "./app/bad.json", false, "", ""},
		{"/home/src", "./app3", true, "name", "app3"},
		{"/home/src", "./appbad", false, "", ""},
		{"/home/src", "./app4", true, "name", "app4"},
		{"/home/src", "./appbad", false, "", ""},
		{"/home/src", "./app5", true, "name", "app5"},
		{"/home/src", "./app6", true, "name", "app6"},
		{"/home/src", "./app7", true, "name", "app7"},
		{"/home/src", "app8", true, "name", "app8"},
		{"/home/src", "./app9/app9", true, "name", "app9"},
		{"/home/src", "app10", true, "name", "app10"},
		{"/home/src", "./app11/app11.js", true, "name", "app11"},
		{"/home/src", "./app11/a/b/c/app11.js", true, "name", "app11"},
		{"/", "./app12", true, "name", "app12"},
		{"/", "./app13/app13", true, "name", "app13"},
		{".", "app14", true, "name", "app14"},
		{"..", "nonexistent", false, "", ""},
	} {
		vm, mod, err := testRequire(tc.src, tc.path, globalFolders, fs)
		if err != nil {
			if tc.ok {
				t.Errorf("%d: require() failed: %v", i, err)
			}
			continue
		} else {
			if !tc.ok {
				t.Errorf("%d: expected to fail, but did not", i)
				continue
			}
		}
		f := mod.ToObject(vm).Get(tc.field)
		if f == nil {
			t.Errorf("%v: field %q not found", i, tc.field)
			continue
		}
		value := f.String()
		if value != tc.value {
			t.Errorf("%v: got %q expected %q", i, value, tc.value)
		}
	}
}

func TestRequireCycle(t *testing.T) {
	vm := js.New()
	r := NewRegistry(WithLoader(mapFileSystemSourceLoader(map[string]string{
		"a.js": `var b = require('./b.js'); exports.done = true;`,
		"b.js": `var a = require('./a.js'); exports.done = true;`,
	})))
	r.Enable(vm)
	res, err := vm.RunString(`
	var a = require('./a.js');
	var b = require('./b.js');
	a.done && b.done;
	`)
	if err != nil {
		t.Fatal(err)
	}
	if v := res.Export(); v != true {
		t.Fatalf("Unexpected result: %v", v)
	}
}

func TestErrorPropagation(t *testing.T) {
	vm := js.New()
	r := NewRegistry(WithLoader(mapFileSystemSourceLoader(map[string]string{
		"m.js": `throw 'test passed';`,
	})))
	rr := r.Enable(vm)
	_, err := rr.Require("./m")
	if err == nil {
		t.Fatal("Expected an error")
	}
	if ex, ok := err.(*js.Exception); ok {
		if !ex.Value().StrictEquals(vm.ToValue("test passed")) {
			t.Fatalf("Unexpected Exception: %v", ex)
		}
	} else {
		t.Fatal(err)
	}
}

func TestSourceMapLoader(t *testing.T) {
	vm := js.New()
	r := NewRegistry(WithLoader(func(p string) ([]byte, error) {
		switch p {
		case "dir/m.js":
			return []byte(`throw 'test passed';
//# sourceMappingURL=m.js.map`), nil
		case "dir/m.js.map":
			return []byte(`{"version":3,"file":"m.js","sourceRoot":"","sources":["m.ts"],"names":[],"mappings":";AAAA"}
`), nil
		}
		return nil, ModuleFileDoesNotExistError
	}))

	rr := r.Enable(vm)
	_, err := rr.Require("./dir/m")
	if err == nil {
		t.Fatal("Expected an error")
	}
	if ex, ok := err.(*js.Exception); ok {
		if !ex.Value().StrictEquals(vm.ToValue("test passed")) {
			t.Fatalf("Unexpected Exception: %v", ex)
		}
	} else {
		t.Fatal(err)
	}
}

func testsetup() (string, func(), error) {
	name, err := os.MkdirTemp("", "goja-nodejs-require-test")
	if err != nil {
		return "", nil, err
	}
	return name, func() {
		os.RemoveAll(name)
	}, nil
}

func TestDefaultModuleLoader(t *testing.T) {
	workdir, teardown, err := testsetup()
	if err != nil {
		t.Fatal(err)
	}
	defer teardown()

	err = os.Chdir(workdir)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir("module", 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile("module/index.js", []byte(`throw 'test passed';`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	vm := js.New()
	r := NewRegistry()
	rr := r.Enable(vm)
	_, err = rr.Require("./module")
	if err == nil {
		t.Fatal("Expected an error")
	}
	if ex, ok := err.(*js.Exception); ok {
		if !ex.Value().StrictEquals(vm.ToValue("test passed")) {
			t.Fatalf("Unexpected Exception: %v", ex)
		}
	} else {
		t.Fatal(err)
	}
}
