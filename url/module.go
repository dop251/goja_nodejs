package url

import (
	"github.com/dop251/goja"
)

func Enable(r *goja.Runtime) {
	r.Set("URL", createURL(r))
}
