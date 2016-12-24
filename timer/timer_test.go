package timer

import (
	"testing"
	"time"

	"github.com/dop251/goja"
)

func TestSetTimeout(t *testing.T) {
	registry := NewRegistry()
	vm := goja.New()
	registry.Enable(vm)

	start := time.Now()

	called := false

	vm.Set("done", func(delay int) {
		et := time.Now().Sub(start)
		t.Logf("delay = %d, elapsed time = %v", delay, et)
		called = true
	})

	_, err := vm.RunString(`
    setTimeout(function(){
        done(1000)
    }, 1000);


    setTimeout(function(){
        done(500)
    }, 500);
    `)

	if err != nil {
		t.Fatal(err)
	}

	registry.Wait()

	if !called {
		t.Error("callback was not called.")
	}
}

func TestClearTimer(t *testing.T) {
	registry := NewRegistry()
	vm := goja.New()
	registry.Enable(vm)

	start := time.Now()

	vm.Set("done", func(delay int) {
		et := time.Now().Sub(start)
		t.Errorf("cancell error: delay = %d elapsed time = %v", delay, et)
	})

	vm.Set("log", t.Log)

	_, err := vm.RunString(`
    var timer = setTimeout(function(){
        done(500)
    }, 500);

    clearTimeout(timer)
    `)

	if err != nil {
		t.Fatal(err)
	}

	registry.Wait()
}

func TestSetInterval(t *testing.T) {
	registry := NewRegistry()
	vm := goja.New()
	registry.Enable(vm)

	start := time.Now()

	called := false

	vm.Set("done", func(delay, i int) {
		et := time.Now().Sub(start)
		t.Logf("delay = %d i = %d elapsed time = %v", delay, i, et)

		if i != 3 {
			t.Errorf("i = %d; want 3", i)
		}

		called = true
	})

	vm.Set("log", t.Log)

	_, err := vm.RunString(`
    var i = 0
    var timer = setInterval(function(){
        i++
        if(i >= 3) {
            clearTimeout(timer)
			done(50, i)
        }
    }, 50);
    `)

	if err != nil {
		t.Fatal(err)
	}

	registry.Wait()

	if !called {
		t.Error("callback was not called.")
	}
}
