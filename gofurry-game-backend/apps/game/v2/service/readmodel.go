package service

import (
	"context"
	"html"
	"math"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/bytedance/sonic"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
)

const (
	defaultLang      = "zh"
	defaultRegion    = "CN"
	onlineUnknown    = "unknown"
	priceUnavailable = "region_price_unavailable"
	priceMissing     = "region_price_missing"

	similarRecommendationAlgorithmVersion = "similar-v2.3.1-hybrid-cbf"
	similarPrecomputeLimit                = 64
)

type gameDetailReader interface {
	GetGameDetailAggregate(ctx context.Context, query v2models.GameV2DetailQuery) (v2models.GameV2Aggregate, common.GFError)
	ListGameAggregates(ctx context.Context, query v2models.GameV2ListQuery) ([]v2models.GameV2Aggregate, common.GFError)
	SearchGames(ctx context.Context, query v2models.GameV2SearchPageQuery) (cm.PageResponse, common.GFError)
	ListTags(ctx context.Context, lang string) ([]v2models.GameV2TagRecord, common.GFError)
	GetGameReviews(ctx context.Context, gameID int64) (v2models.GameV2ReviewList, common.GFError)
	ListLatestReviews(ctx context.Context, lang string, limit int) ([]v2models.GameV2LatestReview, common.GFError)
	GetRandomGameID(ctx context.Context) (string, common.GFError)
	ListSimilarRecommendations(ctx context.Context, query v2models.GameV2SimilarRecommendationQuery) ([]v2models.GameV2RecommendationRow, common.GFError)
	SaveSimilarRecommendations(ctx context.Context, sourceGameID int64, rows []v2models.GfgGameV2Recommendation) common.GFError
	ListRecommendationFeatures(ctx context.Context, lang string, region string) ([]v2models.GameV2RecommendationFeature, common.GFError)
	GetGameNews(ctx context.Context, query v2models.GameV2NewsQuery) ([]v2models.GameV2NewsRow, common.GFError)
	GetLatestGameNews(ctx context.Context, query v2models.GameV2NewsQuery) ([]v2models.GameV2NewsRow, common.GFError)
	ListTopOnlineAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError)
	ListFreeGameAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError)
	ListHighestPriceAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError)
	ListHighestDiscountAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError)
	ListLowPriceAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError)
	GetCollectStatus(ctx context.Context) (v2models.GameV2CollectStatus, common.GFError)
	ListCollectRuns(ctx context.Context, query v2models.GameV2CollectRunQuery) ([]v2models.GfgGameV2CollectRun, common.GFError)
	GetCollectRun(ctx context.Context, runID string) (*v2models.GfgGameV2CollectRun, common.GFError)
	ListCollectTaskResults(ctx context.Context, query v2models.GameV2CollectTaskResultQuery) ([]v2models.GfgGameV2CollectTaskResult, common.GFError)
	GetGameCollectStatus(ctx context.Context, gameID int64, appID int64) (v2models.GameV2CollectGameStatus, common.GFError)
	ListSyncCreators(ctx context.Context, lang string) ([]v2models.GameV2SyncCreatorRow, common.GFError)
}

type ReadModelService struct {
	reader gameDetailReader
}

var (
	htmlBreakRE = regexp.MustCompile(`(?i)<\s*(br|/p|/div|/li)\s*/?>`)
	htmlTagRE   = regexp.MustCompile(`<[^>]+>`)
	spaceRE     = regexp.MustCompile(`[ \t\r\f\v]+`)
	newlineRE   = regexp.MustCompile(`\n{3,}`)
)

func NewReadModelServiceWithReader(reader gameDetailReader) *ReadModelService {
	return &ReadModelService{reader: reader}
}

func (svc *ReadModelService) GetGameList(ctx context.Context, query v2models.GameV2ListQuery) ([]v2models.GameV2ListItem, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	query.Lang = normalizeLang(query.Lang)
	query.Region = normalizeRegion(query.Region)
	query.Limit = clampLimit(query.Limit, 20, 100)
	if query.Offset < 0 {
		query.Offset = 0
	}
	query.Sort = normalizeSort(query.Sort)

	aggregates, err := svc.reader.ListGameAggregates(ctx, query)
	if err != nil {
		return nil, err
	}
	res := make([]v2models.GameV2ListItem, 0, len(aggregates))
	for _, aggregate := range aggregates {
		res = append(res, buildListItem(aggregate, query.Lang, query.Region))
	}
	return res, nil
}

func (svc *ReadModelService) ListSyncGames(ctx context.Context, query v2models.GameV2SyncListQuery) ([]v2models.GameV2SyncGameSummary, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	lang := normalizeLang(query.Lang)
	region := normalizeRegion(query.Region)
	limit := clampLimit(query.Limit, 5000, 5000)
	if query.Offset < 0 {
		query.Offset = 0
	}
	aggregates, err := svc.reader.ListGameAggregates(ctx, v2models.GameV2ListQuery{
		Lang:         lang,
		Region:       region,
		Limit:        limit,
		Offset:       query.Offset,
		Sort:         "weight",
		UpdatedSince: query.UpdatedSince,
	})
	if err != nil {
		return nil, err
	}
	res := make([]v2models.GameV2SyncGameSummary, 0, len(aggregates))
	for _, aggregate := range aggregates {
		item := buildListItem(aggregate, lang, region)
		res = append(res, v2models.GameV2SyncGameSummary{
			ID:          item.ID,
			AppID:       item.AppID,
			Name:        item.Name,
			Info:        cleanSyncText(item.Summary),
			ReleaseDate: item.ReleaseDate,
			Developers:  item.Developers,
			Publishers:  item.Publishers,
			UpdatedAt:   item.UpdatedAt,
		})
	}
	return res, nil
}

func (svc *ReadModelService) GetGameDetail(ctx context.Context, req v2models.GameV2DetailRequest) (v2models.GameV2DetailReadModel, common.GFError) {
	var res v2models.GameV2DetailReadModel
	if svc == nil || svc.reader == nil {
		return res, common.NewServiceError("game v2 read model service is not initialized")
	}

	requestedLang := normalizeLang(req.Lang)
	region := normalizeRegion(req.Region)
	query := v2models.GameV2DetailQuery{
		GameID:    req.GameID,
		AppID:     req.AppID,
		Lang:      requestedLang,
		NewsLimit: req.NewsLimit,
	}

	aggregate, err := svc.reader.GetGameDetailAggregate(ctx, query)
	if err != nil {
		return res, err
	}

	res = buildDetailReadModel(aggregate, requestedLang, region)
	return res, nil
}

