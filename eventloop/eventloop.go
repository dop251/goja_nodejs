package eventloop

import (
	"container/heap"
	"math"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
)

type job struct {
	added     time.Time
	at        time.Time
	interval  time.Duration
	fn        func()
	cancelled bool

	// index of job in jobHeap
	index int
}

type Timer struct {
	job
}

type Interval struct {
	job
}

type EventLoop struct {
	vm            *goja.Runtime
	jobsMu        *sync.Mutex
	jobs          jobHeap
	kickChan      chan struct{}
	quitChan      chan struct{}
	enableConsole bool
}

func NewEventLoop(opts ...Option) *EventLoop {
	vm := goja.New()

	loop := &EventLoop{
		vm:            vm,
		jobsMu:        &sync.Mutex{},
		jobs:          jobHeap{},
		kickChan:      make(chan struct{}, 256),
		enableConsole: true,
	}

	heap.Init(&loop.jobs)

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
		if repeating {
			return loop.vm.ToValue(loop.addInterval(func() { fn(nil, args...) }, time.Duration(delay)*time.Millisecond, false))
		}
		return loop.vm.ToValue(loop.addTimeout(func() { fn(nil, args...) }, time.Duration(delay)*time.Millisecond, false))
	}
	return nil
}

func (loop *EventLoop) setTimeout(call goja.FunctionCall) goja.Value {
	return loop.schedule(call, false)
}

func (loop *EventLoop) setInterval(call goja.FunctionCall) goja.Value {
	return loop.schedule(call, true)
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
	loop.quitChan = make(chan struct{})
	go loop.run(true)
}

// Stop the loop that was started with Start(). After this function returns there will be no more jobs executed
// by the loop. It is possible to call Start() or Run() again after this to resume the execution.
// Note, it does not cancel active timeouts.
func (loop *EventLoop) Stop() {
	loop.quitChan <- struct{}{}
}

// RunOnLoop schedules to run the specified function in the context of the loop as soon as possible.
// The order of the runs is preserved (i.e. the functions will be called in the same order as calls to RunOnLoop())
// The instance of goja.Runtime that is passed to the function and any Values derived from it must not be used outside
// of the function.  RunOnLoop is equivalent to SetTimeout(fn, 0) and is safe to call inside or outside of the loop.
func (loop *EventLoop) RunOnLoop(fn func(*goja.Runtime)) {
	loop.SetTimeout(fn, 0)
}

// SetTimeout schedules to run the specified function in the context
// of the loop as soon as possible after the specified timeout period.
// SetTimeout returns a Timer which can be passed to ClearTimeout.
// The order of the runs is preserved (i.e. the functions will be
// called in the same order as calls to RunOnLoop()) The instance of
// goja.Runtime that is passed to the function and any Values derived
// from it must not be used outside of the function.  SetTimeout is
// safe to call inside or outside of the loop.
func (loop *EventLoop) SetTimeout(fn func(*goja.Runtime), timeout time.Duration) *Timer {
	return loop.addTimeout(func() { fn(loop.vm) }, timeout, true)
}

// ClearTimeout cancels a Timer returned by SetTimeout.  ClearTimeout
// is safe to call inside or outside of the loop.
func (loop *EventLoop) ClearTimeout(t *Timer) {
	loop.clearTimeout(t)
}

// SetInterval schedules to repeatedly run the specified function in
// the context of the loop as soon as possible after every specified
// timeout period.  SetInterval returns an Interval which can be
// passed to ClearInterval.  The order of the runs is preserved
// (i.e. the functions will be called in the same order as calls to
// RunOnLoop()) The instance of goja.Runtime that is passed to the
// function and any Values derived from it must not be used outside of
// the function.  SetInterval is safe to call inside or outside of the
// loop.
func (loop *EventLoop) SetInterval(fn func(*goja.Runtime), timeout time.Duration) *Interval {
	return loop.addInterval(func() { fn(loop.vm) }, timeout, true)
}

// ClearInterval cancels an Interval returned by SetInterval.
// ClearInterval is safe to call inside or outside of the loop.
func (loop *EventLoop) ClearInterval(i *Interval) {
	loop.clearInterval(i)
}

