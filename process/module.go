package process

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

const ModuleName = "process"

type Process struct {
	env map[string]string
}

// hrtime in nodejs returns arbitrary time used for measuring performance between intervals.
// this function uses Go's time.Now()
func (p *Process) hrtime(t []int64) ([]int64, error) {
	var seconds, nanoseconds int64
	if t != nil {
		if len(t) != 2 {
			return nil, fmt.Errorf("the value for time must be of length 2, received %d", len(t))
		}
		seconds, nanoseconds = t[0], t[1]
	}

	now := time.Now().UnixNano() - int64(seconds)*1e9 - nanoseconds
	return []int64{now / 1e9, now % 1e9}, nil
}

// // this is commented out because it's supposed to return bigint, which goja doesn't currently support.
// func (p *Process) hrtime_bigint() int64 {
// 	return time.Now().UnixNano()
// }

func Require(runtime *goja.Runtime, module *goja.Object) {
	p := &Process{
		env: make(map[string]string),
	}

	for _, e := range os.Environ() {
		envKeyValue := strings.SplitN(e, "=", 2)
		p.env[envKeyValue[0]] = envKeyValue[1]
	}

	o := module.Get("exports").(*goja.Object)
	o.Set("env", p.env)
	o.Set("hrtime", p.hrtime)
	// o.Get("hrtime").ToObject(runtime).Set("bigint", p.hrtime_bigint)
}

func Enable(runtime *goja.Runtime) {
	runtime.Set("process", require.Require(runtime, ModuleName))
}

func init() {
	require.RegisterCoreModule(ModuleName, Require)
}
