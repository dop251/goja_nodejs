declare module "buffer" {
    type ImplicitArrayBuffer<T extends WithImplicitCoercion<ArrayBufferLike>> = T extends
        { valueOf(): infer V extends ArrayBufferLike } ? V : T;
    global {
        interface BufferConstructor {
            // see buffer.d.ts for implementation shared with all TypeScript versions

            /**
             * Allocates a new buffer containing the given {str}.
             *
             * @param str String to store in buffer.
             * @param encoding encoding to use, optional.  Default is 'utf8'
             * @deprecated since v10.0.0 - Use `Buffer.from(string[, encoding])` instead.
             */
            new(str: string, encoding?: BufferEncoding): Buffer<ArrayBuffer>;
            /**
             * Allocates a new buffer containing the given {array} of octets.
             *
             * @param array The octets to store.
             * @deprecated since v10.0.0 - Use `Buffer.from(array)` instead.
             */
            new(array: ArrayLike<number>): Buffer<ArrayBuffer>;
            /**
             * Produces a Buffer backed by the same allocated memory as
             * the given {ArrayBuffer}/{SharedArrayBuffer}.
             *
             * @param arrayBuffer The ArrayBuffer with which to share memory.
             * @deprecated since v10.0.0 - Use `Buffer.from(arrayBuffer[, byteOffset[, length]])` instead.
             */
            new<TArrayBuffer extends ArrayBufferLike = ArrayBuffer>(arrayBuffer: TArrayBuffer): Buffer<TArrayBuffer>;
            /**
             * Allocates a new `Buffer` using an `array` of bytes in the range `0` – `255`.
             * Array entries outside that range will be truncated to fit into it.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * // Creates a new Buffer containing the UTF-8 bytes of the string 'buffer'.
             * const buf = Buffer.from([0x62, 0x75, 0x66, 0x66, 0x65, 0x72]);
             * ```
             *
             * If `array` is an `Array`-like object (that is, one with a `length` property of
             * type `number`), it is treated as if it is an array, unless it is a `Buffer` or
             * a `Uint8Array`. This means all other `TypedArray` variants get treated as an
             * `Array`. To create a `Buffer` from the bytes backing a `TypedArray`, use
             * `Buffer.copyBytesFrom()`.
             *
             * A `TypeError` will be thrown if `array` is not an `Array` or another type
             * appropriate for `Buffer.from()` variants.
             *
             * `Buffer.from(array)` and `Buffer.from(string)` may also use the internal
             * `Buffer` pool like `Buffer.allocUnsafe()` does.
             * @since v5.10.0
             */
            from(array: WithImplicitCoercion<ArrayLike<number>>): Buffer<ArrayBuffer>;
            /**
             * This creates a view of the `ArrayBuffer` without copying the underlying
             * memory. For example, when passed a reference to the `.buffer` property of a
             * `TypedArray` instance, the newly created `Buffer` will share the same
             * allocated memory as the `TypedArray`'s underlying `ArrayBuffer`.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const arr = new Uint16Array(2);
             *
             * arr[0] = 5000;
             * arr[1] = 4000;
             *
             * // Shares memory with `arr`.
             * const buf = Buffer.from(arr.buffer);
             *
             * console.log(buf);
             * // Prints: <Buffer 88 13 a0 0f>
             *
             * // Changing the original Uint16Array changes the Buffer also.
             * arr[1] = 6000;
             *
             * console.log(buf);
             * // Prints: <Buffer 88 13 70 17>
             * ```
             *
             * The optional `byteOffset` and `length` arguments specify a memory range within
             * the `arrayBuffer` that will be shared by the `Buffer`.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const ab = new ArrayBuffer(10);
             * const buf = Buffer.from(ab, 0, 2);
             *
             * console.log(buf.length);
             * // Prints: 2
             * ```
             *
             * A `TypeError` will be thrown if `arrayBuffer` is not an `ArrayBuffer` or a
             * `SharedArrayBuffer` or another type appropriate for `Buffer.from()`
             * variants.
             *
             * It is important to remember that a backing `ArrayBuffer` can cover a range
             * of memory that extends beyond the bounds of a `TypedArray` view. A new
             * `Buffer` created using the `buffer` property of a `TypedArray` may extend
             * beyond the range of the `TypedArray`:
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const arrA = Uint8Array.from([0x63, 0x64, 0x65, 0x66]); // 4 elements
             * const arrB = new Uint8Array(arrA.buffer, 1, 2); // 2 elements
             * console.log(arrA.buffer === arrB.buffer); // true
             *
             * const buf = Buffer.from(arrB.buffer);
             * console.log(buf);
             * // Prints: <Buffer 63 64 65 66>
             * ```
             * @since v5.10.0
             * @param arrayBuffer An `ArrayBuffer`, `SharedArrayBuffer`, for example the
             * `.buffer` property of a `TypedArray`.
             * @param byteOffset Index of first byte to expose. **Default:** `0`.
             * @param length Number of bytes to expose. **Default:**
             * `arrayBuffer.byteLength - byteOffset`.
             */
            from<TArrayBuffer extends WithImplicitCoercion<ArrayBufferLike>>(
                arrayBuffer: TArrayBuffer,
                byteOffset?: number,
                length?: number,
            ): Buffer<ImplicitArrayBuffer<TArrayBuffer>>;
            /**
             * Creates a new `Buffer` containing `string`. The `encoding` parameter identifies
             * the character encoding to be used when converting `string` into bytes.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf1 = Buffer.from('this is a tést');
             * const buf2 = Buffer.from('7468697320697320612074c3a97374', 'hex');
             *
             * console.log(buf1.toString());
             * // Prints: this is a tést
             * console.log(buf2.toString());
             * // Prints: this is a tést
             * console.log(buf1.toString('latin1'));
             * // Prints: this is a tÃ©st
             * ```
             *
             * A `TypeError` will be thrown if `string` is not a string or another type
             * appropriate for `Buffer.from()` variants.
             *
             * `Buffer.from(string)` may also use the internal `Buffer` pool like
             * `Buffer.allocUnsafe()` does.
             * @since v5.10.0
             * @param string A string to encode.
             * @param encoding The encoding of `string`. **Default:** `'utf8'`.
             */
            from(string: WithImplicitCoercion<string>, encoding?: BufferEncoding): Buffer<ArrayBuffer>;
            /**
             * Returns a new `Buffer` which is the result of concatenating all the `Buffer` instances in the `list` together.
             *
             * If the list has no items, or if the `totalLength` is 0, then a new zero-length `Buffer` is returned.
             *
             * If `totalLength` is not provided, it is calculated from the `Buffer` instances
             * in `list` by adding their lengths.
             *
             * If `totalLength` is provided, it is coerced to an unsigned integer. If the
             * combined length of the `Buffer`s in `list` exceeds `totalLength`, the result is
             * truncated to `totalLength`. If the combined length of the `Buffer`s in `list` is
             * less than `totalLength`, the remaining space is filled with zeros.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * // Create a single `Buffer` from a list of three `Buffer` instances.
             *
             * const buf1 = Buffer.alloc(10);
             * const buf2 = Buffer.alloc(14);
             * const buf3 = Buffer.alloc(18);
             * const totalLength = buf1.length + buf2.length + buf3.length;
             *
             * console.log(totalLength);
             * // Prints: 42
             *
             * const bufA = Buffer.concat([buf1, buf2, buf3], totalLength);
             *
             * console.log(bufA);
             * // Prints: <Buffer 00 00 00 00 ...>
             * console.log(bufA.length);
             * // Prints: 42
             * ```
             *
             * `Buffer.concat()` may also use the internal `Buffer` pool like `Buffer.allocUnsafe()` does.
             * @since v0.7.11
             * @param list List of `Buffer` or {@link Uint8Array} instances to concatenate.
             * @param totalLength Total length of the `Buffer` instances in `list` when concatenated.
             */
            concat(list: readonly Uint8Array[], totalLength?: number): Buffer<ArrayBuffer>;            /**
             * Allocates a new `Buffer` of `size` bytes. If `fill` is `undefined`, the`Buffer` will be zero-filled.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.alloc(5);
             *
             * console.log(buf);
             * // Prints: <Buffer 00 00 00 00 00>
             * ```
             *
             * If `size` is larger than {@link constants.MAX_LENGTH} or smaller than 0, `ERR_OUT_OF_RANGE` is thrown.
             *
             * If `fill` is specified, the allocated `Buffer` will be initialized by calling `buf.fill(fill)`.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.alloc(5, 'a');
             *
             * console.log(buf);
             * // Prints: <Buffer 61 61 61 61 61>
             * ```
             *
             * If both `fill` and `encoding` are specified, the allocated `Buffer` will be
             * initialized by calling `buf.fill(fill, encoding)`.
             *
             * ```js
             * import { Buffer } from 'node:buffer';
             *
             * const buf = Buffer.alloc(11, 'aGVsbG8gd29ybGQ=', 'base64');
             *
             * console.log(buf);
             * // Prints: <Buffer 68 65 6c 6c 6f 20 77 6f 72 6c 64>
             * ```
             *
             * Calling `Buffer.alloc()` can be measurably slower than the alternative `Buffer.allocUnsafe()` but ensures that the newly created `Buffer` instance
             * contents will never contain sensitive data from previous allocations, including
             * data that might not have been allocated for `Buffer`s.
             *
             * A `TypeError` will be thrown if `size` is not a number.
             * @since v5.10.0
             * @param size The desired length of the new `Buffer`.
             * @param [fill=0] A value to pre-fill the new `Buffer` with.
             * @param [encoding='utf8'] If `fill` is a string, this is its encoding.
             */
            alloc(size: number, fill?: string | Uint8Array | number, encoding?: BufferEncoding): Buffer<ArrayBuffer>;
        }
        interface Buffer<TArrayBuffer extends ArrayBufferLike = ArrayBufferLike> extends Uint8Array<TArrayBuffer> {
            // see buffer.d.ts for implementation shared with all TypeScript versions

        }
    }
}
