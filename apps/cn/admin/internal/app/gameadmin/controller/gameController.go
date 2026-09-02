package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	pkgmodels "github.com/gofurry/gofurry-admin/pkg/models"
	"github.com/gofurry/gofurry-admin/pkg/util"
	steam "github.com/gofurry/steam-go"
	steamassets "github.com/gofurry/steam-go/addons/assets"
	"github.com/gofurry/steam-go/web/storefront"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GameAPI struct{ store *gameStore }

func New(pool *pgxpool.Pool, auditLogger *audit.Logger) *GameAPI {
	return &GameAPI{store: newGameStore(pool, auditLogger)}
}

type steamGameAssetDTO struct {
	AppID    int64               `json:"appid"`
	Kind     string              `json:"kind"`
	URL      string              `json:"url"`
	Digest   string              `json:"digest,omitempty"`
	Filename string              `json:"filename,omitempty"`
	Source   string              `json:"source,omitempty"`
	Assets   []steamGameAssetDTO `json:"assets,omitempty"`
}

type steamGamePrefillDTO struct {
	AppID      int64               `json:"appid"`
	Name       string              `json:"name"`
	NameEn     string              `json:"name_en"`
	Info       string              `json:"info"`
	InfoEn     string              `json:"info_en"`
	Groups     []pkgmodels.KvModel `json:"groups"`
	Developers []string            `json:"developers"`
	Publishers []string            `json:"publishers"`
	Header     string              `json:"header"`
	Links      []pkgmodels.KvModel `json:"links"`
}

func (api *GameAPI) ListGames(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	total, rows, err := api.store.listGames(c.Context(), page)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	list := make([]models.GameDTO, 0, len(rows))
	for _, row := range rows {
		list = append(list, gameDTO(row))
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}

func (api *GameAPI) CreateGame(c fiber.Ctx) error {
	var req models.GamePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := validateGamePayload(req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	created, err := api.store.createGame(c.Context(), audit.MetaFromFiber(c), gameInsertParams(req))
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(gameDTO(created))
}

func (api *GameAPI) GetGame(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	row, daoErr := api.store.getGame(c.Context(), id)
	if daoErr != nil {
		return common.NewResponse(c).Error(daoErr)
	}
	return common.NewResponse(c).SuccessWithData(gameDTO(row))
}

func (api *GameAPI) GetGameWorkspace(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	game, tags, storeErr := api.store.getGameWorkspace(c.Context(), id)
	if storeErr != nil {
		return common.NewResponse(c).Error(storeErr)
	}
	return common.NewResponse(c).SuccessWithData(models.GameWorkspace{Game: gameDTO(game), Tags: tags})
}

func (api *GameAPI) UpdateGame(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.GamePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := validateGamePayload(req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	params := gameUpdateParams(req)
	params.ID = id
	_, _, txErr := api.store.updateGame(c.Context(), audit.MetaFromFiber(c), params)
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetGame(c)
}

func (api *GameAPI) DeleteGame(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := api.store.deleteGame(c.Context(), audit.MetaFromFiber(c), id); err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).Success()
}

func (api *GameAPI) ResolveSteamGameAsset(c fiber.Ctx) error {
	appid, err := strconv.ParseInt(strings.TrimSpace(c.Query("appid", "")), 10, 64)
	if err != nil || appid <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("appid must be a positive integer"))
	}

	kinds := steamAssetKinds(c.Query("kind", "header"))
	client, timeout, err := newAdminSteamClient()
	if err != nil {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("create steam client failed: %v", err)))
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	items, err := steamassets.FetchStoreItemAssetURLs(ctx, client.API.StoreBrowseService, steamassets.StoreItemAssetOptions{
		CountryCode: "CN",
		Language:    "schinese",
		Kinds:       kinds,
		StripQuery:  true,
	}, uint32(appid))
	if err != nil {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("fetch steam asset failed: %v", err)))
	}
	if len(items) == 0 {
		return common.NewResponse(c).Error(common.NewServiceError("steam asset not found"))
	}

	assets := make([]steamGameAssetDTO, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.URL) == "" {
			continue
		}
		assets = append(assets, steamGameAssetDTO{
			AppID:    int64(item.AppID),
			Kind:     string(item.Kind),
			URL:      item.URL,
			Digest:   item.Digest,
			Filename: item.Filename,
			Source:   item.Source,
		})
	}
	if len(assets) == 0 {
		return common.NewResponse(c).Error(common.NewServiceError("steam asset url is empty"))
	}

	result := assets[0]
	result.Assets = assets
	return common.NewResponse(c).SuccessWithData(result)
}

