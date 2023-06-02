package errors

import (
	"fmt"

	"github.com/dop251/goja"
)

const (
	ErrCodeInvalidArgType = "ERR_INVALID_ARG_TYPE"
)

func error_toString(call goja.FunctionCall, r *goja.Runtime) goja.Value {
	this := call.This.ToObject(r)
	var name, msg string
	if n := this.Get("name"); n != nil && !goja.IsUndefined(n) {
		name = n.String()
	} else {
		name = "Error"
	}
	if m := this.Get("message"); m != nil && !goja.IsUndefined(m) {
		msg = m.String()
	}
	if code := this.Get("code"); code != nil && !goja.IsUndefined(code) {
		if name != "" {
			name += " "
		}
		name += "[" + code.String() + "]"
	}
	if msg != "" {
		if name != "" {
			name += ": "
		}
		name += msg
	}
	return r.ToValue(name)
}

func addProps(r *goja.Runtime, e *goja.Object, code string) {
	e.Set("code", code)
	e.DefineDataProperty("toString", r.ToValue(error_toString), goja.FLAG_TRUE, goja.FLAG_TRUE, goja.FLAG_FALSE)
}

func NewTypeError(r *goja.Runtime, code string, params ...interface{}) *goja.Object {
	e := r.NewTypeError(params...)
	addProps(r, e, code)
	return e
}

func NewError(r *goja.Runtime, ctor *goja.Object, code string, args ...interface{}) *goja.Object {
	if ctor == nil {
		ctor, _ = r.Get("Error").(*goja.Object)
	}
	if ctor == nil {
		return nil
	}
	msg := ""
	if len(args) > 0 {
		f, _ := args[0].(string)
		msg = fmt.Sprintf(f, args[1:]...)
	}
	o, err := r.New(ctor, r.ToValue(msg))
	if err != nil {
		panic(err)
	}
	addProps(r, o, code)
	return o
}
