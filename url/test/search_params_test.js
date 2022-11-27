url = new URL("http://www.google.com")

if (url.searchParams.toString() != "") {
  throw new Error("Empty search param should return empty string")
}

// Add
url = new URL("http://www.google.com")
url.searchParams.set("first", "first")
if (url.searchParams.toString() != "first=first") {
  throw new Error(`Failed to add query entry. got: ${url.searchParams.toString()}, expected: first=first`)
}

// Append
url = new URL("http://www.google.com?aaa=bbb")
url.searchParams.append("ccc", "ddd")
if (url.searchParams.toString() != "aaa=bbb&ccc=ddd") {
  throw new Error(`Failed to append query entry. got: ${url.searchParams.toString()}, expected: aaa=bbb&ccc=ddd`)
}

// delete
url = new URL("http://www.google.com?aaa=bbb")
url.searchParams.delete("aaa")
if (url.searchParams.toString() != "") {
  throw new Error(`Failed to delete query entry. got: ${url.searchParams.toString()}, expected: ""`)
}

// get
url = new URL("http://www.google.com?aaa=bbb")
if (url.searchParams.get("aaa") != "bbb") {
  throw new Error(`Failed to get query entry. got: ${aaa}, expected: "bbb"`)
}

// getAll
url = new URL("http://www.google.com?aaa=111&aaa=222&aaa=333")
if (url.searchParams.getAll("aaa").toString() != "111,222,333") {
  throw new Error(`Failed to get all query entries. got: ${url.searchParams.getAll("aaa").toString()}, expected: "111","222","333"`)
}

// has
url = new URL("http://www.google.com?aaa=bbb")
if (!url.searchParams.has("ccc") === false) {
  throw new Error(`Expected not to find "bbb" and it did`)
}
if (url.searchParams.has("aaa") !== true) {
  throw new Error(`Expected to find name "aaa" and didn't find it in ${url.search}`)
}

// Iterator
url = new URL("http://www.google.com?aaa=111&bbb=222")
i = 0
for (const [name, value] of url.searchParams) {
  i += 1
  if ((name != "aaa" && value != "111") &&
      (name != "bbb" && value != "222")) {
    throw new Error(`Matched element we didn't expect. didn't expect [${name}, ${value}]`)
  }
}
if (i != 2) {
  throw new Error(`Expected 2 elements in search parans, got ${i}`)
}

// keys
url = new URL("http://www.google.com?aaa=111&bbb=222")
i = 0
for (n of url.searchParams.keys()) {
  i += 1
  if (n != "aaa" && n != "bbb") {
    throw new Error(`Didn't expect key of value ${n}`)
  }
}
if (i != 2) {
  throw new Error(`Expected 2 elements in search params, got ${i}`)
}

// values
url = new URL("http://www.google.com?aaa=111&bbb=222")
i = 0
for (n of url.searchParams.values()) {
  i += 1
  if (n != "111" && n != "222") {
    throw new Error(`Didn't expect value of value ${n}`)
  }
}
if (i != 2) {
  throw new Error(`Expected 2 elements in search params, got ${i}`)
}

// set
url = new URL("http://www.google.com?aaa=111&bbb=222")
url.searchParams.set("aaa", 222)
if (url.search != "?aaa=222&bbb=222") {
  throw new Error(`Failed to set key of aaa. got: ${url.search}, expected: "?aaa=222&bbb=222"`);
}

// sort
url = new URL("http://www.google.com?bbb=111&ccc=222&aaa=333")
url.searchParams.sort()
if (url.search != "?aaa=333&bbb=111&ccc=222") {
  throw new Error(`Failed to set key of aaa. got: ${url.search}, expected: "?aaa=333&bbb=111&ccc=222"`);
}

// Entries
url = new URL("http://www.google.com?aaa=111&bbb=222")
i = 0
for (const [name, value] of url.searchParams.entries()) {
  i += 1
  if ((name != "aaa" && value != "111") &&
      (name != "bbb" && value != "222")) {
    throw new Error(`Matched element we didn't expect. didn't expect [${name}, ${value}]`)
  }
}
if (i != 2) {
  throw new Error(`Expected 2 elements in search parans, got ${i}`)
}
