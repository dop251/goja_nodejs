package url

import (
	_ "embed"
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

func TestURL(t *testing.T) {
	vm := goja.New()
	new(require.Registry).Enable(vm)
	Enable(vm)

	if c := vm.Get("URL"); c == nil {
		t.Fatal("URL not found")
	}

	script := `const url = new URL("https://user:pass@sub.example.com:8080/p/a/t/h?query=string#hash");`

	if _, err := vm.RunString(script); err != nil {
		t.Fatal("Failed to process url script.", err)
	}
}

func TestGetters(t *testing.T) {
	vm := goja.New()
	new(require.Registry).Enable(vm)
	Enable(vm)

	if c := vm.Get("URL"); c == nil {
		t.Fatal("URL not found")
	}

	script := `
		new URL("https://user:pass@sub.example.com:8080/p/a/t/h?query=string#hashed");
	`

	v, err := vm.RunString(script)
	if err != nil {
		t.Fatal("Failed to process url script.", err)
	}

	url := v.ToObject(vm)

	tests := []struct {
		prop     string
		expected string
	}{
		{
			prop:     "hash",
			expected: "#hashed",
		},
		{
			prop:     "host",
			expected: "sub.example.com:8080",
		},
		{
			prop:     "hostname",
			expected: "sub.example.com",
		},
		{
			prop:     "href",
			expected: "https://user:pass@sub.example.com:8080/p/a/t/h?query=string#hashed",
		},
		{
			prop:     "origin",
			expected: "https://sub.example.com",
		},
		{
			prop:     "password",
			expected: "pass",
		},
		{
			prop:     "username",
			expected: "user",
		},
		{
			prop:     "port",
			expected: "8080",
		},
		{
			prop:     "protocol",
			expected: "https:",
		},
		{
			prop:     "search",
			expected: "?query=string",
		},
	}

	for _, test := range tests {
		v := url.Get(test.prop).String()
		if v != test.expected {
			t.Fatal("failed to match " + test.prop + " property. got: " + v + ", expected: " + test.expected)
		}
	}
}

//go:embed testdata/url_test.js
var urlTest string

func TestJs(t *testing.T) {
	vm := goja.New()
	new(require.Registry).Enable(vm)
	Enable(vm)

	if c := vm.Get("URL"); c == nil {
		t.Fatal("URL not found")
	}

	// Script will throw an error on failed validation

	_, err := vm.RunScript("testdata/url_test.js", urlTest)
	if err != nil {
		if ex, ok := err.(*goja.Exception); ok {
			t.Fatal(ex.String())
		}
		t.Fatal("Failed to process url script.", err)
	}
}
