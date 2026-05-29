package service

import (
	"github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/dao"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/common/util"
)

type sitePageService struct{}

var sitePageSingleton = new(sitePageService)

func GetSitePageService() *sitePageService { return sitePageSingleton }

// 获取单个站点信息
func (svc *sitePageService) GetSiteDetailService(id string, lang string, clientIP string) (siteInfoVo models.SiteInfoVo, err common.GFError) {
	siteId, utilErr := util.String2Int64(id)
	if utilErr != nil {
		return siteInfoVo, common.NewServiceError("string2int64转换错误: " + utilErr.Error())
	}
	record, err := dao.GetSitePageDao().GetSiteById(siteId)
	siteInfoVo.Icon = record.Icon
	siteInfoVo.Nsfw = record.Nsfw
	siteInfoVo.Country = record.Country
	siteInfoVo.Welfare = record.Welfare
	siteInfoVo.ViewCount = svc.touchSiteViewCount(siteId, record.ViewCount, clientIP)
	switch lang {
	case "zh":
		siteInfoVo.Name = record.Name
		siteInfoVo.Info = record.Info
	case "en":
		siteInfoVo.Name = record.NameEn
		siteInfoVo.Info = record.InfoEn
	default:
		siteInfoVo.Name = record.Name
		siteInfoVo.Info = record.Info
	}

	return
}

// 获取单个站点的 HTTP 记录
func (svc *sitePageService) GetSiteHttpRecordService(domain string) (record string, err common.GFError) {
	return cs.GetString("request:" + domain)
}

// 获取单个站点的 DNS 记录
func (svc *sitePageService) GetSiteDnsRecordService(domain string) (dnsVo models.SiteDnsVo, err common.GFError) {
	reason := "dns:" + domain + "缓存不存在."

	dnsVo.A, err = cs.HGet("dns:"+domain, "A")
	if err != nil && err.GetMsg() != reason {
		return
	}
	dnsVo.AAAA, err = cs.HGet("dns:"+domain, "AAAA")
	if err != nil && err.GetMsg() != reason {
		return
	}
	dnsVo.CNAME, err = cs.HGet("dns:"+domain, "CNAME")
	if err != nil && err.GetMsg() != reason {
		return
	}
	dnsVo.TXT, err = cs.HGet("dns:"+domain, "TXT")
	if err != nil && err.GetMsg() != reason {
		return
	}
	dnsVo.MX, err = cs.HGet("dns:"+domain, "MX")
	if err != nil && err.GetMsg() != reason {
		return
	}
	dnsVo.NS, err = cs.HGet("dns:"+domain, "NS")
	if err != nil && err.GetMsg() != reason {
		return
	}
	dnsVo.SOA, err = cs.HGet("dns:"+domain, "SOA")
	if err != nil && err.GetMsg() != reason {
		return
	}
	dnsVo.CAA, err = cs.HGet("dns:"+domain, "CAA")
	if err != nil && err.GetMsg() != reason {
		return
	}
	if err.GetMsg() == reason {
		err = nil
	}
	return
}

// 获取单个站点的 Ping 记录
func (svc *sitePageService) GetSitePingRecordService(domain string) (siteDelayVo models.SiteDelayVo, err common.GFError) {
	delayList, err := dao.GetSitePageDao().GetDelayList(domain)
	if err != nil {
		return
	}
	idx := 0
	var temp models.SiteDelay
	loss20, delay20, count20 := 0, 0, 0
	loss60, delay60, count60 := 0, 0, 0
	loss100, delay100, count100 := 0, 0, 0
	for _, v := range delayList {
		idx++

		temp.Status = v.Status
		temp.Loss = util.ExtractSuffix2Int(v.Loss, "%")
		temp.Delay = util.ExtractSuffix2Int(v.Delay, "ms")
		temp.Time = v.CreateTime

		// 20 次
		if idx <= 20 {
			siteDelayVo.Twenty.DelayModel = append(siteDelayVo.Twenty.DelayModel, temp)
			loss20 += temp.Loss
			delay20 += temp.Delay
			count20++
		}
		// 60 次抽样 20 次
		if idx <= 60 {
			loss60 += temp.Loss
			delay60 += temp.Delay
			count60++
			if idx%3 == 0 {
				siteDelayVo.Sixty.DelayModel = append(siteDelayVo.Sixty.DelayModel, temp)
			}
		}
		// 100 次抽样 20 次
		if idx <= 100 {
			loss100 += temp.Loss
			delay100 += temp.Delay
			count100++
			if idx%5 == 0 {
				siteDelayVo.Hundred.DelayModel = append(siteDelayVo.Hundred.DelayModel, temp)
			}
		}
	}

	siteDelayVo.Twenty.AvgDelay = avgWithUnit(delay20, count20, "ms")
	siteDelayVo.Twenty.AvgLoss = avgWithUnit(loss20, count20, "%")
	siteDelayVo.Sixty.AvgDelay = avgWithUnit(delay60, count60, "ms")
	siteDelayVo.Sixty.AvgLoss = avgWithUnit(loss60, count60, "%")
	siteDelayVo.Hundred.AvgDelay = avgWithUnit(delay100, count100, "ms")
	siteDelayVo.Hundred.AvgLoss = avgWithUnit(loss100, count100, "%")

	return
}

func avgWithUnit(total int, count int, unit string) string {
	if count <= 0 {
		return "0" + unit
	}
	return util.Int642String(int64(total/count)) + unit
}