func (api *GameAPI) ResolveSteamGamePrefill(c fiber.Ctx) error {
	appid, err := strconv.ParseInt(strings.TrimSpace(c.Query("appid", "")), 10, 64)
	if err != nil || appid <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("appid must be a positive integer"))
	}

	client, timeout, err := newAdminSteamClient()
	if err != nil {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("create steam client failed: %v", err)))
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	zhData, err := fetchSteamAppDetails(ctx, client, appid, "schinese")
	if err != nil {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("fetch schinese steam app details failed: %v", err)))
	}
	enData, err := fetchSteamAppDetails(ctx, client, appid, "english")
	if err != nil {
		return common.NewResponse(c).Error(common.NewServiceError(fmt.Sprintf("fetch english steam app details failed: %v", err)))
	}

	header := strings.TrimSpace(zhData.HeaderImage)
	if header == "" {
		header = strings.TrimSpace(enData.HeaderImage)
	}
	items, assetErr := steamassets.FetchStoreItemAssetURLs(ctx, client.API.StoreBrowseService, steamassets.StoreItemAssetOptions{
		CountryCode: "CN",
		Language:    "schinese",
		Kinds:       steamAssetKinds("header"),
		StripQuery:  true,
	}, uint32(appid))
	if assetErr == nil {
		for _, item := range items {
			if value := strings.TrimSpace(item.URL); value != "" {
				header = value
				break
			}
		}
	}

	return common.NewResponse(c).SuccessWithData(steamGamePrefill(appid, zhData, enData, header))
}

func fetchSteamAppDetails(ctx context.Context, client *steam.Client, appid int64, language string) (storefront.AppDetailsData, error) {
	envelope, err := client.Web.Storefront.GetAppDetails(ctx, uint32(appid), &storefront.GetAppDetailsOptions{
		CountryCode: "CN",
		Language:    language,
	})
	if err != nil {
		return storefront.AppDetailsData{}, err
	}

	result, ok := envelope[strconv.FormatInt(appid, 10)]
	if !ok || !result.Success {
		return storefront.AppDetailsData{}, errors.New("steam app details not found")
	}
	return result.Data, nil
}

func steamGamePrefill(appid int64, zhData, enData storefront.AppDetailsData, header string) steamGamePrefillDTO {
	groups := make([]pkgmodels.KvModel, 0, 1)
	website := strings.TrimSpace(zhData.Website)
	if website == "" {
		website = strings.TrimSpace(enData.Website)
	}
	if website != "" {
		groups = append(groups, pkgmodels.KvModel{Key: "official", Value: website})
	}

	developers := normalizeStringArray(zhData.Developers)
	if len(developers) == 0 {
		developers = normalizeStringArray(enData.Developers)
	}
	publishers := normalizeStringArray(zhData.Publishers)
	if len(publishers) == 0 {
		publishers = normalizeStringArray(enData.Publishers)
	}

	return steamGamePrefillDTO{
		AppID:      appid,
		Name:       strings.TrimSpace(zhData.Name),
		NameEn:     strings.TrimSpace(enData.Name),
		Info:       strings.TrimSpace(zhData.ShortDescription),
		InfoEn:     strings.TrimSpace(enData.ShortDescription),
		Groups:     groups,
		Developers: developers,
		Publishers: publishers,
		Header:     strings.TrimSpace(header),
		Links:      normalizeGameLinks(appid, nil),
	}
}

