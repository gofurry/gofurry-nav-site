package service

import (
	"context"
	"net"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/apps/prize/dao"
	"github.com/gofurry/gofurry-game-backend/apps/prize/models"
	"github.com/gofurry/gofurry-game-backend/common"
	ca "github.com/gofurry/gofurry-game-backend/common/abstract"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
	"github.com/gofurry/gofurry-game-backend/common/util"
	"github.com/google/uuid"
)

type prizeService struct{}

var prizeSingleton = new(prizeService)

func GetPrizeService() *prizeService { return prizeSingleton }

func (s prizeService) PrizeParticipation(req models.PrizeParticipationRequest, c fiber.Ctx) common.GFError {
	// 入参校验
	reqErr := ca.ValidateServiceApi.Validate(req)
	if reqErr != nil {
		return common.NewServiceError("入参有误: " + reqErr[0].ErrMsg)
	}

	ip := util.GetClientIP(c)
	if ip == "" {
		ip = "unknown"
	} else {
		if parsedIP := net.ParseIP(ip); parsedIP == nil {
			ip = "invalid"
		}
	}

	emailLockKey := "prize:email:" + util.Int642String(req.ID) + ":" + req.Email
	data, gfsError := cs.GetString(emailLockKey)
	if gfsError != nil {
		return gfsError
	} else if data == "1" {
		return common.NewServiceError("请不要频繁请求邮箱服务")
	}
	var prizeRecord models.GfgPrize
	gfsError = dao.GetPrizeDao().GetById(req.ID, &prizeRecord)
	if gfsError != nil {
		return common.NewServiceError("抽奖活动不存在")
	}
	if !prizeRecord.Status {
		return common.NewServiceError("抽奖活动未开放")
	}
	now := time.Now().UTC().Add(8 * time.Hour)
	if now.Before(time.Time(prizeRecord.StartTime).UTC()) {
		return common.NewServiceError("抽奖尚未开始")
	}
	if now.After(time.Time(prizeRecord.EndTime).UTC()) {
		return common.NewServiceError("抽奖已结束")
	}

	if prizeRecord.Key != req.Key {
		return common.NewServiceError("密钥错误")
	}

	_, gfsError = dao.GetPrizeDao().GetMemberById(req.ID, req.Email)
	if gfsError == nil {
		return common.NewServiceError("您的邮箱已参与了此活动，请勿重复申请")
	} else if gfsError.GetMsg() != common.RETURN_RECORD_NOT_FOUND {
		return common.NewServiceError("数据库查询失败")
	}

	record := models.ParticipationCacheSaveModel{
		PrizeId: req.ID,
		Name:    req.Name,
		Email:   req.Email,
		IP:      ip,
		Agent:   c.Get("User-Agent"),
	}

	v7uuid, err := uuid.NewV7()
	if err != nil {
		return common.NewServiceError("UUID 生成失败")
	}
	key := v7uuid.String()
	prizeKey := "prize:" + util.Int642String(req.ID) + ":" + key
	if jsonRecord, jsonErr := sonic.Marshal(record); jsonErr == nil {
		if !cs.SetNX(prizeKey, string(jsonRecord), 10*time.Minute) {
			return common.NewServiceError("报名失败，请重试")
		}
		cs.SetExpire(emailLockKey, "1", 3*time.Minute)
	} else {
		return common.NewServiceError("设置缓存失败")
	}

	gfsError = cs.GetEmailService().SendActivationEmail(req.Email,
		"gofurry 抽奖服务 - 参与",
		"https://game.go-furry.com/api/v1/game/prize/participation/activation?key="+key+"&id="+util.Int642String(req.ID),
		"点击此处完成参与",
		"您正在申请 gofurry 的抽奖服务，点击下方链接完成报名的最后一步: ",
		"10分钟")
	if gfsError != nil {
		cs.Del(emailLockKey, prizeKey)
	}

	return gfsError
}

