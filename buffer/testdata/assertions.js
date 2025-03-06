/**
 * Assertion helper functions for Buffer tests
 */
"use strict";

const assert = require("../../assert.js");

function assertValueRead(actual, expected) {
    assert.sameValue(actual, expected, "value read does not match; ")
}

function assertBytesWritten(actual, expected) {
    assert.sameValue(actual, expected, "bytesWritten does not match; ")
}

function assertBufferWriteRead(buffer, writeMethod, readMethod, value, offset = 0) {
    const bytesWritten = buffer[writeMethod](value, offset);
    const bytesPerElement = getBufferElementSize(writeMethod);
    assertBytesWritten(bytesWritten, offset + bytesPerElement);

    const readValue = buffer[readMethod](offset);
    assertValueRead(readValue, value);
}

// getBufferElementSize determines the number of bytes per type based on method name
function getBufferElementSize(methodName) {
    if (methodName.includes('64')) return 8;
    if (methodName.includes('32')) return 4;
    if (methodName.includes('16')) return 2;
    if (methodName.includes('8')) return 1;
    return 1;
}