func newAdminSteamClient() (*steam.Client, time.Duration, error) {
	cfg := env.GetServerConfig().ExternalServices.Steam
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	proxySelector, err := steamProxySelector(cfg.Proxy)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid steam proxy config: %w", err)
	}

	client, err := steam.NewClient(
		steam.WithTimeout(timeout),
		steam.WithRateLimit(cfg.RateLimit),
		steam.WithProxySelector(proxySelector),
	)
	if err != nil {
		return nil, 0, err
	}
	return client, timeout, nil
}

func steamProxySelector(raw string) (steam.ProxySelector, error) {
	proxies := splitSteamProxyURLs(raw)
	if len(proxies) == 0 {
		return nil, nil
	}
	if len(proxies) == 1 {
		return steam.NewStaticProxySelector(proxies[0])
	}
	return steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{
			FailureThreshold: 2,
			Cooldown:         5 * time.Minute,
		},
		proxies...,
	)
}

func splitSteamProxyURLs(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t'
	})
	proxies := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			proxies = append(proxies, part)
		}
	}
	return proxies
}

func (api *GameAPI) ListComments(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	total, rows, err := api.store.listComments(c.Context(), page)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, rows))
}

func (api *GameAPI) CreateComment(c fiber.Ctx) error {
	var req models.GameCommentPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Content) == "" || req.GameID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("content and game_id are required"))
	}
	row, txErr := api.store.createComment(c.Context(), audit.MetaFromFiber(c), gamesqlc.InsertGameCommentParams{
		ID:      util.GenerateId(),
		Region:  strings.TrimSpace(req.Region),
		Content: strings.TrimSpace(req.Content),
		Score:   req.Score,
		GameID:  req.GameID,
		Ip:      strings.TrimSpace(req.IP),
		Name:    strings.TrimSpace(req.Name),
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return common.NewResponse(c).SuccessWithData(row)
}

func (api *GameAPI) GetComment(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	row, daoErr := api.store.getComment(c.Context(), id)
	if daoErr != nil {
		return common.NewResponse(c).Error(daoErr)
	}
	return common.NewResponse(c).SuccessWithData(row)
}

func (api *GameAPI) UpdateComment(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.GameCommentPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Content) == "" || req.GameID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("content and game_id are required"))
	}
	txErr := api.store.updateComment(c.Context(), audit.MetaFromFiber(c), gamesqlc.UpdateGameCommentParams{
		ID: id, Region: strings.TrimSpace(req.Region), Content: strings.TrimSpace(req.Content), Score: req.Score,
		GameID: req.GameID, Ip: strings.TrimSpace(req.IP), Name: strings.TrimSpace(req.Name),
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetComment(c)
}

func (api *GameAPI) DeleteComment(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := api.store.deleteComment(c.Context(), audit.MetaFromFiber(c), id); err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).Success()
}

func (api *GameAPI) ListPrizes(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	total, rows, err := api.store.listPrizes(c.Context(), page)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	list := make([]models.PrizeDTO, 0, len(rows))
	for _, row := range rows {
		list = append(list, prizeDTO(row))
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, list))
}

