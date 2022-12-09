let myURL = new URL('https://example.org/');
if (myURL.toString() != "https://example.org/") {
  throw new Error(`Failed comparison during creation. got: ${myURL.toString()}, expected: https://example.org/foo`)
}

myURL = new URL('/foo', 'https://example.org/');
if (myURL.toString() != "https://example.org/foo") {
  throw new Error(`Failed comparison during creation. got: ${myURL.toString()}, expected: https://example.org/foo`)
}

myURL = new URL('http://Example.com/', 'https://example.org/');
if (myURL.toString() != "http://Example.com/") {
  throw new Error(`Failed comparison during creation. got: ${myURL.toString()}, expected: http://Example.com/`)
}

myURL = new URL('https://Example.com/', 'https://example.org/');
if (myURL.toString() != "https://Example.com/") {
  throw new Error(`Failed comparison during creation. got: ${myURL.toString()}, expected: https://Example.com/`)
}

myURL = new URL('foo://Example.com/', 'https://example.org/');
if (myURL.toString() != "foo://Example.com/") {
  throw new Error(`Failed comparison during creation. got: ${myURL.toString()}, expected: foo://Example.com/`)
}

myURL = new URL('foo:Example.com/', 'https://example.org/');
if (myURL.toString() != "foo:Example.com/") {
  throw new Error(`Failed comparison during creation. got: ${myURL.toString()}, expected: foo:Example.com//`)
}

// Hash
myURL = new URL('https://example.org/foo#bar');
myURL.hash = 'baz';
if (myURL.href != "https://example.org/foo#baz") {
  throw new Error(`Failed setting hash. got: ${myURL.href}, expected: https://example.org/foo#baz`)
}

// Host
myURL = new URL('https://example.org:81/foo');
myURL.host = 'example.com:82';
if (myURL.href != "https://example.com:82/foo") {
  throw new Error(`Failed setting host. got: ${myURL.href}, expected: https://example.com:82/foo`)
}

// Hostname
myURL = new URL('https://example.org:81/foo');

myURL.hostname = 'example.com:82';
if (myURL.href != "https://example.com:81/foo") {
  throw new Error(`Failed setting hostname. got: ${myURL.href}, expected: https://example.com:81/foo`)
}

// href
myURL = new URL('https://example.org/foo');
myURL.href = 'https://example.com/bar';
if (myURL.href != "https://example.com/bar") {
  throw new Error(`Failed setting href. got: ${myURL.href}, expected: https://example.com/bar`)
}

// Password
myURL = new URL('https://abc:xyz@example.com');
myURL.password = '123';
if (myURL.href != "https://abc:123@example.com") {
  throw new Error(`Failed setting password. got: ${myURL.href}, expected: https://abc:123@example.com`)
}

// pathname
myURL = new URL('https://example.org/abc/xyz?123');
myURL.pathname = '/abcdef';
if (myURL.href != "https://example.org/abcdef?123") {
  throw new Error(`Failed setting pathname. got: ${myURL.href}, expected: https://example.org/abcdef?123`)
}

// port
myURL = new URL('https://example.org:8888');
if (myURL.port != "8888") {
  throw new Error(`Failed setting port. got: ${myURL.port}, expected: 8888`)
}
// Default ports are automatically transformed to the empty string
// (HTTPS protocol's default port is 443)
myURL.port = "443";
if (myURL.port != "") {
  throw new Error(`Failed setting port. got: ${myURL.port}, expected: ""`)
}
myURL.port = 1234;
if (myURL.port != "1234") {
  throw new Error(`Failed setting port. got: ${myURL.port}, expected: "1234"`)
}
// Completely invalid port strings are ignored
myURL.port = 'abcd';
if (myURL.port != "1234") {
  throw new Error(`Failed setting port. got: ${myURL.port}, expected: "1234"`)
}
// Leading numbers are treated as a port number
myURL.port = '5678abcd';
if (myURL.port != "5678") {
  throw new Error(`Failed setting port. got: ${myURL.port}, expected: "5678"`)
}
// Non-integers are truncated
myURL.port = 1234.5678;
if (myURL.port != "1234") {
  throw new Error(`Failed setting port. got: ${myURL.port}, expected: "1234"`)
}
// Out-of-range numbers which are not represented in scientific notation
// will be ignored.
myURL.port = 1e10;
if (myURL.port != "1234") {
  throw new Error(`Failed setting port. got: ${myURL.port}, expected: "1234"`)
}
myURL.port = 4.567e21;
if (myURL.port != "4") {
  throw new Error(`Failed setting port. got: ${myURL.port}, expected: "4"`)
}

// Protocol
myURL = new URL('https://example.org');
myURL.protocol = 'ftp';
if (myURL.href != "ftp://example.org") {
  throw new Error(`Failed setting protocol. got: ${myURL.href}, expected: ftp://example.org`)
}

// Search
myURL = new URL('https://example.org/abc?123');
myURL.search = 'abc=xyz';
if (myURL.href != "https://example.org/abc?abc=xyz") {
  throw new Error(`Failed setting search. got: ${myURL.href}, expected: https://example.org/abc?abc=xyz`)
}

// Username
myURL = new URL('https://abc:xyz@example.com/');
myURL.username = '123';
if (myURL.href != "https://123:xyz@example.com/") {
  throw new Error(`Failed setting username. got: ${myURL.href}, expected: https://123:xyz@example.com/`)
}
