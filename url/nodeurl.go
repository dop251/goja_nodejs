package url

import (
	"fmt"
	"net/url"
	"strings"
)

type searchParam struct {
	name  string
	value []string
}

func (sp *searchParam) Encode() string {
	return sp.string(true)
}

func (sp *searchParam) string(encode bool) string {
	vals := []string{}
	for _, v := range sp.value {
		if encode {
			vals = append(vals, fmt.Sprintf("%s=%s", url.QueryEscape(sp.name), url.QueryEscape(v)))
		} else {
			vals = append(vals, fmt.Sprintf("%s=%s", sp.name, v))
		}
	}
	return strings.Join(vals, "&")
}

type searchParams []searchParam

func (s searchParams) Len() int {
	return len(s)
}

func (s searchParams) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s searchParams) Less(i, j int) bool {
	return len(s[i].name) > len(s[j].name)
}

func (s searchParams) Encode() string {
	str := ""
	sep := ""
	for _, v := range s {
		str = fmt.Sprintf("%s%s%s", str, sep, v.Encode())
		sep = "&"
	}
	return str
}

func (s searchParams) String() string {
	var b strings.Builder
	sep := ""
	for _, v := range s {
		b.WriteString(sep)
		b.WriteString(v.string(false)) // keep it raw
		sep = "&"
	}
	return b.String()
}

type nodeURL struct {
	url          *url.URL
	searchParams searchParams
}

// This methods ensures that the url.URL has the proper RawQuery based on the searchParam
// structs. If a change is made to the searchParams we need to keep them in sync.
func (nu *nodeURL) syncSearchParams() {
	nu.url.RawQuery = nu.searchParams.Encode()
}

func (nu *nodeURL) String() string {
	return nu.url.String()
}

func (nu *nodeURL) hasName(name string) bool {
	for _, v := range nu.searchParams {
		if v.name == name {
			return true
		}
	}
	return false
}

func (nu *nodeURL) getValues(name string) []string {
	vals := []string{}
	for _, v := range nu.searchParams {
		if v.name == name {
			vals = append(vals, v.value...)
		}
	}

	return vals
}

func parseSearchQuery(query string) searchParams {
	ret := searchParams{}
	if query == "" {
		return ret
	}

	query = strings.TrimPrefix(query, "?")

	for _, v := range strings.Split(query, "&") {
		pair := strings.Split(v, "=")
		name := pair[0]
		sp := searchParam{name: name, value: []string{}}
		if len(pair) > 1 {
			sp.value = append(sp.value, []string{pair[1]}...)
		}
		ret = append(ret, sp)
	}

	return ret
}
