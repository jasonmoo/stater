package stater

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Governor struct {
	sync.Mutex

	blocking   bool
	ready      uint32
	ready_cond *sync.Cond
	do         func()
	interval   time.Duration
}

func NewGovernor(interval time.Duration, f func()) (g *Governor) {
	g = &Governor{
		ready:      1,
		interval:   interval,
		ready_cond: new(sync.Cond),
	}
	g.SetFunc(f)
	return
}

func (g *Governor) SetInterval(interval time.Duration) {
	g.Lock()
	defer g.Unlock()
	g.interval = interval
}
func (g *Governor) SetBlocking(blocking bool) {
	g.Lock()
	defer g.Unlock()
	g.blocking = blocking
}

func (g *Governor) SetFunc(f func()) {
	g.Lock()
	defer g.Unlock()
	g.do = func() {
		// quick check
		if atomic.LoadUint32(&g.ready) == 0 {
			if g.blocking {
				g.ready_cond.Wait()
			}
			return
		}

		// slower blocking
		g.Lock()
		defer g.Unlock()

		if g.ready == 1 {
			// set the quick flag
			atomic.StoreUint32(&g.ready, 0)

			start := time.Now()
			defer func() {
				if r := recover(); r != nil {
					log.Println("governor do panic'd!", r)
				}
				// after our wait, just reset the ready state
				if remainder := g.interval - time.Since(start); remainder > 0 {
					<-time.After(remainder)
				}
				g.Ready()
			}()

			f()
		}
	}
}

func (g *Governor) Do() {
	g.do()
}

// calling this subverts the interval and causes
// the next attempted Do() to execute the supplied function
func (g *Governor) Ready() {
	atomic.StoreUint32(&g.ready, 1)
	g.ready_cond.Broadcast()
}
