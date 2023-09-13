package url

import (
	"reflect"
	"sort"

	"github.com/dop251/goja_nodejs/errors"

	"github.com/dop251/goja"
)

var (
	reflectTypeURLSearchParams         = reflect.TypeOf((*urlSearchParams)(nil))
	reflectTypeURLSearchParamsIterator = reflect.TypeOf((*urlSearchParamsIterator)(nil))
)

func newInvalidTupleError(r *goja.Runtime) *goja.Object {
	return errors.NewTypeError(r, "ERR_INVALID_TUPLE", "Each query pair must be an iterable [name, value] tuple")
}

func newMissingArgsError(r *goja.Runtime, msg string) *goja.Object {
	return errors.NewTypeError(r, errors.ErrCodeMissingArgs, msg)
}

func newInvalidArgsError(r *goja.Runtime) *goja.Object {
	return errors.NewTypeError(r, "ERR_INVALID_ARG_TYPE", `The "callback" argument must be of type function.`)
}

func toUrlSearchParams(r *goja.Runtime, v goja.Value) *urlSearchParams {
	if v.ExportType() == reflectTypeURLSearchParams {
		if u := v.Export().(*urlSearchParams); u != nil {
			return u
		}
	}
	panic(errors.NewTypeError(r, errors.ErrCodeInvalidThis, `Value of "this" must be of type URLSearchParams`))
}

func (m *urlModule) newURLSearchParams(sp *urlSearchParams) *goja.Object {
	v := m.r.ToValue(sp).(*goja.Object)
	v.SetPrototype(m.URLSearchParamsPrototype)
	return v
}

func (m *urlModule) createURLSearchParamsConstructor() goja.Value {
	f := m.r.ToValue(func(call goja.ConstructorCall) *goja.Object {
		var sp searchParams
		v := call.Argument(0)
		if o, ok := v.(*goja.Object); ok {
			sp = m.buildParamsFromObject(o)
		} else if !goja.IsUndefined(v) {
			sp = parseSearchQuery(v.String())
		}

		return m.newURLSearchParams(&urlSearchParams{searchParams: sp})
	}).(*goja.Object)

	m.URLSearchParamsPrototype = m.createURLSearchParamsPrototype()
	f.Set("prototype", m.URLSearchParamsPrototype)
	m.URLSearchParamsPrototype.DefineDataProperty("constructor", f, goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_FALSE)

	return f
}

func (m *urlModule) buildParamsFromObject(o *goja.Object) searchParams {
	var query searchParams

	if o.GetSymbol(goja.SymIterator) != nil {
		return m.buildParamsFromIterable(o)
	}

	for _, k := range o.Keys() {
		val := o.Get(k).String()
		query = append(query, searchParam{name: k, value: val})
	}

	return query
}

func (m *urlModule) buildParamsFromIterable(o *goja.Object) searchParams {
	var query searchParams

	m.r.ForOf(o, func(val goja.Value) bool {
		obj := val.ToObject(m.r)
		var name, value string
		i := 0
		// Use ForOf to determine if the object is iterable
		m.r.ForOf(obj, func(val goja.Value) bool {
			if i == 0 {
				name = val.String()
				i++
				return true
			}
			if i == 1 {
				value = val.String()
				i++
				return true
			}
			// Array isn't a tuple
			panic(newInvalidTupleError(m.r))
		})

		// Ensure we have two values
		if i <= 1 {
			panic(newInvalidTupleError(m.r))
		}

		query = append(query, searchParam{
			name:  name,
			value: value,
		})

		return true
	})

	return query
}

