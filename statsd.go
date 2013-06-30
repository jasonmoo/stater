package stater

import (
	"fmt"
	"log"
	"net"
	"time"
)

type Statsd struct {
	active bool
	stats  chan string

	Addr       string
	BufferSize int
}

const crlf = "\r\n"

var StatTimeout = time.Millisecond * 100

func (d *Statsd) stat(format string, args ...interface{}) {
	if d.active {
		select {
		// write to channel or discard if timeout
		case d.stats <- fmt.Sprintf(format, args...):
		// todo:  notify someone we're discarding stats
		case <-time.After(StatTimeout):
		}
	}
}

func (d *Statsd) Timer(key string, value time.Duration, rate float32) {
	if CanSample(rate) {
		d.stat("%s:%d|t|@%f%s", key, value/time.Nanosecond, rate, crlf)
	}
}
func (d *Statsd) Gauge(key string, value interface{}, rate float32) {
	if CanSample(rate) {
		d.stat("%s:%v|g|@%f%s", key, value, rate, crlf)
	}
}
func (d *Statsd) Increment(key string, value int, rate float32) {
	if CanSample(rate) {
		d.stat("%s:%d|c|@%f%s", key, value, rate, crlf)
	}
}
func (d *Statsd) Init() {

	d.stats = make(chan string, d.BufferSize)
	done := make(chan bool)

	go func() {
		for /*ever*/ {
			go func() {
				d.active = true
				defer func() {
					if r := recover(); r != nil {
						log.Println("stat processor goroutine panicked!", r)
					}

					d.active = false
					done <- true
				}()

				conn, err := net.Dial("udp", d.Addr)
				if err != nil {
					log.Printf("Unable to connect to statsd server: %v", err)
					return // this will restart the goroutine
				}
				defer conn.Close()

				for stat := range d.stats {

					buf := []byte(stat)
					if n, err := conn.Write(buf); n != len(buf) || err != nil {
						log.Printf("Failed to write to statsd connection: %v", err)
						return // this will restart the goroutine
					}

				}
			}()
			// retry connect after a bit
			<-done
			<-time.After(time.Second)
		}
	}()
}
func (d *Statsd) Shutdown() {
	d.active = false
	close(d.stats) // exits the goroutine
}
