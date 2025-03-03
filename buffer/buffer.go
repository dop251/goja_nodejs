package buffer

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/big"
	"reflect"
	"strconv"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/errors"
	"github.com/dop251/goja_nodejs/require"

	"github.com/dop251/base64dec"
	"golang.org/x/text/encoding/unicode"
)

const ModuleName = "buffer"

type Buffer struct {
	r *goja.Runtime

	bufferCtorObj *goja.Object

	uint8ArrayCtorObj *goja.Object
	uint8ArrayCtor    goja.Constructor
}

var (
	symApi = goja.NewSymbol("api")
)

var (
	reflectTypeArrayBuffer = reflect.TypeOf(goja.ArrayBuffer{})
	reflectTypeString      = reflect.TypeOf("")
	reflectTypeInt         = reflect.TypeOf(int64(0))
	reflectTypeFloat       = reflect.TypeOf(0.0)
	reflectTypeBytes       = reflect.TypeOf(([]byte)(nil))
)

func Enable(runtime *goja.Runtime) {
	runtime.Set("Buffer", require.Require(runtime, ModuleName).ToObject(runtime).Get("Buffer"))
}

func Bytes(r *goja.Runtime, v goja.Value) []byte {
	var b []byte
	err := r.ExportTo(v, &b)
	if err != nil {
		return []byte(v.String())
	}
	return b
}

func mod(r *goja.Runtime) *goja.Object {
	res := r.Get("Buffer")
	if res == nil {
		res = require.Require(r, ModuleName).ToObject(r).Get("Buffer")
	}
	m, ok := res.(*goja.Object)
	if !ok {
		panic(r.NewTypeError("Could not extract Buffer"))
	}
	return m
}

func api(mod *goja.Object) *Buffer {
	if s := mod.GetSymbol(symApi); s != nil {
		b, _ := s.Export().(*Buffer)
		return b
	}

	return nil
}

func GetApi(r *goja.Runtime) *Buffer {
	return api(mod(r))
}

func DecodeBytes(r *goja.Runtime, arg, enc goja.Value) []byte {
	switch arg.ExportType() {
	case reflectTypeArrayBuffer:
		return arg.Export().(goja.ArrayBuffer).Bytes()
	case reflectTypeString:
		var codec StringCodec
		if !goja.IsUndefined(enc) {
			codec = stringCodecs[enc.String()]
		}
		if codec == nil {
			codec = utf8Codec
		}
		return codec.DecodeAppend(arg.String(), nil)
	default:
		if o, ok := arg.(*goja.Object); ok {
			if o.ExportType() == reflectTypeBytes {
				return o.Export().([]byte)
			}
		}
	}
	panic(errors.NewTypeError(r, errors.ErrCodeInvalidArgType, "The \"data\" argument must be of type string or an instance of Buffer, TypedArray, or DataView."))
}

func WrapBytes(r *goja.Runtime, data []byte) *goja.Object {
	m := mod(r)
	if api := api(m); api != nil {
		return api.WrapBytes(data)
	}
	if from, ok := goja.AssertFunction(m.Get("from")); ok {
		ab := r.NewArrayBuffer(data)
		v, err := from(m, r.ToValue(ab))
		if err != nil {
			panic(err)
		}
		return v.ToObject(r)
	}
	panic(r.NewTypeError("Buffer.from is not a function"))
}

// EncodeBytes returns the given byte slice encoded as string with the given encoding. If encoding
// is not specified or not supported, returns a Buffer that wraps the data.
func EncodeBytes(r *goja.Runtime, data []byte, enc goja.Value) goja.Value {
	var codec StringCodec
	if !goja.IsUndefined(enc) {
		codec = StringCodecByName(enc.String())
	}
	if codec != nil {
		return r.ToValue(codec.Encode(data))
	}
	return WrapBytes(r, data)
}

func (b *Buffer) WrapBytes(data []byte) *goja.Object {
	return b.fromBytes(data)
}

func (b *Buffer) ctor(call goja.ConstructorCall) (res *goja.Object) {
	arg := call.Argument(0)
	switch arg.ExportType() {
	case reflectTypeInt, reflectTypeFloat:
		panic(b.r.NewTypeError("Calling the Buffer constructor with numeric argument is not implemented yet"))
		// TODO implement
	}
	return b._from(call.Arguments...)
}

