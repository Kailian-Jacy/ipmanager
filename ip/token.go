package ip

import "C"
import (
	"fmt"
	"ipmanager/Config"
	"sync"
	"time"
)

type Token struct {
	ip   *IP
	next time.Time
}

type Cron struct {
	tokens    []*Token
	add       chan *Token
	remove    chan int
	running   bool
	runningMu sync.Mutex
}

var tokenPool sync.Pool
var cron *Cron

func New() *Cron {
	c := &Cron{
		tokens:    nil,
		add:       make(chan *Token),
		remove:    make(chan int),
		running:   false,
		runningMu: sync.Mutex{},
	}
	return c
}

func (c *Cron) Start() {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.running {
		return
	}
	c.running = true
	go c.run()
}

func (c *Cron) run() {
	fmt.Println("Cron Start")
	now := time.Now()
	for {
		var timer *time.Timer
		if len(c.tokens) == 0 {
			timer = time.NewTimer(654321 * time.Hour)
		} else {
			timer = time.NewTimer(c.tokens[0].next.Sub(now))
		}

		for {
			select {
			case now = <-timer.C:
				last := len(c.tokens)
				for i, t := range c.tokens {
					if t.next.After(now) || t.next.IsZero() {
						last = i
						break
					}
					tokenPool.Put(t)
				}
				c.tokens = c.tokens[last:]
			case newToken := <-c.add:
				timer.Stop()
				now = time.Now()
				c.tokens = append(c.tokens, newToken)
			}
			break
		}
	}
}

func GetAvailableIP() (string, bool) {
	for {
		t := tokenPool.Get()
		if t == nil {
			fmt.Println("Token pool is empty")
			return "", false
		}
		token := t.(Token)
		if !token.ip.Banned {
			token.next = time.Now().Add(time.Duration(Config.C.TokenInterval) * time.Second)
			cron.runningMu.Lock()
			cron.add <- &token
			cron.runningMu.Unlock()
			return token.ip.Addr, true
		}
	}
}

func InitPool() {
	for i := len(IPAvailable) - 1; i >= 0; i-- {
		var token Token
		token.ip = IPAll[IPAvailable[i]]
		tokenPool.Put(token)
	}
}