func (s prizeService) ActiveParticipation(id string, key string) (prizeRes models.GfgPrize, memberRes models.GfgPrizeMember, gfsError common.GFError) {
	prizeKey := "prize:" + id + ":" + key
	data, gfsError := cs.GetString(prizeKey)
	if gfsError != nil {
		return prizeRes, memberRes, common.NewServiceError("缓存查询失败, 请重试")
	}
	if data == "" {
		return prizeRes, memberRes, common.NewServiceError("激活链接已失效或已过期")
	}

	var prizeRecord models.ParticipationCacheSaveModel
	if jsonErr := sonic.Unmarshal([]byte(data), &prizeRecord); jsonErr != nil {
		return prizeRes, memberRes, common.NewServiceError("激活链接已失效, 请重试")
	}

	var prize models.GfgPrize
	strId, err := util.String2Int64(id)
	if err != nil {
		return prizeRes, memberRes, common.NewServiceError("参数错误")
	}
	gfsError = dao.GetPrizeDao().GetById(strId, &prize)
	if gfsError != nil {
		return prizeRes, memberRes, common.NewServiceError("激活链接已失效, 请重试")
	}

	now := time.Now().UTC().Add(8 * time.Hour)
	if !prize.Status || now.After(time.Time(prize.EndTime).UTC()) {
		return prizeRes, memberRes, common.NewServiceError("活动已过期, 请重试")
	}

	data, err = cs.GetRedisService().GetDel(context.Background(), prizeKey).Result()
	if err != nil || data == "" {
		return prizeRes, memberRes, common.NewServiceError("激活链接已失效或已过期")
	}

	memberRecord := models.GfgPrizeMember{
		ID:         util.GenerateId(),
		PrizeID:    prizeRecord.PrizeId,
		Name:       prizeRecord.Name,
		Email:      prizeRecord.Email,
		IP:         prizeRecord.IP,
		Agent:      prizeRecord.Agent,
		IsWinner:   false,
		CreateTime: cm.LocalTime(time.Now()),
	}
	gfsError = dao.GetPrizeDao().Add(&memberRecord)
	if gfsError != nil {
		return prizeRes, memberRes, common.NewServiceError("激活失败, 请重试")
	}

	emailLockKey := "prize:email:" + id + ":" + prizeRecord.Email
	cs.Del(emailLockKey)

	return prize, memberRecord, nil
}

func (s prizeService) LotteryInfo() (res models.LotteryResp, err common.GFError) {
	// 往期
	data, err := cs.GetString("prize:history")
	if err != nil {
		return
	}
	jsonErr := sonic.Unmarshal([]byte(data), &res.History)
	if jsonErr != nil {
		return res, common.NewServiceError("json err:" + jsonErr.Error())
	}

	// 本期
	active, err := dao.GetPrizeDao().GetLotteryActive()
	if err != nil {
		return
	}

	for idx := range active {
		members, err := dao.GetPrizeDao().GetMembers(active[idx].ID)
		if err != nil {
			log.Error("GetMembers err:", err.GetMsg())
		}
		// 脱敏
		memberCache := make([]models.WinnerCacheModel, 0, len(members))
		for _, m := range members {
			memberCache = append(memberCache, models.WinnerCacheModel{
				Name:  m.Name,
				Email: util.MaskEmail(m.Email),
			})
		}

		var prizeModels models.PrizeModel
		jsonErr = sonic.Unmarshal([]byte(active[idx].Prize), &prizeModels)
		if jsonErr != nil {
			return res, common.NewServiceError("json err:" + jsonErr.Error())
		}

		newLottery := models.LotteryVo{
			ID:        active[idx].ID,
			Title:     active[idx].Title,
			Desc:      active[idx].Desc,
			StartTime: active[idx].StartTime,
			EndTime:   active[idx].EndTime,
			Prize: struct {
				Title    string `json:"title"`
				Platform string `json:"platform"`
				Count    int    `json:"count"`
			}{Title: prizeModels.Title, Platform: prizeModels.Platform, Count: len(prizeModels.Keys)},
		}

		res.Active = append(res.Active, models.ActiveVo{
			Lottery: newLottery,
			Member:  memberCache,
			Count:   len(memberCache),
		})
	}

	return
}
