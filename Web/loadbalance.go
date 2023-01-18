package web

import (
	"fmt"
	"golang.org/x/exp/rand"
	config "ipmanager/Config"
	IP "ipmanager/ip"
	"sync"
	"time"
)

var tokenQue TokenQue
var tokenPool sync.Pool

type Token struct {
	id int
	// next time the job will run
	next time.Time
}

type TokenQue struct {
	tokens []Token
	lock   sync.RWMutex
}

// push token into queue
func (q *TokenQue) push(t Token) {
	q.lock.Lock()
	q.tokens = append(q.tokens, t)
	q.lock.Unlock()
}

// pop front item from queue
func (q *TokenQue) pop() {
	q.lock.Lock()
	q.tokens = q.tokens[1:]
	q.lock.Unlock()
}

// get the front item of queue
func (q *TokenQue) front() *Token {
	q.lock.Lock()
	t := q.tokens[0]
	q.lock.Unlock()
	return &t
}

// judge whether queue is empty
func (q *TokenQue) isEmpty() bool {
	return len(q.tokens) == 0
}

// LoadBalance handle balancing and return port.
func LoadBalance() string {
	// Rand a hash for load balancing.
	k := config.C.TryTimes
	for i := 0; i < k; i++ {
		if v := tokenPool.Get().(Token); v.id != 0 {
			v.next = time.Now().Add(time.Duration(config.C.TokenInterval) * time.Second)
			tokenQue.push(v)

			t := time.Now().UnixMilli()
			r := rand.New(rand.NewSource(uint64(t)))
			p := r.Intn(len(IP.IPAvailable))
			// "10.76.8.101:19001"
			if config.C.Debug {
				fmt.Printf("Balanced to: %s with token ID %d", IP.IPAvailable[p], v.id)
			}
			return config.C.Next + ":" + IP.IPAll[IP.IPAvailable[p]].Port
		}
		time.Sleep(time.Second)
	}
	return ""
}

func InitPool() {
	tokenPool = sync.Pool{
		New: func() interface{} {
			var t Token
			t.id = 0
			return t
		},
	}
	n := config.C.TokenNumber
	for i := 1; i <= n; i++ {
		var t Token
		t.id = i
		t.next = time.Now()
		tokenPool.Put(t)
	}
}

func WatchToken() {
	for !tokenQue.isEmpty() && tokenQue.front().next.Before(time.Now()) {
		tokenPool.Put(tokenQue.front())
		tokenQue.pop()
	}
}
