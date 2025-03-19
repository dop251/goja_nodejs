'use strict';

const assert = {
    _isSameValue(a, b) {
        if (this._isNumber(a)) {
            return this._numberEquals(a, b);
        }
        
        return a === b;
    },
    
    _isNumber(val) {
        return typeof val === "number";
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
    
    _numberEquals(a, b, precision = 1e-6) {
        if (!this._isNumber(b)) {
            return false;
        }
        // Handle NaN vs. NaN
        if (a !== a && b !== b) {
            return true; // Both are NaN
        }
        // If only one is NaN, they're not equal
        if (a !== a || b !== b) {
            return false;
        }
        if (a === b) {
            // Handle +/-0 vs. -/+0
            return a !== 0 || 1 / a === 1 / b;
        }
        // Use relative error for larger numbers, absolute for smaller ones
        if (Math.abs(a) > 1 || Math.abs(b) > 1) {
            return Math.abs((a - b) / Math.max(Math.abs(a), Math.abs(b))) < precision;
        }
        
        // Absolute error for small numbers
        return Math.abs(a - b) < precision;
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