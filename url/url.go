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

func setURLPort(u *nodeURL, v goja.Value) {
	if u.protocol == "file" {
		return
	}
	portNum, empty := valueToURLPort(v)
	if empty {
		u.port = ""
		return
	}
	if portNum == -1 {
		return
	}
	if isDefaultURLPort(u.protocol, portNum) {
		u.port = ""
	} else {
		u.port = strconv.Itoa(portNum)
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
		return u.host
	}, func(u *nodeURL, arg goja.Value) {
		host := arg.String()
		if _, err := url.ParseRequestURI(u.protocol + "://" + host); err == nil {
			lh := strings.ToLower(host)
			h, err := idna.Punycode.ToASCII(lh)
			if err != nil {
				panic(newInvalidURLError(r, InvalidHostname, lh))
			}
			u.host = h

			// Update hostname
			vals := strings.Split(h, ":")
			if len(vals) > 1 {
				u.hostname = vals[0]
			}
		}
	})

	// hash
	defineURLAccessorProp(r, p, "hash", func(u *nodeURL) interface{} {
		return "#" + u.hash
	}, func(u *nodeURL, arg goja.Value) {
		h := arg.String()
		if len(h) > 0 && h[0] == '#' {
			h = h[1:]
		}
		u.hash = h
	})

	// hostname
	defineURLAccessorProp(r, p, "hostname", func(u *nodeURL) interface{} {
		return u.hostname
	}, func(u *nodeURL, arg goja.Value) {
		h := arg.String()
		if strings.IndexByte(h, ':') >= 0 {
			return
		}
		if _, err := url.ParseRequestURI(u.protocol + "://" + h); err == nil {
			lh := strings.ToLower(h)
			host, err := idna.Punycode.ToASCII(lh)
			if err != nil {
				panic(newInvalidURLError(r, InvalidHostname, lh))
			}
			u.hostname = host

			// Update Host
			if u.port != "" {
				u.host = host + ":" + u.port
			}
		}
	})

	// href
	defineURLAccessorProp(r, p, "href", func(u *nodeURL) interface{} {
		return u.String() // Encoded
	}, func(u *nodeURL, arg goja.Value) {
		url := parseURL(r, arg.String(), true)
		*u = *newFromURL(url)
	})

	// pathname
	defineURLAccessorProp(r, p, "pathname", func(u *nodeURL) interface{} {
		url, _ := url.Parse(u.pathname)
		return url.String()
	}, func(u *nodeURL, arg goja.Value) {
		p := arg.String()
		if _, err := url.Parse(p); err == nil {
			switch u.protocol {
			case "https", "http", "ftp", "ws", "wss":
				if !strings.HasPrefix(p, "/") {
					p = "/" + p
				}
			}
			u.pathname = p
		}
	})

	// origin
	defineURLAccessorProp(r, p, "origin", func(u *nodeURL) interface{} {
		return u.protocol + "://" + u.hostname
	}, nil)

	// password
	defineURLAccessorProp(r, p, "password", func(u *nodeURL) interface{} {
		return u.password
	}, func(u *nodeURL, arg goja.Value) {
		u.password = arg.String()
	})

	// username
	defineURLAccessorProp(r, p, "username", func(u *nodeURL) interface{} {
		return u.username
	}, func(u *nodeURL, arg goja.Value) {
		u.username = arg.String()
	})

	// port
	defineURLAccessorProp(r, p, "port", func(u *nodeURL) interface{} {
		return u.port
	}, func(u *nodeURL, arg goja.Value) {
		setURLPort(u, arg)
	})

	// protocol
	defineURLAccessorProp(r, p, "protocol", func(u *nodeURL) interface{} {
		return u.protocol + ":"
	}, func(u *nodeURL, arg goja.Value) {
		s := arg.String()
		pos := strings.IndexByte(s, ':')
		if pos >= 0 {
			s = s[:pos]
		}
		s = strings.ToLower(s)
		if isSpecialProtocol(u.protocol) == isSpecialProtocol(s) {
			if _, err := url.ParseRequestURI(s + "://" + u.host); err == nil {
				u.protocol = s
			}
		}
	})

	// Search
	defineURLAccessorProp(r, p, "search", func(u *nodeURL) interface{} {
		return u.search
	}, func(u *nodeURL, arg goja.Value) {
		query := arg.String()
		if sp, err := parseSearchQuery(query); err == nil {
			u.search = encodeSearchParams(sp)
			u.searchParams = sp
		}
	})

	// search Params
	defineURLAccessorProp(r, p, "searchParams", func(u *nodeURL) interface{} {
		o := r.ToValue(u).(*goja.Object)
		o.SetPrototype(createURLSearchParamsPrototype(r))
		return o
	}, func(u *nodeURL, arg goja.Value) {
		nu := toURL(r, arg)
		u.searchParams = nu.searchParams
		u.search = encodeSearchParams(nu.searchParams)
	})

	p.Set("toString", r.ToValue(func(call goja.FunctionCall) goja.Value {
		return r.ToValue(toURL(r, call.This).String())
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
		res := r.ToValue(newFromURL(u)).(*goja.Object)
		res.SetPrototype(call.This.Prototype())
		return res
	}).(*goja.Object)

	f.Set("prototype", createURLPrototype(r))
	return f
}