func (svc *ReadModelService) GetSyncGameDetail(ctx context.Context, req v2models.GameV2DetailRequest) (v2models.GameV2SyncGameDetail, common.GFError) {
	var res v2models.GameV2SyncGameDetail
	req.NewsLimit = 0
	detail, err := svc.GetGameDetail(ctx, req)
	if err != nil {
		return res, err
	}
	return v2models.GameV2SyncGameDetail{
		ID:                  detail.ID,
		AppID:               detail.AppID,
		Name:                detail.Name,
		Info:                cleanSyncText(detail.Summary),
		Resources:           detail.Site.Resources,
		Groups:              detail.Site.Groups,
		ReleaseDate:         detail.Release.Date,
		Developers:          detail.Developers,
		Publishers:          detail.Publishers,
		Links:               detail.Site.Links,
		Platform:            platformsText(detail.Platforms),
		Tags:                detail.Tags,
		SupportedLanguages:  cleanSyncText(detail.SupportedLanguages),
		Website:             detail.Website,
		DetailedDescription: cleanSyncText(detail.DetailedDescription),
		AboutTheGame:        cleanSyncText(detail.AboutTheGame),
		PcRequirements:      syncPCRequirements(detail.Requirements.PC),
		UpdatedAt:           detail.UpdatedAt,
	}, nil
}

func (svc *ReadModelService) SimpleSearch(ctx context.Context, req v2models.GameV2SearchRequest) ([]v2models.GameV2SearchItem, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	query := normalizeSearchPageQuery(v2models.GameV2SearchPageQueryRequest{
		PageReq: cm.PageReq{
			PageNum:  1,
			PageSize: 8,
		},
		Content: stringPtr(req.Txt),
		Lang:    req.Lang,
	})
	page, err := svc.reader.SearchGames(ctx, query)
	if err != nil {
		return nil, err
	}
	items, ok := page.Data.([]v2models.GameV2SearchPageItem)
	if !ok {
		return nil, common.NewServiceError("game v2 simple search returned invalid data")
	}
	res := make([]v2models.GameV2SearchItem, 0, len(items))
	for _, item := range items {
		res = append(res, v2models.GameV2SearchItem{
			ID:    item.ID,
			Name:  item.Name,
			Info:  item.Info,
			Cover: item.Cover,
		})
	}
	return res, nil
}

func (svc *ReadModelService) SearchPage(ctx context.Context, req v2models.GameV2SearchPageQueryRequest) (cm.PageResponse, common.GFError) {
	if svc == nil || svc.reader == nil {
		return cm.PageResponse{}, common.NewServiceError("game v2 read model service is not initialized")
	}
	return svc.reader.SearchGames(ctx, normalizeSearchPageQuery(req))
}

func (svc *ReadModelService) ListTags(ctx context.Context, lang string) ([]v2models.GameV2TagRecord, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	return svc.reader.ListTags(ctx, normalizeLang(lang))
}

func (svc *ReadModelService) GetGameReviews(ctx context.Context, id string) (v2models.GameV2ReviewList, common.GFError) {
	var res v2models.GameV2ReviewList
	if svc == nil || svc.reader == nil {
		return res, common.NewServiceError("game v2 read model service is not initialized")
	}
	gameID, parseErr := strconv.ParseInt(strings.TrimSpace(id), 10, 64)
	if parseErr != nil || gameID <= 0 {
		return res, common.NewServiceError("Game ID 转换有误")
	}
	res, err := svc.reader.GetGameReviews(ctx, gameID)
	if err != nil {
		return res, err
	}
	for i := range res.Remarks {
		res.Remarks[i].IP = desensitizeIP(res.Remarks[i].IP)
	}
	return res, nil
}

func (svc *ReadModelService) ListLatestReviews(ctx context.Context, lang string, limit int) ([]v2models.GameV2LatestReview, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	rows, err := svc.reader.ListLatestReviews(ctx, normalizeLang(lang), clampLimit(limit, 5, 20))
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].IP = desensitizeIP(rows[i].IP)
	}
	return rows, nil
}

func (svc *ReadModelService) GetRandomGameID(ctx context.Context) (string, common.GFError) {
	if svc == nil || svc.reader == nil {
		return "", common.NewServiceError("game v2 read model service is not initialized")
	}
	return svc.reader.GetRandomGameID(ctx)
}

func (svc *ReadModelService) GetSimilarRecommendations(ctx context.Context, query v2models.GameV2SimilarRecommendationQuery) ([]v2models.GameV2SimilarRecommendation, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	if query.GameID <= 0 {
		return nil, common.NewServiceError("id 不能为空")
	}
	query.Lang = normalizeLang(query.Lang)
	query.Region = normalizeRegion(query.Region)
	query.Limit = clampLimit(query.Limit, 8, 20)
	query.AlgorithmVersion = similarRecommendationAlgorithmVersion

	rows, err := svc.reader.ListSimilarRecommendations(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		return buildSimilarRecommendationsFromRows(rows), nil
	}

	features, err := svc.reader.ListRecommendationFeatures(ctx, query.Lang, query.Region)
	if err != nil {
		return nil, err
	}
	computed, saveRows, sourceGameID, computeErr := computeSimilarRecommendations(features, query)
	if computeErr != nil {
		return nil, computeErr
	}
	if err := svc.reader.SaveSimilarRecommendations(ctx, sourceGameID, saveRows); err != nil {
		return nil, err
	}
	if len(computed) > query.Limit {
		computed = computed[:query.Limit]
	}
	return computed, nil
}

func (svc *ReadModelService) GetGameNews(ctx context.Context, query v2models.GameV2NewsQuery) ([]v2models.GameV2NewsItem, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	query.Lang = normalizeLang(query.Lang)
	query.Limit = clampLimit(query.Limit, 20, 100)
	if query.Offset < 0 {
		query.Offset = 0
	}
	rows, err := svc.reader.GetGameNews(ctx, query)
	if err != nil {
		return nil, err
	}
	return buildNewsRows(rows, query.Lang), nil
}

func (svc *ReadModelService) GetLatestGameNews(ctx context.Context, query v2models.GameV2NewsQuery) ([]v2models.GameV2NewsItem, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	query.Lang = normalizeLang(query.Lang)
	query.Limit = clampLimit(query.Limit, 20, 100)
	if query.Offset < 0 {
		query.Offset = 0
	}
	rows, err := svc.reader.GetLatestGameNews(ctx, query)
	if err != nil {
		return nil, err
	}
	return buildNewsRows(rows, query.Lang), nil
}

func (svc *ReadModelService) ListSyncGameNews(ctx context.Context, query v2models.GameV2SyncNewsQuery) ([]v2models.GameV2SyncNewsItem, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	lang := normalizeLang(query.Lang)
	limit := clampLimit(query.Limit, 5000, 5000)
	if query.Offset < 0 {
		query.Offset = 0
	}
	rows, err := svc.reader.GetLatestGameNews(ctx, v2models.GameV2NewsQuery{
		Lang:         lang,
		Limit:        limit,
		Offset:       query.Offset,
		UpdatedSince: query.UpdatedSince,
	})
	if err != nil {
		return nil, err
	}
	items := buildNewsRows(rows, lang)
	res := make([]v2models.GameV2SyncNewsItem, 0, len(items))
	for _, item := range items {
		content := firstNonEmptyString(item.PlainText, item.Summary, item.HTML)
		res = append(res, v2models.GameV2SyncNewsItem{
			ID:          item.ID,
			GameID:      item.GameID,
			AppID:       item.AppID,
			Name:        item.GameName,
			PostTime:    formatSyncTime(item.PublishedAt),
			Headline:    item.Headline,
			Author:      "",
			Content:     cleanSyncText(content),
			URL:         item.URL,
			Lang:        item.Lang,
			UpdatedAt:   item.UpdatedAt,
			PublishedAt: item.PublishedAt,
		})
	}
	return res, nil
}

