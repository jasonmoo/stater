package stater

import (
	"bytes"
	"net"
	"testing"
	"time"
)

type (
	UDPTestServer struct {
		Addr  string // address to listen on
		Value []byte // last Value sent to the server

		Received bool
		Ready    chan bool
		active   bool
		t        *testing.T
	}
)

func NewTestPair(addr string, t *testing.T) (client *Statsd, server *UDPTestServer) {
	client, server = &Statsd{Addr: addr}, &UDPTestServer{Addr: addr, t: t}
	server.Init()
	client.Init()
	return
}

func TestIncrement(t *testing.T) {

	client, server := NewTestPair(":11311", t)
	defer server.Shutdown()
	defer client.Shutdown()

	server.Wait()
	client.Increment("test.inc", 1, 1.0)
	server.Assert("test.inc:1|c\n")

	server.Wait()
	client.Increment("test.inc1", -1, 1.0)
	server.Assert("test.inc1:-1|c\n")

	server.Wait()
	client.Increment("test.inc2", -5, 1.0)
	server.Assert("test.inc2:-5|c\n")

	server.Wait()
	client.Increment("test.inc3", 100, 1.0)
	server.Assert("test.inc3:100|c\n")

	// taking our chances this will run
	server.Wait()
	client.Increment("test.inc4", 1, 0.999999)
	server.Assert("test.inc4:1|c|@0.999999\n")

}

func TestGauge(t *testing.T) {

	client, server := NewTestPair(":11312", t)
	defer server.Shutdown()
	defer client.Shutdown()

	server.Wait()
	client.Gauge("test.gauge", 1, 1.0)
	server.Assert("test.gauge:1|g\n")

	// taking our chances this will run
	server.Wait()
	client.Gauge("test.gauge1", 2, 0.999999)
	server.Assert("test.gauge1:2|g|@0.999999\n")

}

func TestTimer(t *testing.T) {

	client, server := NewTestPair(":11313", t)
	defer server.Shutdown()
	defer client.Shutdown()

	server.Wait()
	client.Timer("test.timer", time.Millisecond, 1.0)
	server.Assert("test.timer:1|ms\n")

	server.Wait()
	client.Timer("test.timer1", 10*time.Millisecond, 1.0)
	server.Assert("test.timer1:10|ms\n")

	server.Wait()
	client.Timer("test.timer2", 100*time.Millisecond, 1.0)
	server.Assert("test.timer2:100|ms\n")

	server.Wait()
	client.Timer("test.timer3", 1*time.Second, 1.0)
	server.Assert("test.timer3:1000|ms\n")

	// taking our chances this will run
	server.Wait()
	client.Timer("test.timer4", time.Millisecond, 0.999)
	server.Assert("test.timer4:1|ms|@0.999000\n")

}

func (s *UDPTestServer) Init() {

	addr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		s.t.Fatal(err)
	}

	s.Ready = make(chan bool)

	go func() {
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			s.t.Fatal(err)
		}
		defer conn.Close()

		s.active = true

		for s.active {

			buf := make([]byte, 1024)

			s.Received = false
			s.Ready <- true // to receive a packet

			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				s.t.Fatal(err)
			}

			s.Received = true
			s.Value = buf[:n]

			s.Ready <- true // to evaluate the sent stat

			<-s.Ready // wait for us to signal that we're done asserting
		}

	}()
}
func (s *UDPTestServer) Wait() {
	// to prevent deadlock, set a sensible timeout
	select {
	case <-s.Ready:
		//log.Println("ready!")
	case <-time.After(time.Millisecond * 100):
		s.t.Errorf("Timeout waiting for UDPTestServer.Ready")
	}
}
func (s *UDPTestServer) Assert(expected string) {
	// to prevent deadlock, set a sensible timeout
	select {
	case <-s.Ready:
		//log.Println("ready!")
	case <-time.After(time.Millisecond * 100):
		s.t.Errorf("Timeout waiting for UDPTestServer.Ready")
	}
	if !s.Received {
		//log.Printf("%#v\n", s)
		s.t.Errorf("Expected: %q, did not receive any value via conn\n", expected)
	} else if !bytes.Equal(s.Value, []byte(expected)) {
		s.t.Errorf("Expected: %q, Got: %q\n", expected, s.Value)
	}
	s.Ready <- true
}
func (s *UDPTestServer) Shutdown() {
	s.active = false
}

/*
func BenchmarkExample(b *testing.B) {
	a, aa := 1, 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a, aa = aa, a
	}
}
*/
