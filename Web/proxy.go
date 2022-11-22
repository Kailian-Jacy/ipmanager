package web

import (
	"errors"
	"fmt"
	"io"
	config "ipmanager/Config"
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
		return
	}

	dst, err := net.DialTimeout("tcp", dPort, time.Duration(config.C.DialTimeOut)*time.Second)
	if err != nil {
		fmt.Println("dial failure to service detected: " + err.Error())
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
	timeOutErr: errors.New("HTTP/1.1 504 Gateway Timeout\nProxy connection timeout.\n"),
}

// gate verify the header of connection and transfer to proxy or return.
func (p *TcpProxy) gate(src *net.Conn) bool {
	// No closing src. Because proxy would be using it.
	return src != nil
}

// TcpProxy receive the connection and proxy to target.
func (p *TcpProxy) proxy(src *net.Conn, dst *net.Conn) {
	done := make(chan struct{})

	defer func() {
		(*dst).Close()
		(*src).Close()
	}()

	go func() {
		io.Copy(*dst, *src)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(*src, *dst)
		done <- struct{}{}
	}()

	select {
	case <-done:
		return
	case <-time.After(p.timeOut):
		fmt.Println("Connection timeout.")
		(*src).Write([]byte(p.timeOutErr.Error()))
		(*dst).Write([]byte(p.timeOutErr.Error()))
		return
	}
	// Either side connection close would cause "defer: Send EOF and close connection."
}
