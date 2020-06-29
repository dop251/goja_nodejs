package eventloop

import (
	"testing"
	"time"

	"github.com/dop251/goja"
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

func TestRunNoSchedule(t *testing.T) {
	loop := NewEventLoop()
	fired := false
	loop.Run(func(vm *goja.Runtime) { // should not hang
		fired = true
		// do not schedule anything
	})

	if !fired {
		t.Fatal("Not fired")
	}
}

func TestRunWithConsole(t *testing.T) {
	const SCRIPT = `
	console.log("Started");
	`

	loop := NewEventLoop()
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		_, err = vm.RunProgram(prg)
	})
	if err != nil {
		t.Fatal("Call to console.log generated an error", err)
	}

	loop = NewEventLoop(EnableConsole(true))
	prg, err = goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		_, err = vm.RunProgram(prg)
	})
	if err != nil {
		t.Fatal("Call to console.log generated an error", err)
	}
}

func TestRunNoConsole(t *testing.T) {
	const SCRIPT = `
	console.log("Started");
	`

	loop := NewEventLoop(EnableConsole(false))
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		_, err = vm.RunProgram(prg)
	})
	if err == nil {
		t.Fatal("Call to console.log did not generate an error", err)
	}
}