type StringCodec interface {
	DecodeAppend(string, []byte) []byte
	Encode([]byte) string
	Decode(s string) []byte
}

type hexCodec struct{}

func (hexCodec) DecodeAppend(s string, b []byte) []byte {
	l := hex.DecodedLen(len(s))
	dst, res := expandSlice(b, l)
	n, err := hex.Decode(dst, []byte(s))
	if err != nil {
		res = res[:len(b)+n]
	}
	return res
}

func (hexCodec) Decode(s string) []byte {
	n, _ := hex.DecodeString(s)
	return n
}
func (hexCodec) Encode(b []byte) string {
	return hex.EncodeToString(b)
}

type _utf8Codec struct{}

func (c _utf8Codec) DecodeAppend(s string, b []byte) []byte {
	r := c.Decode(s)
	dst, res := expandSlice(b, len(r))
	copy(dst, r)
	return res
}

func (_utf8Codec) Decode(s string) []byte {
	r, _ := unicode.UTF8.NewEncoder().String(s)
	return []byte(r)
}
func (_utf8Codec) Encode(b []byte) string {
	r, _ := unicode.UTF8.NewDecoder().Bytes(b)
	return string(r)
}

type base64Codec struct{}

type base64UrlCodec struct {
	base64Codec
}

func (base64Codec) DecodeAppend(s string, b []byte) []byte {
	res, _ := Base64DecodeAppend(b, s)
	return res
}

func (base64Codec) Decode(s string) []byte {
	res, _ := base64.StdEncoding.DecodeString(s)
	return res
}
func (base64Codec) Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func (base64UrlCodec) Encode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

var utf8Codec StringCodec = _utf8Codec{}

var stringCodecs = map[string]StringCodec{
	"hex":       hexCodec{},
	"utf8":      utf8Codec,
	"utf-8":     utf8Codec,
	"base64":    base64Codec{},
	"base64Url": base64UrlCodec{},
}

func expandSlice(b []byte, l int) (dst, res []byte) {
	if cap(b)-len(b) < l {
		b1 := make([]byte, len(b)+l)
		copy(b1, b)
		dst = b1[len(b):]
		res = b1
	} else {
		dst = b[len(b) : len(b)+l]
		res = b[:len(b)+l]
	}
	return
}

func Base64DecodeAppend(dst []byte, src string) ([]byte, error) {
	l := base64.RawStdEncoding.DecodedLen(len(src))
	d, res := expandSlice(dst, l)
	n, err := base64dec.DecodeBase64(d, src)

	res = res[:len(dst)+n]
	return res, err
}

func (b *Buffer) fromString(str, enc string) *goja.Object {
	codec := stringCodecs[enc]
	if codec == nil {
		codec = utf8Codec
	}
	return b.fromBytes(codec.DecodeAppend(str, nil))
}

func (b *Buffer) fromBytes(data []byte) *goja.Object {
	o, err := b.uint8ArrayCtor(b.bufferCtorObj, b.r.ToValue(b.r.NewArrayBuffer(data)))
	if err != nil {
		panic(err)
	}
	return o
}

func (b *Buffer) _from(args ...goja.Value) *goja.Object {
	if len(args) == 0 {
		panic(errors.NewTypeError(b.r, errors.ErrCodeInvalidArgType, "The first argument must be of type string or an instance of Buffer, ArrayBuffer, or Array or an Array-like Object. Received undefined"))
	}
	arg := args[0]
	switch arg.ExportType() {
	case reflectTypeArrayBuffer:
		v, err := b.uint8ArrayCtor(b.bufferCtorObj, args...)
		if err != nil {
			panic(err)
		}
		return v
	case reflectTypeString:
		var enc string
		if len(args) > 1 {
			enc = args[1].String()
		}
		return b.fromString(arg.String(), enc)
	default:
		if o, ok := arg.(*goja.Object); ok {
			if o.ExportType() == reflectTypeBytes {
				bb, _ := o.Export().([]byte)
				a := make([]byte, len(bb))
				copy(a, bb)
				return b.fromBytes(a)
			} else {
				if f, ok := goja.AssertFunction(o.Get("valueOf")); ok {
					valueOf, err := f(o)
					if err != nil {
						panic(err)
					}
					if valueOf != o {
						args[0] = valueOf
						return b._from(args...)
					}
				}

				if s := o.GetSymbol(goja.SymToPrimitive); s != nil {
					if f, ok := goja.AssertFunction(s); ok {
						str, err := f(o, b.r.ToValue("string"))
						if err != nil {
							panic(err)
						}
						args[0] = str
						return b._from(args...)
					}
				}
			}
			// array-like
			if v := o.Get("length"); v != nil {
				length := int(v.ToInteger())
				a := make([]byte, length)
				for i := 0; i < length; i++ {
					item := o.Get(strconv.Itoa(i))
					if item != nil {
						a[i] = byte(item.ToInteger())
					}
				}
				return b.fromBytes(a)
			}
		}
	}
	panic(errors.NewTypeError(b.r, errors.ErrCodeInvalidArgType, "The first argument must be of type string or an instance of Buffer, ArrayBuffer, or Array or an Array-like Object. Received %s", arg))
}

