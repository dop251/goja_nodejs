package url

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

const ModuleName = "node:url"

var (
	reflectTypeURL       = reflect.TypeOf((*url.URL)(nil))
	nonAlphanumericRegex = regexp.MustCompile(`[^0-9]`)
	errInvalidPort       = errors.New("invalid port assignment")
	reservedPorts        = map[string]string{
		"ftp":   "21",
		"file":  "",
		"http":  "80",
		"https": "443",
		"ws":    "80",
		"wss":   "443",
	}
)

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

func createURLPrototype(r *goja.Runtime) *goja.Object {
	p := r.NewObject()

	// host
	defineURLAccessorProp(r, p, "host", func(u *url.URL) interface{} {
		return u.Host
	}, func(u *url.URL, arg goja.Value) {
		host := arg.String()
		if _, err := url.ParseRequestURI(u.Scheme + "://" + host); err == nil {
			u.Host = host
		}
	})

	// hash
	defineURLAccessorProp(r, p, "hash", func(u *url.URL) interface{} {
		if u.Fragment != "" {
			return "#" + u.Fragment
		}
		return ""
	}, func(u *url.URL, arg goja.Value) {
		u.Fragment = strings.Replace(arg.String(), "#", "", 1)
	})

	// hostname
	defineURLAccessorProp(r, p, "hostname", func(u *url.URL) interface{} {
		return strings.Split(u.Host, ":")[0]
	}, func(u *url.URL, arg goja.Value) {
		h := arg.String()
		if _, err := url.ParseRequestURI(u.Scheme + "://" + h); err == nil {
			hostname := strings.Split(h, ":")[0]
			u.Host = hostname + ":" + u.Port()
		}
	})

	// href
	defineURLAccessorProp(r, p, "href", func(u *url.URL) interface{} {
		return u.String()
	}, func(u *url.URL, arg goja.Value) {
		if url, err := url.ParseRequestURI(arg.String()); err == nil {
			*u = *url
		}
	})

	// pathname
	defineURLAccessorProp(r, p, "pathname", func(u *url.URL) interface{} {
		return u.Path
	}, func(u *url.URL, arg goja.Value) {
		p := arg.String()
		if _, err := url.Parse(p); err == nil {
			u.Path = p
		}
	})

	// origin
	defineURLAccessorProp(r, p, "origin", func(u *url.URL) interface{} {
		return u.Scheme + "://" + u.Hostname()
	}, func(u *url.URL, arg goja.Value) { /* noop */ })

	// password
	defineURLAccessorProp(r, p, "password", func(u *url.URL) interface{} {
		p, _ := u.User.Password()
		return p
	}, func(u *url.URL, arg goja.Value) {
		user := u.User
		u.User = url.UserPassword(user.Username(), arg.String())
	})

	// username
	defineURLAccessorProp(r, p, "username", func(u *url.URL) interface{} {
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
	defineURLAccessorProp(r, p, "port", func(u *url.URL) interface{} {
		return u.Port()
	}, func(u *url.URL, arg goja.Value) {
		if p, err := parsePort(u.Scheme, arg); err == nil {
			if p == "" {
				u.Host = u.Hostname()
			} else {
				u.Host = u.Hostname() + ":" + p
			}
		}
		// Ignore invalid values
	})

	// protocol
	defineURLAccessorProp(r, p, "protocol", func(u *url.URL) interface{} {
		return u.Scheme + ":"
	}, func(u *url.URL, arg goja.Value) {
		scheme := strings.Replace(arg.String(), ":", "", -1)
		if _, err := url.ParseRequestURI(scheme + "://" + u.Host); err == nil {
			u.Scheme = scheme
		}
	})

	// Search
	defineURLAccessorProp(r, p, "search", func(u *url.URL) interface{} {
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

	p.Set("toString", r.ToValue(func(call goja.FunctionCall) goja.Value {
		return r.ToValue(toURL(r, call.This).String())
	}))

	p.Set("toJSON", r.ToValue(func(call goja.FunctionCall) goja.Value {
		return r.ToValue(toURL(r, call.This).String())
	}))

	return p
}

func parsePort(s string, v goja.Value) (string, error) {
	// Clear for empty string, or reserved ports
	str := v.String()
	if str == "" || reservedPorts[s] == str {
		return "", nil
	}

	// Remove non-alphanumerics
	t := strings.Trim(str, " ")
	t = nonAlphanumericRegex.ReplaceAllString(t, " ")
	t = strings.Split(t, " ")[0]
	if t == "" {
		return "", errInvalidPort
	}

	i, err := strconv.Atoi(t)
	if err != nil {
		return "", errInvalidPort
	}

	// Port bounds
	if i >= 0 && i < math.MaxUint16 {
		return fmt.Sprintf("%d", i), nil
	}

	return "", errInvalidPort
}

func createURLConstructor(r *goja.Runtime) goja.Value {
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

	f.Set("prototype", createURLPrototype(r))
	return f
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	exports := module.Get("exports").(*goja.Object)
	exports.Set("URL", createURLConstructor(runtime))
}

func Enable(runtime *goja.Runtime) {
	runtime.Set("URL", require.Require(runtime, ModuleName).ToObject(runtime).Get("URL"))
}

func init() {
	require.RegisterNativeModule(ModuleName, Require)
}
