package url

import (
	"net/url"
	"strings"
)

type searchParam struct {
	name  string
	value string
}

func (sp *searchParam) Encode() string {
	return sp.string(true)
}

func escapeSearchParam(s string) string {
	return escape(s, &tblEscapeURLQueryParam, true)
}

func (sp *searchParam) string(encode bool) string {
	if encode {
		return escapeSearchParam(sp.name) + "=" + escapeSearchParam(sp.value)
	} else {
		return sp.name + "=" + sp.value
	}
}

type searchParams []searchParam

func (s searchParams) Len() int {
	return len(s)
}

func (s searchParams) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s searchParams) Less(i, j int) bool {
	return strings.Compare(s[i].name, s[j].name) < 0
}

func (s searchParams) Encode() string {
	var sb strings.Builder
	for i, v := range s {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(v.Encode())
	}
	return sb.String()
}

func (s searchParams) String() string {
	var sb strings.Builder
	for i, v := range s {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(v.string(false))
	}
	return sb.String()
}

type nodeURL struct {
	url          *url.URL
	searchParams searchParams
}

type urlSearchParams nodeURL

// This methods ensures that the url.URL has the proper RawQuery based on the searchParam
// structs. If a change is made to the searchParams we need to keep them in sync.
func (nu *nodeURL) syncSearchParams() {
	if nu.rawQueryUpdateNeeded() {
		nu.url.RawQuery = nu.searchParams.Encode()
	}
}

func (nu *nodeURL) rawQueryUpdateNeeded() bool {
	return len(nu.searchParams) > 0 && nu.url.RawQuery == ""
}

func (nu *nodeURL) String() string {
	return nu.url.String()
}

func (sp *urlSearchParams) hasName(name string) bool {
	for _, v := range sp.searchParams {
		if v.name == name {
			return true
		}
	}
	return false
}

func (sp *urlSearchParams) hasValue(name, value string) bool {
	for _, v := range sp.searchParams {
		if v.name == name && v.value == value {
			return true
		}
	}
	return false
}

func (sp *urlSearchParams) getValues(name string) []string {
	vals := make([]string, 0, len(sp.searchParams))
	for _, v := range sp.searchParams {
		if v.name == name {
			vals = append(vals, v.value)
		}
	}

	return vals
}

func (sp *urlSearchParams) getFirstValue(name string) (string, bool) {
	for _, v := range sp.searchParams {
		if v.name == name {
			return v.value, true
		}
	}

	return "", false
}

func parseSearchQuery(query string) (ret searchParams) {
	if query == "" {
		return
	}

	query = strings.TrimPrefix(query, "?")

	for _, v := range strings.Split(query, "&") {
		if v == "" {
			continue
		}
		pair := strings.SplitN(v, "=", 2)
		l := len(pair)
		if l == 1 {
			ret = append(ret, searchParam{name: unescapeSearchParam(pair[0]), value: ""})
		} else if l == 2 {
			ret = append(ret, searchParam{name: unescapeSearchParam(pair[0]), value: unescapeSearchParam(pair[1])})
		}
	}

	return
}
