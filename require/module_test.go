package require

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

	registry := NewRegistryWithLoader(func(name string) ([]byte, error) {
		if name == "m.js" {
			return []byte(MODULE), nil
		}
		return nil, errors.New("Module does not exist")
	})
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

	registry := NewRegistryWithLoader(func(name string) ([]byte, error) {
		if name == "m.js" {
			return []byte(MODULE), nil
		}
		return nil, errors.New("Module does not exist")
	})
	registry.Enable(vm)

	v, err := vm.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	if !v.StrictEquals(vm.ToValue("passed1")) {
		t.Fatalf("Unexpected result: %v", v)
	}
}
