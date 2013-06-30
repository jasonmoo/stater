package stater

import (
	"fmt"
	"io"
	"time"
)

type Debug struct {
	io.Writer
}

func (d *Debug) Timer(key string, value time.Duration, rate float32) {
	if CanSample(rate) {
		fmt.Fprintf(d, "Timer: %v, %s, %f\n", key, value, rate)
	}
}
func (d *Debug) Gauge(key string, value interface{}, rate float32) {
	if CanSample(rate) {
		fmt.Fprintf(d, "Gauge: %v, %v, %f\n", key, value, rate)
	}
}
func (d *Debug) Increment(key string, value int, rate float32) {
	if CanSample(rate) {
		fmt.Fprintf(d, "Increment: %v, %d, %f\n", key, value, rate)
	}
}
func (d *Debug) Init() {
	fmt.Fprintln(d, "Stater Init")
}
func (d *Debug) Shutdown() {
	fmt.Fprintln(d, "Stater Shutdown")
}
