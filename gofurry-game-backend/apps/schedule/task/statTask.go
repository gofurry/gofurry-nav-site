package task

import (
	"encoding/json"
	"time"

	gd "github.com/gofurry/gofurry-game-backend/apps/game/dao"
	gm "github.com/gofurry/gofurry-game-backend/apps/game/models"
	pd "github.com/gofurry/gofurry-game-backend/apps/prize/dao"
	pm "github.com/gofurry/gofurry-game-backend/apps/prize/models"
	rd "github.com/gofurry/gofurry-game-backend/apps/review/dao"
	rm "github.com/gofurry/gofurry-game-backend/apps/review/models"
	"github.com/gofurry/gofurry-game-backend/apps/schedule/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
	"github.com/gofurry/gofurry-game-backend/common/util"
	"github.com/bytedance/sonic"
)

func UpdateMainInfoCache() {
	log.Info("StatTask UpdateMainInfoCache 开始...")

	info := map[string]any{
		"latest": LatestInfo{models.InfoModel{Key: "game-info:latest", Num: 8, Duration: 3 * time.Hour}},
		"recent": RecentInfo{models.InfoModel{Key: "game-info:recent", Num: 8, Duration: 3 * time.Hour}},
		"free":   FreeInfo{models.InfoModel{Key: "game-info:free", Num: 8, Duration: 3 * time.Hour}},
		"hot":    HotInfo{models.InfoModel{Key: "game-info:hot", Num: 8, Duration: 3 * time.Hour}},
	}
	for k, v := range info {
		switch k {
		case "latest":
			latestInfo := v.(LatestInfo)
			latestInfo.cacheGameInfo()
		case "recent":
			recentInfo := v.(RecentInfo)
			recentInfo.cacheGameInfo()
		case "free":
			freeInfo := v.(FreeInfo)
			freeInfo.cacheGameInfo()
		case "hot":
			hotInfo := v.(HotInfo)
			hotInfo.cacheGameInfo()
		}
	}
	log.Info("StatTask UpdateMainInfoCache 结束...")
}

type HotInfo struct {
	models.InfoModel
}

func (r *HotInfo) cacheGameInfo() common.GFError {
	res, err := rd.GetReviewDao().GetHotGame(r.Num)
	if err != nil {
		return err
	}
	if idList, jsonErr := sonic.Marshal(res); jsonErr == nil {
		cs.SetExpire(r.Key, string(idList), r.Duration)
	}
	return nil
}

type FreeInfo struct {
	models.InfoModel
}

func (r *FreeInfo) cacheGameInfo() common.GFError {
	res, err := gd.GetGameDao().GetFreeGame(r.Num)
	if err != nil {
		return err
	}
	infoRecord := []rm.AvgScoreResult{}
	for _, v := range res {
		newRecord, gfError := rd.GetReviewDao().GetScoreById(v)
		if gfError != nil && gfError.GetMsg() == "record not found" {
			newRecord = rm.AvgScoreResult{GameID: util.Int642String(v), AvgScore: 0.0, CommentCount: 0}
			gameRecord := gm.GfgGame{}
			gd.GetGameDao().GetById(v, &gameRecord)
			newRecord.Name = gameRecord.Name
			newRecord.NameEn = gameRecord.NameEn
			newRecord.Info = gameRecord.Info
			newRecord.InfoEn = gameRecord.InfoEn
			newRecord.Header = gameRecord.Header
		} else if gfError != nil {
			return gfError
		}
		infoRecord = append(infoRecord, newRecord)
	}
	if idList, jsonErr := sonic.Marshal(infoRecord); jsonErr == nil {
		cs.SetExpire(r.Key, string(idList), r.Duration)
	}
	return nil
}

type RecentInfo struct {
	models.InfoModel
}

