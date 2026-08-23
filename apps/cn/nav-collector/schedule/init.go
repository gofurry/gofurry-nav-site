package schedule

import (
	"time"

	dnsDAO "github.com/gofurry/gofurry-nav-collector/collector/dns/dao"
	dnsService "github.com/gofurry/gofurry-nav-collector/collector/dns/service"
	httpDAO "github.com/gofurry/gofurry-nav-collector/collector/http/dao"
	httpService "github.com/gofurry/gofurry-nav-collector/collector/http/service"
	lightProbeDAO "github.com/gofurry/gofurry-nav-collector/collector/lightprobe/dao"
	lightProbeService "github.com/gofurry/gofurry-nav-collector/collector/lightprobe/service"
	"github.com/gofurry/gofurry-nav-collector/collector/observation"
	pingDAO "github.com/gofurry/gofurry-nav-collector/collector/ping/dao"
	pingService "github.com/gofurry/gofurry-nav-collector/collector/ping/service"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitSchedule(pool *pgxpool.Pool) (initialized bool) {
	defer func() {
		if err := recover(); err != nil {
			initialized = false
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
	observationDAO := observation.NewDAO(pool)
	pingService.InitPingOnStart(pingDAO.New(pool), observationDAO) // ping
	httpService.InitHTTPOnStart(httpDAO.New(pool), observationDAO) // http
	dnsService.InitDNSOnStart(dnsDAO.New(pool), observationDAO)    // dns
	lightProbeService.InitLightProbeOnStart(lightProbeDAO.New(pool), observationDAO)
	log.InfoFields(map[string]interface{}{
		"component": "scheduler",
		"duration":  time.Since(start),
		"event":     "init_complete",
	}, "采集调度初始化完成")
	return true
}

func StopSchedule() {
	log.InfoFields(map[string]interface{}{
		"component": "scheduler",
		"event":     "stop",
	}, "采集调度正在停止")
	dnsService.CloseGeoDB()
}
