export {};

declare global {
    namespace GojaNodeJS {
        interface Iterator<T, TReturn = any, TNext = any> extends IteratorObject<T, TReturn, TNext> {
            [Symbol.iterator](): GojaNodeJS.Iterator<T, TReturn, TNext>;
        }

        // Polyfill for TS 5.6's instrinsic BuiltinIteratorReturn type, required for DOM-compatible iterators
        type BuiltinIteratorReturn = ReturnType<any[][typeof Symbol.iterator]> extends
            globalThis.Iterator<any, infer TReturn> ? TReturn
            : any;

    }
}

