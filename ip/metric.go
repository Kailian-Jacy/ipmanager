package ip

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	All_Metric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ip_all",
		Help: "All IP. Including banned and available IP.",
	}, []string{"ip", "port"})
	All_Count_Metric = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ip_all_count",
		Help: "All IP count. Including banned and available IP.",
	})
	Available_Metric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ip_available_list",
		Help: "The current usable IPs list.",
	}, []string{"ip", "port"})
	Available_Count_Metric = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ip_available_count",
		Help: "The current usable IPs count.",
	})
	Cooldown_Metric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ip_cooldown_list",
		Help: "The current cooling down IPs list.",
	}, []string{"ip", "port"})
	Cooldown_Count_Metric = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ip_cooldown_count",
		Help: "The current cooling down IPs count.",
	})
)
