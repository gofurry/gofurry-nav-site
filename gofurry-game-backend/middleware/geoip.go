package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gofurry/gofurry-game-backend/common/log"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
	"github.com/gofurry/gofurry-game-backend/roof/env"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/oschwald/geoip2-golang"
)

/*
 * @Desc: GeoIP中间件
 * @author: 福狼
 * @version: v1.0.0
 */

// GeoIP DB 全局变量
var (
	countryDB *geoip2.Reader
	cityDB    *geoip2.Reader
	asnDB     *geoip2.Reader
)

// 初始化 GeoIP DB
func InitGeoIP() {
	var err error
	var url = env.GetServerConfig().Resource.Geolite2Path

	countryDB, err = geoip2.Open(url + "/GeoLite2-Country.mmdb")
	if err != nil {
		log.Error("[GeoLite2] open Country DB fail 打开 Country DB 失败: ", err)
	}

	cityDB, err = geoip2.Open(url + "/GeoLite2-City.mmdb")
	if err != nil {
		log.Error("[GeoLite2] open City DB fail 打开 City DB 失败: ", err)
	}

	asnDB, err = geoip2.Open(url + "/GeoLite2-ASN.mmdb")
	if err != nil {
		log.Error("[GeoLite2] open ASN DB fail 打开 ASN DB 失败: ", err)
	}
}

type baiduResp struct {
	Status string `json:"status"`
	Data   []struct {
		Location string `json:"location"`
	} `json:"data"`
}

type ipInfo struct {
	Country  string `json:"country"`
	Province string `json:"province"`
	City     string `json:"city"`
	ISP      string `json:"isp"`
}

func normalizeISP(raw string) string {
	raw = strings.ToLower(raw)
	switch {
	case strings.Contains(raw, "chinanet"), strings.Contains(raw, "电信"):
		return "电信"
	case strings.Contains(raw, "unicom"), strings.Contains(raw, "联通"):
		return "联通"
	case strings.Contains(raw, "cmcc"), strings.Contains(raw, "移动"):
		return "移动"
	case strings.Contains(raw, "cernet"), strings.Contains(raw, "教育网"):
		return "教育网"
	default:
		return raw
	}
}

func queryBaiduIP(ip string) (country, province, city, isp string) {
	url := "https://opendata.baidu.com/api.php?query=" + ip + "&co=&resource_id=6006&oe=utf8"
	resp, err := http.Get(url)
	if err != nil {
		log.Error("请求百度 IP API 失败: ", err)
		return
	}
	defer resp.Body.Close()

	var result baiduResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error("解析百度 IP API 响应失败: ", err)
		return
	}

	if len(result.Data) == 0 {
		return
	}
	loc := result.Data[0].Location
	if loc == "" {
		return
	}

	if strings.Contains(loc, "省") || strings.Contains(loc, "市") ||
		strings.Contains(loc, "自治区") || strings.Contains(loc, "特别行政区") {

		country = "中国"
		isp = normalizeISP(loc)

		if idx := strings.Index(loc, "省"); idx != -1 {
			province = loc[:idx+len("省")]
			loc = loc[idx+len("省"):]
		} else if idx := strings.Index(loc, "自治区"); idx != -1 {
			province = loc[:idx+len("自治区")]
			loc = loc[idx+len("自治区"):]
		} else if idx := strings.Index(loc, "特别行政区"); idx != -1 {
			province = loc[:idx+len("特别行政区")]
			loc = loc[idx+len("特别行政区"):]
		}

		if idx := strings.LastIndex(loc, "市"); idx != -1 {
			city = loc[:idx+len("市")]
		} else if idx := strings.LastIndex(loc, "地区"); idx != -1 {
			city = loc[:idx+len("地区")]
		}
	} else {
		country = loc
	}
	return
}

// 获取客户端真实 IP
func getClientIP(c fiber.Ctx) string {
	// 先尝试 X-Forwarded-For
	xff := c.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		ip := strings.TrimSpace(ips[0])
		if ip != "" && ip != "::1" && !strings.HasPrefix(ip, "127.") {
			return ip
		}
	}

	// 再尝试 X-Real-IP
	xri := c.Get("X-Real-IP")
	if xri != "" && xri != "::1" && !strings.HasPrefix(xri, "127.") {
		return xri
	}

	// fallback
	ip := c.IP()
	if ip == "" || ip == "::1" || strings.HasPrefix(ip, "127.") {
		return ""
	}
	return ip
}

// 中间件入口
func GeoIPStat(c fiber.Ctx) error {
	ip := getClientIP(c)
	if ip == "" {
		// 无法获取公网 IP 不计数
		return c.Next()
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return c.Next()
	}

	cacheKey := "stat-geoip-cache:" + ip
	var info ipInfo

	// 查缓存
	if cached, err := cs.GetString(cacheKey); err == nil && cached != "" {
		return c.Next()
	}

	// 本地 GeoIP 数据库
	if countryDB != nil {
		if countryInfo, err := countryDB.Country(parsedIP); err == nil && countryInfo != nil {
			if name, ok := countryInfo.Country.Names["zh-CN"]; ok && name != "" {
				info.Country = name
			}
		}
	}
	if info.Country == "中国" && cityDB != nil {
		if cityInfo, err := cityDB.City(parsedIP); err == nil && cityInfo != nil {
			if name, ok := cityInfo.City.Names["zh-CN"]; ok && name != "" {
				info.City = name
			}
		}
	}
	if asnDB != nil {
		if asnInfo, err := asnDB.ASN(parsedIP); err == nil && asnInfo != nil {
			if asnInfo.AutonomousSystemOrganization != "" {
				info.ISP = normalizeISP(asnInfo.AutonomousSystemOrganization)
			}
		}
	}

	// 如果本地数据不全，调用百度 API
	if info.Country == "" || (info.Country == "中国" && (info.City == "" || info.ISP == "")) {
		bCountry, bProvince, bCity, bIsp := queryBaiduIP(ip)
		if info.Country == "" && bCountry != "" {
			info.Country = bCountry
		}
		if info.Province == "" && bProvince != "" {
			info.Province = bProvince
		}
		if info.City == "" && bCity != "" {
			info.City = bCity
		}
		if info.ISP == "" && bIsp != "" {
			info.ISP = bIsp
		}
	}

	// 写入缓存 24h
	if b, err := sonic.Marshal(info); err == nil {
		cs.SetExpire(cacheKey, string(b), 24*time.Hour)
	}

	// 增加统计
	if info.Country != "" {
		cs.Incr("stat-count:total")
		cs.Incr("stat-geoip-country:" + info.Country)
	}
	if info.Country == "中国" && info.Province != "" {
		cs.Incr("stat-geoip-province:" + info.Province)
	}
	if info.Country == "中国" && info.City != "" {
		cs.Incr("stat-geoip-city:" + info.City)
	}
	if info.ISP != "" {
		cs.Incr("stat-geoip-isp:" + info.ISP)
	}

	fmt.Println(info.Country, info.Province, info.City, info.ISP)
	return c.Next()
}
