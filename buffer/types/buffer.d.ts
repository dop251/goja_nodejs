/// <reference types="@dop251/types-goja_nodejs-global" />
/// <reference path="buffer.buffer.d.ts" />
declare module 'buffer' {
    export type WithImplicitCoercion<T> =
        | T
        | { valueOf(): T }
        | (T extends string ? { [Symbol.toPrimitive](hint: "string"): T } : never);

    export { Buffer };

    global {
        type BufferEncoding =
        //            | "ascii"
            | "utf8"
            | "utf-8"
//            | "utf16le"
//            | "utf-16le"
//            | "ucs2"
//            | "ucs-2"
            | "base64"
            | "base64url"
//            | "latin1"
//            | "binary"
            | "hex";

        /**
         * Raw data is stored in instances of the Buffer class.
         * A Buffer is similar to an array of integers but corresponds to a raw memory allocation outside the V8 heap.  A Buffer cannot be resized.
         * Valid string encodings: 'ascii'|'utf8'|'utf16le'|'ucs2'(alias of 'utf16le')|'base64'|'base64url'|'binary'(deprecated)|'hex'
         */
        interface BufferConstructor {
            // see buffer.buffer.d.ts for implementation specific to TypeScript 5.7 and later
            // see ts5.6/buffer.buffer.d.ts for implementation specific to TypeScript 5.6 and earlier

            /**
             * Returns `true` if `obj` is a `Buffer`, `false` otherwise.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * Buffer.isBuffer(Buffer.alloc(10)); // true
             * Buffer.isBuffer(Buffer.from('foo')); // true
             * Buffer.isBuffer('a string'); // false
             * Buffer.isBuffer([]); // false
             * Buffer.isBuffer(new Uint8Array(1024)); // false
             * ```
             * @since v0.1.101
             */
            isBuffer(obj: any): obj is Buffer;

            /**
             * Returns `true` if `encoding` is the name of a supported character encoding,
             * or `false` otherwise.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * console.log(Buffer.isEncoding('utf8'));
             * // Prints: true
             *
             * console.log(Buffer.isEncoding('hex'));
             * // Prints: true
             *
             * console.log(Buffer.isEncoding('utf/8'));
             * // Prints: false
             *
             * console.log(Buffer.isEncoding(''));
             * // Prints: false
             * ```
             * @since v0.9.1
             * @param encoding A character encoding name to check.
             */
            isEncoding(encoding: string): encoding is BufferEncoding;

            /**
             * Returns the byte length of a string when encoded using `encoding`.
             * This is not the same as [`String.prototype.length`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/length), which does not account
             * for the encoding that is used to convert the string into bytes.
             *
             * For `'base64'`, `'base64url'`, and `'hex'`, this function assumes valid input.
             * For strings that contain non-base64/hex-encoded data (e.g. whitespace), the
             * return value might be greater than the length of a `Buffer` created from the
             * string.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const str = '\u00bd + \u00bc = \u00be';
             *
             * console.log(`${str}: ${str.length} characters, ` +
             *             `${Buffer.byteLength(str, 'utf8')} bytes`);
             * // Prints: ½ + ¼ = ¾: 9 characters, 12 bytes
             * ```
             *
             * When `string` is a
             * `Buffer`/[`DataView`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/DataView)/[`TypedArray`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/-
             * Reference/Global_Objects/TypedArray)/[`ArrayBuffer`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/ArrayBuffer)/[`SharedArrayBuffer`](https://develop-
             * er.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/SharedArrayBuffer), the byte length as reported by `.byteLength`is returned.
             * @since v0.1.90
             * @param string A value to calculate the length of.
             * @param [encoding='utf8'] If `string` is a string, this is its encoding.
             * @return The number of bytes contained within `string`.
             */
            byteLength(
                string: string | Buffer | ArrayBuffer,
                encoding?: BufferEncoding,
            ): number;

            /**
             * Compares `buf1` to `buf2`, typically for the purpose of sorting arrays of `Buffer` instances. This is equivalent to calling `buf1.compare(buf2)`.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf1 = Buffer.from('1234');
             * const buf2 = Buffer.from('0123');
             * const arr = [buf1, buf2];
             *
             * console.log(arr.sort(Buffer.compare));
             * // Prints: [ <Buffer 30 31 32 33>, <Buffer 31 32 33 34> ]
             * // (This result is equal to: [buf2, buf1].)
             * ```
             * @since v0.11.13
             * @return Either `-1`, `0`, or `1`, depending on the result of the comparison. See `compare` for details.
             */
            compare(buf1: Uint8Array, buf2: Uint8Array): -1 | 0 | 1;

            /**
             * This is the size (in bytes) of pre-allocated internal `Buffer` instances used
             * for pooling. This value may be modified.
             * @since v0.11.3
             */
            poolSize: number;
        }

        interface Buffer {
            /**
             * Writes `string` to `buf` at `offset` according to the character encoding in`encoding`. The `length` parameter is the number of bytes to write. If `buf` did
             * not contain enough space to fit the entire string, only part of `string` will be
             * written. However, partially encoded characters will not be written.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.alloc(256);
             *
             * const len = buf.write('\u00bd + \u00bc = \u00be', 0);
             *
             * console.log(`${len} bytes: ${buf.toString('utf8', 0, len)}`);
             * // Prints: 12 bytes: ½ + ¼ = ¾
             *
             * const buffer = Buffer.alloc(10);
             *
             * const length = buffer.write('abcd', 8);
             *
             * console.log(`${length} bytes: ${buffer.toString('utf8', 8, 10)}`);
             * // Prints: 2 bytes : ab
             * ```
             * @since v0.1.90
             * @param string String to write to `buf`.
             * @param [offset=0] Number of bytes to skip before starting to write `string`.
             * @param [length=buf.length - offset] Maximum number of bytes to write (written bytes will not exceed `buf.length - offset`).
             * @param [encoding='utf8'] The character encoding of `string`.
             * @return Number of bytes written.
             */
            write(string: string, encoding?: BufferEncoding): number;

            write(string: string, offset: number, encoding?: BufferEncoding): number;

            write(string: string, offset: number, length: number, encoding?: BufferEncoding): number;

            /**
             * Decodes `buf` to a string according to the specified character encoding in`encoding`. `start` and `end` may be passed to decode only a subset of `buf`.
             *
             * If `encoding` is `'utf8'` and a byte sequence in the input is not valid UTF-8,
             * then each invalid byte is replaced with the replacement character `U+FFFD`.
             *
             * The maximum length of a string instance (in UTF-16 code units) is available
             * as {@link constants.MAX_STRING_LENGTH}.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf1 = Buffer.allocUnsafe(26);
             *
             * for (let i = 0; i < 26; i++) {
             *   // 97 is the decimal ASCII value for 'a'.
             *   buf1[i] = i + 97;
             * }
             *
             * console.log(buf1.toString('utf8'));
             * // Prints: abcdefghijklmnopqrstuvwxyz
             * console.log(buf1.toString('utf8', 0, 5));
             * // Prints: abcde
             *
             * const buf2 = Buffer.from('tést');
             *
             * console.log(buf2.toString('hex'));
             * // Prints: 74c3a97374
             * console.log(buf2.toString('utf8', 0, 3));
             * // Prints: té
             * console.log(buf2.toString(undefined, 0, 3));
             * // Prints: té
             * ```
             * @since v0.1.90
             * @param [encoding='utf8'] The character encoding to use.
             * @param [start=0] The byte offset to start decoding at.
             * @param [end=buf.length] The byte offset to stop decoding at (not inclusive).
             */
            toString(encoding?: BufferEncoding, start?: number, end?: number): string;

            /**
             * Returns a JSON representation of `buf`. [`JSON.stringify()`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/JSON/stringify) implicitly calls
             * this function when stringifying a `Buffer` instance.
             *
             * `Buffer.from()` accepts objects in the format returned from this method.
             * In particular, `Buffer.from(buf.toJSON())` works like `Buffer.from(buf)`.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0x1, 0x2, 0x3, 0x4, 0x5]);
             * const json = JSON.stringify(buf);
             *
             * console.log(json);
             * // Prints: {"type":"Buffer","data":[1,2,3,4,5]}
             *
             * const copy = JSON.parse(json, (key, value) => {
             *   return value &#x26;&#x26; value.type === 'Buffer' ?
             *     Buffer.from(value) :
             *     value;
             * });
             *
             * console.log(copy);
             * // Prints: <Buffer 01 02 03 04 05>
             * ```
             * @since v0.9.2
             */
            // NOT IMPLEMENTED
            // toJSON(): {
            //     type: "Buffer";
            //     data: number[];
            // };

            /**
             * Returns `true` if both `buf` and `otherBuffer` have exactly the same bytes,`false` otherwise. Equivalent to `buf.compare(otherBuffer) === 0`.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf1 = Buffer.from('ABC');
             * const buf2 = Buffer.from('414243', 'hex');
             * const buf3 = Buffer.from('ABCD');
             *
             * console.log(buf1.equals(buf2));
             * // Prints: true
             * console.log(buf1.equals(buf3));
             * // Prints: false
             * ```
             * @since v0.11.13
             * @param otherBuffer A `Buffer` or {@link Uint8Array} with which to compare `buf`.
             */
            equals(otherBuffer: Uint8Array): boolean;

            /**
             * Writes `value` to `buf` at the specified `offset` as big-endian.
             *
             * `value` is interpreted and written as a two's complement signed integer.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.allocUnsafe(8);
             *
             * buf.writeBigInt64BE(0x0102030405060708n, 0);
             *
             * console.log(buf);
             * // Prints: <Buffer 01 02 03 04 05 06 07 08>
             * ```
             * @since v12.0.0, v10.20.0
             * @param value Number to be written to `buf`.
             * @param [offset=0] Number of bytes to skip before starting to write. Must satisfy: `0 <= offset <= buf.length - 8`.
             * @return `offset` plus the number of bytes written.
             */
            writeBigInt64BE(value: bigint, offset?: number): number;

            /**
             * Writes `value` to `buf` at the specified `offset` as little-endian.
             *
             * `value` is interpreted and written as a two's complement signed integer.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.allocUnsafe(8);
             *
             * buf.writeBigInt64LE(0x0102030405060708n, 0);
             *
             * console.log(buf);
             * // Prints: <Buffer 08 07 06 05 04 03 02 01>
             * ```
             * @since v12.0.0, v10.20.0
             * @param value Number to be written to `buf`.
             * @param [offset=0] Number of bytes to skip before starting to write. Must satisfy: `0 <= offset <= buf.length - 8`.
             * @return `offset` plus the number of bytes written.
             */
            writeBigInt64LE(value: bigint, offset?: number): number;

            /**
             * Writes `value` to `buf` at the specified `offset` as big-endian.
             *
             * This function is also available under the `writeBigUint64BE` alias.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.allocUnsafe(8);
             *
             * buf.writeBigUInt64BE(0xdecafafecacefaden, 0);
             *
             * console.log(buf);
             * // Prints: <Buffer de ca fa fe ca ce fa de>
             * ```
             * @since v12.0.0, v10.20.0
             * @param value Number to be written to `buf`.
             * @param [offset=0] Number of bytes to skip before starting to write. Must satisfy: `0 <= offset <= buf.length - 8`.
             * @return `offset` plus the number of bytes written.
             */
            writeBigUInt64BE(value: bigint, offset?: number): number;

            /**
             * @alias Buffer.writeBigUInt64BE
             * @since v14.10.0, v12.19.0
             */
            writeBigUint64BE(value: bigint, offset?: number): number;

            /**
             * Writes `value` to `buf` at the specified `offset` as little-endian
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.allocUnsafe(8);
             *
             * buf.writeBigUInt64LE(0xdecafafecacefaden, 0);
             *
             * console.log(buf);
             * // Prints: <Buffer de fa ce ca fe fa ca de>
             * ```
             *
             * This function is also available under the `writeBigUint64LE` alias.
             * @since v12.0.0, v10.20.0
             * @param value Number to be written to `buf`.
             * @param [offset=0] Number of bytes to skip before starting to write. Must satisfy: `0 <= offset <= buf.length - 8`.
             * @return `offset` plus the number of bytes written.
             */
            writeBigUInt64LE(value: bigint, offset?: number): number;

            /**
             * @alias Buffer.writeBigUInt64LE
             * @since v14.10.0, v12.19.0
             */
            writeBigUint64LE(value: bigint, offset?: number): number;

            /**
             * Reads an unsigned, big-endian 64-bit integer from `buf` at the specified`offset`.
             *
             * This function is also available under the `readBigUint64BE` alias.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
             *
             * console.log(buf.readBigUInt64BE(0));
             * // Prints: 4294967295n
             * ```
             * @since v12.0.0, v10.20.0
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy: `0 <= offset <= buf.length - 8`.
             */
            readBigUInt64BE(offset?: number): bigint;

            /**
             * @alias Buffer.readBigUInt64BE
             * @since v14.10.0, v12.19.0
             */
            readBigUint64BE(offset?: number): bigint;

            /**
             * Reads an unsigned, little-endian 64-bit integer from `buf` at the specified`offset`.
             *
             * This function is also available under the `readBigUint64LE` alias.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff]);
             *
             * console.log(buf.readBigUInt64LE(0));
             * // Prints: 18446744069414584320n
             * ```
             * @since v12.0.0, v10.20.0
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy: `0 <= offset <= buf.length - 8`.
             */
            readBigUInt64LE(offset?: number): bigint;

            /**
             * @alias Buffer.readBigUInt64LE
             * @since v14.10.0, v12.19.0
             */
            readBigUint64LE(offset?: number): bigint;

            /**
             * Reads a signed, big-endian 64-bit integer from `buf` at the specified `offset`.
             *
             * Integers read from a `Buffer` are interpreted as two's complement signed
             * values.
             * @since v12.0.0, v10.20.0
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy: `0 <= offset <= buf.length - 8`.
             */
            readBigInt64BE(offset?: number): bigint;

            /**
             * Reads a signed, little-endian 64-bit integer from `buf` at the specified`offset`.
             *
             * Integers read from a `Buffer` are interpreted as two's complement signed
             * values.
             * @since v12.0.0, v10.20.0
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy: `0 <= offset <= buf.length - 8`.
             */
            readBigInt64LE(offset?: number): bigint;

            /**
             * Reads `byteLength` number of bytes from `buf` at the specified `offset` and interprets the result as an unsigned, little-endian integer supporting
             * up to 48 bits of accuracy.
             *
             * This function is also available under the `readUintLE` alias.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
             *
             * console.log(buf.readUIntLE(0, 6).toString(16));
             * // Prints: ab9078563412
             * ```
             * @since v0.11.15
             * @param offset Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - byteLength`.
             * @param byteLength Number of bytes to read. Must satisfy `0 < byteLength <= 6`.
             */
            readUIntLE(offset: number, byteLength: number): number;

            /**
             * @alias Buffer.readUIntLE
             * @since v14.9.0, v12.19.0
             */
            readUintLE(offset: number, byteLength: number): number;

            /**
             * Reads `byteLength` number of bytes from `buf` at the specified `offset` and interprets the result as an unsigned big-endian integer supporting
             * up to 48 bits of accuracy.
             *
             * This function is also available under the `readUintBE` alias.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
             *
             * console.log(buf.readUIntBE(0, 6).toString(16));
             * // Prints: 1234567890ab
             * console.log(buf.readUIntBE(1, 6).toString(16));
             * // Throws ERR_OUT_OF_RANGE.
             * ```
             * @since v0.11.15
             * @param offset Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - byteLength`.
             * @param byteLength Number of bytes to read. Must satisfy `0 < byteLength <= 6`.
             */
            readUIntBE(offset: number, byteLength: number): number;

            /**
             * @alias Buffer.readUIntBE
             * @since v14.9.0, v12.19.0
             */
            readUintBE(offset: number, byteLength: number): number;

            /**
             * Reads `byteLength` number of bytes from `buf` at the specified `offset` and interprets the result as a little-endian, two's complement signed value
             * supporting up to 48 bits of accuracy.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
             *
             * console.log(buf.readIntLE(0, 6).toString(16));
             * // Prints: -546f87a9cbee
             * ```
             * @since v0.11.15
             * @param offset Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - byteLength`.
             * @param byteLength Number of bytes to read. Must satisfy `0 < byteLength <= 6`.
             */
            readIntLE(offset: number, byteLength: number): number;

            /**
             * Reads `byteLength` number of bytes from `buf` at the specified `offset` and interprets the result as a big-endian, two's complement signed value
             * supporting up to 48 bits of accuracy.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0x12, 0x34, 0x56, 0x78, 0x90, 0xab]);
             *
             * console.log(buf.readIntBE(0, 6).toString(16));
             * // Prints: 1234567890ab
             * console.log(buf.readIntBE(1, 6).toString(16));
             * // Throws ERR_OUT_OF_RANGE.
             * console.log(buf.readIntBE(1, 0).toString(16));
             * // Throws ERR_OUT_OF_RANGE.
             * ```
             * @since v0.11.15
             * @param offset Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - byteLength`.
             * @param byteLength Number of bytes to read. Must satisfy `0 < byteLength <= 6`.
             */
            readIntBE(offset: number, byteLength: number): number;

            /**
             * Reads an unsigned 8-bit integer from `buf` at the specified `offset`.
             *
             * This function is also available under the `readUint8` alias.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([1, -2]);
             *
             * console.log(buf.readUInt8(0));
             * // Prints: 1
             * console.log(buf.readUInt8(1));
             * // Prints: 254
             * console.log(buf.readUInt8(2));
             * // Throws ERR_OUT_OF_RANGE.
             * ```
             * @since v0.5.0
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 1`.
             */
            readUInt8(offset?: number): number;

            /**
             * @alias Buffer.readUInt8
             * @since v14.9.0, v12.19.0
             */
            readUint8(offset?: number): number;

            /**
             * Reads an unsigned, little-endian 16-bit integer from `buf` at the specified `offset`.
             *
             * This function is also available under the `readUint16LE` alias.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0x12, 0x34, 0x56]);
             *
             * console.log(buf.readUInt16LE(0).toString(16));
             * // Prints: 3412
             * console.log(buf.readUInt16LE(1).toString(16));
             * // Prints: 5634
             * console.log(buf.readUInt16LE(2).toString(16));
             * // Throws ERR_OUT_OF_RANGE.
             * ```
             * @since v0.5.5
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 2`.
             */
            readUInt16LE(offset?: number): number;

            /**
             * @alias Buffer.readUInt16LE
             * @since v14.9.0, v12.19.0
             */
            readUint16LE(offset?: number): number;

            /**
             * Reads an unsigned, big-endian 16-bit integer from `buf` at the specified`offset`.
             *
             * This function is also available under the `readUint16BE` alias.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0x12, 0x34, 0x56]);
             *
             * console.log(buf.readUInt16BE(0).toString(16));
             * // Prints: 1234
             * console.log(buf.readUInt16BE(1).toString(16));
             * // Prints: 3456
             * ```
             * @since v0.5.5
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 2`.
             */
            readUInt16BE(offset?: number): number;

            /**
             * @alias Buffer.readUInt16BE
             * @since v14.9.0, v12.19.0
             */
            readUint16BE(offset?: number): number;

            /**
             * Reads an unsigned, little-endian 32-bit integer from `buf` at the specified`offset`.
             *
             * This function is also available under the `readUint32LE` alias.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0x12, 0x34, 0x56, 0x78]);
             *
             * console.log(buf.readUInt32LE(0).toString(16));
             * // Prints: 78563412
             * console.log(buf.readUInt32LE(1).toString(16));
             * // Throws ERR_OUT_OF_RANGE.
             * ```
             * @since v0.5.5
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 4`.
             */
            readUInt32LE(offset?: number): number;

            /**
             * @alias Buffer.readUInt32LE
             * @since v14.9.0, v12.19.0
             */
            readUint32LE(offset?: number): number;

            /**
             * Reads an unsigned, big-endian 32-bit integer from `buf` at the specified`offset`.
             *
             * This function is also available under the `readUint32BE` alias.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0x12, 0x34, 0x56, 0x78]);
             *
             * console.log(buf.readUInt32BE(0).toString(16));
             * // Prints: 12345678
             * ```
             * @since v0.5.5
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 4`.
             */
            readUInt32BE(offset?: number): number;

            /**
             * @alias Buffer.readUInt32BE
             * @since v14.9.0, v12.19.0
             */
            readUint32BE(offset?: number): number;

            /**
             * Reads a signed 8-bit integer from `buf` at the specified `offset`.
             *
             * Integers read from a `Buffer` are interpreted as two's complement signed values.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([-1, 5]);
             *
             * console.log(buf.readInt8(0));
             * // Prints: -1
             * console.log(buf.readInt8(1));
             * // Prints: 5
             * console.log(buf.readInt8(2));
             * // Throws ERR_OUT_OF_RANGE.
             * ```
             * @since v0.5.0
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 1`.
             */
            readInt8(offset?: number): number;

            /**
             * Reads a signed, little-endian 16-bit integer from `buf` at the specified`offset`.
             *
             * Integers read from a `Buffer` are interpreted as two's complement signed values.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0, 5]);
             *
             * console.log(buf.readInt16LE(0));
             * // Prints: 1280
             * console.log(buf.readInt16LE(1));
             * // Throws ERR_OUT_OF_RANGE.
             * ```
             * @since v0.5.5
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 2`.
             */
            readInt16LE(offset?: number): number;

            /**
             * Reads a signed, big-endian 16-bit integer from `buf` at the specified `offset`.
             *
             * Integers read from a `Buffer` are interpreted as two's complement signed values.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0, 5]);
             *
             * console.log(buf.readInt16BE(0));
             * // Prints: 5
             * ```
             * @since v0.5.5
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 2`.
             */
            readInt16BE(offset?: number): number;

            /**
             * Reads a signed, little-endian 32-bit integer from `buf` at the specified`offset`.
             *
             * Integers read from a `Buffer` are interpreted as two's complement signed values.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0, 0, 0, 5]);
             *
             * console.log(buf.readInt32LE(0));
             * // Prints: 83886080
             * console.log(buf.readInt32LE(1));
             * // Throws ERR_OUT_OF_RANGE.
             * ```
             * @since v0.5.5
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 4`.
             */
            readInt32LE(offset?: number): number;

            /**
             * Reads a signed, big-endian 32-bit integer from `buf` at the specified `offset`.
             *
             * Integers read from a `Buffer` are interpreted as two's complement signed values.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([0, 0, 0, 5]);
             *
             * console.log(buf.readInt32BE(0));
             * // Prints: 5
             * ```
             * @since v0.5.5
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 4`.
             */
            readInt32BE(offset?: number): number;

            /**
             * Reads a 32-bit, little-endian float from `buf` at the specified `offset`.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([1, 2, 3, 4]);
             *
             * console.log(buf.readFloatLE(0));
             * // Prints: 1.539989614439558e-36
             * console.log(buf.readFloatLE(1));
             * // Throws ERR_OUT_OF_RANGE.
             * ```
             * @since v0.11.15
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 4`.
             */
            readFloatLE(offset?: number): number;

            /**
             * Reads a 32-bit, big-endian float from `buf` at the specified `offset`.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([1, 2, 3, 4]);
             *
             * console.log(buf.readFloatBE(0));
             * // Prints: 2.387939260590663e-38
             * ```
             * @since v0.11.15
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 4`.
             */
            readFloatBE(offset?: number): number;

            /**
             * Reads a 64-bit, little-endian double from `buf` at the specified `offset`.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([1, 2, 3, 4, 5, 6, 7, 8]);
             *
             * console.log(buf.readDoubleLE(0));
             * // Prints: 5.447603722011605e-270
             * console.log(buf.readDoubleLE(1));
             * // Throws ERR_OUT_OF_RANGE.
             * ```
             * @since v0.11.15
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 8`.
             */
            readDoubleLE(offset?: number): number;

            /**
             * Reads a 64-bit, big-endian double from `buf` at the specified `offset`.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.from([1, 2, 3, 4, 5, 6, 7, 8]);
             *
             * console.log(buf.readDoubleBE(0));
             * // Prints: 8.20788039913184e-304
             * ```
             * @since v0.11.15
             * @param [offset=0] Number of bytes to skip before starting to read. Must satisfy `0 <= offset <= buf.length - 8`.
             */
            readDoubleBE(offset?: number): number;
        }

        var Buffer: BufferConstructor;
    }
}

declare module "node:buffer" {
    export * from "buffer";
}