func (r *RecentInfo) cacheGameInfo() common.GFError {
	res, err := gd.GetGameDao().GetRecentGame(r.Num)
	if err != nil {
		return err
	}
	infoRecord := []rm.AvgScoreResult{}
	for _, v := range res {
		newRecord, gfError := rd.GetReviewDao().GetScoreById(v)
		if gfError != nil && gfError.GetMsg() == "record not found" {
			newRecord = rm.AvgScoreResult{GameID: util.Int642String(v), AvgScore: 0.0, CommentCount: 0}
			gameRecord := gm.GfgGame{}
			gd.GetGameDao().GetById(v, &gameRecord)
			newRecord.Name = gameRecord.Name
			newRecord.NameEn = gameRecord.NameEn
			newRecord.Info = gameRecord.Info
			newRecord.InfoEn = gameRecord.InfoEn
			newRecord.Header = gameRecord.Header
		} else if gfError != nil {
			return gfError
		}
		infoRecord = append(infoRecord, newRecord)
	}
	if idList, jsonErr := sonic.Marshal(infoRecord); jsonErr == nil {
		cs.SetExpire(r.Key, string(idList), r.Duration)
	}
	return nil
}

type LatestInfo struct {
	models.InfoModel
}

func (l *LatestInfo) cacheGameInfo() common.GFError {
	res, err := gd.GetGameDao().GetLatestGame(l.Num)
	if err != nil {
		return err
	}
	infoRecord := []rm.AvgScoreResult{}
	for _, v := range res {
		newRecord, gfError := rd.GetReviewDao().GetScoreById(v)
		if gfError != nil && gfError.GetMsg() == "record not found" {
			newRecord = rm.AvgScoreResult{GameID: util.Int642String(v), AvgScore: 0.0, CommentCount: 0}
			gameRecord := gm.GfgGame{}
			gd.GetGameDao().GetById(v, &gameRecord)
			newRecord.Name = gameRecord.Name
			newRecord.NameEn = gameRecord.NameEn
			newRecord.Info = gameRecord.Info
			newRecord.InfoEn = gameRecord.InfoEn
			newRecord.Header = gameRecord.Header
		} else if gfError != nil {
			return gfError
		}
		infoRecord = append(infoRecord, newRecord)
	}
	if idList, jsonErr := sonic.Marshal(infoRecord); jsonErr == nil {
		cs.SetExpire(l.Key, string(idList), l.Duration)
	}
	return nil
}

