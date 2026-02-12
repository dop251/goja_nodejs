package goutil

import (
	"math"
	"math/big"
	"reflect"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/errors"
)

func RequiredIntegerArgument(r *goja.Runtime, call goja.FunctionCall, name string, argIndex int) int64 {
	arg := call.Argument(argIndex)
	if goja.IsNumber(arg) {
		return arg.ToInteger()
	}
	if goja.IsUndefined(arg) {
		panic(errors.NewTypeError(r, errors.ErrCodeInvalidArgType, "The \"%s\" argument is required.", name))
	}

	panic(errors.NewArgumentNotNumberTypeError(r, name))
}

func RequiredStrictIntegerArgument(r *goja.Runtime, call goja.FunctionCall, name string, argIndex int) int64 {
	arg := call.Argument(argIndex)
	if goja.IsNumber(arg) {
		val := arg.ToFloat()
		if val != math.Trunc(val) {
			panic(errors.NewRangeError(r, errors.ErrCodeOutOfRange, "The value of %q is out of range. It must be an integer.", name))
		}
		return int64(val)
	}
	if goja.IsUndefined(arg) {
		panic(errors.NewTypeError(r, errors.ErrCodeInvalidArgType, "The %q argument is required.", name))
	}

	panic(errors.NewArgumentNotNumberTypeError(r, name))
}

func RequiredFloatArgument(r *goja.Runtime, call goja.FunctionCall, name string, argIndex int) float64 {
	arg := call.Argument(argIndex)
	if goja.IsNumber(arg) {
		return arg.ToFloat()
	}
	if goja.IsUndefined(arg) {
		panic(errors.NewTypeError(r, errors.ErrCodeInvalidArgType, "The \"%s\" argument is required.", name))
	}

	panic(errors.NewArgumentNotNumberTypeError(r, name))
}

func CoercedIntegerArgument(call goja.FunctionCall, argIndex int, defaultValue int64, typeMismatchValue int64) int64 {
	arg := call.Argument(argIndex)
	if goja.IsNumber(arg) {
		return arg.ToInteger()
	}
	if goja.IsUndefined(arg) {
		return defaultValue
	}

	return typeMismatchValue
}

func OptionalIntegerArgument(r *goja.Runtime, call goja.FunctionCall, name string, argIndex int, defaultValue int64) int64 {
	arg := call.Argument(argIndex)
	if goja.IsNumber(arg) {
		return arg.ToInteger()
	}
	if goja.IsUndefined(arg) {
		return defaultValue
	}

	panic(errors.NewArgumentNotNumberTypeError(r, name))
}

func RequiredBigIntArgument(r *goja.Runtime, call goja.FunctionCall, name string, argIndex int) *big.Int {
	arg := call.Argument(argIndex)
	if goja.IsUndefined(arg) {
		panic(errors.NewTypeError(r, errors.ErrCodeInvalidArgType, "The \"%s\" argument is required.", name))
	}
	if !goja.IsBigInt(arg) {
		panic(errors.NewArgumentNotBigIntTypeError(r, name))
	}

	n, _ := arg.Export().(*big.Int)
	if n == nil {
		n = new(big.Int)
	}
	return n
}

func RequiredStringArgument(r *goja.Runtime, call goja.FunctionCall, name string, argIndex int) string {
	arg := call.Argument(argIndex)
	if goja.IsString(arg) {
		return arg.String()
	}
	if goja.IsUndefined(arg) {
		panic(errors.NewTypeError(r, errors.ErrCodeInvalidArgType, "The \"%s\" argument is required.", name))
	}

	panic(errors.NewArgumentNotStringTypeError(r, name))
}

func RequiredArrayArgument(r *goja.Runtime, call goja.FunctionCall, name string, argIndex int) goja.Value {
	arg := call.Argument(argIndex)
	if arg.ExportType() != reflect.TypeOf(([]any)(nil)) {
		panic(errors.NewNotCorrectTypeError(r, name, "Array"))
	}
	return arg
}