func (m *urlModule) createURLSearchParamsPrototype() *goja.Object {
	p := m.r.NewObject()

	p.Set("append", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(newMissingArgsError(m.r, `The "name" and "value" arguments must be specified`))
		}

		u := toUrlSearchParams(m.r, call.This)
		u.searchParams = append(u.searchParams, searchParam{
			name:  call.Argument(0).String(),
			value: call.Argument(1).String(),
		})
		u.markUpdated()

		return goja.Undefined()
	}))

	p.Set("delete", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toUrlSearchParams(m.r, call.This)

		if len(call.Arguments) < 1 {
			panic(newMissingArgsError(m.r, `The "name" argument must be specified`))
		}

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

		j := 0
		for i, v := range u.searchParams {
			if isValid(v) {
				if i != j {
					u.searchParams[j] = v
				}
				j++
			}
		}
		u.searchParams = u.searchParams[:j]
		u.markUpdated()

		return goja.Undefined()
	}))

	entries := m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		return m.newURLSearchParamsIterator(toUrlSearchParams(m.r, call.This), urlSearchParamsIteratorEntries)
	})
	p.Set("entries", entries)
	p.DefineDataPropertySymbol(goja.SymIterator, entries, goja.FLAG_TRUE, goja.FLAG_FALSE, goja.FLAG_TRUE)

	p.Set("forEach", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toUrlSearchParams(m.r, call.This)

		if len(call.Arguments) != 1 {
			panic(newInvalidArgsError(m.r))
		}

		if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
			for _, pair := range u.searchParams {
				// name, value, searchParams
				_, err := fn(
					nil,
					m.r.ToValue(pair.name),
					m.r.ToValue(pair.value),
					call.This,
				)

				if err != nil {
					panic(err)
				}
			}
		} else {
			panic(newInvalidArgsError(m.r))
		}

		return goja.Undefined()
	}))

	p.Set("get", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toUrlSearchParams(m.r, call.This)

		if len(call.Arguments) == 0 {
			panic(newMissingArgsError(m.r, `The "name" argument must be specified`))
		}

		if val, exists := u.getFirstValue(call.Argument(0).String()); exists {
			return m.r.ToValue(val)
		}

		return goja.Null()
	}))

	p.Set("getAll", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toUrlSearchParams(m.r, call.This)

		if len(call.Arguments) == 0 {
			panic(newMissingArgsError(m.r, `The "name" argument must be specified`))
		}

		vals := u.getValues(call.Argument(0).String())
		return m.r.ToValue(vals)
	}))

	p.Set("has", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toUrlSearchParams(m.r, call.This)

		if len(call.Arguments) == 0 {
			panic(newMissingArgsError(m.r, `The "name" argument must be specified`))
		}

		name := call.Argument(0).String()
		value := call.Argument(1)
		var res bool
		if goja.IsUndefined(value) {
			res = u.hasName(name)
		} else {
			res = u.hasValue(name, value.String())
		}
		return m.r.ToValue(res)
	}))

	p.Set("keys", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		return m.newURLSearchParamsIterator(toUrlSearchParams(m.r, call.This), urlSearchParamsIteratorKeys)
	}))

	p.Set("set", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toUrlSearchParams(m.r, call.This)

		if len(call.Arguments) < 2 {
			panic(newMissingArgsError(m.r, `The "name" and "value" arguments must be specified`))
		}

		name := call.Argument(0).String()
		found := false
		j := 0
		for i, sp := range u.searchParams {
			if sp.name == name {
				if found {
					continue // Remove all values
				}

				u.searchParams[i].value = call.Argument(1).String()
				found = true
			}
			if i != j {
				u.searchParams[j] = sp
			}
			j++
		}

		if !found {
			u.searchParams = append(u.searchParams, searchParam{
				name:  name,
				value: call.Argument(1).String(),
			})
		} else {
			u.searchParams = u.searchParams[:j]
		}

		u.markUpdated()

		return goja.Undefined()
	}))

	p.Set("sort", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toUrlSearchParams(m.r, call.This)
		sort.Stable(u.searchParams)
		u.markUpdated()
		return goja.Undefined()
	}))

	p.DefineAccessorProperty("size", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toUrlSearchParams(m.r, call.This)
		return m.r.ToValue(len(u.searchParams))
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	p.Set("toString", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toUrlSearchParams(m.r, call.This)
		str := u.searchParams.Encode()
		return m.r.ToValue(str)
	}))

	p.Set("values", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		return m.newURLSearchParamsIterator(toUrlSearchParams(m.r, call.This), urlSearchParamsIteratorValues)
	}))

	return p
}

func (sp *urlSearchParams) markUpdated() {
	if sp.url != nil && sp.url.RawQuery != "" {
		sp.url.RawQuery = ""
	}
}

type urlSearchParamsIteratorType int

const (
	urlSearchParamsIteratorKeys urlSearchParamsIteratorType = iota
	urlSearchParamsIteratorValues
	urlSearchParamsIteratorEntries
)

type urlSearchParamsIterator struct {
	typ urlSearchParamsIteratorType
	sp  *urlSearchParams
	idx int
}

func toURLSearchParamsIterator(r *goja.Runtime, v goja.Value) *urlSearchParamsIterator {
	if v.ExportType() == reflectTypeURLSearchParamsIterator {
		if u := v.Export().(*urlSearchParamsIterator); u != nil {
			return u
		}
	}

	panic(errors.NewTypeError(r, errors.ErrCodeInvalidThis, `Value of "this" must be of type URLSearchParamIterator`))
}

func (m *urlModule) getURLSearchParamsIteratorPrototype() *goja.Object {
	if m.URLSearchParamsIteratorPrototype != nil {
		return m.URLSearchParamsIteratorPrototype
	}

	p := m.r.NewObject()

	p.Set("next", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		it := toURLSearchParamsIterator(m.r, call.This)
		res := m.r.NewObject()
		if it.idx < len(it.sp.searchParams) {
			param := it.sp.searchParams[it.idx]
			switch it.typ {
			case urlSearchParamsIteratorKeys:
				res.Set("value", param.name)
			case urlSearchParamsIteratorValues:
				res.Set("value", param.value)
			default:
				res.Set("value", m.r.NewArray(param.name, param.value))
			}
			res.Set("done", false)
			it.idx++
		} else {
			res.Set("value", goja.Undefined())
			res.Set("done", true)
		}
		return res
	}))

	p.DefineDataPropertySymbol(goja.SymToStringTag, m.r.ToValue("URLSearchParams Iterator"), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_TRUE)

	m.URLSearchParamsIteratorPrototype = p
	return p
}

func (m *urlModule) newURLSearchParamsIterator(sp *urlSearchParams, typ urlSearchParamsIteratorType) goja.Value {
	it := m.r.ToValue(&urlSearchParamsIterator{
		typ: typ,
		sp:  sp,
	}).(*goja.Object)

	it.SetPrototype(m.getURLSearchParamsIteratorPrototype())

	return it
}