func (svc *ReadModelService) ListSyncCreators(ctx context.Context, lang string) ([]v2models.GameV2SyncCreator, common.GFError) {
	return svc.ListCreators(ctx, lang)
}

func (svc *ReadModelService) ListCreators(ctx context.Context, lang string) ([]v2models.GameV2SyncCreator, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	rows, err := svc.reader.ListSyncCreators(ctx, normalizeLang(lang))
	if err != nil {
		return nil, err
	}
	res := make([]v2models.GameV2SyncCreator, 0, len(rows))
	for _, row := range rows {
		res = append(res, v2models.GameV2SyncCreator{
			ID:         strconv.FormatInt(row.ID, 10),
			Name:       row.Name,
			Info:       cleanSyncText(row.Info),
			URL:        row.URL,
			Avatar:     row.Avatar,
			Links:      parseKVList(row.Links),
			Contact:    parseKVList(row.Contact),
			Type:       row.Type,
			CreateTime: row.CreateTime,
			UpdateTime: row.UpdateTime,
		})
	}
	return res, nil
}

func (svc *ReadModelService) GetPanelMain(ctx context.Context, query v2models.GameV2PanelQuery) (v2models.GameV2PanelReadModel, common.GFError) {
	var res v2models.GameV2PanelReadModel
	if svc == nil || svc.reader == nil {
		return res, common.NewServiceError("game v2 read model service is not initialized")
	}
	query.Lang = normalizeLang(query.Lang)
	query.Region = normalizeRegion(query.Region)
	query.Limit = clampLimit(query.Limit, 8, 24)
	query.TopOnlineLimit = clampLimit(query.TopOnlineLimit, 60, 60)
	query.PriceLimit = clampLimit(query.PriceLimit, 120, 200)
	query.NewsLimit = clampLimit(query.NewsLimit, 8, 24)

	latest, err := svc.reader.ListGameAggregates(ctx, v2models.GameV2ListQuery{
		Lang:   query.Lang,
		Region: query.Region,
		Limit:  query.Limit,
		Sort:   "newest",
	})
	if err != nil {
		return res, err
	}
	updated, err := svc.reader.ListGameAggregates(ctx, v2models.GameV2ListQuery{
		Lang:   query.Lang,
		Region: query.Region,
		Limit:  query.Limit,
		Sort:   "updated",
	})
	if err != nil {
		return res, err
	}
	topOnlineQuery := query
	topOnlineQuery.Limit = query.TopOnlineLimit
	topOnline, err := svc.reader.ListTopOnlineAggregates(ctx, topOnlineQuery)
	if err != nil {
		return res, err
	}
	freeGames, err := svc.reader.ListFreeGameAggregates(ctx, query)
	if err != nil {
		return res, err
	}
	priceQuery := query
	priceQuery.Region = "US"
	priceQuery.Limit = query.PriceLimit
	topPriceQuery := priceQuery
	topPriceQuery.Limit = 15
	topPrice, err := svc.reader.ListHighestPriceAggregates(ctx, topPriceQuery)
	if err != nil {
		return res, err
	}
	discountQuery := priceQuery
	discountQuery.Limit = 15
	highestDiscount, err := svc.reader.ListHighestDiscountAggregates(ctx, discountQuery)
	if err != nil {
		return res, err
	}
	lowPrice, err := svc.reader.ListLowPriceAggregates(ctx, priceQuery)
	if err != nil {
		return res, err
	}
	latestNews, err := svc.GetLatestGameNews(ctx, v2models.GameV2NewsQuery{
		Lang:  query.Lang,
		Limit: query.NewsLimit,
	})
	if err != nil {
		return res, err
	}

	res.LatestGames = buildListItems(latest, query.Lang, query.Region)
	res.UpdatedGames = buildListItems(updated, query.Lang, query.Region)
	res.TopOnline = buildListItems(topOnline, query.Lang, query.Region)
	res.FreeGames = buildListItems(freeGames, query.Lang, query.Region)
	res.TopPrice = buildListItems(topPrice, query.Lang, query.Region)
	res.HighestDiscount = buildListItems(highestDiscount, query.Lang, query.Region)
	res.LowPrice = buildListItems(lowPrice, query.Lang, query.Region)
	res.LatestNews = latestNews
	return res, nil
}

func (svc *ReadModelService) GetCollectStatus(ctx context.Context) (v2models.GameV2CollectStatus, common.GFError) {
	var res v2models.GameV2CollectStatus
	if svc == nil || svc.reader == nil {
		return res, common.NewServiceError("game v2 read model service is not initialized")
	}
	return svc.reader.GetCollectStatus(ctx)
}

func (svc *ReadModelService) ListCollectRuns(ctx context.Context, query v2models.GameV2CollectRunQuery) ([]v2models.GfgGameV2CollectRun, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	query.TaskType = strings.TrimSpace(query.TaskType)
	query.Status = strings.TrimSpace(query.Status)
	query.Limit = clampLimit(query.Limit, 20, 100)
	if query.Offset < 0 {
		query.Offset = 0
	}
	return svc.reader.ListCollectRuns(ctx, query)
}

func (svc *ReadModelService) GetCollectRun(ctx context.Context, runID string) (*v2models.GfgGameV2CollectRun, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, common.NewServiceError("run_id is required")
	}
	return svc.reader.GetCollectRun(ctx, runID)
}

func (svc *ReadModelService) ListCollectTaskResults(ctx context.Context, query v2models.GameV2CollectTaskResultQuery) ([]v2models.GfgGameV2CollectTaskResult, common.GFError) {
	if svc == nil || svc.reader == nil {
		return nil, common.NewServiceError("game v2 read model service is not initialized")
	}
	query.RunID = strings.TrimSpace(query.RunID)
	query.TaskType = strings.TrimSpace(query.TaskType)
	query.Status = strings.TrimSpace(query.Status)
	query.Limit = clampLimit(query.Limit, 50, 200)
	if query.Offset < 0 {
		query.Offset = 0
	}
	return svc.reader.ListCollectTaskResults(ctx, query)
}

func (svc *ReadModelService) GetGameCollectStatus(ctx context.Context, gameID int64, appID int64) (v2models.GameV2CollectGameStatus, common.GFError) {
	var res v2models.GameV2CollectGameStatus
	if svc == nil || svc.reader == nil {
		return res, common.NewServiceError("game v2 read model service is not initialized")
	}
	if gameID <= 0 && appID <= 0 {
		return res, common.NewServiceError("game_id or appid is required")
	}
	return svc.reader.GetGameCollectStatus(ctx, gameID, appID)
}

type recommendationFeature struct {
	row            v2models.GameV2RecommendationFeature
	tags           []recommendationTag
	tagWeights     map[string]float64
	tagNames       map[string]string
	developers     []string
	publishers     []string
	creators       map[string]struct{}
	platforms      map[string]bool
	textTokens     map[string]float64
	priceAvailable bool
}

type recommendationTag struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Prefix string `json:"prefix"`
}

