package Config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Config struct {
	// Booting mode.
	// server: Not reading the former logs.
	// parse: Scan all the former logs before starting.
	Mode string `json:"mode"`

	// Debug causes verbose output.
	// Warning: may cause behavior change.
	// - NO load balance. All sent according to http.Headers["X-Balance-Dst"]
	Debug   bool   `json:"debug"`
	FIXPORT string `json:"fix_port"`

	// ProbePort: GET HOST/probe to get full ip details.
	ProbePort string `json:"probe_port"`
	// ProbePort: GET HOST/proxy to use the proxy. You'd be load balanced to any Next:Port in UpstreamConfPath.
	ProxyPort string `json:"proxy_port"`

	/*
		Nginx related conf.
	*/
	// Main config for ip manager service.
	ConfigPath string `json:"config_path"`
	// The nginx proxy service IP or Host. Composed as Next:RandomPort to be sent.
	Next string `json:"next"` // e.g. ip.zjuqsc.internal
	// Time Duration to scan log and renew available IPs.
	ScanInterval int `json:"scan_interval_min"`
	// Nginx AccessLog Path.
	AccessLogPath string `json:"access_log_path"` // e.g. /var/log/nginx/host.access.log
	// Upstream of nginx. Scanned automatically for IP list.
	UpstreamConfPath string `json:"upstream_conf_path"` // e.g. /etc/nginx/conf.d/001-upstream.conf

	/*
		In-memory Log rotation.
	*/
	// For log rotate. In case of OOM.
	MaxHistoryLogEachIP  int `json:"max_history_log_each_ip"`
	MaxCoolDownLogEachIP int `json:"max_cool_down_log_each_ip"`

	/*
		Flow control related configuration.
	*/
	// How long does a banned IP to be added to available list. No less than ScanInterval.
	CoolDown int `json:"cool_down_min"`
	// Dial time out defines the time out of upstream. You need to know the time out of oriented service.
	DialTimeOut int `json:"dial_timeout"`
	// Max connection Timeout for connections to prevent goroutine leak. Advised to set long enough.
	MaxConnectionTimeout int `json:"max_connection_timeout"`

	// Judge unhealthy strategy. Only "consecutive" is supported now.
	Strategy string `json:"strategy"`
	// ConsecutiveFailure defines the number of consecutive failure to judge an IP unhealthy.
	ConsecutiveFailure int `json:"consecutive_failure"`
}

var C = Config{
	Mode: "parse",
	//Mode: "serve",

	Debug:   true,
	FIXPORT: "",

	Next:         "127.0.0.1",
	ProbePort:    "9095",
	ProxyPort:    "9096",
	ScanInterval: 5,

	AccessLogPath:    "/Users/kailianjacy/go/ipmanager/Config/host.access.log",
	UpstreamConfPath: "/Users/kailianjacy/go/ipmanager/Config/upstream.conf",

	MaxHistoryLogEachIP:  100,
	MaxCoolDownLogEachIP: 100,

	CoolDown:             10,
	DialTimeOut:          15,
	MaxConnectionTimeout: 100,

	Strategy:           "consecutive",
	ConsecutiveFailure: 3,
}

func LoadMainConfig(path string) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		panic("Configuration Loading error:" + err.Error())
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(b, &C); err != nil {
		panic(err)
	}
}

func ProbeHandler(w http.ResponseWriter, r *http.Request) {
	if C.Debug {
		fmt.Println("Receiving probe: " + r.RequestURI + " From " + r.RemoteAddr)
	}
	b, _ := json.Marshal(C)
	w.Write(b)
}
