package web

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/rand"
	config "ipmanager/Config"
	IP "ipmanager/ip"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func Init() {
	http.HandleFunc("/api/config", config.ProbeHandler)
	http.HandleFunc("/api/ip/renew", IP.RenewHandler)
	http.HandleFunc("/api/ip/history", IP.HistoryHandler)
	http.Handle("/metrics", promhttp.Handler())
}

func ProxyServeAt(port string) {
	fmt.Println("Proxy is listening at: " + port)
	listener, err := net.Listen("tcp", ":"+port)
	// This is a naive implementation of http proxy. For stubborn, ErrServerClosed should be considered in case.
	// if src.shuttingDown() { return ErrServerClosed }
	if err != nil {
		panic("connection error:" + err.Error())
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("connection accept error:" + err.Error())
			continue
		}
		if config.C.Debug {
			fmt.Println("Receiving proxy.")
		}
		go func() {
			if config.C.FIXPORT != "" {
				Proxy(tp, &conn, config.C.Next+":"+config.C.FIXPORT)
			} else {
				Proxy(tp, &conn, LoadBalance())
			}
		}()
	}
}

func ListenAndServe(port string) {
	fmt.Println("Prober is listening at 127.0.0.1:" + port)
	http.ListenAndServe(":"+port, nil)
}

// LoadBalance handle balancing and return port.
func LoadBalance() string {
	// Rand a hash for load balancing.
	s := rand.NewSource(uint64(time.Now().Unix()))
	r := rand.New(s) // initialize local pseudorandom generator
	r.Intn(len(IP.IPAvailable))
	// "10.76.8.101:19001"
	if config.C.Debug {
		fmt.Println("Balanced to: " + IP.IPAvailable[r.Intn(len(IP.IPAvailable))])
	}
	return config.C.Next + ":" + IP.IPAll[IP.IPAvailable[r.Intn(len(IP.IPAvailable))]].Port
}