func (b *Buffer) from(call goja.FunctionCall) goja.Value {
	return b._from(call.Arguments...)
}

func isNumber(v goja.Value) bool {
	switch v.ExportType() {
	case reflectTypeInt, reflectTypeFloat:
		return true
	}
	return false
}

func isString(v goja.Value) bool {
	return v.ExportType() == reflectTypeString
}

func StringCodecByName(name string) StringCodec {
	return stringCodecs[name]
}

func (b *Buffer) getStringCodec(enc goja.Value) (codec StringCodec) {
	if !goja.IsUndefined(enc) {
		codec = stringCodecs[enc.String()]
		if codec == nil {
			panic(errors.NewTypeError(b.r, "ERR_UNKNOWN_ENCODING", "Unknown encoding: %s", enc))
		}
	} else {
		codec = utf8Codec
	}
	return
}

func (b *Buffer) fill(buf []byte, fill string, enc goja.Value) []byte {
	codec := b.getStringCodec(enc)
	b1 := codec.DecodeAppend(fill, buf[:0])
	if len(b1) > len(buf) {
		return b1[:len(buf)]
	}
	for i := len(b1); i < len(buf); {
		i += copy(buf[i:], buf[:i])
	}
	return buf
}

func (b *Buffer) alloc(call goja.FunctionCall) goja.Value {
	arg0 := call.Argument(0)
	size := -1
	if isNumber(arg0) {
		size = int(arg0.ToInteger())
	}
	if size < 0 {
		panic(b.newArgumentNotNumberTypeError("size"))
	}
	fill := call.Argument(1)
	buf := make([]byte, size)
	if !goja.IsUndefined(fill) {
		if isString(fill) {
			var enc goja.Value
			if a := call.Argument(2); isString(a) {
				enc = a
			} else {
				enc = goja.Undefined()
			}
			buf = b.fill(buf, fill.String(), enc)
		} else {
			fill = fill.ToNumber()
			if !goja.IsNaN(fill) && !goja.IsInfinity(fill) {
				fillByte := byte(fill.ToInteger())
				if fillByte != 0 {
					for i := range buf {
						buf[i] = fillByte
					}
				}
			}
		}
	}
	return b.fromBytes(buf)
}

func (b *Buffer) proto_toString(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	codec := b.getStringCodec(call.Argument(0))
	start := b.getCoercedIntegerArgument(call, 1, 0, 0)

	// Node's Buffer class makes this zero if it is negative
	if start < 0 {
		start = 0
	} else if start >= int64(len(bb)) {
		// returns an empty string if start is beyond the length of the buffer
		return b.r.ToValue("")
	}

	// NOTE that Node will default to the length of the buffer, but uses 0 for type mismatch defaults
	end := b.getCoercedIntegerArgument(call, 2, int64(len(bb)), 0)
	if end < 0 || start >= end {
		// returns an empty string if end < 0 or start >= end
		return b.r.ToValue("")
	} else if end > int64(len(bb)) {
		// and Node ensures you don't go past the Buffer
		end = int64(len(bb))
	}

	return b.r.ToValue(codec.Encode(bb[start:end]))
}

func (b *Buffer) proto_equals(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	other := call.Argument(0)
	if b.r.InstanceOf(other, b.uint8ArrayCtorObj) {
		otherBytes := Bytes(b.r, other)
		return b.r.ToValue(bytes.Equal(bb, otherBytes))
	}
	panic(errors.NewTypeError(b.r, errors.ErrCodeInvalidArgType, "The \"otherBuffer\" argument must be an instance of Buffer or Uint8Array."))
}

