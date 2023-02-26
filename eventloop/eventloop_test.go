package eventloop

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
)

func TestRun(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	const SCRIPT = `
	var count = 0;
	var t = setInterval(function(times) {
		console.log("tick");
		if (++count > times) {
			clearInterval(t);
		}
	}, 1000, 2);
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

func TestImmediate(t *testing.T) {
	t.Parallel()
	const SCRIPT = `
	var cb = function(arg) {
		console.log(arg);
	}
	var i;
	var t = setImmediate(function() {
		console.log("tick");
		setImmediate(cb, "tick 2");
		i = setImmediate(cb, "should not run")
	});
	setImmediate(function() {
		clearImmediate(i);
	});
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

func TestClearIntervalRace(t *testing.T) {
	t.Parallel()
	const SCRIPT = `
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
	// Should not hang
	loop.Run(func(vm *goja.Runtime) {
		vm.Set("sleep", func(ms int) {
			<-time.After(time.Duration(ms) * time.Millisecond)
		})
		vm.RunProgram(prg)
	})
}

func TestNativeTimeout(t *testing.T) {
	t.Parallel()
	fired := false
	loop := NewEventLoop()
	loop.SetTimeout(func(*goja.Runtime) {
		fired = true
	}, 1*time.Second)
	loop.Run(func(*goja.Runtime) {
		// do not schedule anything
	})
	if !fired {
		t.Fatal("Not fired")
	}
}

func TestNativeClearTimeout(t *testing.T) {
	t.Parallel()
	fired := false
	loop := NewEventLoop()
	timer := loop.SetTimeout(func(*goja.Runtime) {
		fired = true
	}, 2*time.Second)
	loop.SetTimeout(func(*goja.Runtime) {
		loop.ClearTimeout(timer)
	}, 1*time.Second)
	loop.Run(func(*goja.Runtime) {
		// do not schedule anything
	})
	if fired {
		t.Fatal("Cancelled timer fired!")
	}
}

func TestNativeInterval(t *testing.T) {
	t.Parallel()
	count := 0
	loop := NewEventLoop()
	var i *Interval
	i = loop.SetInterval(func(*goja.Runtime) {
		t.Log("tick")
		count++
		if count > 2 {
			loop.ClearInterval(i)
		}
	}, 1*time.Second)
	loop.Run(func(*goja.Runtime) {
		// do not schedule anything
	})
	if count != 3 {
		t.Fatal("Expected interval to fire 3 times, got", count)
	}
}

func TestNativeClearInterval(t *testing.T) {
	t.Parallel()
	count := 0
	loop := NewEventLoop()
	loop.Run(func(*goja.Runtime) {
		i := loop.SetInterval(func(*goja.Runtime) {
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

func TestSetTimeoutConcurrent(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop()
	loop.Start()
	ch := make(chan struct{}, 1)
	loop.SetTimeout(func(*goja.Runtime) {
		ch <- struct{}{}
	}, 100*time.Millisecond)
	<-ch
	loop.Stop()
}

func TestClearTimeoutConcurrent(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop()
	loop.Start()
	timer := loop.SetTimeout(func(*goja.Runtime) {
	}, 100*time.Millisecond)
	loop.ClearTimeout(timer)
	loop.Stop()
	if c := loop.jobCount; c != 0 {
		t.Fatalf("jobCount: %d", c)
	}
}

func TestClearIntervalConcurrent(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop()
	loop.Start()
	ch := make(chan struct{}, 1)
	i := loop.SetInterval(func(*goja.Runtime) {
		ch <- struct{}{}
	}, 500*time.Millisecond)

	<-ch
	loop.ClearInterval(i)
	loop.Stop()
	if c := loop.jobCount; c != 0 {
		t.Fatalf("jobCount: %d", c)
	}
}

func TestRunOnStoppedLoop(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop()
	var failed int32
	done := make(chan struct{})
	go func() {
		for atomic.LoadInt32(&failed) == 0 {
			loop.Start()
			time.Sleep(10 * time.Millisecond)
			loop.Stop()
		}
	}()
	go func() {
		for atomic.LoadInt32(&failed) == 0 {
			loop.RunOnLoop(func(*goja.Runtime) {
				if !loop.running {
					atomic.StoreInt32(&failed, 1)
					close(done)
					return
				}
			})
			time.Sleep(10 * time.Millisecond)
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
	if atomic.LoadInt32(&failed) != 0 {
		t.Fatal("running job on stopped loop")
	}
}

func TestPromise(t *testing.T) {
	t.Parallel()
	const SCRIPT = `
	let result;
	const p = new Promise((resolve, reject) => {
		setTimeout(() => {resolve("passed")}, 500);
	});
	p.then(value => {
		result = value;
	});
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
		t.Fatal(err)
	}
	loop.Run(func(vm *goja.Runtime) {
		result := vm.Get("result")
		if !result.SameAs(vm.ToValue("passed")) {
			err = fmt.Errorf("unexpected result: %v", result)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPromiseNative(t *testing.T) {
	t.Parallel()
	const SCRIPT = `
	let result;
	p.then(value => {
		result = value;
		done();
	});
	`

	loop := NewEventLoop()
	prg, err := goja.Compile("main.js", SCRIPT, false)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan error)
	loop.Start()
	defer loop.Stop()

	loop.RunOnLoop(func(vm *goja.Runtime) {
		vm.Set("done", func() {
			ch <- nil
		})
		p, resolve, _ := vm.NewPromise()
		vm.Set("p", p)
		_, err = vm.RunProgram(prg)
		if err != nil {
			ch <- err
			return
		}
		go func() {
			time.Sleep(500 * time.Millisecond)
			loop.RunOnLoop(func(*goja.Runtime) {
				resolve("passed")
			})
		}()
	})
	err = <-ch
	if err != nil {
		t.Fatal(err)
	}
	loop.RunOnLoop(func(vm *goja.Runtime) {
		result := vm.Get("result")
		if !result.SameAs(vm.ToValue("passed")) {
			ch <- fmt.Errorf("unexpected result: %v", result)
		} else {
			ch <- nil
		}
	})
	err = <-ch
	if err != nil {
		t.Fatal(err)
	}
}

func TestEventLoop_StopNoWait(t *testing.T) {
	t.Parallel()
	loop := NewEventLoop()
	var ran int32
	loop.Run(func(runtime *goja.Runtime) {
		loop.SetTimeout(func(*goja.Runtime) {
			atomic.StoreInt32(&ran, 1)
		}, 5*time.Second)

		loop.SetTimeout(func(*goja.Runtime) {
			loop.StopNoWait()
		}, 500*time.Millisecond)
	})

	if atomic.LoadInt32(&ran) != 0 {
		t.Fatal("ran != 0")
	}
}
