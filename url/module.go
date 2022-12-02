package url

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

const ModuleName = "node:URL"

var reflectTypeURL = reflect.TypeOf((*url.URL)(nil))

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

func createURL(r *goja.Runtime) goja.Value {
	proto := r.NewObject()

	// host
	defineURLAccessorProp(r, proto, "host", func(u *url.URL) interface{} {
		return u.Host
	}, func(u *url.URL, arg goja.Value) {
		host := arg.String()
		if _, err := url.ParseRequestURI(u.Scheme + "://" + host); err == nil {
			u.Host = host
		}
	})

	// hash
	defineURLAccessorProp(r, proto, "hash", func(u *url.URL) interface{} {
		if u.Fragment != "" {
			return "#" + u.Fragment
		}
		return ""
	}, func(u *url.URL, arg goja.Value) {
		u.Fragment = strings.Replace(arg.String(), "#", "", 1)
	})

	// hostname
	defineURLAccessorProp(r, proto, "hostname", func(u *url.URL) interface{} {
		return strings.Split(u.Host, ":")[0]
	}, func(u *url.URL, arg goja.Value) {
		h := arg.String()
		if _, err := url.ParseRequestURI(u.Scheme + "://" + h); err == nil {
			hostname := strings.Split(h, ":")[0]
			u.Host = hostname + ":" + u.Port()
		}
	})

	// href
	defineURLAccessorProp(r, proto, "href", func(u *url.URL) interface{} {
		return u.String()
	}, func(u *url.URL, arg goja.Value) {
		if url, err := url.ParseRequestURI(arg.String()); err == nil {
			*u = *url
		}
	})

	// pathname
	defineURLAccessorProp(r, proto, "pathname", func(u *url.URL) interface{} {
		return u.Path
	}, func(u *url.URL, arg goja.Value) {
		p := arg.String()
		if _, err := url.Parse(p); err == nil {
			u.Path = p
		}
	})

	// origin
	defineURLAccessorProp(r, proto, "origin", func(u *url.URL) interface{} {
		return u.Scheme + "://" + u.Hostname()
	}, func(u *url.URL, arg goja.Value) { /* noop */ })

	// password
	defineURLAccessorProp(r, proto, "password", func(u *url.URL) interface{} {
		p, _ := u.User.Password()
		return p
	}, func(u *url.URL, arg goja.Value) {
		user := u.User
		u.User = url.UserPassword(user.Username(), arg.String())
	})

	// username
	defineURLAccessorProp(r, proto, "username", func(u *url.URL) interface{} {
		return u.User.Username()
	}, func(u *url.URL, arg goja.Value) {
		p, has := u.User.Password()
		if !has {
			u.User = url.User(arg.String())
		} else {
			u.User = url.UserPassword(arg.String(), p)
		}
	})

	// port
	defineURLAccessorProp(r, proto, "port", func(u *url.URL) interface{} {
		return u.Port()
	}, func(u *url.URL, arg goja.Value) {
		f, _ := strconv.ParseFloat(arg.String(), 64)
		max := ^uint16(0) // 65535
		if f > float64(max) {
			f = float64(max)
		}
		u.Host = u.Hostname() + ":" + fmt.Sprintf("%d", int(f))
	})

	// protocol
	defineURLAccessorProp(r, proto, "protocol", func(u *url.URL) interface{} {
		return u.Scheme + ":"
	}, func(u *url.URL, arg goja.Value) {
		scheme := strings.Replace(arg.String(), ":", "", -1)
		if _, err := url.ParseRequestURI(scheme + "://" + u.Host); err == nil {
			u.Scheme = scheme
		}
	})

	// Search
	defineURLAccessorProp(r, proto, "search", func(u *url.URL) interface{} {
		s := strings.Split(u.RawQuery, "#")[0]
		if s != "" {
			return "?" + s
		}
		return ""
	}, func(u *url.URL, arg goja.Value) {
		hash := ""
		if u.Fragment != "" {
			hash = "#" + u.Fragment
		}
		u.RawQuery = arg.String() + hash
	})

	proto.Set("toString", r.ToValue(func(call goja.FunctionCall) goja.Value {
		return r.ToValue(toURL(r, call.This).String())
	}))

	proto.Set("toJSON", r.ToValue(func(call goja.FunctionCall) goja.Value {
		return r.ToValue(toURL(r, call.This).String())
	}))

	f := r.ToValue(func(call goja.ConstructorCall) *goja.Object {
		var u *url.URL
		if len(call.Arguments) == 1 {
			if url, err := url.ParseRequestURI(call.Arguments[0].String()); err != nil {
				panic(r.NewTypeError("Failed to construct 'URL': Invalid URL"))
			} else {
				u, _ = url.Parse(call.Arguments[0].String())
			}
		} else {
			if _, err := url.ParseRequestURI(call.Arguments[1].String()); err != nil {
				panic(r.NewTypeError("Failed to construct 'URL': Invalid base URL"))
			} else if input, err := url.Parse(call.Arguments[0].String()); err != nil {
				panic(r.NewTypeError("Failed to construct 'URL': Invalid URL"))
			} else {
				base, _ := url.Parse(call.Arguments[1].String())
				u = base.ResolveReference(input)
			}
		}

		res := r.ToValue(u).(*goja.Object)
		res.SetPrototype(call.This.Prototype())
		return res
	}).(*goja.Object)

	f.Set("prototype", proto)
	return f
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	module.Set("exports", createURL(runtime))
}

func Enable(runtime *goja.Runtime) {
	runtime.Set("URL", require.Require(runtime, ModuleName))
}

func init() {
	require.RegisterNativeModule(ModuleName, Require)
}