// readBigInt64BE reads a big-endian 64-bit signed integer from the buffer
func (b *Buffer) readBigInt64BE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 8)
	value := int64(binary.BigEndian.Uint64(bb[offset : offset+8]))

	return b.r.ToValue(big.NewInt(value))
}

// readBigInt64LE reads a little-endian 64-bit signed integer from the buffer
func (b *Buffer) readBigInt64LE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 8)
	value := int64(binary.LittleEndian.Uint64(bb[offset : offset+8]))

	return b.r.ToValue(big.NewInt(value))
}

// readBigUInt64BE reads a big-endian 64-bit unsigned integer from the buffer
func (b *Buffer) readBigUInt64BE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 8)
	value := binary.BigEndian.Uint64(bb[offset : offset+8])

	return b.r.ToValue(new(big.Int).SetUint64(value))
}

// readBigUInt64LE reads a little-endian 64-bit unsigned integer from the buffer
func (b *Buffer) readBigUInt64LE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 8)
	value := binary.LittleEndian.Uint64(bb[offset : offset+8])

	return b.r.ToValue(new(big.Int).SetUint64(value))
}

// readDoubleBE reads a big-endian 64-bit floating-point number from the buffer
func (b *Buffer) readDoubleBE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 8)
	value := binary.BigEndian.Uint64(bb[offset : offset+8])

	return b.r.ToValue(math.Float64frombits(value))
}

// readDoubleLE reads a little-endian 64-bit floating-point number from the buffer
func (b *Buffer) readDoubleLE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 8)
	value := binary.LittleEndian.Uint64(bb[offset : offset+8])

	return b.r.ToValue(math.Float64frombits(value))
}

// readFloatBE reads a big-endian 32-bit floating-point number from the buffer
func (b *Buffer) readFloatBE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 4)
	value := binary.BigEndian.Uint32(bb[offset : offset+4])

	return b.r.ToValue(math.Float32frombits(value))
}

// readFloatLE reads a little-endian 32-bit floating-point number from the buffer
func (b *Buffer) readFloatLE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 4)
	value := binary.LittleEndian.Uint32(bb[offset : offset+4])

	return b.r.ToValue(math.Float32frombits(value))
}

// readInt8 reads an 8-bit signed integer from the buffer
func (b *Buffer) readInt8(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 1)
	value := int8(bb[offset])

	return b.r.ToValue(value)
}

// readInt16BE reads a big-endian 16-bit signed integer from the buffer
func (b *Buffer) readInt16BE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 2)
	value := int16(binary.BigEndian.Uint16(bb[offset : offset+2]))

	return b.r.ToValue(value)
}

// readInt16LE reads a little-endian 16-bit signed integer from the buffer
func (b *Buffer) readInt16LE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 2)
	value := int16(binary.LittleEndian.Uint16(bb[offset : offset+2]))

	return b.r.ToValue(value)
}

// readInt32BE reads a big-endian 32-bit signed integer from the buffer
func (b *Buffer) readInt32BE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 4)
	value := int32(binary.BigEndian.Uint32(bb[offset : offset+4]))

	return b.r.ToValue(value)
}

// readInt32LE reads a little-endian 32-bit signed integer from the buffer
func (b *Buffer) readInt32LE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 4)
	value := int32(binary.LittleEndian.Uint32(bb[offset : offset+4]))

	return b.r.ToValue(value)
}

// readIntBE reads a big-endian signed integer of variable byte length
func (b *Buffer) readIntBE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset, byteLength := b.getVariableLengthReadArguments(call, bb)

	var value int64
	for i := int64(0); i < byteLength; i++ {
		value = (value << 8) | int64(bb[offset+i])
	}

	value = signExtend(value, byteLength)

	return b.r.ToValue(value)
}

// readIntLE reads a little-endian signed integer of variable byte length
func (b *Buffer) readIntLE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset, byteLength := b.getVariableLengthReadArguments(call, bb)

	var value int64
	for i := byteLength - 1; i >= 0; i-- {
		value = (value << 8) | int64(bb[offset+i])
	}

	value = signExtend(value, byteLength)

	return b.r.ToValue(value)
}

// readUInt8 reads an 8-bit unsigned integer from the buffer
func (b *Buffer) readUInt8(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 1)
	value := bb[offset]

	return b.r.ToValue(value)
}

