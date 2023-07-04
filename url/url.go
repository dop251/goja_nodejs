package url

import (
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"golang.org/x/net/idna"
)

const (
	URLNotAbsolute  = "URL is not absolute"
	InvalidURL      = "Invalid URL"
	InvalidBaseURL  = "Invalid base URL"
	InvalidHostname = "Invalid hostname"
)

var (
	reflectTypeURL = reflect.TypeOf((*nodeURL)(nil))
	reflectTypeInt = reflect.TypeOf(0)
)

func newInvalidURLError(r *goja.Runtime, msg, input string) *goja.Object {
	// when node's error module is added this should return a NodeError
	o := r.NewTypeError(msg)
	o.Set("input", r.ToValue(input))
	return o
}

func valueToURLPort(v goja.Value) (portNum int, empty bool) {
	portNum = -1
	if et := v.ExportType(); et == reflectTypeInt {
		if num := v.ToInteger(); num >= 0 && num <= math.MaxUint16 {
			portNum = int(num)
		}
	} else {
		s := v.String()
		if s == "" {
			return 0, true
		}
		for i := 0; i < len(s); i++ {
			if c := s[i]; c >= '0' && c <= '9' {
				if portNum == -1 {
					portNum = 0
				}
				portNum = portNum*10 + int(c-'0')
				if portNum > math.MaxUint16 {
					portNum = -1
					break
				}
			} else {
				break
			}
		}
	}
	return
}

func isDefaultURLPort(protocol string, port int) bool {
	switch port {
	case 21:
		if protocol == "ftp" {
			return true
		}
	case 80:
		if protocol == "http" || protocol == "ws" {
			return true
		}
	case 443:
		if protocol == "https" || protocol == "wss" {
			return true
		}
	}
	return false
}

func isSpecialProtocol(protocol string) bool {
	switch protocol {
	case "ftp", "file", "http", "https", "ws", "wss":
		return true
	}
	return false
}

func clearURLPort(u *url.URL) {
	u.Host = u.Hostname()
}

func setURLPort(nu *nodeURL, v goja.Value) {
	u := nu.url
	if u.Scheme == "file" {
		return
	}
	portNum, empty := valueToURLPort(v)
	if empty {
		clearURLPort(u)
		return
	}
	if portNum == -1 {
		return
	}
	if isDefaultURLPort(u.Scheme, portNum) {
		clearURLPort(u)
	} else {
		u.Host = u.Hostname() + ":" + strconv.Itoa(portNum)
	}
}

func parseURL(r *goja.Runtime, s string, isBase bool) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		if isBase {
			panic(newInvalidURLError(r, InvalidBaseURL, s))
		} else {
			panic(newInvalidURLError(r, InvalidURL, s))
		}
	}
	if isBase && !u.IsAbs() {
		panic(newInvalidURLError(r, URLNotAbsolute, s))
	}
	if portStr := u.Port(); portStr != "" {
		if port, err := strconv.Atoi(portStr); err != nil || isDefaultURLPort(u.Scheme, port) {
			u.Host = u.Hostname() // Clear port
		}
	}
	fixURL(r, u)
	return u
}

func fixRawQuery(u *url.URL) {
	if u.RawQuery != "" {
		var u1 url.URL
		u1.Fragment = u.RawQuery
		u.RawQuery = u1.EscapedFragment()
	}
}

func fixURL(r *goja.Runtime, u *url.URL) {
	switch u.Scheme {
	case "https", "http", "ftp", "wss", "ws":
		if u.Path == "" {
			u.Path = "/"
		}
		hostname := u.Hostname()
		lh := strings.ToLower(hostname)
		ch, err := idna.Punycode.ToASCII(lh)
		if err != nil {
			panic(newInvalidURLError(r, InvalidHostname, lh))
		}
		if ch != hostname {
			if port := u.Port(); port != "" {
				u.Host = ch + ":" + port
			} else {
				u.Host = ch
			}
		}
	}
}

