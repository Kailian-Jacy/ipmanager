package main

import (
	config "ipmanager/Config"
	web "ipmanager/Web"
	"ipmanager/cmd"
	IP "ipmanager/ip"
	_ "runtime/pprof"
	"time"
)

func main() {
	// Add Command line flag.
	cmd.GetFlags()

	// Fetch config and start serving.
	IP.Init()
	IP.C.Start()

	// Serve for proxy and probe.
	web.Init()
	go web.ListenAndServe(config.C.ProbePort)

	go func() {
		for {
			//  keep looping to update IP availability.
			time.Sleep(time.Duration(config.C.ScanInterval) * time.Minute)
			IP.Watch()
		}
	}()

	web.ProxyServeAt(config.C.ProxyPort)
}
