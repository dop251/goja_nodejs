package url

import (
	"net/url"

	"github.com/dop251/goja"
)

func createURL(r *goja.Runtime) func(goja.ConstructorCall) *goja.Object {
	return func(call goja.ConstructorCall) *goja.Object {
		if len(call.Arguments) == 0 {
			panic(r.NewTypeError("Failed to construct 'URL': 1 argument required, but only 0 present."))
		}

		var u *url.URL
		if len(call.Arguments) == 1 {
			if url, err := url.ParseRequestURI(call.Arguments[0].String()); err != nil {
				panic(r.NewTypeError("Failed to construct 'URL': Invalid URL"))
			} else {
				u, _ = url.Parse(call.Arguments[0].String())
			}
		} else {
			if _, err := url.ParseRequestURI(call.Arguments[1].String()); err != nil {
				panic(r.NewTypeError("Failed to construct 'URL': Invalid base URL"))
			} else if input, err := url.Parse(call.Arguments[0].String()); err != nil {
				panic(r.NewTypeError("Failed to construct 'URL': Invalid URL"))
			} else {
				base, _ := url.Parse(call.Arguments[1].String())
				u = base.ResolveReference(input)
			}
		}

		w := newWrapper(u)
		return polyfill(r, w)
	}
}

func polyfill(r *goja.Runtime, u *urlWrapper) *goja.Object {
	o := r.NewObject()
	p := r.NewObject()

	noop := func(v string) {}

	// Properties
	addStringProperty(r, p, "hash", u.getHash, u.setHash)
	addStringProperty(r, p, "host", u.getHost, u.setHost)
	addStringProperty(r, p, "hostname", u.getHostname, u.setHostname)
	addStringProperty(r, p, "href", u.getHref,
		func(v string) {
			if err := u.setHref(v); err != nil {
				panic(r.NewTypeError("Failed to set href. Invalid URL string specifed: " + v))
			}
		})

	addStringProperty(r, p, "pathname", u.getPathname, u.setPathname)
	addStringProperty(r, p, "origin", u.getOrigin, noop)
	addStringProperty(r, p, "password", u.getPassword, u.setPassword)
	addStringProperty(r, p, "username", u.getUsername, u.setUsername)
	addStringProperty(r, p, "port", u.getPort, u.setPort)
	addStringProperty(r, p, "protocol", u.getProtocol, u.setProtocol)
	addStringProperty(r, p, "search", u.getSearch, u.setSearch)

	p.Set("searchParams", createSearchParams(r, u.url))

	// Functions
	p.Set("toString", u.toString)
	p.Set("toJSON", func() string { return u.toJSON() })

	o.SetPrototype(p)

	return o
}

func addStringProperty(r *goja.Runtime, o *goja.Object, pn string, getter func() string, setter func(string)) {
	o.DefineAccessorProperty(pn,
		r.ToValue(func() string {
			return getter()
		}),
		r.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) == 0 {
				panic(r.NewTypeError("Failed to set " + pn + " on 'URL': 1 argument required, but only 0 present"))
			}
			setter(call.Arguments[0].String())
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE,
	)
}