func (api *GameAPI) CreatePrize(c fiber.Ctx) error {
	var req models.PrizePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	startTime, endTime, valErr := parsePrizeTimes(req)
	if valErr != nil {
		return common.NewResponse(c).Error(valErr)
	}
	created, err := api.store.createPrize(c.Context(), audit.MetaFromFiber(c), gamesqlc.InsertPrizeParams{
		Title: strings.TrimSpace(req.Title), Description: strings.TrimSpace(req.Desc), Prize: []byte(adminutil.MustJSON(normalizePrizeBody(req.Prize))),
		Key: strings.TrimSpace(req.Key), StartTime: gameTimestamp(startTime), EndTime: gameTimestamp(endTime), Status: req.Status,
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(prizeDTO(created))
}

func (api *GameAPI) GetPrize(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	row, daoErr := api.store.getPrize(c.Context(), id)
	if daoErr != nil {
		return common.NewResponse(c).Error(daoErr)
	}
	return common.NewResponse(c).SuccessWithData(prizeDTO(row))
}

func (api *GameAPI) UpdatePrize(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.PrizePayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	startTime, endTime, valErr := parsePrizeTimes(req)
	if valErr != nil {
		return common.NewResponse(c).Error(valErr)
	}
	txErr := api.store.updatePrize(c.Context(), audit.MetaFromFiber(c), gamesqlc.UpdatePrizeParams{
		ID: id, Title: strings.TrimSpace(req.Title), Description: strings.TrimSpace(req.Desc), Prize: []byte(adminutil.MustJSON(normalizePrizeBody(req.Prize))),
		Key: strings.TrimSpace(req.Key), StartTime: gameTimestamp(startTime), EndTime: gameTimestamp(endTime), Status: req.Status,
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetPrize(c)
}

func (api *GameAPI) DeletePrize(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := api.store.deletePrize(c.Context(), audit.MetaFromFiber(c), id); err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).Success()
}

func (api *GameAPI) ListTags(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	total, rows, err := api.store.listTags(c.Context(), page)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, rows))
}

func (api *GameAPI) CreateTag(c fiber.Ctx) error {
	var req models.TagPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.ID <= 0 || strings.TrimSpace(req.Name) == "" {
		return common.NewResponse(c).Error(common.NewValidationError("id and name are required"))
	}
	row, txErr := api.store.createTag(c.Context(), audit.MetaFromFiber(c), gamesqlc.InsertTagParams{
		ID:     req.ID,
		Name:   strings.TrimSpace(req.Name),
		NameEn: strings.TrimSpace(req.NameEn),
		Info:   strings.TrimSpace(req.Info),
		InfoEn: strings.TrimSpace(req.InfoEn),
		Prefix: req.Prefix,
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return common.NewResponse(c).SuccessWithData(row)
}

func (api *GameAPI) GetTag(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	row, daoErr := api.store.getTag(c.Context(), id)
	if daoErr != nil {
		return common.NewResponse(c).Error(daoErr)
	}
	return common.NewResponse(c).SuccessWithData(row)
}

func (api *GameAPI) UpdateTag(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.TagPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if strings.TrimSpace(req.Name) == "" {
		return common.NewResponse(c).Error(common.NewValidationError("name is required"))
	}
	txErr := api.store.updateTag(c.Context(), audit.MetaFromFiber(c), gamesqlc.UpdateTagParams{
		ID: id, Name: strings.TrimSpace(req.Name), NameEn: strings.TrimSpace(req.NameEn), Info: strings.TrimSpace(req.Info), InfoEn: strings.TrimSpace(req.InfoEn), Prefix: req.Prefix,
	})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetTag(c)
}

func (api *GameAPI) DeleteTag(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := api.store.deleteTag(c.Context(), audit.MetaFromFiber(c), id); err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).Success()
}

func (api *GameAPI) ListTagMaps(c fiber.Ctx) error {
	page := adminutil.ParsePageQuery(c)
	total, rows, err := api.store.listTagMaps(c.Context(), page)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(adminutil.BuildPageResponse(total, rows))
}

func (api *GameAPI) CreateTagMap(c fiber.Ctx) error {
	var req models.TagMapPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.GameID <= 0 || req.TagID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("game_id and tag_id are required"))
	}
	created, err := api.store.createTagMap(c.Context(), audit.MetaFromFiber(c), req.GameID, req.TagID)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(created)
}

func (api *GameAPI) GetTagMap(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	row, daoErr := api.store.getTagMap(c.Context(), id)
	if daoErr != nil {
		return common.NewResponse(c).Error(daoErr)
	}
	return common.NewResponse(c).SuccessWithData(row)
}

func (api *GameAPI) UpdateTagMap(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var req models.TagMapPayload
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.GameID <= 0 || req.TagID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("game_id and tag_id are required"))
	}
	txErr := api.store.updateTagMap(c.Context(), audit.MetaFromFiber(c), gamesqlc.UpdateTagMapParams{ID: id, GameID: req.GameID, TagID: req.TagID})
	if txErr != nil {
		return common.NewResponse(c).Error(txErr)
	}
	return api.GetTagMap(c)
}

func (api *GameAPI) DeleteTagMap(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := api.store.deleteTagMap(c.Context(), audit.MetaFromFiber(c), id); err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).Success()
}

