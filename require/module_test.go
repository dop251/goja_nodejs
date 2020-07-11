package require

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	js "github.com/dop251/goja"
)

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
				fmt.Fprint(w, prefix, call.Argument(0).ToString())
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
	mapFileSystemSourceLoader := func(files map[string]string) SourceLoader {
		return func(path string) ([]byte, error) {
			slashPath := filepath.ToSlash(path)
			t.Logf("SourceLoader(%s) [%s]", path, slashPath)
			s, ok := files[filepath.ToSlash(slashPath)]
			if !ok {
				return nil, InvalidModuleError
			}
			return []byte(s), nil
		}
	}

	testRequire := func(src, path string, globalFolders []string, fs map[string]string) (*js.Runtime, js.Value, error) {
		vm := js.New()
		r := NewRegistry(WithGlobalFolders(globalFolders...), WithLoader(mapFileSystemSourceLoader(fs)))
		rr := r.Enable(vm)
		rr.resolveStart = src
		t.Logf("Require(%s)", path)
		ret, err := rr.Require(path)
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
		"/home/src/app11/a/b/c/app11.js":         `exports.name = require('d/file.js').name`,
		"/home/src/app11/node_modules/d/file.js": `exports.name = "app11"`,
		"/app12.js":                              `exports.name = require('a/file.js').name`,
		"/node_modules/a/file.js":                `exports.name = "app12"`,
		"/app13/app13.js":                        `exports.name = require('b/file.js').name`,
		"/node_modules/b/file.js":                `exports.name = "app13"`,
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
		{"/home/src", "./app11/a/b/c/app11.js", true, "name", "app11"},
		{"/", "./app12", true, "name", "app12"},
		{"/", "./app13/app13", true, "name", "app13"},
	} {
		vm, mod, err := testRequire(tc.src, tc.path, globalFolders, fs)
		if err != nil {
			if tc.ok {
				t.Errorf("%v: require() failed: %v", i, err)
			}
			continue
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
