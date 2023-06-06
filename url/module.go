package url

import (
	"net/url"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

const ModuleName = "node:url"

func toURL(r *goja.Runtime, v goja.Value) *url.URL {
	if v.ExportType() == reflectTypeURL {
		if u := v.Export().(*url.URL); u != nil {
			return u
		}
	}
	panic(r.NewTypeError("Expected URL"))
}

func defineURLAccessorProp(r *goja.Runtime, p *goja.Object, name string, getter func(*url.URL) interface{}, setter func(*url.URL, goja.Value)) {
	var getterVal, setterVal goja.Value
	if getter != nil {
		getterVal = r.ToValue(func(call goja.FunctionCall) goja.Value {
			return r.ToValue(getter(toURL(r, call.This)))
		})
	}
	if setter != nil {
		setterVal = r.ToValue(func(call goja.FunctionCall) goja.Value {
			setter(toURL(r, call.This), call.Argument(0))
			return goja.Undefined()
		})
	}
	p.DefineAccessorProperty(name, getterVal, setterVal, goja.FLAG_FALSE, goja.FLAG_TRUE)
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	exports := module.Get("exports").(*goja.Object)
	exports.Set("URL", createURLConstructor(runtime))
	exports.Set("URLSearchParams", createURLSearchParamsConstructor(runtime))
}

func Enable(runtime *goja.Runtime) {
	runtime.Set("URL", require.Require(runtime, ModuleName).ToObject(runtime).Get("URL"))
	runtime.Set("URLSearchParams", require.Require(runtime, ModuleName).ToObject(runtime).Get("URLSearchParams"))
}

func init() {
	require.RegisterNativeModule(ModuleName, Require)
}