// readUInt16BE reads a big-endian 16-bit unsigned integer from the buffer
func (b *Buffer) readUInt16BE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 2)
	value := binary.BigEndian.Uint16(bb[offset : offset+2])

	return b.r.ToValue(value)
}

// readUInt16LE reads a little-endian 16-bit unsigned integer from the buffer
func (b *Buffer) readUInt16LE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 2)
	value := binary.LittleEndian.Uint16(bb[offset : offset+2])

	return b.r.ToValue(value)
}

// readUInt32BE reads a big-endian 32-bit unsigned integer from the buffer
func (b *Buffer) readUInt32BE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 4)
	value := binary.BigEndian.Uint32(bb[offset : offset+4])

	return b.r.ToValue(value)
}

// readUInt32LE reads a little-endian 32-bit unsigned integer from the buffer
func (b *Buffer) readUInt32LE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset := b.getOffsetArgument(call, 0, bb, 4)
	value := binary.LittleEndian.Uint32(bb[offset : offset+4])

	return b.r.ToValue(value)
}

// readUIntBE reads a big-endian unsigned integer of variable byte length
func (b *Buffer) readUIntBE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset, byteLength := b.getVariableLengthReadArguments(call, bb)

	var value uint64
	for i := int64(0); i < byteLength; i++ {
		value = (value << 8) | uint64(bb[offset+i])
	}

	return b.r.ToValue(value)
}

// readUIntLE reads a little-endian unsigned integer of variable byte length
func (b *Buffer) readUIntLE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	offset, byteLength := b.getVariableLengthReadArguments(call, bb)

	var value uint64
	for i := byteLength - 1; i >= 0; i-- {
		value = (value << 8) | uint64(bb[offset+i])
	}

	return b.r.ToValue(value)
}

// write will write a string to the Buffer at offset according to the character encoding. The length parameter is
// the number of bytes to write. If buffer did not contain enough space to fit the entire string, only part of string
// will be written.
func (b *Buffer) write(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	str := b.getRequiredStringArgument(call, "string", 0)
	// note that we are passing in zero for numBytes, since the length parameter, which depends on offset,
	// will dictate the number of bytes
	offset := b.getOffsetArgument(call, 1, bb, 0)
	// the length defaults to the size of the buffer - offset
	maxLength := int64(len(bb)) - offset
	length := b.getOptionalIntegerArgument(call, "length", 2, maxLength)
	codec := b.getStringCodec(call.Argument(3))

	raw := codec.Decode(str)
	if int64(len(raw)) < length {
		// make sure we only write up to raw bytes
		length = int64(len(raw))
	}
	n := copy(bb[offset:], raw[:length])
	return b.r.ToValue(n)
}

// writeBigInt64BE writes a big-endian 64-bit signed integer to the buffer
func (b *Buffer) writeBigInt64BE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	value := b.getRequiredBigIntArgument(call, "value", 0)
	offset := b.getOffsetArgument(call, 1, bb, 8)

	intValue := value.Int64()
	binary.BigEndian.PutUint64(bb[offset:offset+8], uint64(intValue))

	return b.r.ToValue(offset + 8)
}

// writeBigInt64LE writes a little-endian 64-bit signed integer to the buffer
func (b *Buffer) writeBigInt64LE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	value := b.getRequiredBigIntArgument(call, "value", 0)
	offset := b.getOffsetArgument(call, 1, bb, 8)

	intValue := value.Int64()
	binary.LittleEndian.PutUint64(bb[offset:offset+8], uint64(intValue))

	return b.r.ToValue(offset + 8)
}

// writeBigUInt64BE writes a big-endian 64-bit unsigned integer to the buffer
func (b *Buffer) writeBigUInt64BE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	value := b.getRequiredBigIntArgument(call, "value", 0)
	offset := b.getOffsetArgument(call, 1, bb, 8)

	uintValue := value.Uint64()
	binary.BigEndian.PutUint64(bb[offset:offset+8], uintValue)

	return b.r.ToValue(offset + 8)
}

// writeBigUInt64LE writes a little-endian 64-bit unsigned integer to the buffer
func (b *Buffer) writeBigUInt64LE(call goja.FunctionCall) goja.Value {
	bb := Bytes(b.r, call.This)
	value := b.getRequiredBigIntArgument(call, "value", 0)
	offset := b.getOffsetArgument(call, 1, bb, 8)

	uintValue := value.Uint64()
	binary.LittleEndian.PutUint64(bb[offset:offset+8], uintValue)

	return b.r.ToValue(offset + 8)
}

