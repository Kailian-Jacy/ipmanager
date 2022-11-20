package web

import (
	"bufio"
	"fmt"
	config "git.zjuqsc.com/3200100963/ipmanager/Config"
	"io"
	"net"
	"time"
)

// Gate verify the header of connection and transfer to proxy or return.
func Gate(src net.Conn) error {
	// No closing src. Because proxy would be using it.
	return nil
}

var TimeOut = "HTTP/1.1 504 Gateway Timeout\nProxy connection timeout.\n"

// Proxy receive the connection and proxy to target.
func Proxy(src net.Conn) {
	var d string
	if config.C.Debug {
		// TODO: Configurable load balancing: Marking upstream and set upstream by hand.
		d = "127.0.0.1:19106"
	} else {
		d = LoadBalance()
	}

	dst, err := net.DialTimeout("tcp", d, time.Duration(config.C.DialTimeOut)*time.Second)

	if err != nil {
		fmt.Println("dial failure to service detected: " + err.Error())
		// TODO Send back timeout info.
		src.Write([]byte("HTTP/1.1 502 Bad Gateway\n\r[PROXY RESENDING ERROR FROM UPSTREAM:]\n\r" + err.Error() + "\n"))
		src.Close()
		return
	}

	if Gate(src) != nil {
		fmt.Println("Gate.")
		return
	}

	done := make(chan struct{})
	// To make a duplex channel, you may need two goroutines.
	// Copy is streaming. It returns when EOF reaches.

	defer func() {
		dst.Close()
		src.Close()
	}()

	go func() {
		io.Copy(dst, src)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(src, dst)
		done <- struct{}{}
	}()

	select {
	case <-done:
		return
	case <-time.After(time.Duration(config.C.MaxConnectionTimeout) * time.Second):
		fmt.Println("Connection timeout.")
		src.Write([]byte(TimeOut))
		dst.Write([]byte(TimeOut))
	}
	// Either side connection close would cause "defer: Send EOF and close connection."
}

func read(src net.Conn) {
	fmt.Println("Reading Connection: " + src.RemoteAddr().String())
	r := bufio.NewReader(src)
	for {
		var read = make([]byte, 100)
		if _, err := r.Read(read); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		fmt.Print(string(read))
	}
	fmt.Println()
}