type recommendationScore struct {
	target       recommendationFeature
	score        float64
	displayScore float64
	reasons      []v2models.GameV2RecommendationReason
}

/*
v2.3.1 相似推荐选型说明：

 1. 这版不继续沿用 v1 的“请求时标签独热编码 + 余弦相似度”作为主实现，因为它把计算放在详情页请求路径上，
    后续游戏数量增多时容易让详情页性能抖动，也不方便解释“为什么推荐这个游戏”。
 2. 这版也暂不引入协同过滤。当前站点没有足够稳定的用户行为矩阵，强行做 CF 会遇到冷启动和噪声问题，
    对小体量游戏库的收益低于维护成本。
 3. 这版先不把 RAG/embedding 作为主推荐算法。向量适合做语义召回，但需要额外索引、重建流程和质量评估。
    当前阶段的主要目标是移除 v1 包袱、稳定详情页体验，所以先使用可解释、可 SQL 落库、可手动调权的 hybrid CBF。

算法结构：
  - 标签相似度占 60%。标签是最强业务语义，使用 Weighted Jaccard，而不是纯 one-hot cosine。
    权重来源优先使用 gfg_tag.prefix 和站内主/次标签，避免继续硬编码“某个 ID 段一定代表某种含义”的历史做法。
  - 创作者/开发商/发行商占 15%。同工作室或同发行商通常是强相关，但不能压过标签。
  - 文本相似度占 10%。只用清洗后的名称与简介做轻量 token Jaccard，避免在数据库主链路里引入重分词依赖。
  - 平台占 7%。平台是过滤和弱偏好信号，不应该让“都支持 Windows”变成过强推荐理由。
  - 价格占 5%。免费、同价位或同折扣可以作为补充信号，但价格经常变化，所以权重低。
  - 活跃度占 3%。在线人数接近只作为轻微排序信号，避免热门游戏吞掉所有小众相似项。

结果存储：
- service 计算 top 64 写入 gfg_game_v2_recommendations，并记录 algorithm_version。
- API 优先读取预计算结果；开发环境或单个游戏缺失时即时计算一次并回填。
- 未来如果接入 collector/admin 定时全量重算，只需要调用同一套特征与保存逻辑，接口合同不用变。
*/
func computeSimilarRecommendations(features []v2models.GameV2RecommendationFeature, query v2models.GameV2SimilarRecommendationQuery) ([]v2models.GameV2SimilarRecommendation, []v2models.GfgGameV2Recommendation, int64, common.GFError) {
	normalized := make([]recommendationFeature, 0, len(features))
	sourceIndex := -1
	for _, row := range features {
		feature := normalizeRecommendationFeature(row)
		normalized = append(normalized, feature)
		if row.GameID == query.GameID {
			sourceIndex = len(normalized) - 1
		}
	}
	if sourceIndex < 0 {
		return nil, nil, 0, common.NewServiceError("目标游戏不存在或缺少 v2 详情")
	}
	source := normalized[sourceIndex]

	scored := make([]recommendationScore, 0, len(normalized))
	for _, target := range normalized {
		if target.row.GameID == source.row.GameID {
			continue
		}
		score, reasons := scoreRecommendation(source, target, query.Lang)
		if score <= 0 {
			continue
		}
		scored = append(scored, recommendationScore{
			target:       target,
			score:        score,
			displayScore: displayRecommendationScore(score),
			reasons:      reasons,
		})
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].target.row.GameID < scored[j].target.row.GameID
		}
		return scored[i].score > scored[j].score
	})

	precomputeLimit := similarPrecomputeLimit
	if precomputeLimit < query.Limit {
		precomputeLimit = query.Limit
	}
	if len(scored) > precomputeLimit {
		scored = scored[:precomputeLimit]
	}

	now := time.Now()
	res := make([]v2models.GameV2SimilarRecommendation, 0, len(scored))
	rows := make([]v2models.GfgGameV2Recommendation, 0, len(scored))
	for idx, item := range scored {
		rank := idx + 1
		reasonJSON, _ := sonic.Marshal(item.reasons)
		rows = append(rows, v2models.GfgGameV2Recommendation{
			SourceGameID:     source.row.GameID,
			TargetGameID:     item.target.row.GameID,
			Score:            item.score,
			DisplayScore:     item.displayScore,
			Rank:             rank,
			ReasonJSON:       string(reasonJSON),
			AlgorithmVersion: similarRecommendationAlgorithmVersion,
			ComputedAt:       now,
		})
		res = append(res, buildSimilarRecommendationFromFeature(item.target, item.score, item.displayScore, rank, item.reasons, now))
	}
	return res, rows, source.row.GameID, nil
}

func scoreRecommendation(source recommendationFeature, target recommendationFeature, lang string) (float64, []v2models.GameV2RecommendationReason) {
	tagScore, sharedTags := weightedJaccard(source.tagWeights, target.tagWeights, source.tagNames)
	creatorScore, sharedCreators := stringSetJaccard(source.creators, target.creators)
	textScore, _ := weightedJaccard(source.textTokens, target.textTokens, nil)
	platformScore, sharedPlatforms := platformJaccard(source.platforms, target.platforms)
	priceScore, priceReason := priceSimilarity(source, target)
	activityScore := activitySimilarity(source.row.OnlineCount, target.row.OnlineCount)

	score := tagScore*0.60 + creatorScore*0.15 + textScore*0.10 + platformScore*0.07 + priceScore*0.05 + activityScore*0.03
	score = clamp01(score)

	reasons := make([]v2models.GameV2RecommendationReason, 0, 4)
	if tagScore > 0 && len(sharedTags) > 0 {
		reasons = append(reasons, recommendationReason("tag", reasonLabel(lang, "tag"), strings.Join(limitStrings(sharedTags, 3), ", "), 0.60*tagScore))
	}
	if creatorScore > 0 && len(sharedCreators) > 0 {
		reasons = append(reasons, recommendationReason("creator", reasonLabel(lang, "creator"), strings.Join(limitStrings(sharedCreators, 2), ", "), 0.15*creatorScore))
	}
	if platformScore > 0 && len(sharedPlatforms) > 0 {
		reasons = append(reasons, recommendationReason("platform", reasonLabel(lang, "platform"), strings.Join(limitStrings(sharedPlatforms, 3), ", "), 0.07*platformScore))
	}
	if priceScore > 0 && priceReason != "" {
		reasons = append(reasons, recommendationReason("price", reasonLabel(lang, "price"), priceReason, 0.05*priceScore))
	}
	sort.Slice(reasons, func(i, j int) bool {
		return reasons[i].Weight > reasons[j].Weight
	})
	if len(reasons) > 3 {
		reasons = reasons[:3]
	}
	return score, reasons
}

