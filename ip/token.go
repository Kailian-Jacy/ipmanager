package ip

import (
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
var C Cron

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
				for _, t := range c.tokens {
					if t.next.After(now) || t.next.IsZero() {
						break
					}
					tokenPool.Put(t)
				}
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
			return "", false
		}
		token := t.(Token)
		if !token.ip.Banned {
			token.next = time.Now().Add(time.Duration(Config.C.TokenInterval) * time.Second)
			C.add <- &token
			return token.ip.Addr, true
		}
	}
}

func InitPool() {
	for i := len(IPAvailable) - 1; i >= 0; i-- {
		var token Token
		token.ip = IPAll[IPAvailable[i]]
		token.next = time.Now()
		tokenPool.Put(token)
	}
}
