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

func newInvalidTypleError(r *goja.Runtime) *goja.Object {
	return newError(r, "ERR_MISSING_ARGS", "Each query pair must be an iterable [name, value] tuple")
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
//
//   - ctor(iterable): Using function generators
func createURLSearchParamsConstructor(r *goja.Runtime) goja.Value {
	f := r.ToValue(func(call goja.ConstructorCall) *goja.Object {
		u, _ := url.Parse("")
		if len(call.Arguments) > 0 {
			p := call.Arguments[0]
			e := p.Export()
			switch p.ExportType() {
			case reflectTypeString:
				u = buildParamsFromString(e.(string))
			case reflectTypeObject:
				u = buildParamsFromObject(e.(map[string]interface{}))
			case reflectTypeArray:
				u = buildParamsFromArray(r, e.([]interface{}))
			case reflectTypeMap:
				u = buildParamsFromMap(r, e.([][2]interface{}))
			}
		}

		res := r.ToValue(newFromURL(u)).(*goja.Object)
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
	query := searchParams{}
	for k, v := range o {
		if val, ok := v.([]interface{}); ok {
			vals := []string{}
			for _, e := range val {
				vals = append(vals, fmt.Sprintf("%v", e))
			}
			query = append(query, searchParam{name: k, value: vals})
		} else {
			query = append(query, searchParam{name: k, value: []string{fmt.Sprintf("%v", v)}})
		}
	}
	u, _ := url.Parse("")
	u.RawQuery = encodeSearchParams(query)
	return u
}

func buildParamsFromArray(r *goja.Runtime, a []interface{}) *url.URL {
	query := searchParams{}
	for _, v := range a {
		if kv, ok := v.([]interface{}); ok {
			if len(kv) == 2 {
				query = append(query, searchParam{
					name:  fmt.Sprintf("%v", kv[0]),
					value: []string{fmt.Sprintf("%v", kv[1])},
				})
			} else {
				panic(newInvalidTypleError(r))
			}
		} else {
			panic(newInvalidTypleError(r))
		}
	}

	u, _ := url.Parse("")
	u.RawQuery = encodeSearchParams(query)
	return u
}

func buildParamsFromMap(r *goja.Runtime, m [][2]interface{}) *url.URL {
	query := searchParams{}
	for _, e := range m {
		query = append(query, searchParam{
			name:  fmt.Sprintf("%v", e[0]),
			value: []string{fmt.Sprintf("%v", e[1])},
		})
	}

	u, _ := url.Parse("")
	u.RawQuery = encodeSearchParams(query)
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
			name:  call.Arguments[0].String(),
			value: []string{call.Arguments[1].String()},
		})
		u.syncSearchParams()

		return goja.Undefined()
	}))

	p.Set("delete", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(newMissingArgsError(r, `The "name" argument must be specified`))
		}

		u := toURL(r, call.This)
		name := call.Arguments[0].String()
		if len(call.Arguments) > 1 {
			value := call.Arguments[1].String()
			arr := searchParams{}
			for _, v := range u.searchParams {
				if v.name != name {
					arr = append(arr, v)
				} else {
					subArr := []string{}
					for _, val := range v.value {
						if val != value {
							subArr = append(subArr, val)
						}
					}
					if len(subArr) > 0 {
						arr = append(arr, searchParam{name: name, value: subArr})
					}
				}
			}
			u.searchParams = arr
		} else {
			arr := searchParams{}
			for _, v := range u.searchParams {
				if v.name != name {
					arr = append(arr, v)
				}
			}
			u.searchParams = arr
		}
		u.syncSearchParams()

		return goja.Undefined()
	}))

	p.Set("entries", r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(r, call.This)
		entries := [][]string{}
		for _, sp := range u.searchParams {
			entries = append(entries, []string{sp.name, strings.Join(sp.value, ",")})
		}

		return r.ToValue(entries)
	}))

	p.Set("forEach", r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(newInvalidArgsError(r))
		}

		u := toURL(r, call.This)
		if fn, ok := goja.AssertFunction(call.Arguments[0]); ok {
			for _, pair := range u.searchParams {
				// name, value, searchParams
				for _, v := range pair.value {
					query := strings.TrimPrefix(u.url.RawQuery, "?")
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

		p := call.Arguments[0]
		e := p.Export()
		if n, ok := e.(string); ok {
			u := toURL(r, call.This)
			vals, _ := u.getValues(n)
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

		p := call.Arguments[0]
		e := p.Export()
		if n, ok := e.(string); ok {
			u := toURL(r, call.This)
			vals, _ := u.getValues(n)
			if len(vals) > 0 {
				return r.ToValue(vals)
			}
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
			u := toURL(r, call.This)
			vals, contained := u.getValues(n)
			if len(call.Arguments) > 1 {
				for _, v := range vals {
					cmp := call.Arguments[1].String()
					if v == cmp {
						return r.ToValue(true)
					}
				}
				return r.ToValue(false)
			}

			return r.ToValue(contained)
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
		name := call.Arguments[0].String()
		found := false
		sps := searchParams{}
		for _, sp := range u.searchParams {
			if sp.name == name {
				if found {
					continue // Skip duplicates if present.
				}

				sp.value = []string{call.Arguments[1].String()}
				found = true
			}
			sps = append(sps, sp)
		}

		if found {
			u.searchParams = sps
		} else {
			u.searchParams = append(u.searchParams, searchParam{
				name:  name,
				value: []string{call.Arguments[1].String()},
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
		str := strings.TrimPrefix(encodeSearchParams(u.searchParams), "?")
		return r.ToValue(str)
	}))

	p.Set("values", r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(r, call.This)
		values := []string{}
		for _, sp := range u.searchParams {
			values = append(values, sp.value...)
		}

		return r.ToValue(values)
	}))

	return p
}
