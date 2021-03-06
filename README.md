#stater

Stater is a simple lib for managing stat clients in go.

A simple interface defines what a Stater must do.

	Stater interface {

		// stating methods
		Timer(key string, value time.Duration, rate float64)
		Gauge(key string, value interface{}, rate float64)
		Increment(key string, value int, rate float64)

		// stater methods
		Init()
		Shutdown()

	}


A stater can be used individually...

	func main() {

		start := time.Now()
		stat := &stater.Debug{os.Stdout}
		stat.Timer("startup.time", time.Since(start), 1.0) // sample rate range is 0.0-1.0

	}


…or added to a registry, where multiple staters can be published to simultaneously.


	func init() {

		// output to statsd 
		stater.Register(&stater.Statsd{Addr: "stats.d.endpoint:8080"})

		// also to stderr and stdout, devnull does nothing
		stater.Register(&stater.Devnull{})
		stater.Register(&stater.Debug{os.Stderr})
		stater.Register(&stater.Debug{os.Stdout})

	}

	func main() {

		start := time.Now()

		// initialize all staters
		stater.Init()
		defer stater.Shutdown()

		// runs methods on all registered staters
		stater.Increment("name.spaced.value", 1, 0.5)    // only report this 50% of the time
		stater.Increment("name.spaced.value2", 1, 1.0)   // report this every time
		stater.Increment("name.spaced.value3", -1, 0.01) // report 1% of the time

		statsd.Timer("name.spaced.value4", time.Since(start), 1.0)
		statsd.Gauge("name.spaced.value5", "on", 1.0)

	}


Multiple registries can be run with thread-safety, and you can add your own custom staters.

	// a polite stater that does nothing
	type HelloGoodbye stater.DevNull
	
	func (h *HelloGoodbye) Init()     { fmt.Println("Hello") }
	func (h *HelloGoodbye) Shutdown() { fmt.Println("Goodbye") }
	
	func main() {
	
		// use make because the underlying type is a map
		reg, debug := make(stater.Registry), make(stater.Registry)
	
		reg.Register(&HelloGoodbye{})
		reg.Register(&stater.Statsd{
			Addr:              "stats.d.endpoint:8080",
			ReconnectInterval: time.Second,
		})
	
		debug.Register(&stater.Debug{os.Stderr})
	
		// initialize all staters
		reg.Init()
		defer reg.Shutdown()
		debug.Init()
		defer debug.Shutdown()
	
		debug.Gauge("my.value", "not set", 1.0)
		reg.Increment("my.value", 20, 1.0)
		debug.Gauge("my.value", "set", 1.0)
	
	}


Stater is under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html)
