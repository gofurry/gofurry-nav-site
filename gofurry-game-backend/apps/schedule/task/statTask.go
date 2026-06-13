package task

import (
	"encoding/json"
	"time"

	"github.com/bytedance/sonic"
	pd "github.com/gofurry/gofurry-game-backend/apps/prize/dao"
	pm "github.com/gofurry/gofurry-game-backend/apps/prize/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
	"github.com/gofurry/gofurry-game-backend/common/util"
)

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
			subject := "GoFurry 抽奖服务-获奖"
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
