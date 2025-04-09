Nodejs compatibility library for Goja
====

This is a collection of [Goja](https://github.com/dop251/goja) modules that provide nodejs compatibility.

Example:

```go
package main

import (
    "github.com/dop251/goja"
    "github.com/dop251/goja_nodejs/require"
)

func main() {
    registry := new(require.Registry) // this can be shared by multiple runtimes

    runtime := goja.New()
    req := registry.Enable(runtime)

    runtime.RunString(`
    var m = require("./m.js");
    m.test();
    `)

    m, err := req.Require("./m.js")
    _, _ = m, err
}
```

Type Definitions
---

Type definitions are published to https://npmjs.com as @dop251/types-goja_nodejs-MODULE.
They only include what's been implemented so far.

To make use of them you need to install the appropriate modules and add `node_modules/@dop251` to `typeRoots` in `tsconfig.json`.

I didn't want to add those to DefinitelyTyped partly because I don't think they really belong there,
and partly because I'd like to fully control the release cycle, i.e. publish the modules by an automated CI job and
exactly at the same time as the Go code is released.

And the reason for splitting them into different packages is that the modules can be enabled or disabled individually, unlike in nodejs.

More modules will be added. Contributions welcome too.
