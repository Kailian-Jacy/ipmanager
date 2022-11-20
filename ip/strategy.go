package ip

import (
	config "ipmanager/Config"
	"time"
)

// IsHealth defines the strategy to determine whether an IP is healthy.
func (ip *IP) IsHealth() bool {

	// Append others here.
	if config.C.Strategy != "consecutive" {
		return true
	}
	// Consecutive: There are more than 3 consecutive failure in last 5 minutes.
	if ip.Failure <= config.C.ConsecutiveFailure {
		return true
	}
	return false
}

// CoolDownDuration is the interface deciding the cool down for unhealthy IP.
func (ip *IP) CoolDownDuration() time.Duration {
	return time.Duration(config.C.CoolDown) * time.Minute
}
