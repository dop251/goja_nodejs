package buffer

import (
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

type testCase struct {
	name        string
	script      string
	expectedErr string
}

func runTestCases(t *testing.T, tcs []testCase) {
	vm := goja.New()
	new(require.Registry).Enable(vm)
	_, err := vm.RunString(`const Buffer = require("node:buffer").Buffer;`)
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
	}

	runTestCases(t, tcs)
}
