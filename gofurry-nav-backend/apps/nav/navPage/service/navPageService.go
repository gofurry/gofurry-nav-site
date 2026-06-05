package service

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/dao"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/common/util"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type navPageService struct{}

var navPageSingleton = new(navPageService)

func GetNavPageService() *navPageService { return navPageSingleton }

const siteListCacheKey = "site:list:v2"

const (
	searchSuggestTimeout      = 3 * time.Second
	searchSuggestMaxBodyBytes = 64 * 1024
	searchSuggestMaxQueryLen  = 128
)

var (
	baiduSuggestEndpoint  = "http://suggestion.baidu.com/su"
	bingSuggestEndpoint   = "https://api.bing.com/qsonhs.aspx"
	googleSuggestEndpoint = "http://suggestqueries.google.com/complete/search"
	biliSuggestEndpoint   = "https://s.search.bilibili.com/main/suggest"
	duckSuggestEndpoint   = "https://duckduckgo.com/ac"
)

// 获取导航站点信息
func (svc *navPageService) GetSiteList(lang string) (res []models.SiteVo, err common.GFError) {
	var jsonErr error

	// 先从缓存读取
	if cacheStr, _ := cs.GetString(siteListCacheKey); cacheStr != "" {
		var records []models.GfnSite
		if jsonErr = sonic.Unmarshal([]byte(cacheStr), &records); jsonErr == nil {
			return svc.convertRecords(records, lang), nil
		}
		log.Warn("GetSiteList cache unmarshal error:", jsonErr)
	}

	// 缓存不存在或反序列化失败再从 DB 查询
	records, err := dao.GetNavPageDao().GetSiteList()
	if err != nil {
		return nil, err
	}

	// 更新缓存
	go func() {
		if b, jsonErr := sonic.Marshal(records); jsonErr == nil {
			cs.Set(siteListCacheKey, string(b))
		}
	}()

	return svc.convertRecords(records, lang), nil
}

// 获取导航站点分组信息
func (svc *navPageService) GetGroupList(lang string) (res []models.GroupVo, err common.GFError) {

	var (
		groupRecords   []models.GfnSiteGroup
		mappingRecords []models.GfnSiteGroupMap
	)

	// 读缓存
	groupCache, _ := cs.GetString("group:list")
	mapCache, _ := cs.GetString("group:site:map")

	var err1, err2 error
	if groupCache != "" && mapCache != "" {
		if err1 = sonic.Unmarshal([]byte(groupCache), &groupRecords); err1 == nil {
			if err2 = sonic.Unmarshal([]byte(mapCache), &mappingRecords); err2 == nil {
				return svc.convertGroupRecords(groupRecords, mappingRecords, lang), nil
			}
			log.Warn("GetGroupList map cache unmarshal error:", err2)
		}
		log.Warn("GetGroupList group cache unmarshal error:", err1)
	}

	// 缓存失效查 DB
	groupRecords, err = dao.GetNavPageDao().GetGroupList()
	if err != nil {
		return nil, err
	}
	mappingRecords, err = dao.GetNavPageDao().GetGroupMapList()
	if err != nil {
		return nil, err
	}

	// 异步回填缓存
	go func() {
		if b, err := sonic.Marshal(groupRecords); err == nil {
			cs.Set("group:list", string(b))
		}
		if b, err := sonic.Marshal(mappingRecords); err == nil {
			cs.Set("group:site:map", string(b))
		}
	}()

	return svc.convertGroupRecords(groupRecords, mappingRecords, lang), nil
}

// 获取导航站点延迟信息
func (svc *navPageService) GetPingList() (res map[string]string, err common.GFError) {
	return cs.HGetAll("ping:result")
}

func (svc *navPageService) GetBaiduSuggestion(q string) ([]string, common.GFError) {
	q = normalizeSuggestionQuery(q)
	if q == "" {
		return []string{}, nil
	}
	reqURL, err := buildSuggestionURL(baiduSuggestEndpoint, map[string]string{
		"wd": q,
		"p":  "3",
		"cb": "window.bdsug.sug",
	})
	if err != nil {
		return []string{}, nil
	}
	body, err := fetchSuggestionBody(reqURL, nil)
	if err != nil {
		return []string{}, nil
	}

	// 转 GBK -> UTF-8
	reader := transform.NewReader(bytes.NewReader(body), simplifiedchinese.GBK.NewDecoder())
	utf8Body, _ := io.ReadAll(io.LimitReader(reader, searchSuggestMaxBodyBytes))
	strBody := string(utf8Body)

	// 提取 JSON 字符串
	prefix := "window.bdsug.sug("
	suffix := ");"
	start := strings.Index(strBody, prefix)
	end := strings.LastIndex(strBody, suffix)
	if start == -1 || end == -1 || end <= start+len(prefix) {
		return []string{}, nil
	}
	jsonStr := strBody[start+len(prefix) : end]

	// 把非标准 JSON 的键加上双引号
	replacer := strings.NewReplacer(
		"q:", `"q":`,
		"p:", `"p":`,
		"s:", `"s":`,
	)
	jsonStr = replacer.Replace(jsonStr)

	// 定义结构体
	type BaiduResp struct {
		S []string `json:"s"`
	}

	var result BaiduResp
	if jsonErr := sonic.Unmarshal([]byte(jsonStr), &result); jsonErr != nil {
		return []string{}, nil
	}

	return result.S, nil
}

