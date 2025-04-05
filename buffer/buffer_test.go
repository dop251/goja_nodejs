package buffer

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

func TestBufferFrom(t *testing.T) {
	vm := goja.New()
	new(require.Registry).Enable(vm)

	_, err := vm.RunString(`
	const Buffer = require("node:buffer").Buffer;

	function checkBuffer(buf) {
		if (!(buf instanceof Buffer)) {
			throw new Error("instanceof Buffer");
		}
	
		if (!(buf instanceof Uint8Array)) {
			throw new Error("instanceof Uint8Array");
		}
	}

	checkBuffer(Buffer.from(new ArrayBuffer(16)));
	checkBuffer(Buffer.from(new Uint16Array(8)));

	{
		const b = Buffer.from("\xff\xfe\xfd");
		const h = b.toString("hex")
		if (h !== "c3bfc3bec3bd") {
			throw new Error(h);
		}
	}

	{
		const b = Buffer.from("0102fffdXXX", "hex");
		checkBuffer(b);
		if (b.toString("hex") !== "0102fffd") {
			throw new Error(b.toString("hex"));
		}
	}

	{
		const b = Buffer.from('1ag123', 'hex');
		if (b.length !== 1 || b[0] !== 0x1a) {
			throw new Error(b);
		}
	}

	{
		const b = Buffer.from('1a7', 'hex');
		if (b.length !== 1 || b[0] !== 0x1a) {
			throw new Error(b);
		}
	}

	{
		const b = Buffer.from("\uD801", "utf-8");
		if (b.length !== 3 || b[0] !== 0xef || b[1] !== 0xbf || b[2] !== 0xbd) {
			throw new Error(b);
		}
	}
	`)

	if err != nil {
		t.Fatal(err)
	}
}

func TestFromBase64(t *testing.T) {
	vm := goja.New()
	new(require.Registry).Enable(vm)

	_, err := vm.RunString(`
	const Buffer = require("node:buffer").Buffer;

	{
		let b = Buffer.from("AAA_", "base64");
		if (b.length !== 3 || b[0] !== 0 || b[1] !== 0 || b[2] !== 0x3f) {
			throw new Error(b.toString("hex"));
		}

		let r = b.toString("base64");
		if (r !== "AAA/") {
			throw new Error("to base64: " + r);
		}
		for (let i = 0; i < 20; i++) {
			let s = "A".repeat(i) + "_" + "A".repeat(20-i);
			let s1 = "A".repeat(i) + "/" + "A".repeat(20-i);
			let b = Buffer.from(s, "base64");
			let b1 = Buffer.from(s1, "base64");
			if (!b.equals(b1)) {
				throw new Error(s);
			}
		}
	}

	{
		let b = Buffer.from("SQ==???", "base64");
		if (b.length !== 1 || b[0] != 0x49) {
			throw new Error(b.toString("hex"));
		}
	}

	{
		let s = Buffer.from("AyM1SysPpbyDfgZld3umj1qzKObwVMkoqQ-EstJQLr_T-1qS0gZH75aKtMN3Yj0iPS4hcgUuTwjAzZr1Z9CAow", "base64Url").toString("base64");
		if (s !== "AyM1SysPpbyDfgZld3umj1qzKObwVMkoqQ+EstJQLr/T+1qS0gZH75aKtMN3Yj0iPS4hcgUuTwjAzZr1Z9CAow==") {
			throw new Error(s);
		}
	}
	`)

	if err != nil {
		t.Fatal(err)
	}
}

func TestWrapBytes(t *testing.T) {
	vm := goja.New()
	new(require.Registry).Enable(vm)
	b := []byte{1, 2, 3}
	buffer := GetApi(vm)
	vm.Set("b", buffer.WrapBytes(b))
	Enable(vm)
	_, err := vm.RunString(`
		if (typeof Buffer !== "function") {
			throw new Error("Buffer is not a function: " + typeof Buffer);
		}
		if (!(b instanceof Buffer)) {
			throw new Error("instanceof Buffer");
		}
		if (b.toString("hex") !== "010203") {
			throw new Error(b);
		}
	`)

	if err != nil {
		t.Fatal(err)
	}
}

func TestBuffer_alloc(t *testing.T) {
	vm := goja.New()
	new(require.Registry).Enable(vm)

	_, err := vm.RunString(`
	const Buffer = require("node:buffer").Buffer;

	{
		const b = Buffer.alloc(2, "abc");
		if (b.toString() !== "ab") {
			throw new Error(b);
		}
	}

	{
		const b = Buffer.alloc(16, "abc");
		if (b.toString() !== "abcabcabcabcabca") {
			throw new Error(b);
		}
	}

	{
		const fill = {
			valueOf() {
				return 0xac;
			}
		}
		const b = Buffer.alloc(8, fill);
		if (b.toString("hex") !== "acacacacacacacac") {
			throw new Error(b);
		}
	}

	{
		const fill = {
			valueOf() {
				return Infinity;
			}
		}
		const b = Buffer.alloc(2, fill);
		if (b.toString("hex") !== "0000") {
			throw new Error(b);
		}
	}

	{
		const fill = {
			valueOf() {
				return "ac";
			}
		}
		const b = Buffer.alloc(2, fill);
		if (b.toString("hex") !== "0000") {
			throw new Error(b);
		}
	}

	{
		const b = Buffer.alloc(2, -257.4);
		if (b.toString("hex") !== "ffff") {
			throw new Error(b);
		}
	}

	{
		const b = Buffer.alloc(2, Infinity);
		if (b.toString("hex") !== "0000") {
			throw new Error("Infinity: " + b.toString("hex"));
		}
	}

	{
		const b = Buffer.alloc(2, null);
		if (b.toString("hex") !== "0000") {
			throw new Error("Infinity: " + b.toString("hex"));
		}
	}

	`)

	if err != nil {
		t.Fatal(err)
	}
}

//go:embed testdata/assertions.js
var assertionsSource string

type testCase struct {
	name        string
	script      string
	expectedErr string
}

