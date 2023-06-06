package url

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/dop251/goja"
)

func newInvalidTypleError(r *goja.Runtime) *goja.Object {
	return newError(r, "ERR_MISSING_ARGS", "Each query pair must be an iterable [name, value] tuple")
}

func newMissingArgsError(r *goja.Runtime, msg string) *goja.Object {
	return newError(r, "ERR_MISSING_ARGS", msg)
}

func newInvalidArgsError(r *goja.Runtime) *goja.Object {
	return newError(r, "ERR_INVALID_ARG_TYPE", `The "callback" argument must be of type function. Received undefined`)
}

func newUnsupportedArgsError(r *goja.Runtime) *goja.Object {
	return newError(r, "ERR_NOT_SUPPORTED", `The current method call is not supported.`)
}

func newError(r *goja.Runtime, code string, msg string) *goja.Object {
	o := r.NewTypeError("[" + code + "]: " + msg)
	o.Set("code", r.ToValue(code))
	return o
}

func urlAndQuery(r *goja.Runtime, v goja.Value) (*url.URL, url.Values) {
	u := toURL(r, v)
	return u, u.Query()
}

// NOTE:
//
// Order of the parameters will not be maintained based on value passed in.
// This is due to the encoding method on url.Values being backed by a map and not an array.
//
// Currently not supporting the following:
//
//   - ctor(iterable): Using function generators
//
//   - sort():  Since the backing object is a url.URL which backs the data as a Map, we can't reliably sort
//     the entries
//
//   - [] operator: TODO, need to figure out if we can override this with goja
func createURLSearchParamsConstructor(r *goja.Runtime) goja.Value {
	f := r.ToValue(func(call goja.ConstructorCall) *goja.Object {
		u, _ := url.Parse("")
		if len(call.Arguments) > 0 {
			p := call.Arguments[0]
			e := p.Export()
			if s, ok := e.(string); ok { // String
				u = buildParamsFromString(s)
			} else if o, ok := e.(map[string]interface{}); ok { // Object
				u = buildParamsFromObject(o)
			} else if a, ok := e.([]interface{}); ok { // Array
				u = buildParamsFromArray(r, a)
			} else if m, ok := e.([][2]interface{}); ok { // Map
				u = buildParamsFromMap(r, m)
			}
		}

		res := r.ToValue(u).(*goja.Object)
		res.SetPrototype(call.This.Prototype())
		return res
	}).(*goja.Object)

	f.Set("prototype", createURLSearchParamsPrototype(r))
	return f
}

// If Parsing results in a path, we move this to the RawQuery
func buildParamsFromString(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		return nil
	}

	if u.Path != "" && u.RawQuery == "" {
		v, err := url.Parse(fmt.Sprintf("?%s", u.Path))
		if err != nil {
			return nil
		}
		return v
	}

	return u
}

func buildParamsFromObject(o map[string]interface{}) *url.URL {
	query := url.Values{}
	for k, v := range o {
		if val, ok := v.([]interface{}); ok {
			vals := []string{}
			for _, e := range val {
				vals = append(vals, fmt.Sprintf("%v", e))
			}
			query.Add(k, strings.Join(vals, ","))
		} else {
			query.Add(k, fmt.Sprintf("%v", v))
		}
	}
	u, _ := url.Parse("")
	u.RawQuery = query.Encode()
	return u
}

func buildParamsFromArray(r *goja.Runtime, a []interface{}) *url.URL {
	query := url.Values{}
	for _, v := range a {
		if kv, ok := v.([]interface{}); ok {
			if len(kv) == 2 {
				query.Add(fmt.Sprintf("%v", kv[0]), fmt.Sprintf("%v", kv[1]))
			} else {
				panic(newInvalidTypleError(r))
			}
		} else {
			panic(newInvalidTypleError(r))
		}
	}

	u, _ := url.Parse("")
	u.RawQuery = query.Encode()
	return u
}

func buildParamsFromMap(r *goja.Runtime, m [][2]interface{}) *url.URL {
	query := url.Values{}
	for _, e := range m {
		query.Add(fmt.Sprintf("%v", e[0]), fmt.Sprintf("%v", e[1]))
	}

	u, _ := url.Parse("")
	u.RawQuery = query.Encode()
	return u
}

