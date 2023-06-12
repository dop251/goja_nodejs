"use strict";

const assert = require("../../assert.js");

function testCtor(value, expected) {
  assert.sameValue(new URLSearchParams(value).toString(), expected);
}

testCtor("user=abc&query=xyz", "user=abc&query=xyz");
testCtor("?user=abc&query=xyz", "user=abc&query=xyz");
testCtor(
  {
    user: "abc",
    query: ["first", "second"],
  },
  "user=abc&query=first,second"
);

const map = new Map();
map.set("user", "abc");
map.set("query", "xyz");
testCtor(map, "user=abc&query=xyz");

testCtor(
  [
    ["user", "abc"],
    ["query", "first"],
    ["query", "second"],
  ],
  "user=abc&query=first&query=second"
);

// Each key-value pair must have exactly two elements
assert.throws(() => new URLSearchParams([["single_value"]]), TypeError);
assert.throws(() => new URLSearchParams([["too", "many", "values"]]), TypeError);

let params;

params = new URLSearchParams("https://example.org/?a=b&c=d");
params.forEach((value, name, searchParams) => {
  if (name === "a") {
    assert.sameValue(value, "b");
  }
  if (name === "c") {
    assert.sameValue(value, "d");
  }
  assert.sameValue(searchParams, "a=b&c=d");
});

params = new URLSearchParams("?user=abc");
assert.throws(() => params.append(), TypeError);
assert.throws(() => params.append(), TypeError);
params.append("query", "first");
assert.sameValue(params.toString(), "user=abc&query=first");

params = new URLSearchParams("first=one&second=two&third=three");
assert.throws(() => params.delete(), TypeError);
params.delete("second", "fake-value");
assert.sameValue(params.toString(), "first=one&second=two&third=three");
params.delete("third", "three");
assert.sameValue(params.toString(), "first=one&second=two");
params.delete("second");
assert.sameValue(params.toString(), "first=one");

params = new URLSearchParams("user=abc&query=xyz");
assert.throws(() => params.get(), TypeError);
assert.sameValue(params.get("user"), "abc");
assert.sameValue(params.get("non-existant"), null);

params = new URLSearchParams("query=first&query=second");
assert.throws(() => params.getAll(), TypeError);
const all = params.getAll("query");
assert.sameValue(all.includes("first"), true);
assert.sameValue(all.includes("second"), true);
assert.sameValue(all.length, 2);

params = new URLSearchParams("user=abc&query=xyz");
assert.throws(() => params.has(), TypeError);
assert.sameValue(params.has("user"), true);
assert.sameValue(params.has("user", "abc"), true);
assert.sameValue(params.has("user", "abc", "extra-param"), true);
assert.sameValue(params.has("user", "efg"), false);

params = new URLSearchParams();
params.append("foo", "bar");
params.append("foo", "baz");
params.append("abc", "def");
assert.sameValue(params.toString(), "foo=bar&foo=baz&abc=def");
params.set("foo", "def");
params.set("xyz", "opq");
assert.sameValue(params.toString(), "foo=def&abc=def&xyz=opq");

params = new URLSearchParams("query=first&query=second&user=abc&double=first,second");
const entries = params.entries();
assert.sameValue(entries.length, 4);
assert.sameValue(entries[0].toString(), ["query", "first"].toString());
assert.sameValue(entries[1].toString(), ["query", "second"].toString());
assert.sameValue(entries[2].toString(), ["user", "abc"].toString());
assert.sameValue(entries[3].toString(), ["double", "first,second"].toString());

params = new URLSearchParams("query=first&query=second&user=abc");
const keys = params.keys();
assert.sameValue(keys.length, 3);
assert.sameValue(keys[0], "query");
assert.sameValue(keys[1], "query");
assert.sameValue(keys[2], "user");

params = new URLSearchParams("query=first&query=second&user=abc");
const values = params.values();
assert.sameValue(values.length, 3);
assert.sameValue(values[0], "first");
assert.sameValue(values[1], "second");
assert.sameValue(values[2], "abc");

params = new URLSearchParams("query[]=abc&type=search&query[]=123");
params.sort();
assert.sameValue(params.toString(), "query%5B%5D=abc&query%5B%5D=123&type=search");

params = new URLSearchParams("query=first&query=second&user=abc");
assert.sameValue(params.size, 3);