func UpdateGamePanelCache() {
	log.Info("StatTask UpdateGamePanelCache 开始...")
	// 在线人数 [1,15]
	record, err := gd.GetGameDao().GetPlayerPeak(15, 0)
	if err != nil {
		log.Error("GetPlayerPeak 1st err:", err)
		return
	}
	if jsonRecord, jsonErr := sonic.Marshal(record); jsonErr == nil {
		cs.SetExpire("game-panel:top-player-count-1st", string(jsonRecord), 3*time.Hour)
	}
	// 在线人数 [16,30]
	record, err = gd.GetGameDao().GetPlayerPeak(15, 15)
	if err != nil {
		log.Error("GetPlayerPeak 2st err:", err)
		return
	}
	if jsonRecord, jsonErr := sonic.Marshal(record); jsonErr == nil {
		cs.SetExpire("game-panel:top-player-count-2st", string(jsonRecord), 3*time.Hour)
	}
	// 在线人数 [31,45]
	record, err = gd.GetGameDao().GetPlayerPeak(15, 30)
	if err != nil {
		log.Error("GetPlayerPeak 3st err:", err)
		return
	}
	if jsonRecord, jsonErr := sonic.Marshal(record); jsonErr == nil {
		cs.SetExpire("game-panel:top-player-count-3st", string(jsonRecord), 3*time.Hour)
	}
	// 在线人数 [46,60]
	record, err = gd.GetGameDao().GetPlayerPeak(15, 45)
	if err != nil {
		log.Error("GetPlayerPeak 4st err:", err)
		return
	}
	if jsonRecord, jsonErr := sonic.Marshal(record); jsonErr == nil {
		cs.SetExpire("game-panel:top-player-count-4st", string(jsonRecord), 3*time.Hour)
	}
	// 最高售价
	priceRecord, err := gd.GetGameDao().GetTopPrice(15)
	if err != nil {
		log.Error("GetTopPrice err:", err)
		return
	}
	if jsonRecord, jsonErr := sonic.Marshal(priceRecord); jsonErr == nil {
		cs.SetExpire("game-panel:top-price", string(jsonRecord), 3*time.Hour)
	}
	// 最低售价 >7USD
	priceRecord, err = gd.GetGameDao().GetLowestPrice(1000, 15)
	if err != nil {
		log.Error("GetBottomPrice>10USD err:", err)
		return
	}
	if jsonRecord, jsonErr := sonic.Marshal(priceRecord); jsonErr == nil {
		cs.SetExpire("game-panel:bottom-price-1st", string(jsonRecord), 3*time.Hour)
	}
	// 最低售价 >10USD
	priceRecord, err = gd.GetGameDao().GetLowestPrice(1500, 15)
	if err != nil {
		log.Error("GetBottomPrice>15USD err:", err)
		return
	}
	if jsonRecord, jsonErr := sonic.Marshal(priceRecord); jsonErr == nil {
		cs.SetExpire("game-panel:bottom-price-2st", string(jsonRecord), 3*time.Hour)
	}
	// 最低售价 >15USD
	priceRecord, err = gd.GetGameDao().GetLowestPrice(2000, 15)
	if err != nil {
		log.Error("GetBottomPrice>10USD err:", err)
		return
	}
	if jsonRecord, jsonErr := sonic.Marshal(priceRecord); jsonErr == nil {
		cs.SetExpire("game-panel:bottom-price-3st", string(jsonRecord), 3*time.Hour)
	}
	// 最低售价 >20USD
	priceRecord, err = gd.GetGameDao().GetLowestPrice(2500, 15)
	if err != nil {
		log.Error("GetBottomPrice>25USD err:", err)
		return
	}
	if jsonRecord, jsonErr := sonic.Marshal(priceRecord); jsonErr == nil {
		cs.SetExpire("game-panel:bottom-price-4st", string(jsonRecord), 3*time.Hour)
	}
	// 最高折扣
	priceRecord, err = gd.GetGameDao().GetHighestDiscount(15)
	if err != nil {
		log.Error("GetHighestDiscount err:", err)
		return
	}
	if jsonRecord, jsonErr := sonic.Marshal(priceRecord); jsonErr == nil {
		cs.SetExpire("game-panel:top-discount", string(jsonRecord), 3*time.Hour)
	}

	log.Info("StatTask UpdateGamePanelCache 结束...")
}

func UpdateGameNewsCache() {
	log.Info("StatTask UpdateGameNewsCache 开始...")

	newRecord := gm.UpdateNewsVo{}

	// 最新新闻
	record, err := gd.GetGameDao().GetUpdateNews(15, "zh", "300")
	if err != nil {
		log.Error("GetUpdateNews err:", err)
		return
	}
	newRecord.NewsZh = record

	record, err = gd.GetGameDao().GetUpdateNews(15, "en", "300")
	if err != nil {
		log.Error("GetUpdateNews err:", err)
		return
	}
	newRecord.NewsEn = record

	if jsonRecord, jsonErr := sonic.Marshal(newRecord); jsonErr == nil {
		cs.SetExpire("game-news:latest", string(jsonRecord), 3*time.Hour)
	}

	log.Info("StatTask UpdateGameNewsCache 结束...")
}

func UpdateGameCreatorCache() {
	log.Info("StatTask UpdateGameCreatorCache 开始...")

	newRecord := gm.UpdateCreatorVo{}

	records, err := gd.GetGameCreatorDao().GetGameCreator("zh")
	if err != nil {
		log.Error("GetGameCreator err:", err)
		return
	}
	res, jsonErr := parseGameCreator(records)
	if jsonErr != nil {
		log.Error("GetGameCreator err:", jsonErr)
		return
	}
	newRecord.CreatorZh = res

	records, err = gd.GetGameCreatorDao().GetGameCreator("en")
	if err != nil {
		log.Error("GetGameCreator err:", err)
		return
	}
	res, jsonErr = parseGameCreator(records)
	if jsonErr != nil {
		log.Error("GetGameCreator err:", jsonErr)
		return
	}
	newRecord.CreatorEn = res

	if jsonRecord, jsonErr := sonic.Marshal(newRecord); jsonErr == nil {
		cs.SetExpire("game-creator:list", string(jsonRecord), 12*time.Hour)
	}

	log.Info("StatTask UpdateGameCreatorCache 结束...")
}