func (b *Buffer) getOffsetArgument(call goja.FunctionCall, argIndex int, bb []byte, numBytes int64) int64 {
	offset := b.getOptionalIntegerArgument(call, "offset", argIndex, 0)

	if offset < 0 || offset+numBytes > int64(len(bb)) {
		panic(b.newArgumentOutOfRangeError("offset", offset))
	}

	return offset
}

func (b *Buffer) getVariableLengthReadArguments(call goja.FunctionCall, bb []byte) (int64, int64) {
	offset := b.getRequiredIntegerArgument(call, "offset", 0)
	byteLength := b.getRequiredIntegerArgument(call, "byteLength", 1)

	if byteLength < 1 || byteLength > 6 {
		panic(b.newArgumentOutOfRangeError("byteLength", byteLength))
	}
	if offset < 0 || offset+byteLength > int64(len(bb)) {
		panic(b.newArgumentOutOfRangeError("offset", offset))
	}

	return offset, byteLength
}

func signExtend(value int64, numBytes int64) int64 {
	// we don't have to turn this to a uint64 first because numBytes < 8 so
	// the sign bit will never pushed out of the int64 range
	return (value << (64 - 8*numBytes)) >> (64 - 8*numBytes)
}

func (b *Buffer) getRequiredIntegerArgument(call goja.FunctionCall, name string, argIndex int) int64 {
	arg := call.Argument(argIndex)
	if isNumber(arg) {
		return arg.ToInteger()
	}
	if goja.IsUndefined(arg) {
		panic(errors.NewTypeError(b.r, errors.ErrCodeInvalidArgType, "The \"%s\" argument is required.", name))
	}

	panic(b.newArgumentNotNumberTypeError(name))
}

func (b *Buffer) getCoercedIntegerArgument(call goja.FunctionCall, argIndex int, defaultValue int64, typeMistMatchValue int64) int64 {
	arg := call.Argument(argIndex)
	if isNumber(arg) {
		return arg.ToInteger()
	}
	if goja.IsUndefined(arg) {
		return defaultValue
	}

	return typeMistMatchValue
}

func (b *Buffer) getOptionalIntegerArgument(call goja.FunctionCall, name string, argIndex int, defaultValue int64) int64 {
	arg := call.Argument(argIndex)
	if isNumber(arg) {
		return arg.ToInteger()
	}
	if goja.IsUndefined(arg) {
		return defaultValue
	}

	panic(b.newArgumentNotNumberTypeError(name))
}

var typeBigInt = reflect.TypeOf((*big.Int)(nil))

func (b *Buffer) getRequiredBigIntArgument(call goja.FunctionCall, name string, argIndex int) *big.Int {
	arg := call.Argument(argIndex)
	if goja.IsUndefined(arg) {
		panic(errors.NewTypeError(b.r, errors.ErrCodeInvalidArgType, "The \"%s\" argument is required.", name))
	}
	if arg.ExportType() != typeBigInt {
		panic(b.newArgumentNotBigIntTypeError(name))
	}

	return arg.Export().(*big.Int)
}

func (b *Buffer) getRequiredStringArgument(call goja.FunctionCall, name string, argIndex int) string {
	arg := call.Argument(argIndex)
	if isString(arg) {
		return arg.String()
	}
	if goja.IsUndefined(arg) {
		panic(errors.NewTypeError(b.r, errors.ErrCodeInvalidArgType, "The \"%s\" argument is required.", name))
	}

	panic(b.newArgumentNotStringTypeError(name))
}

func (b *Buffer) newArgumentNotBigIntTypeError(name string) *goja.Object {
	return b.newNotCorrectTypeError(name, "BigInt")
}

func (b *Buffer) newArgumentNotStringTypeError(name string) *goja.Object {
	return b.newNotCorrectTypeError(name, "string")
}

func (b *Buffer) newArgumentNotNumberTypeError(name string) *goja.Object {
	return b.newNotCorrectTypeError(name, "number")
}

func (b *Buffer) newNotCorrectTypeError(name, _type string) *goja.Object {
	return errors.NewTypeError(b.r, errors.ErrCodeInvalidArgType, "The \"%s\" argument must be of type %s.", name, _type)
}

