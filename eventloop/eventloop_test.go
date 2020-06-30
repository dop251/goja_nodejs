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

func TestNativeTimeout(t *testing.T) {
	fired := false
	loop := NewEventLoop()
	loop.SetTimeout(func(_ *goja.Runtime) {
		fired = true
	}, 1*time.Second)
	loop.Run(func(_ *goja.Runtime) {
		// do not schedule anything
	})
	if !fired {
		t.Fatal("Not fired")
	}
}

func TestNativeClearTimeout(t *testing.T) {
	fired := false
	loop := NewEventLoop()
	timer := loop.SetTimeout(func(_ *goja.Runtime) {
		fired = true
	}, 2*time.Second)
	loop.SetTimeout(func(_ *goja.Runtime) {
		loop.ClearTimeout(timer)
	}, 1*time.Second)
	loop.Run(func(_ *goja.Runtime) {
		// do not schedule anything
	})
	if fired {
		t.Fatal("Cancelled timer fired!")
	}
}

func TestNativeInterval(t *testing.T) {
	count := 0
	loop := NewEventLoop()
	var i *Interval
	i = loop.SetInterval(func(_ *goja.Runtime) {
		t.Log("tick")
		count++
		if count > 2 {
			loop.ClearInterval(i)
		}
	}, 1*time.Second)
	loop.Run(func(_ *goja.Runtime) {
		// do not schedule anything
	})
	if count != 3 {
		t.Fatal("Expected interval to fire 3 times, got", count)
	}
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

func TestNativeClearInterval(t *testing.T) {
	count := 0
	loop := NewEventLoop()
	loop.Run(func(_ *goja.Runtime) {
		i := loop.SetInterval(func(_ *goja.Runtime) {
			t.Log("tick")
			count++
		}, 500*time.Millisecond)
		<-time.After(2 * time.Second)
		loop.ClearInterval(i)
	})
	if count != 0 {
		t.Fatal("Expected interval to fire 0 times, got", count)
	}
}

func TestClearInterval(t *testing.T) {
	const SCRIPT = `
	var count = 0;
	console.log("calling setInterval");
	var t = setInterval(function() {
		console.log("tick");
	}, 500);
	console.log("calling sleep");
        sleep(2000);
	console.log("calling clearInterval");
        clearInterval(t);
	`

	loop := NewEventLoop()
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	var count int64
	loop.Run(func(vm *goja.Runtime) {
		vm.Set("sleep", func(ms int) {
			<-time.After(time.Duration(ms) * time.Millisecond)
		})
		vm.RunProgram(prg)
		count = vm.Get("count").ToInteger()
	})

	if count != 0 {
		t.Fatal("Expected count 0, got", 0)
	}
}
