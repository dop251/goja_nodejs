package timer

import (
	"time"

	"sync"

	"github.com/dop251/goja"
)

type (
	Registry struct {
		mu     sync.Mutex
		queue  chan *timer
		vm     *goja.Runtime
		timers map[*timer]*timer
	}

	timer struct {
		timer    *time.Timer
		duration time.Duration
		interval bool
		call     goja.Callable
	}
)

func NewRegistry() (r *Registry) {
	r = &Registry{
		queue:  make(chan *timer),
		timers: map[*timer]*timer{},
	}
	return
}

func (r *Registry) newTimer(call goja.Callable, delay int64, interval bool) *timer {
	t := &timer{
		call:     call,
		duration: time.Duration(delay) * time.Millisecond,
		interval: interval,
	}
	r.mu.Lock()
	r.timers[t] = t
	r.mu.Unlock()

	t.timer = time.AfterFunc(t.duration, func() {
		r.queue <- t
	})

	return t
}

func (r *Registry) setTimeout(c goja.FunctionCall) goja.Value {
	call, ok := goja.AssertFunction(c.Argument(0))

	if !ok {
		panic("argument 1 is not function")
	}

	delay := c.Argument(1).ToInteger()

	return r.vm.ToValue(r.newTimer(call, delay, false))
}

func (r *Registry) setInterval(c goja.FunctionCall) goja.Value {
	call, ok := goja.AssertFunction(c.Argument(0))

	if !ok {
		panic("argument 1 is not function")
	}

	delay := c.Argument(1).ToInteger()

	return r.vm.ToValue(r.newTimer(call, delay, true))
}

func (r *Registry) clearTimer(t *timer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	timer, ok := r.timers[t]
	if !ok {
		return
	}

	timer.timer.Stop()
	delete(r.timers, timer)
	tl := len(r.timers)
	if tl != 0 {
		return
	}
}

func (r *Registry) clearTimeout(c goja.FunctionCall) goja.Value {
	t, ok := c.Argument(0).Export().(*timer)
	if ok {
		r.clearTimer(t)
	}

	return goja.Undefined()
}

func (r *Registry) Enable(vm *goja.Runtime) {
	r.vm = vm
	vm.Set("setTimeout", r.setTimeout)
	vm.Set("setInterval", r.setInterval)
	vm.Set("clearTimeout", r.clearTimeout)
	vm.Set("clearInterval", r.clearTimeout)
}

func (r *Registry) Wait() {
	tl := func() (i int) {
		r.mu.Lock()
		i = len(r.timers)
		r.mu.Unlock()
		return
	}

	if tl() <= 0 {
		return
	}

	for t := range r.queue {
		_, err := t.call(nil)
		if err != nil {
			return
		}

		if t.interval {
			t.timer.Reset(t.duration)
		} else {
			r.clearTimer(t)
		}

		if tl() <= 0 {
			return
		}
	}

	return
}
