package ip

import "C"
import (
	"bufio"
	"fmt"
	config "ipmanager/Config"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

var re = regexp.MustCompile(`"[0-9]+->[0-9.]+";`)
var re2 = regexp.MustCompile(`[0-9.]+`)

var re_success = regexp.MustCompile(`[2-3][0-9]{2}`)
var re_failure = regexp.MustCompile(`[4-5][0-9]{2}`)

type Log struct {
	Path  string
	F     *os.File
	Count int64
}

func (l *Log) Tail(mode string) ([]*Entry, error) {
	// Reopen each time.
	var err error
	if l.F, err = os.Open(l.Path); err != nil {
		return nil, err
	}
	defer l.F.Close()

	if l.Count == 0 {
		// Normal Mode: Newly opened file. Record the tail and return.
		if mode != "parse" {
			fmt.Println("Parsing access log...")
			l.Count, err = l.F.Seek(0, 2)
			return nil, nil
		}
	}

	// Read from the last tail.
	if _, err = l.F.Seek(l.Count, 0); err != nil {
		log.Println("tail", err.Error())
	}
	scanner := bufio.NewScanner(l.F)
	var entries []*Entry
	for {
		if scanner.Scan() {
			if e, valid := BuildEntry(scanner.Text()); valid {
				entries = append(entries, e)
			}
			if len(entries) > config.C.MaxHistoryLogEachIP {
				// In case loading history caused OOM.
				entries = entries[config.C.MaxHistoryLogEachIP/2:]
			}
			continue
		}
		// Record the very last place and return.
		l.Count, _ = l.F.Seek(0, 2)
		break
	}
	return entries, nil
}

type Entry struct {
	Time       time.Time
	StatusCode string
	ExgressKey string
	Target     string
	Port       string
}

func (e *Entry) IsSuccess() bool {
	if re_success.MatchString(e.StatusCode) {
		return true
	}
	if re_failure.MatchString(e.StatusCode) {
		return false
	}
	if e.Port == "" {
		fmt.Println("Error parsing Entry.Port: ", e.Port)
	}
	return true
}

// Access log Parsing regex.
var timeReA = regexp.MustCompile(`[0-9]*/[a-zA-Z]{3,4}/20[0-9]{2}:[0-9:]+:[0-9]+:[0-9]+ [+-][0-9]+`)
var statusCodeReA = regexp.MustCompile(` [0-9]{3} `)
var exgressKeyReA = regexp.MustCompile(`Exgress_key: [0-9a-zA-Z]+`)
var portReA = regexp.MustCompile(`127.0.0.1:[0-9]{5}`)
var portReB = regexp.MustCompile(`[0-9]{5}`)
var targetReA = regexp.MustCompile(`-> Request Host: [^/]* <-`)

func BuildEntry(line string) (*Entry, bool) {
	var l Entry
	var err error
	t := timeReA.FindString(line)
	if l.Time, err = time.Parse("02/Jan/2006:15:04:05 -0700", t); err != nil {
		log.Println(err.Error())
	}
	l.StatusCode = statusCodeReA.FindString(line)
	l.ExgressKey = exgressKeyReA.FindString(line)
	l.Target = targetReA.FindString(line)
	l.Port = strings.TrimSpace(portReA.FindString(line))
	if l.Port == "" {
		return nil, false
	}
	l.Port = strings.Trim(portReB.FindString(l.Port), " ")

	return &l, true
}
