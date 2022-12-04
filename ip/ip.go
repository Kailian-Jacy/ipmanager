package ip

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slices"
	"io"
	config "ipmanager/Config"
	"net/http"
	"os"
	"time"
)

var IPAll = make(map[string]*IP, 0)
var IPAvailable = make([]string, 0)

// Translate port into IP.
var port2IP = make(map[string]string, 0)

//var watcher = sync.Mutex{}

type IP struct {
	Port   string
	Addr   string
	Banned bool

	History []*Entry

	// Last cool down.
	Failure   int // Consecutive failures
	CoolDowns [][]time.Time
}

func (ip *IP) Ban(idx int) {
	cooldown := []time.Time{
		time.Now(),
		time.Now().Add(ip.CoolDownDuration()),
	}
	ip.Banned = true

	// Disable
	Available_Metric.With(prometheus.Labels{
		"ip":   ip.Addr,
		"port": ip.Port,
	}).Set(0)
	Available_Count_Metric.Dec()

	ip.CoolDowns = append(ip.CoolDowns, cooldown)
	Cooldown_Metric.With(prometheus.Labels{
		"ip":   ip.Addr,
		"port": ip.Port,
	}).Set(1)
	Cooldown_Count_Metric.Inc()
	IPAvailable = slices.Delete(IPAvailable, idx, idx+1)
}

func (ip *IP) Release() {
	ip.Banned = false
	Cooldown_Metric.With(prometheus.Labels{
		"ip":   ip.Addr,
		"port": ip.Port,
	}).Set(0)
	Cooldown_Count_Metric.Dec()

	IPAvailable = append(IPAvailable, ip.Addr)
	Available_Metric.With(prometheus.Labels{
		"ip":   ip.Addr,
		"port": ip.Port,
	}).Set(1)
	Available_Count_Metric.Inc()
}

func Init() {
	re2.Longest()

	// Fetch all reserved IP.
	Construct(AllIP())

	// Detect available IPs crontab job.
	Watch()
}

// Construct IP to struct.
func Construct(ips [][]string) {
	IPAvailable = make([]string, len(ips))
	for idx, ip := range ips {
		// ip 1 = ip addr, ip 0 = Port.
		IPAvailable[idx] = ip[1]
		IPAll[ip[1]] = &IP{
			Port:   ip[0],
			Addr:   ip[1],
			Banned: false,
		}
		port2IP[ip[0]] = ip[1]
		// Construct Metrics.
		Available_Metric.With(prometheus.Labels{
			"ip":   ip[1],
			"port": ip[0],
		}).Set(1)
		All_Metric.With(prometheus.Labels{
			"ip":   ip[1],
			"port": ip[0],
		}).Set(1)
	}
	if config.C.Debug {
		fmt.Println("Constructed IP:", IPAll)
	}
}

// AllIP parse the nginx conf to get all configured IPs.
func AllIP() [][]string {
	ipconf, err := os.OpenFile(config.C.UpstreamConfPath, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer ipconf.Close()

	all, err := io.ReadAll(ipconf)
	if err != nil {
		panic(err)
	}

	ips := re.FindAllString(string(all), -1)
	ipall := make([][]string, len(ips))
	for i := range ips {
		ipall[i] = re2.FindAllString(ips[i], 2)
	}
	return ipall
}

// Watch and update IP availability and history.
func Watch() {
	//if !watcher.TryLock() {
	//	return
	//}
	//watcher.Lock()
	//defer watcher.Unlock()
	// Scan access log to compose success history.
	ae := AccessLog.Tail(config.C.AccessLogPath, config.C.Mode)
	for _, e := range ae {
		AccessLog.LogIP(e)
	}

	// Limit failure IP.
	for idx := len(IPAvailable) - 1; idx >= 0; idx-- {
		if !IPAll[IPAvailable[idx]].IsHealth() {
			IPAll[IPAvailable[idx]].Ban(idx)
		}
	}

	// Check for minutes and add IP to IPAvailable.
	for idx, ip := range IPAll {
		// Release IP history dynamically.
		if len(ip.CoolDowns) > config.C.MaxCoolDownLogEachIP {
			ip.CoolDowns = ip.CoolDowns[config.C.MaxCoolDownLogEachIP/2:]
		}
		if ip.Banned && ip.CoolDowns[len(ip.CoolDowns)-1][1].Before(time.Now()) {
			// reuse the ip.
			ip.Release()
		}
		IPAll[idx] = ip
	}
}

func Probe() map[string]IP {
	var ips = make(map[string]IP, len(IPAll))
	for ip, ipn := range IPAll {
		ips[ip] = *ipn
	}
	return ips
}

func RenewHandler(w http.ResponseWriter, r *http.Request) {
	Watch()
	w.Write([]byte("Renewed"))
}

func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	if config.C.Debug {
		fmt.Println("Receiving probe: " + r.RequestURI + " From " + r.RemoteAddr)
	}
	Message := struct {
		AvailableIP []string
		IPHistory   map[string]IP
	}{
		AvailableIP: IPAvailable,
		IPHistory:   Probe(),
	}
	b, _ := json.Marshal(Message)
	w.Write(b)
}
