package ip

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	All_Metric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ipmanager_all_list",
		Help: "All IP. Including banned and available IP.",
	}, []string{"ip", "port"})
	Available_Metric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ipmanager_available_list",
		Help: "The current usable IPs list.",
	}, []string{"ip", "port"})
	Cooldown_Metric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ipmanager_cooldown_list",
		Help: "The current cooling down IPs list.",
	}, []string{"ip", "port"})
	ConsecutiveFailure_Metric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ipmanager_consecutive_failure",
		Help: "Consecutive failure number of some certain IP.",
	}, []string{"ip", "port"})
	History_Metric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ipmanager_history_response",
		Help: "Number of responses in last five minutes of some certain IP.",
	}, []string{"ip", "port", "status_code"})
)