type BingResponse struct {
	AS struct {
		Results []struct {
			Suggests []struct {
				Txt string `json:"Txt"`
			} `json:"Suggests"`
		} `json:"Results"`
	} `json:"AS"`
}

func (svc *navPageService) GetBingSuggestion(q string) ([]string, common.GFError) {
	q = normalizeSuggestionQuery(q)
	if q == "" {
		return []string{}, nil
	}
	reqURL, err := buildSuggestionURL(bingSuggestEndpoint, map[string]string{
		"type": "cb",
		"q":    q,
	})
	if err != nil {
		return []string{}, nil
	}
	body, err := fetchSuggestionBodyWithConfiguredProxy(reqURL)
	if err != nil {
		return []string{}, nil
	}
	strBody := string(body)

	// 固定前缀和固定尾部
	prefix := "if(typeof  == 'function') ("
	suffix := "/* pageview_candidate */);"

	start := strings.Index(strBody, prefix)
	end := strings.LastIndex(strBody, suffix)
	if start == -1 || end == -1 || end <= start+len(prefix) {
		return []string{}, nil
	}

	jsonStr := strBody[start+len(prefix) : end]

	// 定义结构体解析
	type Suggest struct {
		Txt string `json:"Txt"`
	}
	type Result struct {
		Suggests []Suggest `json:"Suggests"`
	}
	type AS struct {
		Results []Result `json:"Results"`
	}
	type BingResp struct {
		AS AS `json:"AS"`
	}

	var bingResp BingResp
	if err := sonic.Unmarshal([]byte(jsonStr), &bingResp); err != nil {
		return []string{}, nil
	}

	items := []string{}
	for _, r := range bingResp.AS.Results {
		for _, s := range r.Suggests {
			items = append(items, s.Txt)
		}
	}

	return items, nil
}

type GoogleXML struct {
	XMLName             xml.Name `xml:"toplevel"`
	CompleteSuggestions []struct {
		Suggestion struct {
			Data string `xml:"data,attr"`
		} `xml:"suggestion"`
	} `xml:"CompleteSuggestion"`
}

func (svc *navPageService) GetGoogleSuggestion(q string) ([]string, common.GFError) {
	q = normalizeSuggestionQuery(q)
	if q == "" {
		return []string{}, nil
	}
	reqURL, err := buildSuggestionURL(googleSuggestEndpoint, map[string]string{
		"output": "toolbar",
		"hl":     "zh",
		"q":      q,
	})
	if err != nil {
		return []string{}, nil
	}
	body, err := fetchSuggestionBodyWithConfiguredProxy(reqURL)
	if err != nil {
		return []string{}, nil
	}

	type GoogleXML struct {
		XMLName             xml.Name `xml:"toplevel"`
		CompleteSuggestions []struct {
			Suggestion struct {
				Data string `xml:"data,attr"`
			} `xml:"suggestion"`
		} `xml:"CompleteSuggestion"`
	}

	var xmlResp GoogleXML
	if err := xml.Unmarshal(body, &xmlResp); err != nil {
		return []string{}, nil
	}

	items := []string{}
	for _, s := range xmlResp.CompleteSuggestions {
		items = append(items, s.Suggestion.Data)
	}
	return items, nil
}

func (svc *navPageService) GetBiliBiliSuggestion(q string) ([]string, common.GFError) {
	q = normalizeSuggestionQuery(q)
	if q == "" {
		return []string{}, nil
	}

	reqURL, err := buildSuggestionURL(biliSuggestEndpoint, map[string]string{
		"func":         "suggest",
		"suggest_type": "accurate",
		"sub_type":     "tag",
		"main_ver":     "v1",
		"term":         q,
	})
	if err != nil {
		return []string{}, nil
	}
	body, err := fetchSuggestionBody(reqURL, nil)
	if err != nil {
		return []string{}, nil
	}

	// 定义响应结构体
	type TagItem struct {
		Value string `json:"value"`
		Term  string `json:"term"`
		Name  string `json:"name"`
	}
	type BiliResp struct {
		Result struct {
			Tag []TagItem `json:"tag"`
		} `json:"result"`
	}

	var result BiliResp
	if err := sonic.Unmarshal(body, &result); err != nil {
		return []string{}, nil
	}

	// 只返回 value
	suggestions := make([]string, len(result.Result.Tag))
	for i, item := range result.Result.Tag {
		suggestions[i] = item.Value
	}

	return suggestions, nil
}

