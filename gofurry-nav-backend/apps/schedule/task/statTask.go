package task

import (
	"context"
	"sort"
	"strings"
	"time"

	navDao "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/dao"
	statDao "github.com/gofurry/gofurry-nav-backend/apps/system/stat/dao"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/common/util"
	"github.com/bytedance/sonic"
)

// 更新访问量最多的几个键
func UpdateTopCountCache() {
	start := time.Now()
	log.Debug("[StatTask UpdateTopCountCache] start...")

	type regionType struct {
		Prefix   string
		CacheKey string
	}

	// 最多的 国家、省份、城市
	regions := []regionType{
		{"stat-geoip-country:", "top"},
		{"stat-geoip-province:", "top"},
		{"stat-geoip-city:", "top"},
	}

	for _, r := range regions {
		// 最多的 20 个
		topMap := getTopRegion(r.Prefix, 20)
		// 缓存一天
		if b, err := sonic.Marshal(topMap); err == nil {
			cs.SetExpire(r.Prefix+r.CacheKey, string(b), 24*time.Hour)
		}
	}

	log.Debugf("[StatTask UpdateTopCountCache] update top count cache finished, cost: %v", time.Since(start))
}

func getTopRegion(prefix string, top int) map[string]int64 {
	res := make(map[string]int64)

	// Redis 扫描所有相关 key
	ctx := context.Background()
	iter := cs.GetRedisService().Scan(ctx, 0, prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		valStr, err := cs.GetString(key)
		if err != nil {
			continue
		}
		val, convErr := util.String2Int64(valStr)
		if convErr != nil {
			continue
		}
		// 区域名
		region := strings.TrimPrefix(key, prefix)
		res[region] = val
	}
	if err := iter.Err(); err != nil {
		log.Error("[getTopRegion] redis scan fail 扫描redis键失败:", err)
	}

	// 排序取前 top 个
	type kv struct {
		Key string
		Val int64
	}
	var kvs []kv
	for k, v := range res {
		kvs = append(kvs, kv{k, v})
	}
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Val > kvs[j].Val
	})

	topMap := make(map[string]int64)
	for i := 0; i < len(kvs) && i < top; i++ {
		topMap[kvs[i].Key] = kvs[i].Val
	}
	return topMap
}

func UpdateLatestPingLog() {
	start := time.Now()
	log.Debug("[StatTask UpdateLatestPingLog] start...")
	recordList, err := statDao.GetStatDao().GetLatestPingLog()
	if err != nil {
		log.Error("[StatTask UpdateLatestPingLog] GetLatestPingLog err:", err)
	}

	if b, jsonErr := sonic.Marshal(recordList); jsonErr == nil {
		cs.SetExpire("stat-common:latest-ping-log", string(b), 24*time.Hour)
	}

	log.Debug("[StatTask UpdateLatestPingLog] update latest ping log finished, cost: %v", time.Since(start))
}

func UpdateSiteListCache() {
	start := time.Now()
	log.Debug("[StatTask UpdateSiteListCache] start...")
	records, err := navDao.GetNavPageDao().GetSiteList() // 所有站点记录
	if err != nil {
		log.Error("[StatTask UpdateSiteListCache] GetSiteList err:", err)
	}

	if b, jsonErr := sonic.Marshal(records); jsonErr == nil {
		cs.Set("site:list", string(b))
	}
	log.Debug("[StatTask UpdateSiteListCache] update site list finished, cost: %v", time.Since(start))
}

func UpdateGroupListCache() {
	start := time.Now()
	log.Debug("[StatTask UpdateGroupListCache] start...")
	groupRecords, err := navDao.GetNavPageDao().GetGroupList()
	if err != nil {
		log.Error("[StatTask UpdateGroupListCache] GetGroupList err:", err)
		return
	}

	mappingRecords, err := navDao.GetNavPageDao().GetGroupMapList()
	if err != nil {
		log.Error("[StatTask UpdateGroupListCache] GetGroupMapList err:", err)
		return
	}

	if b, err := sonic.Marshal(groupRecords); err == nil {
		cs.Set("group:list", string(b))
	}
	if b, err := sonic.Marshal(mappingRecords); err == nil {
		cs.Set("group:site:map", string(b))
	}
	log.Debug("[StatTask UpdateGroupListCache] update site group list finished, cost: %v", time.Since(start))
}
