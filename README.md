Nodejs compatibility library for Goja
====

This is a collection of modules Goja modules that provide nodejs compatibility.

Example:

```go
package main

import (
    "github.com/dop251/goja"
    "github.com/dop251/goja_nodejs/require"
)

func main() {
    require := new(require.Require) // this can be shared by multiple runtimes

    runtime := goja.New()
    req := require.Enable(runtime)

    runtime.RunString(`
    var m = require("m.js");
    m.test();
    `)

    m, err := req.Require("m.js")
    _, _ = m, err
}
```

More modules will be added. Contributions welcome too.
