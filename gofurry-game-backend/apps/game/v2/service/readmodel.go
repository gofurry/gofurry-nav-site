package service

import (
	"context"
	"strconv"
	"strings"
	"time"

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
)

type gameDetailReader interface {
	GetGameDetailAggregate(ctx context.Context, query v2models.GameV2DetailQuery) (v2models.GameV2Aggregate, common.GFError)
	ListGameAggregates(ctx context.Context, query v2models.GameV2ListQuery) ([]v2models.GameV2Aggregate, common.GFError)
	GetGameNews(ctx context.Context, query v2models.GameV2NewsQuery) ([]v2models.GameV2NewsRow, common.GFError)
	GetLatestGameNews(ctx context.Context, query v2models.GameV2NewsQuery) ([]v2models.GameV2NewsRow, common.GFError)
}

type ReadModelService struct {
	reader gameDetailReader
}

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

func buildListItem(aggregate v2models.GameV2Aggregate, lang string, region string) v2models.GameV2ListItem {
	detail := buildDetailReadModel(aggregate, lang, region)
	return v2models.GameV2ListItem{
		ID:          detail.ID,
		AppID:       detail.AppID,
		Name:        detail.Name,
		Summary:     detail.Summary,
		HeaderURL:   detail.HeaderURL,
		CapsuleURL:  detail.Media.CapsuleURL,
		ReleaseDate: detail.Release.Date,
		Developers:  detail.Developers,
		Publishers:  detail.Publishers,
		Platforms:   detail.Platforms,
		Price:       detail.Price,
		OnlineCount: detail.OnlineCount,
		Tags:        detail.Tags,
		UpdatedAt:   detail.UpdatedAt,
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
