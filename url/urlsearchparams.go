package url

import (
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"

	"github.com/dop251/goja"
)

var (
	reflectTypeString = reflect.TypeOf("")
	reflectTypeObject = reflect.TypeOf(map[string]interface{}{})
	reflectTypeArray  = reflect.TypeOf([]interface{}{})
	reflectTypeMap    = reflect.TypeOf([][2]interface{}{})
)

func newInvalidTupleError(r *goja.Runtime) *goja.Object {
	return newError(r, "ERR_INVALID_TUPLE", "Each query pair must be an iterable [name, value] tuple")
}

func newMissingArgsError(r *goja.Runtime, msg string) *goja.Object {
	return newError(r, "ERR_MISSING_ARGS", msg)
}

func newInvalidArgsError(r *goja.Runtime) *goja.Object {
	return newError(r, "ERR_INVALID_ARG_TYPE", `The "callback" argument must be of type function. Received undefined`)
}

func newError(r *goja.Runtime, code string, msg string) *goja.Object {
	o := r.NewTypeError("[" + code + "]: " + msg)
	o.Set("code", r.ToValue(code))
	return o
}

// Currently not supporting the following:
//   - ctor(iterable): Using function generators
func createURLSearchParamsConstructor(r *goja.Runtime) goja.Value {
	f := r.ToValue(func(call goja.ConstructorCall) *goja.Object {
		u, _ := url.Parse("")
		v := call.Argument(0)
		if !goja.IsUndefined(v) {
			switch v.ExportType() {
			case reflectTypeString:
				u = buildParamsFromString(v.String())
			case reflectTypeObject:
				u = buildParamsFromObject(r, v)
			case reflectTypeArray:
				u = buildParamsFromArray(r, v)
			case reflectTypeMap:
				u = buildParamsFromMap(r, v)
			}
		}

		sp := parseSearchQuery(u.RawQuery)
		res := r.ToValue(&nodeURL{url: u, searchParams: sp}).(*goja.Object)
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

func buildParamsFromObject(r *goja.Runtime, v goja.Value) *url.URL {
	query := searchParams{}

	o := v.ToObject(r)
	for _, k := range o.Keys() {
		val := o.Get(k).String()
		query = append(query, searchParam{name: k, value: val})
	}

	u, _ := url.Parse("")
	u.RawQuery = query.String()
	return u
}

func buildParamsFromArray(r *goja.Runtime, v goja.Value) *url.URL {
	query := searchParams{}

	o := v.ToObject(r)

	r.ForOf(o, func(val goja.Value) bool {
		obj := val.ToObject(r)
		var name, value string
		i := 0
		// Use ForOf to determine if the object is iterable
		r.ForOf(obj, func(val goja.Value) bool {
			if i == 0 {
				name = fmt.Sprintf("%v", val)
				i++
				return true
			}
			if i == 1 {
				value = fmt.Sprintf("%v", val)
				i++
				return true
			}
			// Array isn't a tuple
			panic(newInvalidTupleError(r))
		})

		// Ensure we have two values
		if i <= 1 {
			panic(newInvalidTupleError(r))
		}

		query = append(query, searchParam{
			name:  name,
			value: value,
		})

		return true
	})

	u, _ := url.Parse("")
	u.RawQuery = query.String()
	return u
}

func buildParamsFromMap(r *goja.Runtime, v goja.Value) *url.URL {
	query := searchParams{}
	o := v.ToObject(r)
	r.ForOf(o, func(val goja.Value) bool {
		obj := val.ToObject(r)
		query = append(query, searchParam{
			name:  obj.Get("0").String(),
			value: obj.Get("1").String(),
		})
		return true
	})

	u, _ := url.Parse("")
	u.RawQuery = query.String()
	return u
}

func createURLSearchParamsPrototype(r *goja.Runtime) *goja.Object {
	p := r.NewObject()

	p.Set("append", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(newMissingArgsError(r, `The "name" and "value" arguments must be specified`))
		}

		u := toURL(r, call.This)
		u.searchParams = append(u.searchParams, searchParam{
			name:  call.Argument(0).String(),
			value: call.Argument(1).String(),
		})
		u.syncSearchParams()

		return goja.Undefined()
	}))

	p.Set("delete", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(newMissingArgsError(r, `The "name" argument must be specified`))
		}

		u := toURL(r, call.This)
		name := call.Argument(0).String()
		isValid := func(v searchParam) bool {
			if len(call.Arguments) == 1 {
				return v.name != name
			} else if v.name == name {
				arg := call.Argument(1)
				if !goja.IsUndefined(arg) && v.value == arg.String() {
					return false
				}
			}
			return true
		}

		i := 0
		for _, v := range u.searchParams {
			if isValid(v) {
				u.searchParams[i] = v
				i++
			}
		}
		u.searchParams = u.searchParams[:i]
		u.syncSearchParams()

		return goja.Undefined()
	}))

	p.Set("entries", r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(r, call.This)
		entries := [][]string{}
		for _, sp := range u.searchParams {
			entries = append(entries, []string{sp.name, sp.value})
		}

		return r.ToValue(entries)
	}))

	p.Set("forEach", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(newInvalidArgsError(r))
		}

		u := toURL(r, call.This)
		if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
			for _, pair := range u.searchParams {
				// name, value, searchParams
				for _, v := range pair.value {
					query := u.url.RawQuery
					_, err := fn(
						nil,
						r.ToValue(pair.name),
						r.ToValue(v),
						r.ToValue(query),
					)

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

		p := call.Argument(0)
		e := p.Export()
		if n, ok := e.(string); ok {
			u := toURL(r, call.This)
			vals := u.getValues(n)
			if len(vals) > 0 {
				return r.ToValue(vals[0])
			}
		}

		return goja.Null()
	}))

	p.Set("getAll", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(newMissingArgsError(r, `The "name" argument must be specified`))
		}

		p := call.Argument(0)
		e := p.Export()
		if n, ok := e.(string); ok {
			u := toURL(r, call.This)
			vals := u.getValues(n)
			if len(vals) > 0 {
				return r.ToValue(vals)
			}
		}

		return r.ToValue([]string{})
	}))

	p.Set("has", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(newMissingArgsError(r, `The "name" argument must be specified`))
		}

		p := call.Argument(0)
		e := p.Export()
		if n, ok := e.(string); ok {
			u := toURL(r, call.This)
			vals := u.getValues(n)
			param := call.Argument(1)
			if !goja.IsUndefined(param) {
				cmp := param.String()
				for _, v := range vals {
					if v == cmp {
						return r.ToValue(true)
					}
				}
				return r.ToValue(false)
			}

			return r.ToValue(u.hasName(n))
		}

		return r.ToValue(false)
	}))

	p.Set("keys", r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(r, call.This)
		keys := []string{}
		for _, sp := range u.searchParams {
			keys = append(keys, sp.name)
		}

		return r.ToValue(keys)
	}))

	p.Set("set", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(newMissingArgsError(r, `The "name" and "value" arguments must be specified`))
		}

		u := toURL(r, call.This)
		name := call.Argument(0).String()
		found := false
		sps := searchParams{}
		for _, sp := range u.searchParams {
			if sp.name == name {
				if found {
					continue // Skip duplicates if present.
				}

				sp.value = call.Argument(1).String()
				found = true
			}
			sps = append(sps, sp)
		}

		if found {
			u.searchParams = sps
		} else {
			u.searchParams = append(u.searchParams, searchParam{
				name:  name,
				value: call.Argument(1).String(),
			})
		}
		u.syncSearchParams()

		return goja.Undefined()
	}))

	p.Set("sort", r.ToValue(func(call goja.FunctionCall) goja.Value {
		sort.Sort(toURL(r, call.This).searchParams)
		return goja.Undefined()
	}))

	defineURLAccessorProp(r, p, "size", func(u *nodeURL) interface{} {
		return len(u.searchParams)
	}, nil)

	p.Set("toString", r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(r, call.This)
		str := u.searchParams.Encode()
		return r.ToValue(str)
	}))

	p.Set("values", r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(r, call.This)
		values := []string{}
		for _, sp := range u.searchParams {
			values = append(values, sp.value)
		}

		return r.ToValue(values)
	}))

	return p
}