func (api *GameAPI) BulkReplaceTagMaps(c fiber.Ctx) error {
	var req adminutil.BulkReplaceRequest
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.OwnerID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("owner_id is required"))
	}
	req.IDs = uniqueInt64s(req.IDs)
	err := api.store.replaceGameTags(c.Context(), audit.MetaFromFiber(c), req.OwnerID, req.IDs)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).Success()
}

func (api *GameAPI) ListTagMapGameIDs(c fiber.Ctx) error {
	tagID, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	gameIDs, daoErr := api.store.listGameIDsByTag(c.Context(), tagID)
	if daoErr != nil {
		return common.NewResponse(c).Error(daoErr)
	}
	return common.NewResponse(c).SuccessWithData(gameIDs)
}

func (api *GameAPI) BulkReplaceTagGameMaps(c fiber.Ctx) error {
	var req adminutil.BulkReplaceRequest
	if err := adminutil.DecodeBody(c, &req); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if req.OwnerID <= 0 {
		return common.NewResponse(c).Error(common.NewValidationError("owner_id is required"))
	}
	req.IDs = uniqueInt64s(req.IDs)

	err := api.store.replaceTagGames(c.Context(), audit.MetaFromFiber(c), req.OwnerID, req.IDs)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).Success()
}

func validateGamePayload(req models.GamePayload) common.Error {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.NameEn) == "" {
		return common.NewValidationError("name and name_en are required")
	}
	return nil
}

func gameInsertParams(req models.GamePayload) gamesqlc.InsertGameParams {
	return gamesqlc.InsertGameParams{
		Name: strings.TrimSpace(req.Name), NameEn: strings.TrimSpace(req.NameEn), Info: strings.TrimSpace(req.Info), InfoEn: strings.TrimSpace(req.InfoEn),
		Resources: []byte(adminutil.MustJSON(normalizeKV(req.Resources))), Groups: []byte(adminutil.MustJSON(normalizeKV(req.Groups))),
		Developers: []byte(adminutil.MustJSON(normalizeStringArray(req.Developers))),
		Publishers: []byte(adminutil.MustJSON(normalizeStringArray(req.Publishers))), Appid: req.Appid, Header: strings.TrimSpace(req.Header),
		Links: []byte(adminutil.MustJSON(normalizeGameLinks(req.Appid, req.Links))), Weight: req.Weight, PrimaryTag: req.PrimaryTag, SecondaryTag: req.SecondaryTag,
	}
}

func gameUpdateParams(req models.GamePayload) gamesqlc.UpdateGameParams {
	insert := gameInsertParams(req)
	return gamesqlc.UpdateGameParams{
		Name: insert.Name, NameEn: insert.NameEn, Info: insert.Info, InfoEn: insert.InfoEn, Resources: insert.Resources, Groups: insert.Groups,
		Developers: insert.Developers, Publishers: insert.Publishers, Appid: insert.Appid, Header: insert.Header,
		Links: insert.Links, Weight: insert.Weight, PrimaryTag: insert.PrimaryTag, SecondaryTag: insert.SecondaryTag,
	}
}

func gameDTO(row models.Game) models.GameDTO {
	return models.GameDTO{
		ID:           row.ID,
		Name:         row.Name,
		NameEn:       row.NameEn,
		Info:         row.Info,
		InfoEn:       row.InfoEn,
		CreateTime:   row.CreateTime,
		UpdateTime:   row.UpdateTime,
		Resources:    adminutil.ParseKVArray(row.Resources),
		Groups:       adminutil.ParseKVArray(row.Groups),
		Developers:   adminutil.ParseStringArray(row.Developers),
		Publishers:   adminutil.ParseStringArray(row.Publishers),
		Appid:        row.Appid,
		Header:       row.Header,
		Links:        adminutil.ParseKVArray(row.Links),
		Weight:       row.Weight,
		PrimaryTag:   row.PrimaryTag,
		SecondaryTag: row.SecondaryTag,
	}
}