func normalizeRecommendationFeature(row v2models.GameV2RecommendationFeature) recommendationFeature {
	tags := decodeJSON[[]recommendationTag](row.Tags, []recommendationTag{})
	developers := normalizeNameList(decodeJSON[[]string](row.Developers, []string{}))
	publishers := normalizeNameList(decodeJSON[[]string](row.Publishers, []string{}))
	platforms := decodeJSON[map[string]bool](row.Platforms, map[string]bool{})

	tagWeights := make(map[string]float64, len(tags))
	tagNames := make(map[string]string, len(tags))
	for _, tag := range tags {
		if tag.ID == "" {
			continue
		}
		tagWeights[tag.ID] = recommendationTagWeight(tag, row.PrimaryTagID, row.SecondaryTagID)
		tagNames[tag.ID] = tag.Name
	}

	creators := make(map[string]struct{}, len(developers)+len(publishers))
	for _, value := range developers {
		creators[value] = struct{}{}
	}
	for _, value := range publishers {
		creators[value] = struct{}{}
	}

	textTokens := tokenizeRecommendationText(row.Name + " " + row.Summary)
	return recommendationFeature{
		row:            row,
		tags:           tags,
		tagWeights:     tagWeights,
		tagNames:       tagNames,
		developers:     developers,
		publishers:     publishers,
		creators:       creators,
		platforms:      platforms,
		textTokens:     textTokens,
		priceAvailable: row.PriceAvailable,
	}
}

func buildSimilarRecommendationsFromRows(rows []v2models.GameV2RecommendationRow) []v2models.GameV2SimilarRecommendation {
	res := make([]v2models.GameV2SimilarRecommendation, 0, len(rows))
	for _, row := range rows {
		reasons := decodeJSON[[]v2models.GameV2RecommendationReason](&row.ReasonJSON, []v2models.GameV2RecommendationReason{})
		tags := decodeJSON[[]recommendationTag](row.Tags, []recommendationTag{})
		res = append(res, v2models.GameV2SimilarRecommendation{
			ID:               strconv.FormatInt(row.TargetGameID, 10),
			AppID:            strconv.FormatInt(row.AppID, 10),
			Name:             row.Name,
			Summary:          row.Summary,
			HeaderURL:        row.HeaderURL,
			CapsuleURL:       row.CapsuleURL,
			Score:            row.Score,
			DisplayScore:     row.DisplayScore,
			Rank:             row.Rank,
			Reasons:          reasons,
			AlgorithmVersion: row.AlgorithmVersion,
			ComputedAt:       row.ComputedAt,
			Tags:             recommendationTagsToView(tags),
			Price:            recommendationRowPrice(row),
			OnlineCount:      recommendationRowOnline(row),
		})
	}
	return res
}

func buildSimilarRecommendationFromFeature(feature recommendationFeature, score float64, displayScore float64, rank int, reasons []v2models.GameV2RecommendationReason, computedAt time.Time) v2models.GameV2SimilarRecommendation {
	return v2models.GameV2SimilarRecommendation{
		ID:               strconv.FormatInt(feature.row.GameID, 10),
		AppID:            strconv.FormatInt(feature.row.AppID, 10),
		Name:             feature.row.Name,
		Summary:          feature.row.Summary,
		HeaderURL:        feature.row.HeaderURL,
		CapsuleURL:       feature.row.CapsuleURL,
		Score:            score,
		DisplayScore:     displayScore,
		Rank:             rank,
		Reasons:          reasons,
		AlgorithmVersion: similarRecommendationAlgorithmVersion,
		ComputedAt:       computedAt,
		Tags:             recommendationTagsToView(feature.tags),
		Price:            recommendationFeaturePrice(feature.row),
		OnlineCount:      recommendationFeatureOnline(feature.row),
	}
}

func recommendationTagWeight(tag recommendationTag, primaryTagID int64, secondaryTagID int64) float64 {
	id, _ := strconv.ParseInt(tag.ID, 10, 64)
	switch id {
	case primaryTagID:
		return 2.0
	case secondaryTagID:
		return 1.5
	}
	switch tag.Prefix {
	case "1000":
		return 1.2
	case "2000":
		return 1.4
	case "3000":
		return 0.4
	case "9000":
		return 1.3
	default:
		return 1.0
	}
}

func weightedJaccard(a map[string]float64, b map[string]float64, names map[string]string) (float64, []string) {
	if len(a) == 0 || len(b) == 0 {
		return 0, nil
	}
	keys := make(map[string]struct{}, len(a)+len(b))
	for key := range a {
		keys[key] = struct{}{}
	}
	for key := range b {
		keys[key] = struct{}{}
	}
	var intersection float64
	var union float64
	shared := make([]string, 0)
	for key := range keys {
		av := a[key]
		bv := b[key]
		intersection += math.Min(av, bv)
		union += math.Max(av, bv)
		if av > 0 && bv > 0 && names != nil {
			if name := strings.TrimSpace(names[key]); name != "" {
				shared = append(shared, name)
			}
		}
	}
	sort.Strings(shared)
	if union <= 0 {
		return 0, shared
	}
	return clamp01(intersection / union), shared
}

func stringSetJaccard(a map[string]struct{}, b map[string]struct{}) (float64, []string) {
	if len(a) == 0 || len(b) == 0 {
		return 0, nil
	}
	intersection := 0
	union := make(map[string]struct{}, len(a)+len(b))
	shared := make([]string, 0)
	for key := range a {
		union[key] = struct{}{}
		if _, ok := b[key]; ok {
			intersection++
			shared = append(shared, key)
		}
	}
	for key := range b {
		union[key] = struct{}{}
	}
	sort.Strings(shared)
	return float64(intersection) / float64(len(union)), shared
}

func platformJaccard(a map[string]bool, b map[string]bool) (float64, []string) {
	aSet := make(map[string]struct{})
	bSet := make(map[string]struct{})
	for key, enabled := range a {
		if enabled {
			aSet[strings.ToLower(key)] = struct{}{}
		}
	}
	for key, enabled := range b {
		if enabled {
			bSet[strings.ToLower(key)] = struct{}{}
		}
	}
	return stringSetJaccard(aSet, bSet)
}

func priceSimilarity(source recommendationFeature, target recommendationFeature) (float64, string) {
	if source.row.IsFree && target.row.IsFree {
		return 1, "free"
	}
	if !source.priceAvailable || !target.priceAvailable {
		return 0, ""
	}
	if source.row.FinalAmount <= 0 || target.row.FinalAmount <= 0 {
		return 0, ""
	}
	minPrice := math.Min(float64(source.row.FinalAmount), float64(target.row.FinalAmount))
	maxPrice := math.Max(float64(source.row.FinalAmount), float64(target.row.FinalAmount))
	if maxPrice <= 0 {
		return 0, ""
	}
	return clamp01(minPrice / maxPrice), "similar price"
}

func activitySimilarity(source int64, target int64) float64 {
	if source <= 0 || target <= 0 {
		return 0
	}
	diff := math.Abs(math.Log1p(float64(source)) - math.Log1p(float64(target)))
	return clamp01(1 / (1 + diff))
}

func displayRecommendationScore(score float64) float64 {
	return clamp01(0.50 + 0.45*math.Sqrt(clamp01(score)))
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func recommendationReason(reasonType string, label string, value string, weight float64) v2models.GameV2RecommendationReason {
	return v2models.GameV2RecommendationReason{
		Type:   reasonType,
		Label:  label,
		Value:  value,
		Weight: math.Round(weight*1000) / 1000,
	}
}

func reasonLabel(lang string, reasonType string) string {
	if lang == "en" {
		switch reasonType {
		case "tag":
			return "Shared tags"
		case "creator":
			return "Related creators"
		case "platform":
			return "Shared platforms"
		case "price":
			return "Similar price"
		}
	}
	switch reasonType {
	case "tag":
		return "共同标签"
	case "creator":
		return "相关创作者"
	case "platform":
		return "共同平台"
	case "price":
		return "价格相近"
	default:
		return "推荐理由"
	}
}

func normalizeNameList(values []string) []string {
	res := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		res = append(res, value)
	}
	return res
}

