package service

import (
	"fmt"
	"math"
	"sort"
	"time"

	logModels "github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/models"
	"github.com/gofurry/gofurry-nav-backend/apps/system/stat/dao"
	"github.com/gofurry/gofurry-nav-backend/apps/system/stat/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/common/util"
	"github.com/bytedance/sonic"
)

type statService struct{}

var statSingleton = new(statService)

func GetStatService() *statService { return statSingleton }

// 增加访问量
func (s statService) AddViewCount() common.GFError {
	cs.Incr("stat-count:total") // 总访问量
	year := util.Int642String(int64(time.Now().Year()))
	month := util.Int642String(int64(time.Now().Month()))
	day := util.Int642String(int64(time.Now().Day()))
	cs.Incr("stat-count:" + year)                           // 年访问量
	cs.Incr("stat-count:" + year + "-" + month)             // 月访问量
	cs.Incr("stat-count:" + year + "-" + month + "-" + day) // 日访问量

	return nil
}

// 返回访问量统计
func (s statService) ViewsCount() (res models.ViewsCountVo, err common.GFError) {
	// 获取时间
	now := time.Now()
	year := now.Year()
	month := now.Month()

	// 获取 redis 缓存
	getViewsCount(&res, year, month)

	// 获取最近7日浏览量
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		key := fmt.Sprintf("stat-count:%d-%d-%d", day.Year(), int(day.Month()), day.Day())

		val, gfsError := cs.GetString(key)
		if gfsError != nil {
			log.Warn("redis未找到key:", key)
			res.Date = append(res.Date, day.Format("2006-01-02"))
			res.Count = append(res.Count, 0)
			continue
		}

		intVal, utilErr := util.String2Int64(val)
		if utilErr != nil {
			intVal = 0
		}

		res.Date = append(res.Date, day.Format("2006-01-02"))
		res.Count = append(res.Count, intVal)
	}

	return
}

// 获取数量最多的分组
func (s statService) GroupCount(lang string) (res []models.GroupCountVo, err common.GFError) {
	return dao.GetStatDao().GetGroupCount(lang)
}

// 获取访问国家统计
func (s statService) CountryCount() (res models.RegionCountVo, err common.GFError) {
	res.RegionMap = readTopCache("stat-geoip-country:top")
	return
}

// 获取访问省份统计
func (s statService) ProvinceCount() (res models.RegionCountVo, err common.GFError) {
	res.RegionMap = readTopCache("stat-geoip-province:top")
	return
}

// 获取访问城市统计
func (s statService) CityCount() (res models.RegionCountVo, err common.GFError) {
	res.RegionMap = readTopCache("stat-geoip-city:top")
	return
}

// 获取近日收录站点列表
func (s statService) SiteList(lang string) (res []models.SiteListVo, err common.GFError) {
	return dao.GetStatDao().GetSiteList(lang)
}

// 获取导航站点的基本数据
func (s statService) SiteCommonInfo() (res models.SiteCommonInfoVo, err common.GFError) {
	res.CommonCountModel, err = dao.GetStatDao().GetCommonCount()
	if err != nil {
		common.NewServiceError(err.GetMsg())
	}

	var nsfwCnt, welfareCnt int64
	siteTypeList, err := dao.GetStatDao().GetSiteCommon()
	if err != nil {
		common.NewServiceError(err.GetMsg())
	}
	for _, v := range siteTypeList {
		if v.NSFW == "1" {
			nsfwCnt++
		}
		if v.Welfare == "1" {
			welfareCnt++
		}
	}

	res.SfwNsfwRatio = math.Round(float64(nsfwCnt)/float64(res.CommonCountModel.SiteCount)*100) / 100
	res.NonProfitRatio = math.Round(float64(welfareCnt)/float64(res.CommonCountModel.SiteCount)*100) / 100

	var pingCnt, pingSum int64
	jsonPingList, gfsError := cs.GetString("stat-common:latest-ping-log")
	if gfsError != nil {
		common.NewServiceError(err.GetMsg())
	}
	var pingRecord []logModels.GfnCollectorLogPing
	sonic.Unmarshal([]byte(jsonPingList), &pingRecord)

	if err != nil {
		common.NewServiceError(err.GetMsg())
	}
	for _, v := range pingRecord {
		pingSum++
		if v.Status == "up" {
			pingCnt++
		}
	}

	if pingSum == 0 {
		res.SiteReachRate = 0.0 // 或根据业务需求设置默认值
	} else {
		res.SiteReachRate = math.Round(float64(pingCnt)/float64(pingSum)*100) / 100
	}

	return res, nil
}

// 获取最近的最高延迟的 ping 记录
func (s statService) SitePingList() (res []models.PingLogVo, err common.GFError) {

	jsonStr, gfsError := cs.GetString("stat-common:latest-ping-log")
	if gfsError != nil {
		common.NewServiceError(gfsError.GetMsg())
	}
	sonic.Unmarshal([]byte(jsonStr), &res)

	sort.Slice(res, func(i, j int) bool {
		delayI := util.ExtractSuffix2Int(res[i].Delay, "ms")
		delayJ := util.ExtractSuffix2Int(res[j].Delay, "ms")
		return delayI > delayJ
	})

	return
}

// 读取缓存
func readTopCache(cacheKey string) map[string]int64 {
	var res map[string]int64
	if val, err := cs.GetString(cacheKey); err == nil && val != "" {
		_ = sonic.Unmarshal([]byte(val), &res)
	}
	if res == nil {
		res = make(map[string]int64)
	}
	return res
}

func getViewsCount(res *models.ViewsCountVo, year int, month time.Month) {
	total, gfsError := cs.GetString("stat-count:total")
	if gfsError != nil {
		log.Error("stat-count:total获取失败: ", gfsError)
	}
	intTotal, utilErr := util.String2Int64(total)
	if utilErr != nil {
		log.Error("String转换Int64获取失败: ", gfsError)
	}
	res.Total = intTotal

	strYear := util.Int642String(int64(year))
	yearCount, gfsError := cs.GetString("stat-count:" + strYear)
	if gfsError != nil {
		log.Error("stat-count:"+strYear+"获取失败: ", strYear, gfsError)
	}
	intYearCount, utilErr := util.String2Int64(yearCount)
	if utilErr != nil {
		log.Error("String转换Int64获取失败: ", gfsError)
	}
	res.YearCount = intYearCount

	strMonth := util.Int642String(int64(month))
	monthCount, gfsError := cs.GetString("stat-count:" + strYear + "-" + strMonth)
	if gfsError != nil {
		log.Error("stat-count:"+strYear+"-"+strMonth+"获取失败: ", gfsError)
	}
	intMonthCount, utilErr := util.String2Int64(monthCount)
	if utilErr != nil {
		log.Error("String转换Int64获取失败: ", gfsError)
	}
	res.MonthCount = intMonthCount
}
