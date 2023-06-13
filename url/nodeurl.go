package url

import (
	"fmt"
	"net/url"
	"strings"
)

type nodeURL struct {
	url          *url.URL
	searchParams searchParams
}

type searchParam struct {
	name  string
	value []string
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

func (sp *searchParam) Encode() string {
	vals := []string{}
	for _, v := range sp.value {
		vals = append(vals, url.QueryEscape(v))
	}

	str := url.QueryEscape(sp.name)
	if len(vals) > 0 {
		str = fmt.Sprintf("%s=%s", str, strings.Join(vals, ","))
	}
	return str
}

func (nu *nodeURL) syncSearchParams() {
	nu.url.RawQuery = strings.TrimPrefix(encodeSearchParams(nu.searchParams), "?")
}

func (nu *nodeURL) String() string {
	return nu.url.String()
}

// Second return value determines if name was found in the search
func (nu *nodeURL) getValues(name string) ([]string, bool) {
	contained := false

	vals := []string{}
	for _, v := range nu.searchParams {
		if v.name == name {
			contained = true
			vals = append(vals, v.value...)
		}
	}

	return vals, contained
}

func encodeSearchParams(sp searchParams) string {
	str := ""
	sep := "?"
	for _, v := range sp {
		str = fmt.Sprintf("%s%s%s", str, sep, v.Encode())
		sep = "&"
	}
	return str
}

func newFromURL(u *url.URL) *nodeURL {
	sp, _ := parseSearchQuery(u.RawQuery)
	nu := nodeURL{url: u, searchParams: sp}
	nu.syncSearchParams()
	return &nu
}

func parseSearchQuery(query string) (searchParams, error) {
	ret := searchParams{}
	if query == "" {
		return ret, nil
	}

	query = strings.TrimPrefix(query, "?")

	for _, v := range strings.Split(query, "&") {
		pair := strings.Split(v, "=")
		name := pair[0]
		sp := searchParam{name: name, value: []string{}}
		if len(pair) > 1 {
			sp.value = append(sp.value, strings.Split(pair[1], ",")...)
		}
		ret = append(ret, sp)
	}

	return ret, nil
}
