package eventloop

import (
	"context"
	"testing"
	"time"

	"github.com/dop251/goja"
)

type testContextKey string

const (
	someContextKey   testContextKey = "someContext"
	someContextValue                = "someContextValue"
)

func TestRun(t *testing.T) {
	const SCRIPT = `
	setTimeout(function() {
		console.log("ok");
	}, 1000);
	console.log("Started");
	`

	loop := NewEventLoop()
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		vm.RunProgram(prg)
	})
}

func TestStart(t *testing.T) {
	const SCRIPT = `
	setTimeout(function() {
		console.log("ok");
	}, 1000);
	console.log("Started");
	`

	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}

	loop := NewEventLoop()
	loop.Start()

	loop.RunOnLoop(func(vm *goja.Runtime) {
		vm.RunProgram(prg)
	})

	time.Sleep(2 * time.Second)
	loop.Stop()
}

func TestStartWithContext(t *testing.T) {
	const SCRIPT = `
	setTimeout(function() {
		console.log("ok");
	}, 1000);
	console.log("Started");
	`

	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}

	loop := NewEventLoop()
	loop.Start()

	ctx := generateContext()
	loop.RunOnLoopWithContext(ctx, func(vm *goja.Runtime) {
		vm.RunProgram(prg)
		verifyContext(t, loop)
	})

	time.Sleep(2 * time.Second)
	loop.Stop()
}

func verifyContext(t *testing.T, eventLoop *EventLoop) {
	ctx := eventLoop.GetContext()
	if ctx == nil {
		t.Error("expected EventLoop context but none was found")
	}

	result := ctx.Value(someContextKey)
	if result != someContextValue {
		t.Errorf("expected context %s to have value %s, but it was %v", someContextKey, someContextValue, result)
	}
}

func generateContext() (ctx context.Context) {
	ctx = context.Background()
	ctx = context.WithValue(ctx, someContextKey, someContextValue)
	return
}

func TestInterval(t *testing.T) {
	const SCRIPT = `
	var count = 0;
	var t = setInterval(function() {
		console.log("tick");
		if (++count > 2) {
			clearInterval(t);
		}
	}, 1000);
	console.log("Started");
	`

	loop := NewEventLoop()
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		vm.RunProgram(prg)
	})
}
