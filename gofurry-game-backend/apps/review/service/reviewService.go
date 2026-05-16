package service

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-game-backend/apps/review/dao"
	"github.com/gofurry/gofurry-game-backend/apps/review/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
	"github.com/gofurry/gofurry-game-backend/common/util"
)

type reviewService struct{}

var reviewSingleton = new(reviewService)

func GetReviewService() *reviewService { return reviewSingleton }

func (s reviewService) GetLatestReviewList(lang string) (res []models.AnonymousReviewResponse, err common.GFError) {
	res, err = dao.GetReviewDao().GetListByLimit(5, lang)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// 脱敏 IP
	for i := range res {
		res[i].IP = util.DesensitizeIP(res[i].IP)
	}
	return
}

func (s reviewService) AddAnonymousReview(req models.AnonymousReviewRequest, c fiber.Ctx) common.GFError {
	if req.ID == "" || req.Content == "" || strings.TrimSpace(req.Name) == "" {
		return common.NewServiceError("入参不能为空")
	}
	if req.Score < 0.0 || req.Score > 5.0 {
		return common.NewServiceError("评分有误")
	}
	if utf8.RuneCountInString(req.Content) > 500 {
		return common.NewServiceError("您的评论已超过500字长度限制")
	}

	ip := util.GetClientIP(c)
	// 无法获取公网 IP 不计数
	//if ip == "" {
	//	return common.NewServiceError("IP 为空")
	//}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return common.NewServiceError("IP 解析失败")
	}

	// 检验记录是否存在
	_, err := dao.GetReviewDao().GetReviewByIPAndName(req.ID, parsedIP.String(), req.Name)
	if err != nil {
		if err.GetMsg() != common.RETURN_RECORD_NOT_FOUND {
			log.Error(err)
			return common.NewServiceError(err.GetMsg())
		}
	} else {
		return common.NewServiceError("您的 IP + 名称已评论过该游戏, 需要修改请联系官方人员")
	}

	region, queryErr := queryBaiduIP(parsedIP.String())
	if queryErr != nil {
		log.Error(queryErr)
		return common.NewServiceError(queryErr.Error())
	}
	i64ID, parseErr := util.String2Int64(req.ID)
	if parseErr != nil {
		log.Error(parseErr)
		return common.NewServiceError(parseErr.Error())
	}

	// 转换精度为小数后1位
	formattedStr := fmt.Sprintf("%.1f", req.Score)
	formattedScore, parseErr := util.String2Float64(formattedStr)
	if parseErr != nil {
		log.Error(parseErr)
		return common.NewServiceError(parseErr.Error())
	}

	newRecord := models.GfgGameComment{
		ID:         util.GenerateId(),
		Region:     region,
		Content:    req.Content,
		Score:      formattedScore,
		CreateTime: cm.LocalTime(time.Now()),
		GameID:     i64ID,
		IP:         parsedIP.String(),
		Name:       req.Name,
	}

	return dao.GetReviewDao().Add(&newRecord)
}

type baiduResp struct {
	Status string `json:"status"`
	Data   []struct {
		Location string `json:"location"`
	} `json:"data"`
}

func queryBaiduIP(ip string) (region string, err error) {
	url := "https://opendata.baidu.com/api.php?query=" + ip + "&co=&resource_id=6006&oe=utf8"
	resp, err := http.Get(url)
	if err != nil {
		log.Error("请求百度 IP API 失败: ", err)
		return
	}
	defer resp.Body.Close()

	var result baiduResp
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error("解析百度 IP API 响应失败: ", err)
		return
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("IP 查询结果为空")
	}
	region = result.Data[0].Location
	if region == "" {
		return "Unknown Region/未知地区", nil
	}
	return
}
