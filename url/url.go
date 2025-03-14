package url

import (
	"math"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/errors"

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
	reflectTypeInt = reflect.TypeOf(int64(0))
)

func toURL(r *goja.Runtime, v goja.Value) *nodeURL {
	if v.ExportType() == reflectTypeURL {
		if u := v.Export().(*nodeURL); u != nil {
			return u
		}
	}

	panic(errors.NewTypeError(r, errors.ErrCodeInvalidThis, `Value of "this" must be of type URL`))
}

func (m *urlModule) newInvalidURLError(msg, input string) *goja.Object {
	o := errors.NewTypeError(m.r, "ERR_INVALID_URL", msg)
	o.Set("input", m.r.ToValue(input))
	return o
}

func (m *urlModule) defineURLAccessorProp(p *goja.Object, name string, getter func(*nodeURL) interface{}, setter func(*nodeURL, goja.Value)) {
	var getterVal, setterVal goja.Value
	if getter != nil {
		getterVal = m.r.ToValue(func(call goja.FunctionCall) goja.Value {
			return m.r.ToValue(getter(toURL(m.r, call.This)))
		})
	}
	if setter != nil {
		setterVal = m.r.ToValue(func(call goja.FunctionCall) goja.Value {
			setter(toURL(m.r, call.This), call.Argument(0))
			return goja.Undefined()
		})
	}
	p.DefineAccessorProperty(name, getterVal, setterVal, goja.FLAG_FALSE, goja.FLAG_TRUE)
}

