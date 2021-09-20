package eventloop

import (
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
)

type job struct {
	cancelled bool
	fn        func()
}

type Timer struct {
	job
	timer *time.Timer
}

type Interval struct {
	job
	ticker   *time.Ticker
	stopChan chan struct{}
}

type EventLoop struct {
	vm       *goja.Runtime
	jobChan  chan func()
	jobCount int32
	canRun   bool

	auxJobs     []func()
	auxJobsLock sync.Mutex
	wakeup      chan struct{}

	stopCond *sync.Cond
	running  bool

	enableConsole bool
}

func NewEventLoop(opts ...Option) *EventLoop {
	vm := goja.New()

	loop := &EventLoop{
		vm:            vm,
		jobChan:       make(chan func()),
		wakeup:        make(chan struct{}, 1),
		stopCond:      sync.NewCond(&sync.Mutex{}),
		enableConsole: true,
	}

	for _, opt := range opts {
		opt(loop)
	}

	new(require.Registry).Enable(vm)
	if loop.enableConsole {
		console.Enable(vm)
	}
	vm.Set("setTimeout", loop.setTimeout)
	vm.Set("setInterval", loop.setInterval)
	vm.Set("clearTimeout", loop.clearTimeout)
	vm.Set("clearInterval", loop.clearInterval)

	return loop
}

type Option func(*EventLoop)

// EnableConsole controls whether the "console" module is loaded into
// the runtime used by the loop.  By default, loops are created with
// the "console" module loaded, pass EnableConsole(false) to
// NewEventLoop to disable this behavior.
func EnableConsole(enableConsole bool) Option {
	return func(loop *EventLoop) {
		loop.enableConsole = enableConsole
	}
}

func (loop *EventLoop) schedule(call goja.FunctionCall, repeating bool) goja.Value {
	if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
		delay := call.Argument(1).ToInteger()
		var args []goja.Value
		if len(call.Arguments) > 2 {
			args = call.Arguments[2:]
		}
		f := func() { fn(nil, args...) }
		loop.jobCount++
		if repeating {
			return loop.vm.ToValue(loop.addInterval(f, time.Duration(delay)*time.Millisecond))
		} else {
			return loop.vm.ToValue(loop.addTimeout(f, time.Duration(delay)*time.Millisecond))
		}
	}
	return nil
}

func (loop *EventLoop) setTimeout(call goja.FunctionCall) goja.Value {
	return loop.schedule(call, false)
}

func (loop *EventLoop) setInterval(call goja.FunctionCall) goja.Value {
	return loop.schedule(call, true)
}

// SetTimeout schedules to run the specified function in the context
// of the loop as soon as possible after the specified timeout period.
// SetTimeout returns a Timer which can be passed to ClearTimeout.
// The instance of goja.Runtime that is passed to the function and any Values derived
// from it must not be used outside of the function. SetTimeout is
// safe to call inside or outside of the loop.
func (loop *EventLoop) SetTimeout(fn func(*goja.Runtime), timeout time.Duration) *Timer {
	t := loop.addTimeout(func() { fn(loop.vm) }, timeout)
	loop.addAuxJob(func() {
		loop.jobCount++
	})
	return t
}

// ClearTimeout cancels a Timer returned by SetTimeout if it has not run yet.
// ClearTimeout is safe to call inside or outside of the loop.
func (loop *EventLoop) ClearTimeout(t *Timer) {
	loop.addAuxJob(func() {
		loop.clearTimeout(t)
	})
}

// SetInterval schedules to repeatedly run the specified function in
// the context of the loop as soon as possible after every specified
// timeout period.  SetInterval returns an Interval which can be
// passed to ClearInterval. The instance of goja.Runtime that is passed to the
// function and any Values derived from it must not be used outside of
// the function. SetInterval is safe to call inside or outside of the
// loop.
func (loop *EventLoop) SetInterval(fn func(*goja.Runtime), timeout time.Duration) *Interval {
	i := loop.addInterval(func() { fn(loop.vm) }, timeout)
	loop.addAuxJob(func() {
		loop.jobCount++
	})
	return i
}

// ClearInterval cancels an Interval returned by SetInterval.
// ClearInterval is safe to call inside or outside of the loop.
func (loop *EventLoop) ClearInterval(i *Interval) {
	loop.addAuxJob(func() {
		loop.clearInterval(i)
	})
}

func (loop *EventLoop) setRunning() {
	loop.stopCond.L.Lock()
	if loop.running {
		panic("Loop is already started")
	}
	loop.running = true
	loop.stopCond.L.Unlock()
}

