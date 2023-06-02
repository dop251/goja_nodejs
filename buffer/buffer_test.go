package buffer

import (
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