func parseGameCreator(records []gm.TempCreator) (res []gm.CreatorVo, err error) {
	for _, r := range records {
		vo := gm.CreatorVo{
			ID:         util.Int642String(r.ID),
			Name:       r.Name,
			Info:       r.Info,
			URL:        r.URL,
			Avatar:     r.Avatar,
			Type:       r.Type,
			CreateTime: r.CreateTime,
			UpdateTime: r.UpdateTime,
		}

		// 解析 Links JSON 字段
		if r.Links != nil && *r.Links != "" {
			var links []cm.KvModel
			if jsonErr := json.Unmarshal([]byte(*r.Links), &links); jsonErr != nil {
				return res, jsonErr
			} else {
				vo.Links = links
			}
		}

		// 解析 Contact JSON 字段
		if r.Contact != nil && *r.Contact != "" {
			var contact []cm.KvModel
			if jsonErr := json.Unmarshal([]byte(*r.Contact), &contact); jsonErr != nil {
				return res, jsonErr
			} else {
				vo.Contact = contact
			}
		}

		res = append(res, vo)
	}
	return
}

func UpdateMoreGameNewsCache() {
	log.Info("StatTask UpdateMoreGameNewsCache 开始...")

	newRecord := gm.UpdateNewsVo{}

	// 最新新闻
	record, err := gd.GetGameDao().GetUpdateNews(100, "zh", "1000")
	if err != nil {
		log.Error("GetUpdateMoreNews err:", err)
		return
	}
	newRecord.NewsZh = record

	record, err = gd.GetGameDao().GetUpdateNews(100, "en", "1000")
	if err != nil {
		log.Error("GetUpdateMoreNews err:", err)
		return
	}
	newRecord.NewsEn = record

	if jsonRecord, jsonErr := sonic.Marshal(newRecord); jsonErr == nil {
		cs.SetExpire("game-news:latest-more", string(jsonRecord), 3*time.Hour)
	}

	log.Info("StatTask UpdateMoreGameNewsCache 结束...")
}

func UpdatePrizeWinner() {
	log.Info("StatTask UpdatePrizeWinner 开始...")

	prizeList, err := pd.GetPrizeDao().GetActivePrizeList()
	if err != nil {
		if err.GetMsg() != common.RETURN_RECORD_NOT_FOUND {
			log.Error("GetUpdateMoreNews err:", err)
		}
		return
	}

	now := time.Now().UTC().Add(8 * time.Hour)
	for _, v := range prizeList {
		if now.After(time.Time(v.EndTime).UTC()) {
			newRecord := v
			performLottery(newRecord)
		}
	}

	// 缓存往期抽奖记录
	cachePrizeWinner()

	log.Info("StatTask UpdatePrizeWinner 结束...")
}