// Run calls the specified function, starts the event loop and waits until there are no more delayed jobs to run
// after which it stops the loop and returns.
// The instance of goja.Runtime that is passed to the function and any Values derived from it must not be used outside
// of the function.
// Do NOT use this function while the loop is already running. Use RunOnLoop() instead.
// If the loop is already started it will panic.
func (loop *EventLoop) Run(fn func(*goja.Runtime)) {
	loop.setRunning()
	fn(loop.vm)
	loop.run(false)
}

// Start the event loop in the background. The loop continues to run until Stop() is called.
// If the loop is already started it will panic.
func (loop *EventLoop) Start() {
	loop.setRunning()
	go loop.run(true)
}

// Stop the loop that was started with Start(). After this function returns there will be no more jobs executed
// by the loop. It is possible to call Start() or Run() again after this to resume the execution.
// Note, it does not cancel active timeouts.
// It is not allowed to run Start() and Stop() concurrently.
// Calling Stop() on an already stopped loop or inside the loop will hang.
func (loop *EventLoop) Stop() {
	loop.jobChan <- func() {
		loop.canRun = false
	}

	loop.stopCond.L.Lock()
	for loop.running {
		loop.stopCond.Wait()
	}
	loop.stopCond.L.Unlock()
}

// RunOnLoop schedules to run the specified function in the context of the loop as soon as possible.
// The order of the runs is preserved (i.e. the functions will be called in the same order as calls to RunOnLoop())
// The instance of goja.Runtime that is passed to the function and any Values derived from it must not be used outside
// of the function. It is safe to call inside or outside of the loop.
func (loop *EventLoop) RunOnLoop(fn func(*goja.Runtime)) {
	loop.addAuxJob(func() { fn(loop.vm) })
}

func (loop *EventLoop) runAux() {
	loop.auxJobsLock.Lock()
	jobs := loop.auxJobs
	loop.auxJobs = nil
	loop.auxJobsLock.Unlock()
	for _, job := range jobs {
		job()
	}
}

func (loop *EventLoop) run(inBackground bool) {
	loop.canRun = true
	loop.runAux()

	for loop.canRun && (inBackground || loop.jobCount > 0) {
		select {
		case job := <-loop.jobChan:
			job()
			if loop.canRun {
				select {
				case <-loop.wakeup:
					loop.runAux()
				default:
				}
			}
		case <-loop.wakeup:
			loop.runAux()
		}
	}
	loop.stopCond.L.Lock()
	loop.running = false
	loop.stopCond.L.Unlock()
	loop.stopCond.Broadcast()
}

func (loop *EventLoop) addAuxJob(fn func()) {
	loop.auxJobsLock.Lock()
	loop.auxJobs = append(loop.auxJobs, fn)
	loop.auxJobsLock.Unlock()
	select {
	case loop.wakeup <- struct{}{}:
	default:
	}
}

func (loop *EventLoop) addTimeout(f func(), timeout time.Duration) *Timer {
	t := &Timer{
		job: job{fn: f},
	}
	t.timer = time.AfterFunc(timeout, func() {
		loop.jobChan <- func() {
			loop.doTimeout(t)
		}
	})

	return t
}

func (loop *EventLoop) addInterval(f func(), timeout time.Duration) *Interval {
	// https://nodejs.org/api/timers.html#timers_setinterval_callback_delay_args
	if timeout <= 0 {
		timeout = time.Millisecond
	}

	i := &Interval{
		job:      job{fn: f},
		ticker:   time.NewTicker(timeout),
		stopChan: make(chan struct{}),
	}

	go i.run(loop)
	return i
}

func (loop *EventLoop) doTimeout(t *Timer) {
	if !t.cancelled {
		t.fn()
		t.cancelled = true
		loop.jobCount--
	}
}

func (loop *EventLoop) doInterval(i *Interval) {
	if !i.cancelled {
		i.fn()
	}
}

func (loop *EventLoop) clearTimeout(t *Timer) {
	if t != nil && !t.cancelled {
		t.timer.Stop()
		t.cancelled = true
		loop.jobCount--
	}
}

func (loop *EventLoop) clearInterval(i *Interval) {
	if i != nil && !i.cancelled {
		i.cancelled = true
		close(i.stopChan)
		loop.jobCount--
	}
}

func (i *Interval) run(loop *EventLoop) {
L:
	for {
		select {
		case <-i.stopChan:
			i.ticker.Stop()
			break L
		case <-i.ticker.C:
			loop.jobChan <- func() {
				loop.doInterval(i)
			}
		}
	}
}
