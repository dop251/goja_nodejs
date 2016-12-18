// Copyright 2016 Dual Inventive. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package redis

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"testing"
)

func TestRedisEnable(t *testing.T) {
	vm := goja.New()

	new(require.Registry).Enable(vm)
	Enable(vm)

	if c := vm.Get("redis"); c == nil {
		t.Fatal("redis not found")
	}
}

func TestRedisSetGet(t *testing.T) {
	vm := goja.New()

	new(require.Registry).Enable(vm)
	Enable(vm)

	// TODO: hack...
	redis := vm.Get("redis").Export().(map[string]interface{})

	if c := redis; c == nil {
		t.Fatal("redis not found")
	}

	fmt.Println(redis)

	// TODO: this is far to much typing
	set := redis["set"].(func(goja.FunctionCall) goja.Value)
	args := goja.FunctionCall{Arguments: []goja.Value{vm.ToValue("boem"), vm.ToValue("UUUU")}}
	set(args)

	// TODO: this is far to much typing
	get := redis["get"].(func(goja.FunctionCall) goja.Value)
	args = goja.FunctionCall{Arguments: []goja.Value{vm.ToValue("boem")}}
	rep := get(args)
	fmt.Printf("%+v\n", rep.ExportType())
}
