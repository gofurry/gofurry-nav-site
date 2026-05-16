package task

import (
	"time"

	siteDao "github.com/gofurry/gofurry-nav-backend/apps/system/site/dao"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/bytedance/sonic"
)

func UpdateChangeLog() {
	start := time.Now()
	log.Debug("[StatTask UpdateChangeLog] start...")
	recordList, err := siteDao.GetSiteDao().GetChangeLogList()
	if err != nil {
		log.Error("[StatTask UpdateChangeLog] GetLatestPingLog err:", err)
	}

	if b, jsonErr := sonic.Marshal(recordList); jsonErr == nil {
		cs.SetExpire("site-common:changelog", string(b), 72*time.Hour)
	}

	log.Debug("[StatTask UpdateChangeLog] update latest ping log finished, cost: %v", time.Since(start))
}