func createURLPrototype(r *goja.Runtime) *goja.Object {
	p := r.NewObject()

	// host
	defineURLAccessorProp(r, p, "host", func(u *nodeURL) interface{} {
		return u.url.Host
	}, func(u *nodeURL, arg goja.Value) {
		host := arg.String()
		if _, err := url.ParseRequestURI(u.url.Scheme + "://" + host); err == nil {
			u.url.Host = host
			fixURL(r, u.url)
		}
	})

	// hash
	defineURLAccessorProp(r, p, "hash", func(u *nodeURL) interface{} {
		if u.url.Fragment != "" {
			return "#" + u.url.EscapedFragment()
		}
		return ""
	}, func(u *nodeURL, arg goja.Value) {
		h := arg.String()
		if len(h) > 0 && h[0] == '#' {
			h = h[1:]
		}
		u.url.Fragment = h
	})

	// hostname
	defineURLAccessorProp(r, p, "hostname", func(u *nodeURL) interface{} {
		return strings.Split(u.url.Host, ":")[0]
	}, func(u *nodeURL, arg goja.Value) {
		h := arg.String()
		if strings.IndexByte(h, ':') >= 0 {
			return
		}
		if _, err := url.ParseRequestURI(u.url.Scheme + "://" + h); err == nil {
			if port := u.url.Port(); port != "" {
				u.url.Host = h + ":" + port
			} else {
				u.url.Host = h
			}
			fixURL(r, u.url)
		}
	})

	// href
	defineURLAccessorProp(r, p, "href", func(u *nodeURL) interface{} {
		return u.String()
	}, func(u *nodeURL, arg goja.Value) {
		url := parseURL(r, arg.String(), true)
		*u.url = *url
	})

	// pathname
	defineURLAccessorProp(r, p, "pathname", func(u *nodeURL) interface{} {
		return u.url.EscapedPath()
	}, func(u *nodeURL, arg goja.Value) {
		p := arg.String()
		if _, err := url.Parse(p); err == nil {
			switch u.url.Scheme {
			case "https", "http", "ftp", "ws", "wss":
				if !strings.HasPrefix(p, "/") {
					p = "/" + p
				}
			}
			u.url.Path = p
		}
	})

	// origin
	defineURLAccessorProp(r, p, "origin", func(u *nodeURL) interface{} {
		return u.url.Scheme + "://" + u.url.Hostname()
	}, nil)

	// password
	defineURLAccessorProp(r, p, "password", func(u *nodeURL) interface{} {
		p, _ := u.url.User.Password()
		return p
	}, func(u *nodeURL, arg goja.Value) {
		user := u.url.User
		u.url.User = url.UserPassword(user.Username(), arg.String())
	})

	// username
	defineURLAccessorProp(r, p, "username", func(u *nodeURL) interface{} {
		return u.url.User.Username()
	}, func(u *nodeURL, arg goja.Value) {
		p, has := u.url.User.Password()
		if !has {
			u.url.User = url.User(arg.String())
		} else {
			u.url.User = url.UserPassword(arg.String(), p)
		}
	})

	// port
	defineURLAccessorProp(r, p, "port", func(u *nodeURL) interface{} {
		return u.url.Port()
	}, func(u *nodeURL, arg goja.Value) {
		setURLPort(u, arg)
	})

	// protocol
	defineURLAccessorProp(r, p, "protocol", func(u *nodeURL) interface{} {
		return u.url.Scheme + ":"
	}, func(u *nodeURL, arg goja.Value) {
		s := arg.String()
		pos := strings.IndexByte(s, ':')
		if pos >= 0 {
			s = s[:pos]
		}
		s = strings.ToLower(s)
		if isSpecialProtocol(u.url.Scheme) == isSpecialProtocol(s) {
			if _, err := url.ParseRequestURI(s + "://" + u.url.Host); err == nil {
				u.url.Scheme = s
			}
		}
	})

	// Search
	defineURLAccessorProp(r, p, "search", func(u *nodeURL) interface{} {
		if u.url.RawQuery != "" {
			return "?" + u.url.RawQuery
		}
		return ""
	}, func(u *nodeURL, arg goja.Value) {
		u.url.RawQuery = arg.String()
		fixRawQuery(u.url)
	})

	// search Params
	defineURLAccessorProp(r, p, "searchParams", func(u *nodeURL) interface{} {
		if u.url.RawQuery != "" && len(u.searchParams) == 0 {
			sp, _ := parseSearchQuery(u.url.RawQuery)
			u.searchParams = sp
		}

		o := r.ToValue(u).(*goja.Object)
		o.SetPrototype(createURLSearchParamsPrototype(r))
		return o
	}, func(u *nodeURL, arg goja.Value) {
		nu := toURL(r, arg)
		u.searchParams = nu.searchParams
		u.syncSearchParams()
	})

	p.Set("toString", r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(r, call.This)

		// Search Parameters are lazy loaded
		if u.url.RawQuery != "" && len(u.searchParams) == 0 {
			sp, _ := parseSearchQuery(u.url.RawQuery)
			u.searchParams = sp
		}
		copy := u.url
		copy.RawQuery = u.searchParams.Encode()
		return r.ToValue(u.url.String())
	}))

	p.Set("toJSON", r.ToValue(func(call goja.FunctionCall) goja.Value {
		return r.ToValue(toURL(r, call.This).String())
	}))

	return p
}

func createURLConstructor(r *goja.Runtime) goja.Value {
	f := r.ToValue(func(call goja.ConstructorCall) *goja.Object {
		var u *url.URL
		if baseArg := call.Argument(1); !goja.IsUndefined(baseArg) {
			base := parseURL(r, baseArg.String(), true)
			ref := parseURL(r, call.Arguments[0].String(), false)
			u = base.ResolveReference(ref)
		} else {
			u = parseURL(r, call.Argument(0).String(), true)
		}
		res := r.ToValue(&nodeURL{url: u}).(*goja.Object)
		res.SetPrototype(call.This.Prototype())
		return res
	}).(*goja.Object)

	f.Set("prototype", createURLPrototype(r))
	return f
}
