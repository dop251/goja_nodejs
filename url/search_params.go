package url

import (
	"net/url"
	"sort"

	"github.com/dop251/goja"
)

func createSearchParams(r *goja.Runtime, u *url.URL) *goja.Object {
	o := r.NewObject()
	p := r.NewObject()

	p.SetSymbol(goja.SymIterator, iterator(r, u, func(ps [][2]string, i int) interface{} { return ps[i] }))

	p.Set("append", func(name string, value string) {
		v := u.Query()
		v.Add(name, value)
		u.RawQuery = v.Encode()
	})

	p.Set("delete", func(name string) {
		v := u.Query()
		v.Del(name)
		u.RawQuery = v.Encode()
	})

	p.Set("entries", func() *goja.Object {
		o := r.NewObject()
		o.SetSymbol(goja.SymIterator, iterator(r, u, func(ps [][2]string, i int) interface{} { return ps[i] }))
		return o
	})

	p.Set("forEach", func(f func(string, string)) {
		for _, e := range paramQueryToArray(u) {
			f(e[0], e[1])
		}
	})

	p.Set("get", func(name string) string {
		v := u.Query().Get(name)
		return v
	})

	p.Set("getAll", func(name string) []string {
		if u.Query().Has(name) {
			return u.Query()[name]
		}
		return []string{}
	})

	p.Set("has", func(name string) bool {
		has := u.Query().Has(name)
		return has
	})

	p.Set("keys", func() *goja.Object {
		o := r.NewObject()
		o.SetSymbol(goja.SymIterator, iterator(r, u, func(ps [][2]string, i int) interface{} { return ps[i][0] }))
		return o
	})

	p.Set("values", func() *goja.Object {
		o := r.NewObject()
		o.SetSymbol(goja.SymIterator, iterator(r, u, func(ps [][2]string, i int) interface{} { return ps[i][1] }))
		return o
	})

	p.Set("set", func(name string, value string) {
		v := u.Query()
		v.Set(name, value)
		u.RawQuery = v.Encode()
	})

	p.Set("sort", func() {
		q := u.Query()
		ks := searchParamKeys(q)
		sort.Strings(ks)

		v := url.Values{}
		for _, val := range ks {
			v.Set(val, q.Get(val))
		}
		u.RawQuery = v.Encode()
	})

	p.Set("toString", func() string {
		return u.RawQuery
	})

	o.SetPrototype(p)
	return o
}

func searchParamKeys(v url.Values) []string {
	var ks = []string{}
	for k, _ := range v {
		ks = append(ks, k)
	}
	return ks
}

// Flattens the query map[string][]string into an array of [name, value] entries
func paramQueryToArray(u *url.URL) [][2]string {
	r := [][2]string{}
	for k, vs := range u.Query() {
		for _, v := range vs {
			r = append(r, [2]string{k, v})
		}
	}
	return r
}

func iterator(r *goja.Runtime, u *url.URL, extractor func([][2]string, int) interface{}) func() *goja.Object {
	return func() *goja.Object {
		io := r.NewObject()
		i := 0

		io.Set("next", func() *goja.Object {
			ps := paramQueryToArray(u)
			n := r.NewObject()

			if i < len(ps) {
				n.Set("value", extractor(ps, i))
				n.Set("done", false)
			} else {
				n.Set("done", true)
			}
			i += 1
			return n
		})
		return io
	}
}
