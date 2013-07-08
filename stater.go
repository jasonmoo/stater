package stater

import (
	"math/rand"
	"sync"
	"time"
)

type (
	Stater interface {
		// stating methods
		Timer(key string, value time.Duration, rate float32)
		Gauge(key string, value interface{}, rate float32)
		Increment(key string, value int, rate float32)

		// stater methods
		Init()
		Shutdown()
	}

	Registry []Stater
)

var (
	Rand            = rand.New(rand.NewSource(time.Now().UnixNano()))
	DefaultRegistry = Registry{}
)

// float 0.0 - 1.0
func CanSample(rate float32) bool {
	return rate == 1.0 || Rand.Float32() < rate
}

func Register(s Stater) {
	DefaultRegistry.Register(s)
}

// helper methods for using the default registry
func Timer(key string, value time.Duration, rate float32) {
	DefaultRegistry.Timer(key, value, rate)
}
func Gauge(key string, value interface{}, rate float32) {
	DefaultRegistry.Gauge(key, value, rate)
}
func Increment(key string, value int, rate float32) {
	DefaultRegistry.Increment(key, value, rate)
}
func Init() {
	DefaultRegistry.Init()
}
func Shutdown() {
	DefaultRegistry.Shutdown()
}

func (r Registry) Register(s ...Stater) {
	DefaultRegistry = append(DefaultRegistry, s...)
}
func (r Registry) Increment(key string, value int, rate float32) {
	for _, s := range r {
		// run each in separate goroutine to ensure staters are not
		// affected by eachother's length of run (all should report
		// at roughly the same time)
		go s.Increment(key, value, rate)
	}
}
func (r Registry) Gauge(key string, value interface{}, rate float32) {
	for _, s := range r {
		go s.Gauge(key, value, rate)
	}
}
func (r Registry) Timer(key string, value time.Duration, rate float32) {
	for _, s := range r {
		go s.Timer(key, value, rate)
	}
}

func (r Registry) Init() {
	wg := new(sync.WaitGroup)
	wg.Add(len(r))
	for _, s := range r {
		go func() {
			s.Init()
			wg.Done()
		}()
	}
	wg.Wait()
}
func (r Registry) Shutdown() {
	wg := new(sync.WaitGroup)
	wg.Add(len(r))
	for _, s := range r {
		go func() {
			s.Shutdown()
			wg.Done()
		}()
	}
	wg.Wait()
}