func performLottery(record pm.GfgPrize) {

	// 获取参与者
	members, err := pd.GetPrizeDao().GetMembers(record.ID)
	if err != nil {
		log.Error("GetMembers err:", err)
		return
	}
	if len(members) == 0 {
		log.Info("No participants for prize:", record.ID)
		return
	}

	// 获取奖品列表
	var prizeRecord pm.PrizeModel
	jsonErr := json.Unmarshal([]byte(record.Prize), &prizeRecord)
	if jsonErr != nil {
		log.Error("PrizeModel Unmarshal err:", jsonErr)
		return
	}
	prizes := prizeRecord.Keys // 假设是 []string
	winnerCount := len(prizes)
	if winnerCount > len(members) {
		winnerCount = len(members)
	}

	// 抽奖
	winners := make([]pm.GfgPrizeMember, 0, winnerCount)
	selected := map[int]bool{}
	for len(winners) < winnerCount {
		i := util.CryptoRandInt(len(members))
		if selected[i] {
			continue
		}
		selected[i] = true
		member := members[i]
		member.IsWinner = true
		member.PrizeKey = &prizes[len(winners)]
		winners = append(winners, member)
	}

	// 更新数据库
	for _, winner := range winners {
		_, err = pd.GetPrizeDao().Update(winner.ID, winner)
		if err != nil {
			log.Error("UpdateWinner err:", err)
			continue
		}
	}

	// 异步发送中奖通知
	for _, winner := range winners {
		go func(m pm.GfgPrizeMember) {
			subject := "gofurry 抽奖服务-获奖"
			body := "您已中奖，奖品为 [" + prizeRecord.Platform + "] 平台的 [" +
				prizeRecord.Title + "] 请不要忘记自行兑换~"

			if sendErr := cs.GetEmailService().SendLotteryEmail(m.Email, subject, *m.PrizeKey, body); sendErr != nil {
				log.Error("Send email err:", err)
			}
		}(winner)
	}

	record.Status = false
	retry := 0
	for retry < 3 {
		retry++
		_, err = pd.GetPrizeDao().Save(record.ID, record) // 全量更新, GORM的updates隐式忽略false的字段
		if err != nil {
			log.Error("UpdatePrizeStatus err:", err)
		} else {
			break
		}
	}
}

func cachePrizeWinner() {
	log.Info("StatTask CachePrizeWinner 开始...")

	records, err := pd.GetPrizeDao().GetLotteryHistory()
	if err != nil {
		log.Error("GetLotteryHistory err:", err)
		return
	}

	cacheRecords := make([]pm.PrizeCacheSaveModel, 0, len(records))

	for _, v := range records {

		// 反序列化 prize json
		var prizeModels pm.PrizeModel
		if jsonErr := sonic.Unmarshal([]byte(v.Prize), &prizeModels); jsonErr != nil {
			log.Error("prize json 解析失败:", jsonErr)
			continue
		}

		// 不缓存 keys
		var prizeDisplay struct {
			Title    string `json:"title"`
			Platform string `json:"platform"`
		}
		prizeDisplay.Title = prizeModels.Title
		prizeDisplay.Platform = prizeModels.Platform

		// 查询中奖者
		winners, wErr := pd.GetPrizeDao().GetWinners(v.ID)
		if wErr != nil {
			log.Error("GetWinners err:", wErr)
			continue
		}

		winnerCache := make([]pm.WinnerCacheModel, 0, len(winners))
		for _, w := range winners {
			winnerCache = append(winnerCache, pm.WinnerCacheModel{
				Name:  w.Name,
				Email: util.MaskEmail(w.Email),
			})
		}

		// 查询参与人数
		count, cErr := pd.GetPrizeDao().GetMemberCount(v.ID)
		if cErr != nil {
			log.Error("GetMemberCount err:", cErr)
			continue
		}

		newRecord := pm.PrizeCacheSaveModel{
			Name:    v.Title,
			Desc:    v.Desc,
			EndTime: v.EndTime,
			Winner:  winnerCache,
			Count:   int(count),
		}
		newRecord.Prize.Title = prizeDisplay.Title
		newRecord.Prize.Platform = prizeDisplay.Platform
		newRecord.Prize.Count = len(prizeModels.Keys)

		cacheRecords = append(cacheRecords, newRecord)
	}

	cacheModel := pm.PrizeWinnerCacheSaveModel{
		Prize:      cacheRecords,
		PrizeCount: len(cacheRecords),
	}

	jData, jsonErr := sonic.Marshal(cacheModel)
	if jsonErr != nil {
		log.Error("sonic.Marshal err:", jsonErr)
		return
	}

	cs.SetExpire("prize:history", jData, 144*time.Hour)

	log.Info("StatTask CachePrizeWinner 结束...")
}
