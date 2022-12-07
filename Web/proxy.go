package web

import (
	"errors"
	"io"
	config "ipmanager/Config"
	"log"
	"net"
	"time"
)

// ProxyHandler Extend proxy type with gate and proxy.
type ProxyHandler interface {
	gate(src *net.Conn) bool
	proxy(src *net.Conn, dst *net.Conn)
}

func Proxy(p ProxyHandler, src *net.Conn, dPort string) {
	if !p.gate(src) {
		if config.C.Debug {
			log.Println("gate.")
		}
		return
	}
	if config.C.Debug {
		log.Println("Proxy: proxy to: ", dPort)
	}

	dst, err := net.DialTimeout("tcp", dPort, time.Duration(config.C.DialTimeOut)*time.Second)
	if err != nil {
		log.Println("dial failure to service detected: " + err.Error())
		// TODO: Return http error.
		(*src).Write([]byte("HTTP/1.1 502 Bad Gateway\n\r[PROXY RESENDING ERROR FROM UPSTREAM:]\n\r" + err.Error() + "\n"))
		(*src).Close()
		return
	}

	p.proxy(src, &dst)
}

type TcpProxy struct {
	timeOut    time.Duration
	timeOutErr error
}

var tp = &TcpProxy{
	timeOut:    time.Duration(config.C.MaxConnectionTimeout) * time.Second,
	timeOutErr: errors.New("HTTP/1.1 504 Gateway Timeout\n\nProxy connection timeout.\n"),
}

// gate verify the header of connection and transfer to proxy or return.
func (p *TcpProxy) gate(src *net.Conn) bool {
	// No closing src. Because proxy would be using it.
	return src != nil
}

// TcpProxy receive the connection and proxy to target.
func (p *TcpProxy) proxy(src *net.Conn, dst *net.Conn) {
	done := make(chan bool, 2)

	defer func() {
		err := (*dst).Close()
		if config.C.Debug && err != nil {
			log.Println("tcpproxy.defer: dst close.", err)
		}
		err = (*src).Close()
		if config.C.Debug && err != nil {
			log.Println("tcpproxy.defer: src close.", err)
		}
	}()

	go func() {
		defer func() {
			err := (*dst).Close()
			if config.C.Debug && err != nil {
				log.Println("tcpproxy.func2.proxy: dst close.", err)
			}
			err = (*src).Close()
			if config.C.Debug && err != nil {
				log.Println("tcpproxy.func2.proxy: src close.", err)
			}
		}()
		_, err := io.Copy(*dst, *src)
		if config.C.Debug && err != nil {
			log.Println("tcpproxy.func2.proxy.Copy.", err)
		}
		done <- true
	}()
	go func() {
		defer func() {
			err := (*dst).Close()
			if config.C.Debug && err != nil {
				log.Println("tcpproxy.func3.proxy: dst close.", err)
			}
			err = (*src).Close()
			if config.C.Debug && err != nil {
				log.Println("tcpproxy.func3.proxy: src close.", err)
			}
		}()
		_, err := io.Copy(*src, *dst)
		if config.C.Debug && err != nil {
			log.Println("tcpproxy.func3.proxy.Copy.", err)
		}
		done <- true
	}()

	select {
	case <-done:
		if config.C.Debug {
			log.Println("tcpproxy.proxy: done.")
		}
		return
	case <-time.After(p.timeOut):
		if config.C.Debug {
			log.Println("tcpproxy.proxy.main: timeout.")
		}
		return
	}
}