func prizeDTO(row models.Prize) models.PrizeDTO {
	var prize models.PrizeBody
	_ = json.Unmarshal([]byte(strings.TrimSpace(row.Prize)), &prize)
	return models.PrizeDTO{
		ID:         row.ID,
		Title:      row.Title,
		Desc:       row.Desc,
		Prize:      prize,
		Key:        row.Key,
		StartTime:  row.StartTime,
		EndTime:    row.EndTime,
		CreateTime: row.CreateTime,
		Status:     row.Status,
	}
}

func normalizeStringArray(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func normalizeKV(items []pkgmodels.KvModel) []pkgmodels.KvModel {
	result := make([]pkgmodels.KvModel, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		value := strings.TrimSpace(item.Value)
		if key == "" && value == "" {
			continue
		}
		result = append(result, pkgmodels.KvModel{Key: key, Value: value})
	}
	return result
}

func normalizeGameLinks(appid int64, items []pkgmodels.KvModel) []pkgmodels.KvModel {
	result := normalizeKV(items)
	if appid <= 0 {
		return result
	}

	defaults := []pkgmodels.KvModel{
		{Key: "steamdb", Value: fmt.Sprintf("https://steamdb.info/app/%d/", appid)},
		{Key: "gamalytic", Value: fmt.Sprintf("https://gamalytic.com/game/%d", appid)},
	}
	indexByKey := make(map[string]int, len(result))
	for index, item := range result {
		key := strings.ToLower(strings.TrimSpace(item.Key))
		if key == "" {
			continue
		}
		if _, exists := indexByKey[key]; !exists {
			indexByKey[key] = index
		}
	}

	for _, item := range defaults {
		if index, exists := indexByKey[item.Key]; exists {
			if strings.TrimSpace(result[index].Value) == "" {
				result[index].Key = item.Key
				result[index].Value = item.Value
			}
			continue
		}
		result = append(result, item)
	}
	return result
}

func steamAssetKinds(kind string) []steamassets.Kind {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "library", "library_cover", "library_capsule":
		return []steamassets.Kind{
			steamassets.KindLibraryCapsule,
			steamassets.KindLibraryCapsule2x,
		}
	case "capsule", "capsule_main":
		return []steamassets.Kind{
			steamassets.KindCapsuleMain,
			steamassets.KindCapsuleMain2x,
			steamassets.KindHeroCapsule,
		}
	case "hero", "library_hero":
		return []steamassets.Kind{
			steamassets.KindLibraryHero,
			steamassets.KindLibraryHero2x,
			steamassets.KindHeroCapsule,
			steamassets.KindHeroCapsule2x,
		}
	case "header_2x":
		return []steamassets.Kind{
			steamassets.KindHeader2x,
			steamassets.KindHeader,
		}
	case "header":
		fallthrough
	default:
		return []steamassets.Kind{
			steamassets.KindHeader,
			steamassets.KindHeader2x,
		}
	}
}

func normalizePrizeBody(body models.PrizeBody) models.PrizeBody {
	body.Keys = normalizeStringArray(body.Keys)
	body.Title = strings.TrimSpace(body.Title)
	body.Platform = strings.TrimSpace(body.Platform)
	return body
}

func parsePrizeTimes(req models.PrizePayload) (time.Time, time.Time, common.Error) {
	if strings.TrimSpace(req.Title) == "" {
		return time.Time{}, time.Time{}, common.NewValidationError("title is required")
	}
	start, err := parseDateTime(req.StartTime)
	if err != nil {
		return time.Time{}, time.Time{}, common.NewValidationError("invalid start_time")
	}
	end, err := parseDateTime(req.EndTime)
	if err != nil {
		return time.Time{}, time.Time{}, common.NewValidationError("invalid end_time")
	}
	return start, end, nil
}

func parseDateTime(value string) (time.Time, error) {
	layouts := []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02 15:04:05", "2006-01-02 15:04"}
	value = strings.TrimSpace(value)
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fiber.ErrBadRequest
}

func uniqueInt64s(input []int64) []int64 {
	result := make([]int64, 0, len(input))
	seen := make(map[int64]struct{}, len(input))
	for _, item := range input {
		if item <= 0 {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
