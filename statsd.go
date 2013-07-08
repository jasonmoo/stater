package stater

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"time"
)

type Statsd struct {
	Addr              string
	BufferSize        int
	ReconnectInterval time.Duration

	active    uint32
	addr      *net.UDPAddr
	conn      *net.UDPConn
	reconnect *Governor
	stats     chan string
}

const defaultReconnectInterval = time.Second

func (d *Statsd) Timer(key string, value time.Duration, rate float32) {
	d.sendf("%s:%d|ms", key, value/time.Millisecond, rate)
}
func (d *Statsd) Gauge(key string, value interface{}, rate float32) {
	d.sendf("%s:%v|g", key, value, rate)
}
func (d *Statsd) Increment(key string, value int, rate float32) {
	d.sendf("%s:%d|c", key, value, rate)
}

func (d *Statsd) setActiveFlag(active int) {
	atomic.StoreUint32(&d.active, uint32(active))
}

func (d *Statsd) Init() {

	var err error

	d.addr, err = net.ResolveUDPAddr("udp", d.Addr)
	if err != nil {
		log.Fatal("unable to resolve udp address", err)
	}

	d.conn, err = net.DialUDP("udp", nil, d.addr)
	if err != nil {
		log.Fatal("unable to connect to statsd addr", err)
	}

	// set up a reconnection function that tries to
	// reconnect to the statsd server at-most once
	// per interval
	d.reconnect = NewGovernor(d.ReconnectInterval, func() {
		d.Shutdown()
		d.Init()
	})

	d.setActiveFlag(1)
}

func (d *Statsd) Shutdown() {
	d.setActiveFlag(0)
	if d.conn != nil {
		d.conn.Close()
	}
}

func (d *Statsd) Reconnect() {
	// governed to once per interval
	d.reconnect.Do()
}

func (d *Statsd) sendf(format string, key string, value interface{}, rate float32) {

	if atomic.LoadUint32(&d.active) == 1 && CanSample(rate) {

		stat := bytes.NewBuffer(make([]byte, 0, 32))

		if rate != 1.0 {
			fmt.Fprintf(stat, format+"|@%f\n", key, value, rate)
		} else {
			fmt.Fprintf(stat, format+"\n", key, value)
		}

		//log.Println("stat to send", stat.String())

		n, err := d.conn.Write(stat.Bytes())

		if n != stat.Len() || err != nil {
			log.Printf("Failed to write to statsd connection: %v, %q\n", err, stat)
			// don't block this stat send
			go d.Reconnect()
		}

	}

}
