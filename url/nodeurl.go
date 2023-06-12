package url

import (
	"fmt"
	"net/url"
	"strings"
)

type nodeURL struct {
	href         string
	origin       string
	protocol     string
	username     string
	password     string
	host         string
	hostname     string
	port         string
	pathname     string
	search       string
	searchParams searchParams
	hash         string
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

func (nu *nodeURL) String() string {
	if nu.host == "" && nu.hostname == "" {
		return nu.href
	}

	str := ""
	if nu.protocol != "" {
		str = fmt.Sprintf("%s%s://", str, nu.protocol)
	}

	if nu.username != "" {
		str = fmt.Sprintf("%s%s:%s@", str, nu.username, nu.password)
	}

	if nu.host != "" {
		str = fmt.Sprintf("%s%s", str, url.PathEscape(nu.host))
	}

	if nu.pathname != "" {
		u, err := url.Parse(nu.pathname)
		if err == nil {
			str = fmt.Sprintf("%s%s", str, u.EscapedPath())
		}
	}

	if nu.search != "" {
		str = fmt.Sprintf("%s%s", str, encodeSearchParams(nu.searchParams))
	}

	if nu.hash != "" {
		str = fmt.Sprintf("%s#%s", str, url.PathEscape(nu.hash))
	}

	nu.href = str

	return str
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
	p, _ := u.User.Password()
	sp, _ := parseSearchQuery(u.RawQuery)

	nu := nodeURL{
		href:         u.String(),
		origin:       u.Scheme + "://" + u.Hostname(),
		protocol:     u.Scheme,
		username:     u.User.Username(),
		password:     p,
		host:         u.Host,
		hostname:     strings.Split(u.Host, ":")[0],
		port:         u.Port(),
		pathname:     u.Path,
		search:       encodeSearchParams(sp),
		searchParams: sp,
		hash:         u.Fragment,
	}

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
