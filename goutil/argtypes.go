package goutil

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/errors"
	"math/big"
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

func CoercedIntegerArgument(call goja.FunctionCall, argIndex int, defaultValue int64, typeMistMatchValue int64) int64 {
	arg := call.Argument(argIndex)
	if goja.IsNumber(arg) {
		return arg.ToInteger()
	}
	if goja.IsUndefined(arg) {
		return defaultValue
	}

	return typeMistMatchValue
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