func valueToURLPort(v goja.Value) (portNum int, empty bool) {
	portNum = -1
	if et := v.ExportType(); et == reflectTypeInt {
		num := v.ToInteger()
		if num < 0 {
			empty = true
		} else if num <= math.MaxUint16 {
			portNum = int(num)
		}
	} else {
		s := v.String()
		if s == "" {
			return 0, true
		}
		firstDigitIdx := -1
		for i := 0; i < len(s); i++ {
			if c := s[i]; c >= '0' && c <= '9' {
				firstDigitIdx = i
				break
			}
		}

		if firstDigitIdx == -1 {
			return -1, false
		}

		if firstDigitIdx > 0 {
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

func isSpecialNetProtocol(protocol string) bool {
	switch protocol {
	case "https", "http", "ftp", "wss", "ws":
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

func (m *urlModule) parseURL(s string, isBase bool) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		if isBase {
			panic(m.newInvalidURLError(InvalidBaseURL, s))
		} else {
			panic(m.newInvalidURLError(InvalidURL, s))
		}
	}
	if isBase && !u.IsAbs() {
		panic(m.newInvalidURLError(URLNotAbsolute, s))
	}
	if isSpecialNetProtocol(u.Scheme) && u.Host == "" && u.Path == "" {
		panic(m.newInvalidURLError(InvalidURL, s))
	}
	if portStr := u.Port(); portStr != "" {
		if port, err := strconv.Atoi(portStr); err != nil || isDefaultURLPort(u.Scheme, port) {
			u.Host = u.Hostname() // Clear port
		}
	}
	m.fixURL(u)
	return u
}

func fixRawQuery(u *url.URL) {
	if u.RawQuery != "" {
		u.RawQuery = escape(u.RawQuery, &tblEscapeURLQuery, false)
	}
}

func cleanPath(p, proto string) string {
	if !strings.HasPrefix(p, "/") && (isSpecialProtocol(proto) || p != "") {
		p = "/" + p
	}
	if p != "" {
		return path.Clean(p)
	}
	return ""
}

func (m *urlModule) fixURL(u *url.URL) {
	u.Path = cleanPath(u.Path, u.Scheme)
	if isSpecialNetProtocol(u.Scheme) {
		hostname := u.Hostname()
		lh := strings.ToLower(hostname)
		ch, err := idna.Punycode.ToASCII(lh)
		if err != nil {
			panic(m.newInvalidURLError(InvalidHostname, lh))
		}
		if ch != hostname {
			if port := u.Port(); port != "" {
				u.Host = ch + ":" + port
			} else {
				u.Host = ch
			}
		}
	}
	fixRawQuery(u)
}

func (m *urlModule) createURLPrototype() *goja.Object {
	p := m.r.NewObject()

	// host
	m.defineURLAccessorProp(p, "host", func(u *nodeURL) interface{} {
		return u.url.Host
	}, func(u *nodeURL, arg goja.Value) {
		host := arg.String()
		if _, err := url.ParseRequestURI(u.url.Scheme + "://" + host); err == nil {
			u.url.Host = host
			m.fixURL(u.url)
		}
	})

	// hash
	m.defineURLAccessorProp(p, "hash", func(u *nodeURL) interface{} {
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
	m.defineURLAccessorProp(p, "hostname", func(u *nodeURL) interface{} {
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
			m.fixURL(u.url)
		}
	})

	// href
	m.defineURLAccessorProp(p, "href", func(u *nodeURL) interface{} {
		return u.String()
	}, func(u *nodeURL, arg goja.Value) {
		u.url = m.parseURL(arg.String(), true)
	})

	// pathname
	m.defineURLAccessorProp(p, "pathname", func(u *nodeURL) interface{} {
		return u.url.EscapedPath()
	}, func(u *nodeURL, arg goja.Value) {
		u.url.Path = cleanPath(arg.String(), u.url.Scheme)
	})

	// origin
	m.defineURLAccessorProp(p, "origin", func(u *nodeURL) interface{} {
		return u.url.Scheme + "://" + u.url.Hostname()
	}, nil)

	// password
	m.defineURLAccessorProp(p, "password", func(u *nodeURL) interface{} {
		p, _ := u.url.User.Password()
		return p
	}, func(u *nodeURL, arg goja.Value) {
		user := u.url.User
		u.url.User = url.UserPassword(user.Username(), arg.String())
	})

	// username
	m.defineURLAccessorProp(p, "username", func(u *nodeURL) interface{} {
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
	m.defineURLAccessorProp(p, "port", func(u *nodeURL) interface{} {
		return u.url.Port()
	}, func(u *nodeURL, arg goja.Value) {
		setURLPort(u, arg)
	})

	// protocol
	m.defineURLAccessorProp(p, "protocol", func(u *nodeURL) interface{} {
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
	m.defineURLAccessorProp(p, "search", func(u *nodeURL) interface{} {
		u.syncSearchParams()
		if u.url.RawQuery != "" {
			return "?" + u.url.RawQuery
		}
		return ""
	}, func(u *nodeURL, arg goja.Value) {
		u.url.RawQuery = arg.String()
		fixRawQuery(u.url)
		if u.searchParams != nil {
			u.searchParams = parseSearchQuery(u.url.RawQuery)
			if u.searchParams == nil {
				u.searchParams = make(searchParams, 0)
			}
		}
	})

	// search Params
	m.defineURLAccessorProp(p, "searchParams", func(u *nodeURL) interface{} {
		if u.searchParams == nil {
			sp := parseSearchQuery(u.url.RawQuery)
			if sp == nil {
				sp = make(searchParams, 0)
			}
			u.searchParams = sp
		}
		return m.newURLSearchParams((*urlSearchParams)(u))
	}, nil)

	p.Set("toString", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(m.r, call.This)
		u.syncSearchParams()
		return m.r.ToValue(u.url.String())
	}))

	p.Set("toJSON", m.r.ToValue(func(call goja.FunctionCall) goja.Value {
		u := toURL(m.r, call.This)
		u.syncSearchParams()
		return m.r.ToValue(u.url.String())
	}))

	return p
}

func (m *urlModule) createURLConstructor() goja.Value {
	f := m.r.ToValue(func(call goja.ConstructorCall) *goja.Object {
		var u *url.URL
		if baseArg := call.Argument(1); !goja.IsUndefined(baseArg) {
			base := m.parseURL(baseArg.String(), true)
			ref := m.parseURL(call.Argument(0).String(), false)
			u = base.ResolveReference(ref)
		} else {
			u = m.parseURL(call.Argument(0).String(), true)
		}
		res := m.r.ToValue(&nodeURL{url: u}).(*goja.Object)
		res.SetPrototype(call.This.Prototype())
		return res
	}).(*goja.Object)

	proto := m.createURLPrototype()
	f.Set("prototype", proto)
	proto.DefineDataProperty("constructor", f, goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_FALSE)
	return f
}

func (m *urlModule) domainToASCII(domUnicode string) string {
	res, err := idna.ToASCII(domUnicode)
	if err != nil {
		return ""
	}
	return res
}

func (m *urlModule) domainToUnicode(domASCII string) string {
	res, err := idna.ToUnicode(domASCII)
	if err != nil {
		return ""
	}
	return res
}
