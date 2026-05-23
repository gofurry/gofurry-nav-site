package schedule

import (
	"time"

	dnsService "github.com/gofurry/gofurry-nav-collector/collector/dns/service"
	httpService "github.com/gofurry/gofurry-nav-collector/collector/http/service"
	pingService "github.com/gofurry/gofurry-nav-collector/collector/ping/service"
	"github.com/gofurry/gofurry-nav-collector/common/log"
)

func InitSchedule() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorFields(map[string]interface{}{
				"component": "scheduler",
				"event":     "init_recovered",
			}, err)
		}
	}()

	start := time.Now()
	log.InfoFields(map[string]interface{}{
		"component": "scheduler",
		"event":     "init_start",
	}, "Collector schedule initialization started")
	pingService.InitPingOnStart() // ping
	httpService.InitHTTPOnStart() // http
	dnsService.InitDNSOnStart()   // dns
	log.InfoFields(map[string]interface{}{
		"component": "scheduler",
		"duration":  time.Since(start),
		"event":     "init_complete",
	}, "Collector schedule initialization completed")
}

func StopSchedule() {
	log.InfoFields(map[string]interface{}{
		"component": "scheduler",
		"event":     "stop",
	}, "Collector schedule stopping")
	dnsService.CloseGeoDB()
}
