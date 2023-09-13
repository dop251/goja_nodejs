package url

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

const ModuleName = "url"

type urlModule struct {
	r *goja.Runtime

	URLSearchParamsPrototype         *goja.Object
	URLSearchParamsIteratorPrototype *goja.Object
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	exports := module.Get("exports").(*goja.Object)
	m := &urlModule{
		r: runtime,
	}
	exports.Set("URL", m.createURLConstructor())
	exports.Set("URLSearchParams", m.createURLSearchParamsConstructor())
}

func Enable(runtime *goja.Runtime) {
	runtime.Set("URL", require.Require(runtime, ModuleName).ToObject(runtime).Get("URL"))
	runtime.Set("URLSearchParams", require.Require(runtime, ModuleName).ToObject(runtime).Get("URLSearchParams"))
}

func init() {
	require.RegisterCoreModule(ModuleName, Require)
}