func (svc *navPageService) GetDuckDuckGoSuggestion(q string) ([]string, common.GFError) {
	q = normalizeSuggestionQuery(q)
	if q == "" {
		return []string{}, nil
	}

	reqURL, err := buildSuggestionURL(duckSuggestEndpoint, map[string]string{
		"q": q,
	})
	if err != nil {
		return []string{}, nil
	}
	body, err := fetchSuggestionBodyWithConfiguredProxy(reqURL)
	if err != nil {
		return []string{}, nil
	}

	type DuckItem struct {
		Phrase string `json:"phrase"`
	}
	var result []DuckItem
	if err := sonic.Unmarshal(body, &result); err != nil {
		return []string{}, nil
	}

	suggestions := make([]string, 0, len(result))
	for _, item := range result {
		if item.Phrase != "" {
			suggestions = append(suggestions, item.Phrase)
		}
	}
	return suggestions, nil
}

func normalizeSuggestionQuery(q string) string {
	q = strings.TrimSpace(q)
	if len([]rune(q)) <= searchSuggestMaxQueryLen {
		return q
	}
	return string([]rune(q)[:searchSuggestMaxQueryLen])
}

func buildSuggestionURL(endpoint string, params map[string]string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	values := u.Query()
	for key, value := range params {
		values.Set(key, value)
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}

func suggestionProxyURL() *url.URL {
	raw := strings.TrimSpace(env.GetServerConfig().Proxy.Url)
	if raw == "" {
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil
	}
	return parsed
}

func fetchSuggestionBodyWithConfiguredProxy(reqURL string) ([]byte, error) {
	return fetchSuggestionBody(reqURL, suggestionProxyURL())
}

func fetchSuggestionBody(reqURL string, proxyURL *url.URL) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), searchSuggestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", common.USER_AGENT)
	req.Header.Set("Accept-Language", common.ACCEPT_LANGUAGE)

	transport := http.DefaultTransport
	if proxyURL != nil {
		transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}
	client := &http.Client{
		Timeout:   searchSuggestTimeout,
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, searchSuggestMaxBodyBytes))
}

func (svc *navPageService) GetSayingService() (res models.SayingModel, err common.GFError) {
	record, err := dao.GetNavPageDao().GetSayingByRandom()
	if err != nil || record == nil {
		return res, common.NewServiceError(fmt.Sprintf("查询金句记录失败: %v", err))
	}
	res.Author = record.Author
	res.Content = record.Saying
	return res, nil
}

func (svc *navPageService) GetImageUrl(t string) string {
	rand.Seed(time.Now().UnixNano())
	addr := "https://qcdn.go-furry.com/nav/bg/"
	res := env.GetServerConfig().Resource
	num := res.NavImageNum
	if t == "standard" {
		addr += "standard-"
		num = res.NavResizedImageNum
	}
	if t == "mobile" {
		addr += "mobile-"
		num = res.NavResizedImageNum
	}
	return addr + "bg-" + util.Int2String(rand.Intn(num)+1) + ".avif"
}

func (svc *navPageService) convertRecords(records []models.GfnSite, lang string) []models.SiteVo {
	res := make([]models.SiteVo, 0, len(records))
	for _, v := range records {
		r := models.SiteVo{
			ID:         util.Int642String(v.ID),
			Domain:     v.Domain,
			Country:    v.Country,
			Nsfw:       v.Nsfw,
			Welfare:    v.Welfare,
			Icon:       v.Icon,
			UpdateTime: v.UpdateTime.String(),
		}
		if lang == "en" {
			r.Name = v.NameEn
			r.Info = v.InfoEn
		} else {
			r.Name = v.Name
			r.Info = v.Info
		}
		res = append(res, r)
	}
	return res
}

func (svc *navPageService) convertGroupRecords(
	groupRecords []models.GfnSiteGroup,
	mappingRecords []models.GfnSiteGroupMap,
	lang string,
) (res []models.GroupVo) {
	sort.SliceStable(mappingRecords, func(i, j int) bool {
		if mappingRecords[i].GroupID != mappingRecords[j].GroupID {
			return mappingRecords[i].GroupID < mappingRecords[j].GroupID
		}
		if mappingRecords[i].ID != mappingRecords[j].ID {
			return mappingRecords[i].ID < mappingRecords[j].ID
		}
		return mappingRecords[i].SiteID < mappingRecords[j].SiteID
	})

	idList := make([]int64, 0, len(groupRecords))
	voMap := make(map[int64]*models.GroupVo, len(groupRecords))

	// 初始化分组
	for _, v := range groupRecords {
		vo := &models.GroupVo{
			ID:       util.Int642String(v.ID),
			Priority: v.Priority,
			Sites:    []string{},
		}

		switch lang {
		case "en":
			vo.Name = v.NameEn
			vo.Info = v.InfoEn
		default:
			vo.Name = v.Name
			vo.Info = v.Info
		}

		voMap[v.ID] = vo
		idList = append(idList, v.ID)
	}

	// 绑定站点
	for _, v := range mappingRecords {
		if vo, ok := voMap[v.GroupID]; ok {
			vo.Sites = append(vo.Sites, util.Int642String(v.SiteID))
		}
	}

	// 保持 DB 顺序
	res = make([]models.GroupVo, 0, len(idList))
	for _, id := range idList {
		res = append(res, *voMap[id])
	}

	return
}
