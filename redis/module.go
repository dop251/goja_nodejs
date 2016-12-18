// Copyright 2016 Dual Inventive. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// TODO: add del, hget, hset, hdel

package redis

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/garyburd/redigo/redis"
)

type Redis struct {
	runtime *goja.Runtime
	util    *goja.Object
	redis   redis.Conn
}

func (c *Redis) set(call goja.FunctionCall) goja.Value {
	// TODO: argument/error checking
	_, _ = c.redis.Do("SET", call.Argument(0), call.Argument(1))
	return nil
}

func (c *Redis) get(call goja.FunctionCall) goja.Value {
	// TODO: error/argument checking and return JS runtime error
	rep, err := c.redis.Do("GET", call.Argument(0))
	if err != nil {
		return nil
	}

	return c.runtime.ToValue(rep)
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	r := &Redis{
		runtime: runtime,
	}

	o := module.Get("exports").(*goja.Object)
	o.Set("set", r.set)
	o.Set("get", r.get)

	// TODO: configurable
	r.redis, _ = redis.Dial("tcp", "localhost:6379")
}

func Enable(runtime *goja.Runtime) {
	runtime.Set("redis", require.Require(runtime, "redis"))
}

func init() {
	require.RegisterNativeModule("redis", Require)
}