func tokenizeRecommendationText(value string) map[string]float64 {
	tokens := make(map[string]float64)
	var builder strings.Builder
	flush := func() {
		token := strings.TrimSpace(builder.String())
		builder.Reset()
		if len([]rune(token)) >= 2 {
			tokens[token] = 1
		}
	}
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return tokens
}

func limitStrings(values []string, limit int) []string {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func recommendationTagsToView(tags []recommendationTag) []v2models.GameV2Tag {
	res := make([]v2models.GameV2Tag, 0, len(tags))
	for _, tag := range tags {
		res = append(res, v2models.GameV2Tag{
			ID:   tag.ID,
			Name: tag.Name,
			Desc: tag.Desc,
		})
	}
	return res
}

func recommendationFeaturePrice(row v2models.GameV2RecommendationFeature) v2models.GameV2PriceView {
	updatedAt := time.Time{}
	if row.PriceUpdatedAt != nil {
		updatedAt = *row.PriceUpdatedAt
	}
	view := v2models.GameV2PriceView{
		Region:           row.PriceRegion,
		Available:        row.PriceAvailable,
		IsFree:           row.IsFree,
		Currency:         row.Currency,
		InitialAmount:    row.InitialAmount,
		FinalAmount:      row.FinalAmount,
		DiscountPercent:  row.DiscountPercent,
		InitialFormatted: row.InitialFormatted,
		FinalFormatted:   row.FinalFormatted,
		UpdatedAt:        updatedAt,
		CollectedAt:      updatedAt,
	}
	if !view.Available {
		view.UnavailableReason = priceMissing
	}
	return view
}

func recommendationRowPrice(row v2models.GameV2RecommendationRow) v2models.GameV2PriceView {
	updatedAt := time.Time{}
	if row.PriceUpdatedAt != nil {
		updatedAt = *row.PriceUpdatedAt
	}
	view := v2models.GameV2PriceView{
		Region:           row.PriceRegion,
		Available:        row.PriceAvailable,
		IsFree:           row.IsFree,
		Currency:         row.Currency,
		InitialAmount:    row.InitialAmount,
		FinalAmount:      row.FinalAmount,
		DiscountPercent:  row.DiscountPercent,
		InitialFormatted: row.InitialFormatted,
		FinalFormatted:   row.FinalFormatted,
		UpdatedAt:        updatedAt,
		CollectedAt:      updatedAt,
	}
	if !view.Available {
		view.UnavailableReason = priceMissing
	}
	return view
}

func recommendationFeatureOnline(row v2models.GameV2RecommendationFeature) v2models.GameV2OnlineCount {
	collectedAt := time.Time{}
	if row.OnlineCollectedAt != nil {
		collectedAt = *row.OnlineCollectedAt
	}
	status := row.OnlineStatus
	if status == "" {
		status = onlineUnknown
	}
	return v2models.GameV2OnlineCount{Count: row.OnlineCount, Status: status, CollectedAt: collectedAt}
}

func recommendationRowOnline(row v2models.GameV2RecommendationRow) v2models.GameV2OnlineCount {
	collectedAt := time.Time{}
	if row.OnlineCollectedAt != nil {
		collectedAt = *row.OnlineCollectedAt
	}
	status := row.OnlineStatus
	if status == "" {
		status = onlineUnknown
	}
	return v2models.GameV2OnlineCount{Count: row.OnlineCountValue, Status: status, CollectedAt: collectedAt}
}

func buildListItems(aggregates []v2models.GameV2Aggregate, lang string, region string) []v2models.GameV2ListItem {
	res := make([]v2models.GameV2ListItem, 0, len(aggregates))
	for _, aggregate := range aggregates {
		res = append(res, buildListItem(aggregate, lang, region))
	}
	return res
}

func buildListItem(aggregate v2models.GameV2Aggregate, lang string, region string) v2models.GameV2ListItem {
	detail := buildDetailReadModel(aggregate, lang, region)
	return v2models.GameV2ListItem{
		ID:           detail.ID,
		AppID:        detail.AppID,
		Name:         detail.Name,
		Summary:      detail.Summary,
		HeaderURL:    detail.HeaderURL,
		CapsuleURL:   detail.Media.CapsuleURL,
		ReleaseDate:  detail.Release.Date,
		Developers:   detail.Developers,
		Publishers:   detail.Publishers,
		Platforms:    detail.Platforms,
		Prices:       detail.Prices,
		Price:        detail.Price,
		OnlineCount:  detail.OnlineCount,
		Tags:         detail.Tags,
		AvgScore:     aggregate.ReviewStats.AvgScore,
		CommentCount: aggregate.ReviewStats.CommentCount,
		UpdatedAt:    detail.UpdatedAt,
	}
}

func buildDetailReadModel(aggregate v2models.GameV2Aggregate, requestedLang string, region string) v2models.GameV2DetailReadModel {
	lang := requestedLang
	if aggregate.Localized != nil && aggregate.Localized.Lang != "" {
		lang = aggregate.Localized.Lang
	}

	name := localizedName(aggregate, lang)
	summary := localizedSummary(aggregate, lang)
	headerURL := aggregate.Site.Header
	siteName := aggregate.Site.Name
	siteInfo := aggregate.Site.Info
	if lang == "en" {
		if aggregate.Site.NameEn != "" {
			siteName = aggregate.Site.NameEn
		}
		if aggregate.Site.InfoEn != "" {
			siteInfo = aggregate.Site.InfoEn
		}
	}

	res := v2models.GameV2DetailReadModel{
		ID:            strconv.FormatInt(aggregate.Site.ID, 10),
		AppID:         strconv.FormatInt(aggregate.Site.AppID, 10),
		RequestedLang: requestedLang,
		Lang:          lang,
		Name:          name,
		Summary:       summary,
		HeaderURL:     headerURL,
		Prices:        buildPrices(aggregate.Prices),
		Price:         selectPrice(aggregate.Prices, region),
		Media:         buildMedia(aggregate.Media),
		Requirements:  buildRequirements(aggregate.Requirements),
		News:          buildNews(aggregate.News),
		OnlineCount:   buildOnlineCount(aggregate.OnlineCount),
		Tags:          append([]v2models.GameV2Tag(nil), aggregate.Tags...),
		Site: v2models.GameV2SiteInfo{
			ID:         strconv.FormatInt(aggregate.Site.ID, 10),
			Name:       siteName,
			Info:       siteInfo,
			Header:     aggregate.Site.Header,
			ViewCount:  aggregate.Site.ViewCount,
			Resources:  parseKVList(aggregate.Site.Resources),
			Groups:     parseKVList(aggregate.Site.Groups),
			Links:      parseKVList(aggregate.Site.Links),
			CreateTime: aggregate.Site.CreateTime,
			UpdateTime: aggregate.Site.UpdateTime,
		},
		UpdatedAt: getUpdatedAt(aggregate),
	}

	if aggregate.Details != nil {
		details := aggregate.Details
		res.Type = details.Type
		res.IsFree = details.IsFree
		res.Website = strValue(details.Website)
		if strValue(details.HeaderURL) != "" {
			res.HeaderURL = strValue(details.HeaderURL)
		}
		res.Release = v2models.GameV2Release{ComingSoon: details.ReleaseComingSoon, Date: strValue(details.ReleaseDateText)}
		res.Developers = decodeJSON[[]string](details.Developers, []string{})
		res.Publishers = decodeJSON[[]string](details.Publishers, []string{})
		res.Platforms = decodeJSON[map[string]bool](details.Platforms, map[string]bool{})
		res.SupportedLanguages = strValue(details.SupportedLanguages)
		res.SupportInfo = decodeJSON[map[string]string](details.SupportInfo, map[string]string{})
		res.CollectedAt = details.CollectedAt
		res.Extra.ContentDescriptors = decodeAny(details.ContentDescriptors)
		res.Extra.Ratings = decodeAny(details.Ratings)
	}

	if aggregate.Localized != nil {
		res.ShortDescription = strValue(aggregate.Localized.ShortDescription)
		res.DetailedDescription = strValue(aggregate.Localized.DetailedDescription)
		res.AboutTheGame = strValue(aggregate.Localized.AboutTheGame)
		if res.ShortDescription != "" {
			res.Summary = res.ShortDescription
		}
	}

	if res.Media.HeaderURL != "" {
		res.HeaderURL = res.Media.HeaderURL
	}
	if res.Name == "" {
		res.Name = siteName
	}
	if res.Summary == "" {
		res.Summary = siteInfo
	}
	return res
}

func normalizeLang(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "en", "en-us", "en_us":
		return "en"
	default:
		return defaultLang
	}
}