func (b *Buffer) newArgumentOutOfRangeError(name string, v int64) *goja.Object {
	return errors.NewRangeError(b.r, errors.ErrCodeOutOfRange, "The value of \"%s\" %d is out of range", name, v)
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	b := &Buffer{r: runtime}
	uint8Array := runtime.Get("Uint8Array")
	if c, ok := goja.AssertConstructor(uint8Array); ok {
		b.uint8ArrayCtor = c
	} else {
		panic(runtime.NewTypeError("Uint8Array is not a constructor"))
	}
	uint8ArrayObj := uint8Array.ToObject(runtime)

	ctor := runtime.ToValue(b.ctor).ToObject(runtime)
	ctor.SetPrototype(uint8ArrayObj)
	ctor.DefineDataPropertySymbol(symApi, runtime.ToValue(b), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_FALSE)
	b.bufferCtorObj = ctor
	b.uint8ArrayCtorObj = uint8ArrayObj

	proto := runtime.NewObject()
	proto.SetPrototype(uint8ArrayObj.Get("prototype").ToObject(runtime))
	proto.DefineDataProperty("constructor", ctor, goja.FLAG_TRUE, goja.FLAG_TRUE, goja.FLAG_FALSE)
	proto.Set("equals", b.proto_equals)
	proto.Set("toString", b.proto_toString)
	proto.Set("readBigInt64BE", b.readBigInt64BE)
	proto.Set("readBigInt64LE", b.readBigInt64LE)
	proto.Set("readBigUInt64BE", b.readBigUInt64BE)
	// aliases for readBigUInt64BE
	proto.Set("readBigUint64BE", b.readBigUInt64BE)

	proto.Set("readBigUInt64LE", b.readBigUInt64LE)
	// aliases for readBigUInt64LE
	proto.Set("readBigUint64LE", b.readBigUInt64LE)

	proto.Set("readDoubleBE", b.readDoubleBE)
	proto.Set("readDoubleLE", b.readDoubleLE)
	proto.Set("readFloatBE", b.readFloatBE)
	proto.Set("readFloatLE", b.readFloatLE)
	proto.Set("readInt8", b.readInt8)
	proto.Set("readInt16BE", b.readInt16BE)
	proto.Set("readInt16LE", b.readInt16LE)
	proto.Set("readInt32BE", b.readInt32BE)
	proto.Set("readInt32LE", b.readInt32LE)
	proto.Set("readIntBE", b.readIntBE)
	proto.Set("readIntLE", b.readIntLE)
	proto.Set("readUInt8", b.readUInt8)
	// aliases for readUInt8
	proto.Set("readUint8", b.readUInt8)

	proto.Set("readUInt16BE", b.readUInt16BE)
	// aliases for readUInt16BE
	proto.Set("readUint16BE", b.readUInt16BE)

	proto.Set("readUInt16LE", b.readUInt16LE)
	// aliases for readUInt16LE
	proto.Set("readUint16LE", b.readUInt16LE)

	proto.Set("readUInt32BE", b.readUInt32BE)
	// aliases for readUInt32BE
	proto.Set("readUint32BE", b.readUInt32BE)

	proto.Set("readUInt32LE", b.readUInt32LE)
	// aliases for readUInt32LE
	proto.Set("readUint32LE", b.readUInt32LE)

	proto.Set("readUIntBE", b.readUIntBE)
	// aliases for readUIntBE
	proto.Set("readUintBE", b.readUIntBE)

	proto.Set("readUIntLE", b.readUIntLE)
	// aliases for readUIntLE
	proto.Set("readUintLE", b.readUIntLE)
	proto.Set("write", b.write)
	proto.Set("writeBigInt64BE", b.writeBigInt64BE)
	proto.Set("writeBigInt64LE", b.writeBigInt64LE)
	proto.Set("writeBigUInt64BE", b.writeBigUInt64BE)
	// aliases for writeBigUInt64BE
	proto.Set("writeBigUint64BE", b.writeBigUInt64BE)

	proto.Set("writeBigUInt64LE", b.writeBigUInt64LE)
	// aliases for writeBigUInt64LE
	proto.Set("writeBigUint64LE", b.writeBigUInt64LE)

	ctor.Set("prototype", proto)
	ctor.Set("poolSize", 8192)
	ctor.Set("from", b.from)
	ctor.Set("alloc", b.alloc)

	exports := module.Get("exports").(*goja.Object)
	exports.Set("Buffer", ctor)
}

func init() {
	require.RegisterCoreModule(ModuleName, Require)
}
