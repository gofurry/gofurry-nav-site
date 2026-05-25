package schedule

import (
	"time"

	dnsService "github.com/gofurry/gofurry-nav-collector/collector/dns/service"
	httpService "github.com/gofurry/gofurry-nav-collector/collector/http/service"
	lightProbeService "github.com/gofurry/gofurry-nav-collector/collector/lightprobe/service"
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
	}, "采集调度初始化开始")
	pingService.InitPingOnStart() // ping
	httpService.InitHTTPOnStart() // http
	dnsService.InitDNSOnStart()   // dns
	lightProbeService.InitLightProbeOnStart()
	log.InfoFields(map[string]interface{}{
		"component": "scheduler",
		"duration":  time.Since(start),
		"event":     "init_complete",
	}, "采集调度初始化完成")
}

func StopSchedule() {
	log.InfoFields(map[string]interface{}{
		"component": "scheduler",
		"event":     "stop",
	}, "采集调度正在停止")
	dnsService.CloseGeoDB()
}