func normalizeRegion(region string) string {
	region = strings.ToUpper(strings.TrimSpace(region))
	if region == "" {
		return defaultRegion
	}
	return region
}

func normalizeSort(sort string) string {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "newest", "updated", "weight":
		return strings.ToLower(strings.TrimSpace(sort))
	default:
		return "weight"
	}
}

func clampLimit(limit int, defaultValue int, maxValue int) int {
	if limit <= 0 {
		return defaultValue
	}
	if limit > maxValue {
		return maxValue
	}
	return limit
}

func normalizeSearchPageQuery(req v2models.GameV2SearchPageQueryRequest) v2models.GameV2SearchPageQuery {
	req.InitPageIfAbsent()
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	content := ""
	if req.Content != nil {
		content = strings.TrimSpace(*req.Content)
	}
	return v2models.GameV2SearchPageQuery{
		Lang:            normalizeLang(req.Lang),
		Content:         content,
		PubStartTime:    req.PubStartTime.Time(),
		PubEndTime:      req.PubEndTime.Time(),
		UpdateStartTime: req.UpdateStartTime.Time(),
		UpdateEndTime:   req.UpdateEndTime.Time(),
		ScoreOrder:      req.ScoreOrder,
		RemarkOrder:     req.RemarkOrder,
		TimeOrder:       req.TimeOrder,
		TagList:         append([]int64(nil), req.TagList...),
		PageNum:         req.PageNum,
		PageSize:        req.PageSize,
	}
}

func stringPtr(value string) *string {
	return &value
}

func desensitizeIP(ip string) string {
	if ip == "" {
		return ""
	}
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return "***"
	}
	if ipAddr.To4() != nil {
		segments := strings.Split(ip, ".")
		if len(segments) == 4 {
			return strings.Join(segments[:3], ".") + ".***"
		}
		return "***"
	}
	if strings.Contains(ip, "::") {
		parts := strings.Split(ip, "::")
		if len(parts) == 2 {
			switch {
			case parts[1] == "":
				return parts[0] + "::*"
			case parts[0] == "":
				return "::*"
			default:
				return parts[0] + "::****"
			}
		}
	}
	segments := strings.Split(ip, ":")
	if len(segments) == 0 {
		return "***"
	}
	keepSegCount := len(segments) - 1
	if keepSegCount < 1 {
		keepSegCount = 0
	}
	return strings.Join(segments[:keepSegCount], ":") + ":****"
}

func localizedName(aggregate v2models.GameV2Aggregate, lang string) string {
	if aggregate.Localized != nil && aggregate.Localized.Name != "" {
		return aggregate.Localized.Name
	}
	if aggregate.Details != nil && aggregate.Details.Name != "" {
		return aggregate.Details.Name
	}
	if lang == "en" && aggregate.Site.NameEn != "" {
		return aggregate.Site.NameEn
	}
	return aggregate.Site.Name
}

func localizedSummary(aggregate v2models.GameV2Aggregate, lang string) string {
	if aggregate.Localized != nil && strValue(aggregate.Localized.ShortDescription) != "" {
		return strValue(aggregate.Localized.ShortDescription)
	}
	if lang == "en" && aggregate.Site.InfoEn != "" {
		return aggregate.Site.InfoEn
	}
	return aggregate.Site.Info
}

func buildPrices(prices []v2models.GfgGameV2Price) []v2models.GameV2PriceView {
	res := make([]v2models.GameV2PriceView, 0, len(prices))
	for _, price := range prices {
		res = append(res, buildPriceView(price))
	}
	return res
}

func selectPrice(prices []v2models.GfgGameV2Price, region string) v2models.GameV2PriceView {
	for _, price := range prices {
		if strings.EqualFold(price.Region, region) {
			return buildPriceView(price)
		}
	}
	return v2models.GameV2PriceView{
		Region:            region,
		Available:         false,
		UnavailableReason: priceMissing,
	}
}

func buildPriceView(price v2models.GfgGameV2Price) v2models.GameV2PriceView {
	view := v2models.GameV2PriceView{
		Region:           price.Region,
		Available:        true,
		IsFree:           price.IsFree,
		Currency:         strValue(price.Currency),
		InitialAmount:    price.InitialAmount,
		FinalAmount:      price.FinalAmount,
		DiscountPercent:  price.DiscountPercent,
		InitialFormatted: strValue(price.InitialFormatted),
		FinalFormatted:   strValue(price.FinalFormatted),
		CollectedAt:      price.CollectedAt,
		UpdatedAt:        price.UpdatedAt,
	}
	if isUnavailableRegionalPrice(price) {
		view.Available = false
		view.UnavailableReason = priceUnavailable
	}
	return view
}

func isUnavailableRegionalPrice(price v2models.GfgGameV2Price) bool {
	return !price.IsFree &&
		strings.TrimSpace(strValue(price.Currency)) == "" &&
		price.FinalAmount == 0 &&
		strings.TrimSpace(strValue(price.FinalFormatted)) == ""
}

