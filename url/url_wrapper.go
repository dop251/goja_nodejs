package url

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Wrapper structure meant to allow the same features as the NodeJS URL API.
// Source: https://nodejs.org/api/url.html

type urlWrapper struct {
	url *url.URL
}

func newWrapper(u *url.URL) *urlWrapper {
	return &urlWrapper{
		url: u,
	}
}

// Hash
func (u *urlWrapper) getHash() string {
	if u.url.Fragment != "" {
		return "#" + u.url.Fragment
	}
	return ""
}

func (u *urlWrapper) setHash(v string) {
	u.url.Fragment = strings.Replace(v, "#", "", 1)
}

// Host
func (u *urlWrapper) getHost() string {
	return u.url.Host
}

func (u *urlWrapper) setHost(v string) {
	u.url.Host = v
}

// Hostname
func (u *urlWrapper) getHostname() string {
	return u.url.Hostname()
}

func (u *urlWrapper) setHostname(v string) {
	hostname := strings.Split(v, ":")[0]
	u.url.Host = hostname + ":" + u.url.Port()
}

// Href
func (u *urlWrapper) getHref() string {
	return u.url.String()
}

func (u *urlWrapper) setHref(v string) error {
	url, err := url.ParseRequestURI(v)
	if err != nil {
		return err
	}
	u.url = url
	return nil
}

// Pathname
func (u *urlWrapper) getPathname() string {
	return u.url.Path
}

func (u *urlWrapper) setPathname(v string) {
	u.url.Path = v
}

// Origin
func (u *urlWrapper) getOrigin() string {
	return u.url.Scheme + "://" + u.url.Hostname()
}

// Password
func (u *urlWrapper) getPassword() string {
	v, _ := u.url.User.Password()
	return v
}

func (u *urlWrapper) setPassword(v string) {
	user := u.url.User
	u.url.User = url.UserPassword(user.Username(), v)
}

// Username
func (u *urlWrapper) getUsername() string {
	return u.url.User.Username()
}

func (u *urlWrapper) setUsername(v string) {
	p, has := u.url.User.Password()
	if !has {
		u.url.User = url.User(v)
	} else {
		u.url.User = url.UserPassword(v, p)
	}
}

// Port
func (u *urlWrapper) getPort() string {
	return u.url.Port()
}

func (u *urlWrapper) setPort(v string) {
	f, _ := strconv.ParseFloat(v, 64)
	i := int(f)
	if i > 65535 {
		i = 65535
	}
	u.url.Host = u.url.Hostname() + ":" + fmt.Sprintf("%d", i)
}

// Protocol
func (u *urlWrapper) getProtocol() string {
	return u.url.Scheme + ":"
}

func (u *urlWrapper) setProtocol(v string) {
	u.url.Scheme = strings.Replace(v, ":", "", -1)
}

// Search
func (u *urlWrapper) getSearch() string {
	s := strings.Split(u.url.RawQuery, "#")[0]
	if s != "" {
		return "?" + s
	}
	return ""
}

func (u *urlWrapper) setSearch(v string) {
	u.url.RawQuery = v + u.getHash()
}

func (u *urlWrapper) toString() string {
	return u.url.String()
}

func (u *urlWrapper) toJSON() string {
	return u.toString()
}
