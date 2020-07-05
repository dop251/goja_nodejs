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
	running  bool

	immediateJobs     []func()
	immediateJobsLock sync.Mutex
	wakeup            chan struct{}

	enableConsole bool
}

func NewEventLoop(opts ...Option) *EventLoop {
	vm := goja.New()

	loop := &EventLoop{
		vm:            vm,
		jobChan:       make(chan func()),
		wakeup:        make(chan struct{}, 1),
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

func (loop *EventLoop) SetTimeout(fn func(*goja.Runtime), timeout time.Duration) *Timer {
	t := loop.addTimeout(func() { fn(loop.vm) }, timeout)
	loop.addImmediateJob(func() {
		loop.jobCount++
	})
	return t
}

func (loop *EventLoop) ClearTimeout(t *Timer) {
	loop.addImmediateJob(func() {
		loop.clearTimeout(t)
	})
}

func (loop *EventLoop) SetInterval(fn func(*goja.Runtime), timeout time.Duration) *Interval {
	i := loop.addInterval(func() { fn(loop.vm) }, timeout)
	loop.addImmediateJob(func() {
		loop.jobCount++
	})
	return i
}

func (loop *EventLoop) ClearInterval(i *Interval) {
	loop.addImmediateJob(func() {
		loop.clearInterval(i)
	})
}

// Run calls the specified function, starts the event loop and waits until there are no more delayed jobs to run
// after which it stops the loop and returns.
// The instance of goja.Runtime that is passed to the function and any Values derived from it must not be used outside
// of the function.
// Do NOT use this function while the loop is already running. Use RunOnLoop() instead.
func (loop *EventLoop) Run(fn func(*goja.Runtime)) {
	fn(loop.vm)
	loop.run(false)
}

// Start the event loop in the background. The loop continues to run until Stop() is called.
func (loop *EventLoop) Start() {
	go loop.run(true)
}

// Stop the loop that was started with Start(). After this function returns there will be no more jobs executed
// by the loop. It is possible to call Start() or Run() again after this to resume the execution.
// Note, it does not cancel active timeouts.
func (loop *EventLoop) Stop() {
	ch := make(chan struct{})

	loop.jobChan <- func() {
		loop.running = false
		ch <- struct{}{}
	}

	<-ch
}

// RunOnLoop schedules to run the specified function in the context of the loop as soon as possible.
// The order of the runs is preserved (i.e. the functions will be called in the same order as calls to RunOnLoop())
// The instance of goja.Runtime that is passed to the function and any Values derived from it must not be used outside
// of the function.
func (loop *EventLoop) RunOnLoop(fn func(*goja.Runtime)) {
	loop.addImmediateJob(func() { fn(loop.vm) })
}

func (loop *EventLoop) run(inBackground bool) {
	loop.running = true
L:
	for {
		for {
			loop.immediateJobsLock.Lock()
			jobs := loop.immediateJobs
			loop.immediateJobs = nil
			loop.immediateJobsLock.Unlock()
			if len(jobs) == 0 {
				break
			}
			for _, job := range jobs {
				job()
			}
		}
		if !loop.running || (!inBackground && loop.jobCount <= 0) {
			break
		}
		select {
		case job, ok := <-loop.jobChan:
			if !ok {
				break L
			}
			job()
		case <-loop.wakeup:
			break
		}
	}
}

func (loop *EventLoop) addImmediateJob(fn func()) {
	loop.immediateJobsLock.Lock()
	loop.immediateJobs = append(loop.immediateJobs, fn)
	loop.immediateJobsLock.Unlock()
	select {
	case loop.wakeup <- struct{}{}:
	default:
	}
}

func (loop *EventLoop) addTimeout(f func(), timeout time.Duration) *Timer {
	t := &Timer{
		job: job{
			fn: f,
		},
	}
	t.timer = time.AfterFunc(timeout, func() {
		loop.jobChan <- func() {
			loop.doTimeout(t)
		}
	})

	return t
}

func (loop *EventLoop) addInterval(f func(), timeout time.Duration) *Interval {
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
	if !t.cancelled {
		t.timer.Stop()
		t.cancelled = true
		loop.jobCount--
	}
}

func (loop *EventLoop) clearInterval(i *Interval) {
	if !i.cancelled {
		i.cancelled = true
		close(i.stopChan)
		loop.jobCount--
	}
}

func (i *Interval) run(loop *EventLoop) {
	for {
		select {
		case <-i.stopChan:
			i.ticker.Stop()
			break
		case <-i.ticker.C:
			loop.jobChan <- func() {
				loop.doInterval(i)
			}
		}
	}
}