func runTestCases(t *testing.T, tcs []testCase) {
	vm := goja.New()
	new(require.Registry).Enable(vm)
	_, err := vm.RunScript("testdata/assertions.js", assertionsSource)
	if err != nil {
		t.Fatal(err)
	}
	_, err = vm.RunString(`const Buffer = require("node:buffer").Buffer;`)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			template := `
            {
				%s
			}
            `
			_, err := vm.RunString(fmt.Sprintf(template, tc.script))
			if tc.expectedErr != "" {
				if err == nil {
					t.Errorf("expected error %q, but got none", tc.expectedErr)
				} else {
					if !strings.HasPrefix(err.Error(), tc.expectedErr) {
						t.Errorf("expected error %q, got %q", tc.expectedErr, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got %q", err.Error())
				}
			}
		})
	}
}

func TestBuffer_readBigInt64BE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
				if (b.readBigInt64BE(0) !== BigInt(4294967295)) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
				if (b.readBigInt64BE() !== BigInt(4294967295)) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readBigInt64LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
				if (b.readBigInt64LE(0) !== BigInt(-4294967296)) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with out of range offset",
			script: `
				const b =  Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
				// this should error 
				b.readBigInt64LE(1);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 1 is out of range`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readBigUInt64BE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
				if (b.readBigUInt64BE(0) !== BigInt(4294967295)) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
				if (b.readBigUInt64BE() !== BigInt(4294967295)) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "use alias",
			script: `
				const b = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
				if (b.readBigUint64BE() !== BigInt(4294967295)) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readBigUInt64LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
				if (b.readBigUInt64LE(0) !== BigInt(18446744069414584320)) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with out of range offset",
			script: `
				const b =  Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
				// this should error 
				b.readBigUInt64LE(1);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 1 is out of range`,
		},
		{
			name: "use alias",
			script: `
				const b = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
				if (b.readBigUint64LE(0) !== BigInt(18446744069414584320)) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readDoubleBE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([1, 2, 3, 4, 5, 6, 7, 8]);
				if (b.readDoubleBE(0) !== 8.20788039913184e-304) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([1, 2, 3, 4, 5, 6, 7, 8]);
				if (b.readDoubleBE() !== 8.20788039913184e-304) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readDoubleLE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([1, 2, 3, 4, 5, 6, 7, 8]);
				if (b.readDoubleLE(0) !== 5.447603722011605e-270) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with out of range offset",
			script: `
				const b = Buffer.from([1, 2, 3, 4, 5, 6, 7, 8]);
				// this should error 
				b.readDoubleLE(1);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 1 is out of range`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readFloatBE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([1, 2, 3, 4]);
				if (b.readFloatBE(0) !== 2.387939260590663e-38) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([1, 2, 3, 4]);
				if (b.readFloatBE() !== 2.387939260590663e-38) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readFloatLE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([1, 2, 3, 4]);
				if (b.readFloatLE(0) !== 1.539989614439558e-36) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with out of range offset",
			script: `
				const b = Buffer.from([1, 2, 3, 4]);
				// this should error 
				b.readFloatLE(1);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 1 is out of range`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readInt8(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([-1, 5]);
				if (b.readInt8(0) !== -1) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with last offset",
			script: `
				const b = Buffer.from([-1, 5]);
				if (b.readInt8(1) !== 5) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([-1, 5]);
				if (b.readInt8() !== -1) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readInt16BE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0, 5]);
				if (b.readInt16BE(0) !== 5) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with out of range offset",
			script: `
				const b = Buffer.from([0xA, 0x5]);
				// this should error 
				b.readInt16BE(1);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 1 is out of range`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([0, 5]);
				if (b.readInt16BE() !== 5) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readInt16LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0, 5]);
				if (b.readInt16LE(0) !== 1280) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with out of range offset",
			script: `
				const b = Buffer.from([0, 5]);
				// this should error 
				b.readInt16LE(1);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 1 is out of range`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readInt32BE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0, 0, 0, 5]);
				if (b.readInt32BE(0) !== 5) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with out of range offset",
			script: `
				const b = Buffer.from([0, 0, 0, 5]);
				// this should error 
				b.readInt32BE(1);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 1 is out of range`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([0, 0, 0, 5]);
				if (b.readInt32BE() !== 5) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readInt32LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0, 0, 0, 5]);
				if (b.readInt32LE(0) !== 83886080) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with out of range offset",
			script: `
				const b = Buffer.from([0, 0, 0, 5]);
				// this should error 
				b.readInt32LE(1);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 1 is out of range`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([0, 0, 0, 5]);
				if (b.readInt32LE() !== 83886080) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestName(t *testing.T) {

}

func TestBuffer_readIntBE(t *testing.T) {
	tcs := []testCase{
		{
			name: "6 byte positive integer",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				if (b.readIntBE(0, 6) !== 20015998341291) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "1 byte negative integer",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				if (b.readIntBE(4, 1) !== -112) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no parameters",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readIntBE();
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "offset" argument is required`,
		},
		{
			name: "type mismatch for offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readIntBE("1");
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "offset" argument must be of type number`,
		},
		{
			name: "with no byteLength parameter",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readIntBE(0);
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "byteLength" argument is required`,
		},
		{
			name: "byteLength less than 1",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readIntBE(0,0);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "byteLength" 0 is out of range`,
		},
		{
			name: "byteLength greater than 7",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readIntBE(0,7);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "byteLength" 7 is out of range`,
		},
		{
			name: "offset plus byteLength out of range",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readIntBE(4,3);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 4 is out of range`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readIntLE(t *testing.T) {
	tcs := []testCase{
		{
			name: "6 byte negative integer",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				if (b.readIntLE(0, 6) !== -92837994154990) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "1 byte positive integer",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				if (b.readIntLE(0, 1) !== 18) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no parameters",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readIntLE();
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "offset" argument is required`,
		},
		{
			name: "with no byteLength parameter",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readIntLE(0);
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "byteLength" argument is required`,
		},
		{
			name: "offset plus byteLength out of range",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readIntLE(4,3);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 4 is out of range`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readUInt8(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([1, -2]);
				if (b.readUInt8(0) !== 1) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with last offset",
			script: `
				const b = Buffer.from([1, -2]);
				if (b.readUInt8(1) !== 254) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([1, -2]);
				if (b.readUInt8() !== 1) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "use alias",
			script: `
				const b = Buffer.from([1, -2]);
				if (b.readUint8() !== 1) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readUInt16BE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56]);
				if (b.readUInt16BE(0).toString(16) !== "1234") {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with last offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56]);
				if (b.readUInt16BE(1).toString(16) !== "3456") {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56]);
				if (b.readUInt16BE().toString(16) !== "1234") {
					throw new Error(b);
				}
			`,
		},
		{
			name: "use alias",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56]);
				if (b.readUint16BE().toString(16) !== "1234") {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readUInt16LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56]);
				if (b.readUInt16LE(0).toString(16) !== "3412") {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with last offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56]);
				if (b.readUInt16LE(1).toString(16) !== "5634") {
					throw new Error(b);
				}
			`,
		},
		{
			name: "out of range offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56]);
				// this should error
				b.readUInt16LE(2);	
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 2 is out of range`,
		},
		{
			name: "use alias",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56]);
				if (b.readUint16LE(1).toString(16) !== "5634") {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readUInt32BE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78]);
				if (b.readUInt32BE(0).toString(16) !== "12345678") {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78]);
				if (b.readUInt32BE().toString(16) !== "12345678") {
					throw new Error(b);
				}
			`,
		},
		{
			name: "use alias",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78]);
				if (b.readUint32BE(0).toString(16) !== "12345678") {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readUInt32LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "with zero offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78]);
				if (b.readUInt32LE(0).toString(16) !== "78563412") {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no/default offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78]);
				if (b.readUInt32LE().toString(16) !== "78563412") {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with string offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78]);
				// this should error
				b.readUInt32LE("foo"); 
				throw new Error("should not get here");// this should error
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "offset" argument must be of type number`,
		},
		{
			name: "with negative offset",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78]);
				// this should error
				b.readUInt32LE(-1);
				throw new Error("should not get here");// this should error
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" -1 is out of range`,
		},
		{
			name: "use alias",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78]);
				if (b.readUint32LE(0).toString(16) !== "78563412") {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readUIntBE(t *testing.T) {
	tcs := []testCase{
		{
			name: "6 byte integer",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				if (b.readUIntBE(0, 6) !== 20015998341291) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "1 byte integer",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				if (b.readUIntBE(1, 1) !== 52) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no parameters",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readUIntBE();
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "offset" argument is required`,
		},
		{
			name: "with no byteLength parameter",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readUIntBE(0);
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "byteLength" argument is required`,
		},
		{
			name: "offset plus byteLength out of range",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readUIntBE(4,3);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 4 is out of range`,
		},
		{
			name: "use alias",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				if (b.readUintBE(1, 1) !== 52) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_readUIntLE(t *testing.T) {
	tcs := []testCase{
		{
			name: "6 byte integer",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				if (b.readUIntLE(0, 6) !== 188636982555666) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "1 byte integer",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				if (b.readUIntLE(1, 1) !== 52) {
					throw new Error(b);
				}
			`,
		},
		{
			name: "with no parameters",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readUIntLE();
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "offset" argument is required`,
		},
		{
			name: "with no byteLength parameter",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readUIntLE(0);
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "byteLength" argument is required`,
		},
		{
			name: "offset plus byteLength out of range",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				// this should error 
				b.readUIntLE(4,3);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 4 is out of range`,
		},
		{
			name: "use alias",
			script: `
				const b = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
				if (b.readUintLE(1, 1) !== 52) {
					throw new Error(b);
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_toString(t *testing.T) {
	tcs := []testCase{
		{
			name: "with no parameters",
			script: `
				const buf = Buffer.alloc(5);
  				buf.write('hello');

				if (buf.toString() !== 'hello') {
					throw new Error('should return "hello"');
				}
			`,
		},
		{
			name: "utf8 encoding",
			script: `
				const buf = Buffer.from([0x7E]);
				if (buf.toString('utf8') !== '~') {
					throw new Error('should return "~"');
				}
			`,
		},
		{
			name: "with valid start and valid end",
			script: `
				const buf = Buffer.alloc(10);
				buf.write('hello world');

				if (buf.toString('utf8', 6, 10) !== 'worl') {
					throw new Error('should return "worl"');
				}
			`,
		},
		{
			name: "with start=0 and end=0",
			script: `
				const buf = Buffer.alloc(10);
				buf.write('hello world');

				if (buf.toString('utf8', 0, 0) !== '') {
					throw new Error('should return empty');
				}
			`,
		},
		{
			name: "with start > end",
			script: `
				const buf = Buffer.alloc(10);
				buf.write('hello world');

				if (buf.toString('utf8', 3, 2) !== '') {
					throw new Error('should return empty');
				}
			`,
		},
		{
			name: "with start < 0",
			script: `
				const buf = Buffer.alloc(10);
				buf.write('hello world');

				if (buf.toString('utf8', -1, 2) !== 'he') {
					throw new Error('should return "he"');
				}
			`,
		},
		{
			name: "with start > buffer length",
			script: `
				const buf = Buffer.alloc(10);
				buf.write('hello world');

				if (buf.toString('utf8', 100, 2) !== '') {
					throw new Error('should return empty');
				}
			`,
		},
		{
			name: "with end > buffer length",
			script: `
				const buf = Buffer.alloc(10);
				buf.write('hello world');

				if (buf.toString('utf8', 1, 100) !== 'ello worl') {
					throw new Error('should return "ello worl"');
				}
			`,
		},
		{
			name: "with start == buffer length",
			script: `
				const buf = Buffer.alloc(10);
				buf.write('hello world');

				if (buf.toString('utf8', 10, 2) !== '') {
					throw new Error('should return empty');
				}
			`,
		},
		{
			name: "with non-numeric start",
			script: `
				const buf = Buffer.alloc(10);
				buf.write('hello world');

				if (buf.toString('utf8', {}, 2) !== 'he') {
					throw new Error('should return "he"');
				}
			`,
		},
		{
			name: "with float start",
			script: `
				const buf = Buffer.alloc(10);
				buf.write('hello world');

				if (buf.toString('utf8', 3.5, 10) !== 'lo worl') {
					throw new Error('should return "lo worl"');
				}
			`,
		},
		{
			name: "with float end",
			script: `
				const buf = Buffer.alloc(10);
				buf.write('hello world');

				if (buf.toString('utf8', 0, 4.9) !== 'hell') {
					throw new Error('should return "hell"');
				}
			`,
		},
		{
			name: "with multi-byte character",
			script: `
				const buf = Buffer.from([0xE2, 0x82, 0xAC]);
				
				if (buf.toString('utf8') !== '€') {
					throw new Error('should return "€"');
				}
			`,
		},
		{
			name: "with partitial multi-byte character",
			script: `
				const buf = Buffer.from([0xE2, 0x82, 0xAC, 0xE2, 0x82, 0xAC]);
				
				if (buf.toString('utf8',0, 4) !== '€�') {
					throw new Error('should return "€�"');
				}
			`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_write(t *testing.T) {
	tcs := []testCase{
		{
			name: "write string with defaults",
			script: `
				const buf = Buffer.alloc(10);
				const bytesWritten = buf.write('hello');

				if (bytesWritten !== 5) {
					throw new Error('bytesWritten should be 5');

  				} else if (buf.toString('utf8', 0, 5) !== 'hello') {
					throw new Error('buffer content should be "hello"');

  				} else if (buf.toString('utf8', 5, 10) !== '\0\0\0\0\0') {
					throw new Error('remaining buffer should be zeros');
  				} 
			`,
		},
		{
			name: "write at offset",
			script: `
				const buf = Buffer.alloc(10);
				const bytesWritten = buf.write('world', 5);

				if (bytesWritten !== 5) {
					throw new Error('bytesWritten should be 5');

  				} else if (buf.toString('utf8', 5, 10) !== 'world') {
					throw new Error('buffer content should be "world"');

  				} else if (buf.toString('utf8', 0, 5) !== '\0\0\0\0\0') {
					throw new Error('first 5 bytes should be zeros');
  				} 
			`,
		},
		{
			name: "write with offset and length",
			script: `
				const buf = Buffer.alloc(10);
				const bytesWritten = buf.write('hello world', 0, 5);

				if (bytesWritten !== 5) {
					throw new Error('bytesWritten should be 5');

  				} else if (buf.toString('utf8', 0, 5) !== 'hello') {
					throw new Error('buffer content should be "hello"');

  				} else if (buf.toString('utf8', 5, 10) !== '\0\0\0\0\0') {
					throw new Error('remaining buffer should be zeros');
  				} 
			`,
		},
		{
			name: "write at offset zero",
			script: `
				const buf = Buffer.alloc(5);
				const bytesWritten = buf.write('abc', 0);

				if (bytesWritten !== 3) {
					throw new Error('bytesWritten should be 3');

  				} else if (buf.toString('utf8', 0, 3) !== 'abc') {
					throw new Error('buffer content should be "abc"');
  				} 
			`,
		},
		{
			name: "write at last offset",
			script: `
				const buf = Buffer.alloc(5);
				const bytesWritten = buf.write('a', 4);

				if (bytesWritten !== 1) {
					throw new Error('bytesWritten should be 3');

  				} else if (buf[4] !== 'a'.charCodeAt(0)) {
					throw new Error('buf[4] should be "a"');
  				} 
			`,
		},
		{
			name: "write with length zero",
			script: `
				const buf = Buffer.alloc(5);
				const bytesWritten = buf.write('abc', 0, 0);

				if (bytesWritten !== 0) {
					throw new Error('bytesWritten should be 0');

  				} else if (buf.toString('utf8', 0, 5) !== '\0\0\0\0\0') {
					throw new Error('buffer should remain zeros');
  				} 
			`,
		},
		{
			name: "write with length greater than string length",
			script: `
				const buf = Buffer.alloc(5);
				const bytesWritten = buf.write('abc', 0, 5);

				if (bytesWritten !== 3) {
					throw new Error('bytesWritten should be 3');

  				} else if (buf.toString('utf8', 0, 3) !== 'abc') {
					throw new Error('buffer content should be "abc"');
  				} 
			`,
		},
		{
			name: "write with offset + length exceeding buffer length",
			script: `
				const buf = Buffer.alloc(5);
				const bytesWritten = buf.write('abcde', 3);

				if (bytesWritten !== 2) {
					throw new Error('bytesWritten should be 2');

  				} else if (buf.toString('utf8', 3, 5) !== 'ab') {
					throw new Error('buffer content at 3-5 should be "ab"');
  				} 
			`,
		},
		{
			name: "invalid encoding",
			script: `
				const buf = Buffer.alloc(5);

				// this should error
				buf.write('a', 0, 1, 'invalid');
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_UNKNOWN_ENCODING]: Unknown encoding: invalid`,
		},
		{
			name: "offset out of range",
			script: `
				const buf = Buffer.alloc(5);

				// this should error
				buf.write('abc', 10);
				throw new Error("should not get here");
			`,
			expectedErr: `RangeError [ERR_OUT_OF_RANGE]: The value of "offset" 10 is out of range`,
		},
		{
			name: "with no parameters",
			script: `
				const buf = Buffer.alloc(5);
				// this should error 
				buf.write();
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "string" argument is required`,
		},
		{
			name: "argument not string type",
			script: `
				const buf = Buffer.alloc(5);
				// this should error 
				buf.write(1);
				throw new Error("should not get here");
			`,
			expectedErr: `TypeError [ERR_INVALID_ARG_TYPE]: The "string" argument must be of type string`,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeBigInt64BE(t *testing.T) {

	tcs := []testCase{
		{
			name: "write with default offset",
			script: `
				const buf = Buffer.alloc(16);
				assertBufferWriteRead(buf, 'writeBigInt64BE', 'readBigInt64BE', BigInt(123456789)); 
            `,
		},
		{
			name: "write negative number, zero offset",
			script: `
				const buf = Buffer.alloc(16);
                assertBufferWriteRead(buf, 'writeBigInt64BE', 'readBigInt64BE', BigInt(123456789), 0);
            `,
		},
		{
			name: "write at non-zero offset",
			script: `
				const buf = Buffer.alloc(16);
                assertBufferWriteRead(buf, 'writeBigInt64BE', 'readBigInt64BE', BigInt(123456789), 4);
            `,
		},
		{
			name: "write with max int64 value",
			script: `
				const buf = Buffer.alloc(16);
				assertBufferWriteRead(buf, 'writeBigInt64BE', 'readBigInt64BE', BigInt("9223372036854775807")); // 2^63 - 1
            `,
		},
		{
			name: "write with min int64 value",
			script: `
				const buf = Buffer.alloc(16);
				assertBufferWriteRead(buf, 'writeBigInt64BE', 'readBigInt64BE', BigInt("-9223372036854775808")); // -2^63
            `,
		},
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(8);
				assert.throwsNodeErrorWithMessage(() => buf.writeBigInt64BE(BigInt(1), 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');
            `,
		},
		{
			name: "write with negative offset",
			script: `
				const buf = Buffer.alloc(16);
				assert.throwsNodeErrorWithMessage(() => buf.writeBigInt64BE(BigInt(1), -1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" -1 is out of range.');
            `,
		},
		{
			name: "write without required value",
			script: `
				const buf = Buffer.alloc(16);
				assert.throwsNodeErrorWithMessage(() => buf.writeBigInt64BE(), TypeError, "ERR_INVALID_ARG_TYPE", 'The "value" argument is required.');
            `,
		},
		{
			name: "write with number instead of BigInt",
			script: `
				const buf = Buffer.alloc(16);
				assert.throwsNodeErrorWithMessage(() => buf.writeBigInt64BE(123456789, 0), TypeError, "ERR_INVALID_ARG_TYPE", 'The "value" argument must be of type BigInt.'); 
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(24);
				buf.writeBigInt64BE(BigInt(123));
				buf.writeBigInt64BE(BigInt(456), 8);
				buf.writeBigInt64BE(BigInt(789), 16);
				
				assertValueRead(buf.readBigInt64BE(0), BigInt(123));
				assertValueRead(buf.readBigInt64BE(8), BigInt(456));
				assertValueRead(buf.readBigInt64BE(16), BigInt(789));
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeBigInt64LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with default offset",
			script: `
				const buf = Buffer.alloc(16);
				assertBufferWriteRead(buf, 'writeBigInt64LE', 'readBigInt64LE', BigInt(123456789)); 	
            `,
		},
		{
			name: "write negative number, zero offset",
			script: `
				const buf = Buffer.alloc(16);
				assertBufferWriteRead(buf, 'writeBigInt64LE', 'readBigInt64LE', BigInt(-123456789), 0); 	
            `,
		},
		{
			name: "write at non-zero offset",
			script: `
				const buf = Buffer.alloc(16);
                assertBufferWriteRead(buf, 'writeBigInt64LE', 'readBigInt64LE', BigInt(123456789), 4);
            `,
		},
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(8);
				assert.throwsNodeErrorWithMessage(() => buf.writeBigInt64LE(BigInt(1), 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');	
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(24);
				buf.writeBigInt64LE(BigInt(123));
				buf.writeBigInt64LE(BigInt(456), 8);
				buf.writeBigInt64LE(BigInt(789), 16);
			
				assertValueRead(buf.readBigInt64LE(0), BigInt(123));
				assertValueRead(buf.readBigInt64LE(8), BigInt(456));
				assertValueRead(buf.readBigInt64LE(16), BigInt(789));
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeBigUInt64BE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with default offset",
			script: `
				const buf = Buffer.alloc(16);
				assertBufferWriteRead(buf, 'writeBigUInt64BE', 'readBigUInt64BE', BigInt(123456789));
            `,
		},
		{
			name: "read/write using alias",
			script: `
				const buf = Buffer.alloc(16);	
				assertBufferWriteRead(buf, 'writeBigUint64BE', 'readBigUint64BE', BigInt(123456789));	
            `,
		},
		{
			name: "write at non-zero offset",
			script: `
				const buf = Buffer.alloc(16);
				assertBufferWriteRead(buf, 'writeBigUInt64BE', 'readBigUInt64BE', BigInt(123456789), 4);
            `,
		},
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(8);
				assert.throwsNodeErrorWithMessage(() => buf.writeBigUInt64BE(BigInt(1), 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');	  
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(24);
				buf.writeBigUInt64BE(BigInt(123));
				buf.writeBigUInt64BE(BigInt(456), 8);
				buf.writeBigUInt64BE(BigInt(789), 16);
				
				assertValueRead(buf.readBigUInt64BE(0), BigInt(123));
				assertValueRead(buf.readBigUInt64BE(8), BigInt(456));
				assertValueRead(buf.readBigUInt64BE(16), BigInt(789));
            `,
		},
		{
			name: "write large unsigned value",
			script: `
				const buf = Buffer.alloc(16);
				assertBufferWriteRead(buf, 'writeBigUInt64BE', 'readBigUInt64BE',  BigInt("9007199254740991")); // MAX_SAFE_INTEGER
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeBigUInt64LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with default offset",
			script: `
				const buf = Buffer.alloc(16);
				assertBufferWriteRead(buf, 'writeBigUInt64LE', 'readBigUInt64LE',  BigInt(123456789));
            `,
		},
		{
			name: "read/write using alias",
			script: `
				const buf = Buffer.alloc(16);
				assertBufferWriteRead(buf, 'writeBigUint64LE', 'readBigUint64LE',  BigInt(123456789));
            `,
		},
		{
			name: "write at non-zero offset",
			script: `
				const buf = Buffer.alloc(16);	
				assertBufferWriteRead(buf, 'writeBigUint64LE', 'readBigUint64LE',  BigInt(123456789), 4);
            `,
		},
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(8);
				assert.throwsNodeErrorWithMessage(() => buf.writeBigUInt64LE(BigInt(1), 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');	  
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(24);
				buf.writeBigUInt64LE(BigInt(123));
				buf.writeBigUInt64LE(BigInt(456), 8);
				buf.writeBigUInt64LE(BigInt(789), 16);
				
				assertValueRead(buf.readBigUInt64LE(0), BigInt(123));
				assertValueRead(buf.readBigUInt64LE(8), BigInt(456));
				assertValueRead(buf.readBigUInt64LE(16), BigInt(789));
            `,
		},
		{
			name: "write max uint64 value",
			script: `
				const buf = Buffer.alloc(16);
                assertBufferWriteRead(buf, 'writeBigUint64LE', 'readBigUint64LE', BigInt("18446744073709551615")); // 2^64 - 1
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeDoubleBE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(8);
				assert.throwsNodeErrorWithMessage(() => buf.writeDoubleBE(123.456, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(24);
				buf.writeDoubleBE(123.1);    // default offset of zero
				buf.writeDoubleBE(456.4, 8);
				buf.writeDoubleBE(789.7, 16);
				
				assertValueRead(buf.readDoubleBE(0), 123.1);
				assertValueRead(buf.readDoubleBE(8), 456.4);
				assertValueRead(buf.readDoubleBE(16), 789.7);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeDoubleLE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(8);
				assert.throwsNodeErrorWithMessage(() => buf.writeDoubleLE(123.456, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(24);
				buf.writeDoubleLE(123.1);    // default offset of zero
				buf.writeDoubleLE(456.4, 8);
				buf.writeDoubleLE(789.7, 16);
				
				assertValueRead(buf.readDoubleLE(0), 123.1);
				assertValueRead(buf.readDoubleLE(8), 456.4);
				assertValueRead(buf.readDoubleLE(16), 789.7);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeFloatBE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(4);
				assert.throwsNodeErrorWithMessage(() => buf.writeFloatBE(123.456, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(4);
				// above the max
				assert.throwsNodeErrorWithMessage(() => buf.writeFloatBE(3.5e+38, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 3.5e+38 is out of range.');
				// below the min
				assert.throwsNodeErrorWithMessage(() => buf.writeFloatBE(-3.5e+38, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -3.5e+38 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(12);
				buf.writeFloatBE(123.1);    // default offset of zero
				buf.writeFloatBE(456.4, 4);
				buf.writeFloatBE(789.7, 8);
				
				assertValueRead(buf.readFloatBE(0), 123.1);
				assertValueRead(buf.readFloatBE(4), 456.4);
				assertValueRead(buf.readFloatBE(8), 789.7);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeFloatLE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(4);
				assert.throwsNodeErrorWithMessage(() => buf.writeFloatLE(123.456, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(4);
				// above the max
				assert.throwsNodeErrorWithMessage(() => buf.writeFloatLE(3.5e+38, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 3.5e+38 is out of range.');
				// below the min
				assert.throwsNodeErrorWithMessage(() => buf.writeFloatLE(-3.5e+38, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -3.5e+38 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(12);
				buf.writeFloatLE(123.1);    // default offset of zero
				buf.writeFloatLE(456.4, 4);
				buf.writeFloatLE(789.7, 8);
				
				assertValueRead(buf.readFloatLE(0), 123.1);
				assertValueRead(buf.readFloatLE(4), 456.4);
				assertValueRead(buf.readFloatLE(8), 789.7);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeInt8(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(2);
				assert.throwsNodeErrorWithMessage(() => buf.writeInt8(2, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 2 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(4);
				// above the max
				assert.throwsNodeErrorWithMessage(() => buf.writeInt8(128), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 128 is out of range.');
				// below the min
				assert.throwsNodeErrorWithMessage(() => buf.writeInt8(-129), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -129 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(3);
				buf.writeInt8(3);    // default offset of zero
				buf.writeInt8(2, 1);
				buf.writeInt8(1, 2);
				
				assertValueRead(buf.readInt8(0), 3);
				assertValueRead(buf.readInt8(1), 2);
				assertValueRead(buf.readInt8(2), 1);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeInt16BE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(2);
				assert.throwsNodeErrorWithMessage(() => buf.writeInt16BE(2, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(4);
				// above the max
				assert.throwsNodeErrorWithMessage(() => buf.writeInt16BE(32768), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 32768 is out of range.');
				// below the min
				assert.throwsNodeErrorWithMessage(() => buf.writeInt16BE(-32769), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -32769 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(6);
				buf.writeInt16BE(3);    // default offset of zero
				buf.writeInt16BE(2, 2);
				buf.writeInt16BE(1, 4);
				
				assertValueRead(buf.readInt16BE(0), 3);
				assertValueRead(buf.readInt16BE(2), 2);
				assertValueRead(buf.readInt16BE(4), 1);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeInt16LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(2);
				assert.throwsNodeErrorWithMessage(() => buf.writeInt16LE(2, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(4);
				// above the max
				assert.throwsNodeErrorWithMessage(() => buf.writeInt16LE(32768), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 32768 is out of range.');
				// below the min
				assert.throwsNodeErrorWithMessage(() => buf.writeInt16LE(-32769), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -32769 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(6);
				buf.writeInt16LE(3);    // default offset of zero
				buf.writeInt16LE(2, 2);
				buf.writeInt16LE(1, 4);
				
				assertValueRead(buf.readInt16LE(0), 3);
				assertValueRead(buf.readInt16LE(2), 2);
				assertValueRead(buf.readInt16LE(4), 1);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeInt32BE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(4);
				assert.throwsNodeErrorWithMessage(() => buf.writeInt32BE(2, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(4);
				// above the max
				assert.throwsNodeErrorWithMessage(() => buf.writeInt32BE(2147483648), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 2147483648 is out of range.');
				// below the min
				assert.throwsNodeErrorWithMessage(() => buf.writeInt32BE(-2147483649), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -2147483649 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(12);
				buf.writeInt32BE(3);    // default offset of zero
				buf.writeInt32BE(2, 4);
				buf.writeInt32BE(1, 8);
				
				assertValueRead(buf.readInt32BE(0), 3);
				assertValueRead(buf.readInt32BE(4), 2);
				assertValueRead(buf.readInt32BE(8), 1);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeInt32LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(4);
				assert.throwsNodeErrorWithMessage(() => buf.writeInt32LE(2, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 1 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(4);
				// above the max
				assert.throwsNodeErrorWithMessage(() => buf.writeInt32LE(2147483648), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 2147483648 is out of range.');
				// below the min
				assert.throwsNodeErrorWithMessage(() => buf.writeInt32LE(-2147483649), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -2147483649 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(12);
				buf.writeInt32LE(3);    // default offset of zero
				buf.writeInt32LE(2, 4);
				buf.writeInt32LE(1, 8);
				
				assertValueRead(buf.readInt32LE(0), 3);
				assertValueRead(buf.readInt32LE(4), 2);
				assertValueRead(buf.readInt32LE(8), 1);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeIntBE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write out of range offset",
			script: `
				const buf = Buffer.alloc(6);
				assert.throwsNodeErrorWithMessage(() => buf.writeIntBE(127, 6, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 6 is out of range.');
            `,
		},
		{
			name: "byteLength out of range",
			script: `
				const buf = Buffer.alloc(6);
				
				// above the max
				assert.throwsNodeErrorWithMessage(() => buf.writeIntBE(0, 0, 7), RangeError, "ERR_OUT_OF_RANGE", 'The value of "byteLength" 7 is out of range.');
				// below the min
				assert.throwsNodeErrorWithMessage(() => buf.writeIntBE(0, 0, 0), RangeError, "ERR_OUT_OF_RANGE", 'The value of "byteLength" 0 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(6);
				// above the 6-byte max
				assert.throwsNodeErrorWithMessage(() => buf.writeIntBE(140737488355328, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 140737488355328 is out of range.');
				// below the 6-byte min
				assert.throwsNodeErrorWithMessage(() => buf.writeIntBE(-140737488355329, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -140737488355329 is out of range.');
				// above the 2-byte max
				assert.throwsNodeErrorWithMessage(() => buf.writeIntBE(32768, 0, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 32768 is out of range.');
				// below the 2-byte min
				assert.throwsNodeErrorWithMessage(() => buf.writeIntBE(-32769, 0, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -32769 is out of range.');
            `,
		},
		{
			name: "writing and reading with different byte lengths",
			script: `
				const buf = Buffer.alloc(6);
				buf.writeIntBE(-1, 0, 1);       // 1 byte
				buf.writeIntBE(256, 1, 2);      // 2 bytes
				buf.writeIntBE(97328, 3, 3);    // 3 bytes
				
				assertValueRead(buf.readIntBE(0, 1), -1);
				assertValueRead(buf.toString('hex', 0, 1), "ff");
				assertValueRead(buf.readIntBE(1, 2), 256);
				assertValueRead(buf.toString('hex', 1, 3), "0100");
				assertValueRead(buf.readIntBE(3, 3), 97328);
				assertValueRead(buf.toString('hex', 3, 6), "017c30");
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeIntLE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write out of range offset",
			script: `
				const buf = Buffer.alloc(6);
				assert.throwsNodeErrorWithMessage(() => buf.writeIntLE(127, 6, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 6 is out of range.');
            `,
		},
		{
			name: "byteLength out of range",
			script: `
				const buf = Buffer.alloc(6);
				
				// above the max
				assert.throwsNodeErrorWithMessage(() => buf.writeIntLE(0, 0, 7), RangeError, "ERR_OUT_OF_RANGE", 'The value of "byteLength" 7 is out of range.');
				// below the min
				assert.throwsNodeErrorWithMessage(() => buf.writeIntLE(0, 0, 0), RangeError, "ERR_OUT_OF_RANGE", 'The value of "byteLength" 0 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(6);
				// above the 6-byte max
				assert.throwsNodeErrorWithMessage(() => buf.writeIntLE(140737488355328, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 140737488355328 is out of range.');
				// below the 6-byte min
				assert.throwsNodeErrorWithMessage(() => buf.writeIntLE(-140737488355329, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -140737488355329 is out of range.');
				// above the 2-byte max
				assert.throwsNodeErrorWithMessage(() => buf.writeIntLE(32768, 0, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 32768 is out of range.');
				// below the 2-byte min
				assert.throwsNodeErrorWithMessage(() => buf.writeIntLE(-32769, 0, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -32769 is out of range.');
            `,
		},
		{
			name: "writing and reading with different byte lengths",
			script: `
				const buf = Buffer.alloc(6);
				buf.writeIntLE(-1, 0, 1);       // 1 byte
				buf.writeIntLE(256, 1, 2);      // 2 bytes
				buf.writeIntLE(97328, 3, 3);    // 3 bytes
				
				assertValueRead(buf.readIntLE(0, 1), -1);
				assertValueRead(buf.toString('hex', 0, 1), "ff");
				assertValueRead(buf.readIntLE(1, 2), 256);
				assertValueRead(buf.toString('hex', 1, 3), "0001");
				assertValueRead(buf.readIntLE(3, 3), 97328);
				assertValueRead(buf.toString('hex', 3, 6), "307c01");
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeUInt8(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(2);
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt8(2, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 2 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(2);
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt8(256), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 256 is out of range.');
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt8(256), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 256 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(3);
				buf.writeUInt8(3);    // default offset of zero
				buf.writeUint8(2, 1); // using alias
				buf.writeUInt8(1, 2);
				
				assertValueRead(buf.readUInt8(0), 3);
				assertValueRead(buf.readUInt8(1), 2);
				assertValueRead(buf.readUInt8(2), 1);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeUInt16BE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(2);
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt16BE(2, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 2 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(2);
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt16BE(65536), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 65536 is out of range.');
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt16BE(-1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -1 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(6);
				buf.writeUInt16BE(3);    // default offset of zero
				buf.writeUInt16BE(2, 2); // using method alias
				buf.writeUInt16BE(1, 4);
				
				assertValueRead(buf.readUInt16BE(0), 3);
				assertValueRead(buf.readUInt16BE(2), 2);
				assertValueRead(buf.readUInt16BE(4), 1);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeUInt16LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(2);
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt16LE(2, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 2 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(2);
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt16LE(65536), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 65536 is out of range.');
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt16LE(-1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -1 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(6);
				buf.writeUInt16LE(3);    // default offset of zero
				buf.writeUint16LE(2, 2); // using method alias
				buf.writeUInt16LE(1, 4);
				
				assertValueRead(buf.readUInt16LE(0), 3);
				assertValueRead(buf.readUInt16LE(2), 2);
				assertValueRead(buf.readUInt16LE(4), 1);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeUInt32BE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(4);
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt32BE(2, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 2 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(4);
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt32BE(4294967296), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 4294967296 is out of range.');
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt32BE(-1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -1 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(12);
				buf.writeUInt32BE(3);    // default offset of zero
				buf.writeUint32BE(2, 4); // using method alias
				buf.writeUInt32BE(1, 8);
				
				assertValueRead(buf.readUInt32BE(0), 3);
				assertValueRead(buf.readUInt32BE(4), 2);
				assertValueRead(buf.readUInt32BE(8), 1);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeUInt32LE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write with out of range offset",
			script: `
				const buf = Buffer.alloc(4);
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt32LE(2, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 2 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(4);
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt32LE(4294967296), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 4294967296 is out of range.');
				assert.throwsNodeErrorWithMessage(() => buf.writeUInt32LE(-1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -1 is out of range.');
            `,
		},
		{
			name: "writing and then reading different sections of buffer",
			script: `
				const buf = Buffer.alloc(12);
				buf.writeUInt32LE(3);    // default offset of zero
				buf.writeUInt32LE(2, 4); // using method alias
				buf.writeUInt32LE(1, 8);
				
				assertValueRead(buf.readUInt32LE(0), 3);
				assertValueRead(buf.readUInt32LE(4), 2);
				assertValueRead(buf.readUInt32LE(8), 1);
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeUIntBE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write out of range offset",
			script: `
				const buf = Buffer.alloc(6);
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntBE(127, 6, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 6 is out of range.');
            `,
		},
		{
			name: "byteLength out of range",
			script: `
				const buf = Buffer.alloc(6);
				
				// above the max
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntBE(0, 0, 7), RangeError, "ERR_OUT_OF_RANGE", 'The value of "byteLength" 7 is out of range.');
				// below the min
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntBE(0, 0, 0), RangeError, "ERR_OUT_OF_RANGE", 'The value of "byteLength" 0 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(6);
				// above the 6-byte max
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntBE(281474976710656, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 281474976710656 is out of range.');
				// below the 6-byte min
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntBE(-1, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -1 is out of range.');
				// above the 2-byte max
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntBE(65536, 0, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 65536 is out of range.');
				// below the 2-byte min
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntBE(-1, 0, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -1 is out of range.');
            `,
		},
		{
			name: "writing and reading with different byte lengths",
			script: `
				const buf = Buffer.alloc(6);
				buf.writeUIntBE(1, 0, 1);       // 1 byte
				buf.writeUintBE(256, 1, 2);     // 2 bytes
				buf.writeUIntBE(97328, 3, 3);   // 3 bytes
				
				assertValueRead(buf.readUIntBE(0, 1), 1);
				assertValueRead(buf.toString('hex', 0, 1), "01");
				assertValueRead(buf.readUIntBE(1, 2), 256);
				assertValueRead(buf.toString('hex', 1, 3), "0100");
				assertValueRead(buf.readUIntBE(3, 3), 97328);
				assertValueRead(buf.toString('hex', 3, 6), "017c30");
            `,
		},
	}

	runTestCases(t, tcs)
}

func TestBuffer_writeUIntLE(t *testing.T) {
	tcs := []testCase{
		{
			name: "write out of range offset",
			script: `
				const buf = Buffer.alloc(6);
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntLE(127, 6, 1), RangeError, "ERR_OUT_OF_RANGE", 'The value of "offset" 6 is out of range.');
            `,
		},
		{
			name: "byteLength out of range",
			script: `
				const buf = Buffer.alloc(6);
				
				// above the max
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntLE(0, 0, 7), RangeError, "ERR_OUT_OF_RANGE", 'The value of "byteLength" 7 is out of range.');
				// below the min
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntLE(0, 0, 0), RangeError, "ERR_OUT_OF_RANGE", 'The value of "byteLength" 0 is out of range.');
            `,
		},
		{
			name: "number exceeds ranges",
			script: `
				const buf = Buffer.alloc(6);
				// above the 6-byte max
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntLE(281474976710656, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 281474976710656 is out of range.');
				// below the 6-byte min
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntLE(-1, 0, 6), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -1 is out of range.');
				// above the 2-byte max
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntLE(65536, 0, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" 65536 is out of range.');
				// below the 2-byte min
				assert.throwsNodeErrorWithMessage(() => buf.writeUIntLE(-1, 0, 2), RangeError, "ERR_OUT_OF_RANGE", 'The value of "value" -1 is out of range.');
            `,
		},
		{
			name: "writing and reading with different byte lengths",
			script: `
				const buf = Buffer.alloc(6);
				buf.writeUIntLE(1, 0, 1);       // 1 byte
				buf.writeUintLE(256, 1, 2);     // 2 bytes
				buf.writeUIntLE(97328, 3, 3);   // 3 bytes
				
				assertValueRead(buf.readUIntLE(0, 1), 1);
				assertValueRead(buf.toString('hex', 0, 1), "01");
				assertValueRead(buf.readUIntLE(1, 2), 256);
				assertValueRead(buf.toString('hex', 1, 3), "0001");
				assertValueRead(buf.readUIntLE(3, 3), 97328);
				assertValueRead(buf.toString('hex', 3, 6), "307c01");
            `,
		},
	}

	runTestCases(t, tcs)
}