func buildMedia(items []v2models.GfgGameV2Media) v2models.GameV2MediaView {
	res := v2models.GameV2MediaView{
		Screenshots: []v2models.GameV2Screenshot{},
		Movies:      []v2models.GameV2Movie{},
	}
	for _, item := range items {
		switch item.MediaType {
		case "header":
			res.HeaderURL = strValue(item.URL)
		case "capsule":
			res.CapsuleURL = strValue(item.URL)
		case "capsule_v5":
			res.CapsuleV5URL = strValue(item.URL)
		case "background":
			res.BackgroundURL = strValue(item.URL)
		case "background_raw":
			res.BackgroundRawURL = strValue(item.URL)
		case "screenshot":
			res.Screenshots = append(res.Screenshots, v2models.GameV2Screenshot{
				ID:           item.MediaKey,
				URL:          strValue(item.URL),
				ThumbnailURL: strValue(item.ThumbnailURL),
			})
		case "movie":
			res.Movies = append(res.Movies, v2models.GameV2Movie{
				ID:           item.MediaKey,
				Name:         strValue(item.Title),
				URL:          strValue(item.URL),
				ThumbnailURL: strValue(item.ThumbnailURL),
				Extra:        decodeAny(item.Extra),
			})
		}
	}
	return res
}

func buildRequirements(requirements *v2models.GfgGameV2Requirements) v2models.GameV2RequirementsView {
	if requirements == nil {
		return v2models.GameV2RequirementsView{
			PC:    map[string]string{},
			Mac:   map[string]string{},
			Linux: map[string]string{},
		}
	}
	return v2models.GameV2RequirementsView{
		PC:    decodeJSON[map[string]string](requirements.PC, map[string]string{}),
		Mac:   decodeJSON[map[string]string](requirements.Mac, map[string]string{}),
		Linux: decodeJSON[map[string]string](requirements.Linux, map[string]string{}),
	}
}

func buildNews(news []v2models.GfgGameV2News) []v2models.GameV2NewsItem {
	res := make([]v2models.GameV2NewsItem, 0, len(news))
	for _, item := range news {
		res = append(res, buildNewsItem(item, "", ""))
	}
	return res
}

func buildNewsRows(rows []v2models.GameV2NewsRow, requestedLang string) []v2models.GameV2NewsItem {
	res := make([]v2models.GameV2NewsItem, 0, len(rows))
	for _, row := range rows {
		gameName := row.GameName
		if requestedLang == "en" && row.GameNameEn != "" {
			gameName = row.GameNameEn
		}
		res = append(res, buildNewsItem(row.GfgGameV2News, gameName, row.HeaderURL))
	}
	return res
}

func buildNewsItem(item v2models.GfgGameV2News, gameName string, headerURL string) v2models.GameV2NewsItem {
	return v2models.GameV2NewsItem{
		ID:            strconv.FormatInt(item.ID, 10),
		GameID:        strconv.FormatInt(item.GameID, 10),
		AppID:         strconv.FormatInt(item.AppID, 10),
		Lang:          item.Lang,
		GameName:      gameName,
		HeaderURL:     headerURL,
		EventGID:      item.EventGID,
		Headline:      item.Headline,
		Summary:       strValue(item.Summary),
		PlainText:     strValue(item.PlainText),
		HTML:          strValue(item.HTML),
		URL:           strValue(item.URL),
		Tags:          decodeJSON[[]string](item.Tags, []string{}),
		PublishedAt:   item.PublishedAt,
		UpdatedAt:     item.UpdatedAt,
		CommentCount:  item.CommentCount,
		VoteUpCount:   item.VoteUpCount,
		VoteDownCount: item.VoteDownCount,
	}
}

func buildOnlineCount(online *v2models.GfgGameV2PlayerCount) v2models.GameV2OnlineCount {
	if online == nil {
		return v2models.GameV2OnlineCount{Status: onlineUnknown}
	}
	return v2models.GameV2OnlineCount{
		Count:       online.Count,
		Status:      online.Status,
		CollectedAt: online.CollectedAt,
	}
}

func parseKVList(raw *string) []cm.KvModel {
	return decodeJSON[[]cm.KvModel](raw, []cm.KvModel{})
}

func decodeJSON[T any](raw *string, fallback T) T {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return fallback
	}
	var value T
	if err := sonic.Unmarshal([]byte(*raw), &value); err != nil {
		return fallback
	}
	return value
}

func decodeAny(raw *string) any {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil
	}
	var value any
	if err := sonic.Unmarshal([]byte(*raw), &value); err != nil {
		return nil
	}
	return value
}

func strValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func getUpdatedAt(aggregate v2models.GameV2Aggregate) time.Time {
	var updatedAt time.Time
	if aggregate.Details != nil && aggregate.Details.UpdatedAt.After(updatedAt) {
		updatedAt = aggregate.Details.UpdatedAt
	}
	if aggregate.Localized != nil && aggregate.Localized.UpdatedAt.After(updatedAt) {
		updatedAt = aggregate.Localized.UpdatedAt
	}
	for _, price := range aggregate.Prices {
		if price.UpdatedAt.After(updatedAt) {
			updatedAt = price.UpdatedAt
		}
	}
	for _, item := range aggregate.Media {
		if item.UpdatedAt.After(updatedAt) {
			updatedAt = item.UpdatedAt
		}
	}
	if aggregate.Requirements != nil && aggregate.Requirements.UpdatedAt.After(updatedAt) {
		updatedAt = aggregate.Requirements.UpdatedAt
	}
	for _, item := range aggregate.News {
		if item.UpdatedAt.After(updatedAt) {
			updatedAt = item.UpdatedAt
		}
	}
	if aggregate.OnlineCount != nil && aggregate.OnlineCount.CollectedAt.After(updatedAt) {
		updatedAt = aggregate.OnlineCount.CollectedAt
	}
	return updatedAt
}

func platformsText(platforms map[string]bool) string {
	if len(platforms) == 0 {
		return ""
	}
	order := []string{"windows", "mac", "linux"}
	res := make([]string, 0, len(platforms))
	for _, key := range order {
		if platforms[key] {
			res = append(res, key)
		}
	}
	for key, enabled := range platforms {
		if !enabled || containsString(res, key) {
			continue
		}
		res = append(res, key)
	}
	return strings.Join(res, ", ")
}

func syncPCRequirements(values map[string]string) v2models.GameV2SyncPCRequirements {
	return v2models.GameV2SyncPCRequirements{
		Minimum:     cleanSyncText(firstNonEmptyString(values["minimum"], values["Minimum"], values["最低配置"])),
		Recommended: cleanSyncText(firstNonEmptyString(values["recommended"], values["Recommended"], values["推荐配置"])),
	}
}

func cleanSyncText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = html.UnescapeString(value)
	value = htmlBreakRE.ReplaceAllString(value, "\n")
	value = htmlTagRE.ReplaceAllString(value, " ")
	value = strings.NewReplacer(
		"[b]", "", "[/b]", "",
		"[i]", "", "[/i]", "",
		"[u]", "", "[/u]", "",
		"[list]", "", "[/list]", "",
		"[*]", "\n",
	).Replace(value)
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(spaceRE.ReplaceAllString(line, " "))
	}
	value = strings.Join(lines, "\n")
	value = newlineRE.ReplaceAllString(value, "\n\n")
	return strings.TrimSpace(value)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func formatSyncTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}
