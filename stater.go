package stater

import (
	"math/rand"
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

	Registry map[*Stater]bool
)

var (
	Rand            = rand.New(rand.NewSource(time.Now().UnixNano()))
	DefaultRegistry = make(Registry)
)

// float 0.0 - 1.0
func CanSample(rate float32) bool {
	return Rand.Float32() < rate
}

func Register(s *Stater) {
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

func (r Registry) Register(s *Stater) {
	DefaultRegistry[s] = true
}
func (r Registry) Timer(key string, value time.Duration, rate float32) {
	for s, active := range r {
		if active {
			// run each in separate goroutine to ensure each stater reports at roughly the same time
			go (*s).Timer(key, value, rate)
		}
	}
}
func (r Registry) Gauge(key string, value interface{}, rate float32) {
	for s, active := range r {
		if active {
			// run each in separate goroutine to ensure each stater reports at roughly the same time
			go (*s).Gauge(key, value, rate)
		}
	}
}
func (r Registry) Increment(key string, value int, rate float32) {
	for s, active := range r {
		if active {
			// run each in separate goroutine to ensure each stater reports at roughly the same time
			go (*s).Increment(key, value, rate)
		}
	}
}
func (r Registry) Init() {
	for s, active := range r {
		if active {
			// run each in separate goroutine to ensure each stater reports at roughly the same time
			go (*s).Init()
		}
	}
}
func (r Registry) Shutdown() {
	for s, active := range r {
		if active {
			// run each in separate goroutine to ensure each stater reports at roughly the same time
			go (*s).Shutdown()
		}
	}
}
