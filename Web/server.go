package web

import (
	"encoding/json"
	"fmt"
	config "git.zjuqsc.com/3200100963/ipmanager/Config"
	IP "git.zjuqsc.com/3200100963/ipmanager/ip"
	"golang.org/x/exp/rand"
	"net"
	"net/http"
	"time"
)

func Init() {
	http.HandleFunc("/probe", ProbeHandler)
	http.HandleFunc("/config", config.ProbeHandler)
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
		go Proxy(conn)
	}
}

func ProbeHandler(w http.ResponseWriter, r *http.Request) {
	if config.C.Debug {
		fmt.Println("Receiving probe: " + r.RequestURI + " From " + r.RemoteAddr)
	}
	Message := struct {
		AvailableIP []string
		IPHistory   map[string]IP.IP
	}{
		AvailableIP: IP.IPAvailable,
		IPHistory:   IP.Probe(),
	}
	b, _ := json.Marshal(Message)
	w.Write(b)
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