func (loop *EventLoop) addTimeout(fn func(), timeout time.Duration, kick bool) *Timer {
	t := &Timer{
		job: job{
			at: loop.now().Add(timeout),
			fn: fn,
		},
	}
	loop.pushJob(&t.job, kick)
	return t
}

func (loop *EventLoop) addInterval(fn func(), interval time.Duration, kick bool) *Interval {
	i := &Interval{
		job: job{
			at:       loop.now().Add(interval),
			interval: interval,
			fn:       fn,
		},
	}
	loop.pushJob(&i.job, kick)
	return i
}

func (loop *EventLoop) clearTimeout(t *Timer) {
	t.cancelled = true
}

func (loop *EventLoop) clearInterval(i *Interval) {
	i.cancelled = true
}

func (loop *EventLoop) run(inBackground bool) {
	loop.drainKicks()

	// math.MaxInt64 = time.maxDuration = ~290 years
	nearestJobTimer := time.NewTimer(math.MaxInt64)

	for {
		nearestJob := loop.nearestJob(nearestJobTimer)
		if nearestJob == nil && !inBackground {
			return
		}

		select {
		case <-nearestJob:
			loop.runNext()
		case <-loop.kickChan:
			loop.drainKicks()
		case <-loop.quitChan:
			return
		}
	}

}

func (loop *EventLoop) nearestJob(nearestJobTimer *time.Timer) <-chan time.Time {
	nearestJobTimer.Stop()
	j := loop.peekJob()
	if j == nil {
		return nil
	}
	d := j.at.Sub(loop.now())
	if d < 0 {
		// we're behind schedule, run immediately
		d = 0
	}
	nearestJobTimer.Reset(d)
	return nearestJobTimer.C
}

func (loop *EventLoop) runNext() {
	j := loop.popJob()
	if j == nil {
		return
	}
	if j.interval > 0 {
		j.at = loop.now().Add(j.interval)
		loop.pushJob(j, false)
	}
	j.fn()
}

func (loop *EventLoop) drainKicks() {
	for {
		select {
		case <-loop.kickChan:
		default:
			return
		}
	}
}

func (loop *EventLoop) pushJob(j *job, kick bool) *job {
	j.added = loop.now()
	loop.jobsMu.Lock()
	heap.Push(&loop.jobs, j)
	loop.jobsMu.Unlock()
	if kick {
		go func() { loop.kickChan <- struct{}{} }()
	}
	return j
}

func (loop *EventLoop) popJob() *job {
	loop.jobsMu.Lock()
	defer loop.jobsMu.Unlock()
	for {
		j := loop.jobs.peek()
		if j == nil || j.at.After(loop.now()) {
			return nil
		}
		heap.Pop(&loop.jobs)
		if !j.cancelled {
			return j
		}
	}
}

func (loop *EventLoop) peekJob() *job {
	loop.jobsMu.Lock()
	defer loop.jobsMu.Unlock()
	return loop.jobs.peek()
}

func (loop *EventLoop) now() time.Time {
	return time.Now()
}

type jobHeap []*job

func (js jobHeap) Len() int { return len(js) }

func (js jobHeap) Less(i, j int) bool {
	ii, jj := js[i], js[j]
	if ii.at.Before(jj.at) {
		return true
	}
	if ii.at.After(jj.at) {
		return false
	}
	return ii.added.Before(jj.added)
}

func (js jobHeap) Swap(i, j int) {
	js[i], js[j] = js[j], js[i]
	js[i].index = i
	js[j].index = j
}

func (js *jobHeap) Push(x interface{}) {
	n := len(*js)
	j := x.(*job)
	j.index = n
	*js = append(*js, j)
}

func (js *jobHeap) Pop() interface{} {
	old := *js
	n := len(old)
	j := old[n-1]
	old[n-1] = nil // avoid memory leak
	j.index = -1   // for safety
	*js = old[0 : n-1]
	return j
}

func (js *jobHeap) peek() *job {
	n := len(*js)
	if n == 0 {
		return nil
	}
	return (*js)[0]
}