func createURLSearchParamsPrototype(r *goja.Runtime) *goja.Object {
	p := r.NewObject()

	p.Set("append", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(newMissingArgsError(r, `The "name" and "value" arguments must be specified`))
		}

		u, q := urlAndQuery(r, call.This)
		q.Add(call.Arguments[0].String(), call.Arguments[1].String())
		u.RawQuery = q.Encode()

		return goja.Undefined()
	}))

	p.Set("delete", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(newMissingArgsError(r, `The "name" argument must be specified`))
		}

		u, q := urlAndQuery(r, call.This)
		name := call.Arguments[0].String()
		if len(call.Arguments) > 1 {
			value := call.Arguments[1].String()
			if q.Has(name) && q.Get(name) == value {
				q.Del(name)
				u.RawQuery = q.Encode()
			}
		} else {
			q.Del(name)
			u.RawQuery = q.Encode()
		}

		return goja.Undefined()
	}))

	p.Set("entries", r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(r, call.This)
		entries := [][]string{}
		for k, e := range u.Query() {
			for _, v := range e {
				entries = append(entries, []string{k, v})
			}
		}

		return r.ToValue(entries)
	}))

	p.Set("forEach", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(newInvalidArgsError(r))
		}

		u, q := urlAndQuery(r, call.This)
		if fn, ok := goja.AssertFunction(call.Arguments[0]); ok {
			for k, e := range q {
				// name, value, searchParams
				for _, v := range e {
					_, err := fn(
						nil,
						r.ToValue(k),
						r.ToValue(v),
						r.ToValue(u.RawQuery))

					if err != nil {
						panic(err)
					}
				}
			}
		}

		return goja.Undefined()
	}))

	p.Set("get", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(newMissingArgsError(r, `The "name" argument must be specified`))
		}

		p := call.Arguments[0]
		e := p.Export()
		if n, ok := e.(string); ok {
			_, q := urlAndQuery(r, call.This)
			if !q.Has(n) {
				return goja.Null()
			}

			return r.ToValue(q.Get(n))
		}

		return goja.Null()
	}))

	p.Set("getAll", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(newMissingArgsError(r, `The "name" argument must be specified`))
		}

		p := call.Arguments[0]
		e := p.Export()
		if n, ok := e.(string); ok {
			_, q := urlAndQuery(r, call.This)
			if !q.Has(n) {
				return goja.Null()
			}

			return r.ToValue(q[n])
		}

		return goja.Null()
	}))

	p.Set("has", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(newMissingArgsError(r, `The "name" argument must be specified`))
		}

		p := call.Arguments[0]
		e := p.Export()
		if n, ok := e.(string); ok {
			_, q := urlAndQuery(r, call.This)

			if !q.Has(n) {
				return r.ToValue(false)
			}

			if len(call.Arguments) > 1 {
				value := call.Arguments[1].String()
				if value == q.Get(n) {
					return r.ToValue(true)
				}
			} else {
				return r.ToValue(true)
			}
		}

		return r.ToValue(false)
	}))

	p.Set("keys", r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(r, call.This)
		keys := []string{}
		for k := range u.Query() {
			keys = append(keys, fmt.Sprintf("%v", k))
		}

		return r.ToValue(keys)
	}))

	p.Set("set", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(newMissingArgsError(r, `The "name" and "value" arguments must be specified`))
		}

		u, q := urlAndQuery(r, call.This)
		q.Set(call.Arguments[0].String(), call.Arguments[1].String())
		u.RawQuery = q.Encode()

		return goja.Undefined()
	}))

	p.Set("sort", r.ToValue(func(call goja.FunctionCall) goja.Value {
		panic(newUnsupportedArgsError(r))
	}))

	defineURLAccessorProp(r, p, "size", func(u *url.URL) interface{} {
		q := u.Query()
		t := 0
		for _, v := range q {
			t += len(v)
		}
		return t
	}, nil)

	// toString()
	p.Set("toString", r.ToValue(func(call goja.FunctionCall) goja.Value {
		return r.ToValue(toURL(r, call.This).RawQuery)
	}))

	p.Set("values", r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(r, call.This)
		values := []string{}
		for _, e := range u.Query() {
			for _, v := range e {
				values = append(values, fmt.Sprintf("%v", v))
			}
		}

		return r.ToValue(values)
	}))

	return p
}
