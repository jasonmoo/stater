package stater

import "time"

type DevNull struct{}

// dev null stater for discarding stats
func (d *DevNull) Timer(key string, value time.Duration, rate float32) {}
func (d *DevNull) Gauge(key string, value interface{}, rate float32)   {}
func (d *DevNull) Increment(key string, value int, rate float32)       {}
func (d *DevNull) Init()                                               {}
func (d *DevNull) Shutdown()                                           {}
