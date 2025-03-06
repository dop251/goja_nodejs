'use strict';

const assert = {
    _isSameValue(a, b) {
        if (a === b) {
            // Handle +/-0 vs. -/+0
            return a !== 0 || 1 / a === 1 / b;
        }

        // Handle NaN vs. NaN
        return a !== a && b !== b;
    },

    _toString(value) {
        try {
            if (value === 0 && 1 / value === -Infinity) {
                return '-0';
            }

            return String(value);
        } catch (err) {
            if (err.name === 'TypeError') {
                return Object.prototype.toString.call(value);
            }

            throw err;
        }
    },

    sameValue(actual, expected, message) {
        if (assert._isSameValue(actual, expected)) {
            return;
        }
        if (message === undefined) {
            message = '';
        } else {
            message += ' ';
        }

        message += 'Expected SameValue(«' + assert._toString(actual) + '», «' + assert._toString(expected) + '») to be true';

        throw new Error(message);
    },

    _throws(f, checks, message) {
        if (message === undefined) {
            message = '';
        } else {
            message += ' ';
        }
        try {
            f();
        } catch (e) {
            for (const check of checks) {
                check(e, message);
            }
            return;
        }
        throw new Error(message + "No exception was thrown");
    },

    _sameErrorType(expected){
        return function(e, message) {
            assert.sameValue(e.constructor, expected, `${message}Wrong exception type was thrown:`);
        }
    },

    _sameErrorCode(expected){
        return function(e, message) {
            assert.sameValue(e.code, expected, `${message}Wrong exception code was thrown:`);
        }
    },

    _sameErrorMessage(expected){
        return function(e, message) {
            assert.sameValue(e.message, expected, `${message}Wrong exception message was thrown:`);
        }
    },

    throws(f, ctor, message) {
        return this._throws(f, [
            this._sameErrorType(ctor)
        ], message);
    },

    throwsNodeError(f, ctor, code, message) {
        return this._throws(f, [
            this._sameErrorType(ctor),
            this._sameErrorCode(code)
        ], message);
    },

    throwsNodeErrorWithMessage(f, ctor, code, errorMessage, message) {
        return this._throws(f, [
            this._sameErrorType(ctor),
            this._sameErrorCode(code),
            this._sameErrorMessage(errorMessage)
        ], message);
    }
}

module.exports = assert;