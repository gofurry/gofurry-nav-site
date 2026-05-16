package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-game-backend/apps/recommend/dao"
	"github.com/gofurry/gofurry-game-backend/apps/recommend/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
	"github.com/gofurry/gofurry-game-backend/common/util"
	"golang.org/x/sync/errgroup"

	gd "github.com/gofurry/gofurry-game-backend/apps/game/dao"
	gm "github.com/gofurry/gofurry-game-backend/apps/game/models"
)

type recommendService struct{}

var recommendSingleton = new(recommendService)

func GetRecommendService() *recommendService { return recommendSingleton }

// Redis 定义
const (
	redisTagMappingKey = "recommend:tag-mapping"
	redisTagIDsKey     = "recommend:tag-ids"
	cacheExpireTime    = 1 * time.Hour // 缓存过期时间
)

const (
	// 推荐计算的超时时间
	recommendCalcTimeout = 3 * time.Second
	// 任务池大小
	calcPoolSize = 4
)

// 任务池, 限制并发计算任务数
var calcPool = make(chan struct{}, calcPoolSize)

var once sync.Once

// 随机获取一个 GameID
func (s recommendService) GetRandomGameID() (string, common.GFError) {
	count, err := gd.GetGameDao().Count(gm.GfgGame{})
	if err != nil {
		return "", common.NewServiceError("统计数量出错")
	}
	intCnt, _ := util.String2Int(util.Int642String(count))

	// Go 1.20 + 无需初始化
	once.Do(func() {
		rand.Seed(time.Now().UnixNano())
	})

	// 生成一个 0 到 intCnt-1 之间的随机整数
	randomInt := rand.Intn(intCnt)
	gameRecord, gfError := gd.GetGameDao().GetByNum(randomInt)
	if gfError != nil {
		return "", gfError
	}

	return util.Int642String(gameRecord.ID), nil
}

// RecommendByCBF Content-based Filter 返回物品A的余弦相似度最高的物品
func (s recommendService) RecommendByCBF(id string, lang string) (gameListVo []models.GameRecommendVo, err common.GFError) {
	intID, parseErr := util.String2Int64(id)
	if parseErr != nil {
		return nil, common.NewServiceError(parseErr.Error())
	}

	// 创建根上下文
	rootCtx, rootCancel := context.WithTimeout(context.Background(), recommendCalcTimeout)
	defer rootCancel()

	// 用 errgroup 管理异步任务
	g, ctx := errgroup.WithContext(rootCtx)
	resultChan := make(chan []models.GameRecommendVo, 1)
	errChan := make(chan common.GFError, 1)

	// 异步执行计算任务
	g.Go(func() error {
		select {
		case calcPool <- struct{}{}: // 获取令牌
			defer func() {
				<-calcPool // 释放令牌
			}()
		case <-ctx.Done(): // 超时/取消时
			return ctx.Err()
		}

		// 执行计算密集型任务
		select {
		case <-ctx.Done(): // 任务还没开始就超时
			return ctx.Err()
		default:
			res, e := getGameCBF(intID, lang)

			if e != nil {
				errChan <- e
				return fmt.Errorf("计算推荐失败: %s", e.GetMsg())
			}
			resultChan <- res
			return nil
		}
	})

	// 等待任务完成
	waitErr := g.Wait()
	// 先判断是否是根上下文超时
	if rootCtx.Err() == context.DeadlineExceeded {
		return nil, common.NewServiceError(fmt.Sprintf("推荐计算超时(超时时间: %v)", recommendCalcTimeout))
	}

	// 处理其他错误
	if waitErr != nil {
		log.Error("推荐请求执行失败: id=%s, err=%v", id, waitErr)
		select {
		case e := <-errChan:
			return nil, e
		default:
			return nil, common.NewServiceError("推荐计算失败: " + waitErr.Error())
		}
	}

	// 读取结果时增加超时兜底
	select {
	case res := <-resultChan:
		return res, nil
	case e := <-errChan:
		return nil, e
	case <-rootCtx.Done():
		return nil, common.NewServiceError("推荐计算超时")
	}
}

// CBF 获取一组推荐的游戏记录
func getGameCBF(id int64, lang string) (recommendContent []models.GameRecommendVo, err common.GFError) {
	// 执行 CBF
	similarities, err := processContentBasedFilter(id)
	if err != nil {
		return nil, err
	}

	// 从相似度结果生成推荐视图 前12随机选8
	const topN = 8
	const candidateN = 12

	filtered := make([]models.ContentSimilarities, 0, candidateN)
	for _, sim := range similarities {
		if sim.ID == id || sim.Similarity <= 0 {
			continue
		}
		filtered = append(filtered, sim)
		if len(filtered) >= candidateN {
			break
		}
	}

	// 打乱候选列表 增加多样性
	//rand.Shuffle(len(filtered), func(i, j int) {
	//	filtered[i], filtered[j] = filtered[j], filtered[i]
	//})
	if len(filtered) > topN {
		filtered = filtered[:topN]
	}

	// 转换为 GameRecommendVo
	if len(filtered) == 0 {
		return recommendContent, nil
	}

	var gameIDs []int64
	idToSimilarity := make(map[int64]float64)
	for _, item := range filtered {
		gameIDs = append(gameIDs, item.ID)
		idToSimilarity[item.ID] = item.Similarity
	}
	gameList, err := dao.GetRecommendDao().GetRecommend(gameIDs, lang)
	if err != nil {
		return nil, common.NewServiceError(err.GetMsg())
	}

	for _, game := range gameList {
		vo := models.GameRecommendVo{
			ID:         util.Int642String(game.ID),
			Similarity: idToSimilarity[game.ID],
			Appid:      game.Appid,
		}

		if lang == "en" {
			vo.Name = game.NameEn
			vo.Info = game.InfoEn
		} else {
			vo.Name = game.NameZh
			vo.Info = game.InfoZh
		}

		recommendContent = append(recommendContent, vo)
	}

	// 排序
	sort.Slice(recommendContent, func(i, j int) bool {
		return recommendContent[i].Similarity > recommendContent[j].Similarity
	})

	return recommendContent, nil
}

// CBF 算法
func processContentBasedFilter(gameID int64) ([]models.ContentSimilarities, common.GFError) {
	// 获取标签映射和标签 ID 列表
	tagMappingMap, tagIDs, err := getTagToMap()
	if err != nil {
		return nil, err
	}

	// 初始化标签 ID 到维度索引的映射
	tagIDToIndex, indexToTagID := buildTagIndexMap(tagIDs)

	// 特征提取 - 独热编码
	targetContent, contentFeatures := execFeature(tagMappingMap, tagIDToIndex, gameID)

	// 校验目标游戏是否存在有效特征
	if len(targetContent.Tag) == 0 {
		// 游戏 ID 不在映射中
		if _, exists := tagMappingMap[gameID]; !exists {
			return nil, common.NewServiceError("目标游戏不存在或未关联标签")
		}
		// 游戏存在但无标签
		log.Warn("游戏ID=", gameID, "未关联任何标签，无法生成推荐")
		return []models.ContentSimilarities{}, nil
	}

	// 计算相似度
	similarities := execSimilarity(targetContent, contentFeatures, indexToTagID)
	return similarities, nil
}

// 构建标签 ID到维度索引的映射 map[tagID]index
func buildTagIndexMap(tagIDs []int64) (map[int64]int, map[float64]int64) {
	tagIDToIndex := make(map[int64]int, len(tagIDs))
	indexToTagID := make(map[float64]int64, len(tagIDs))
	for idx, tagID := range tagIDs {
		tagIDToIndex[tagID] = idx
		indexToTagID[float64(idx)] = tagID
	}

	return tagIDToIndex, indexToTagID
}

// 特征提取 - 独热编码
func execFeature(tagMapping map[int64][]int64, tagIDToIndex map[int64]int, targetGameID int64) (models.ContentSimilarities, []models.ContentSimilarities) {
	var targetContent models.ContentSimilarities
	var contentFeatures []models.ContentSimilarities

	for gameID, tagIDs := range tagMapping {
		// 构建独热特征
		feature := make([]float64, 0, len(tagIDs))
		seen := make(map[int]struct{}) // 去重标签

		for _, tagID := range tagIDs {
			idx, ok := tagIDToIndex[tagID]
			if !ok {
				continue // 忽略未注册的标签
			}
			if _, exists := seen[idx]; exists {
				continue // 跳过重复标签
			}
			seen[idx] = struct{}{}
			feature = append(feature, float64(idx)) // 存储维度索引
		}

		// 区分目标游戏和其他游戏
		if gameID == targetGameID {
			targetContent = models.ContentSimilarities{
				ID:  gameID,
				Tag: feature,
			}
		} else {
			contentFeatures = append(contentFeatures, models.ContentSimilarities{
				ID:  gameID,
				Tag: feature,
			})
		}
	}

	return targetContent, contentFeatures
}

// 计算相似度
func execSimilarity(target models.ContentSimilarities, others []models.ContentSimilarities, idx2tag map[float64]int64) []models.ContentSimilarities {
	similarities := make([]models.ContentSimilarities, 0, len(others))

	// 将目标特征转换为字典
	targetSpeciesCount := 0.0
	targetSet := make(map[float64]struct{}, len(target.Tag))
	for _, idx := range target.Tag {
		targetSet[idx] = struct{}{}
		if idx2tag[idx] >= 2000 && idx2tag[idx] < 3000 {
			targetSpeciesCount++
		}
	}

	var targetRecord = gm.GfgGame{}
	err := gd.GetGameDao().GetById(target.ID, &targetRecord)
	if err != nil {
		log.Error("execSimilarity err:", err.GetMsg())
	}

	// 计算每个游戏与目标的相似度
	for _, other := range others {
		// 跳过没有特征的项
		if len(other.Tag) == 0 {
			continue
		}

		// 计算共同标签数量
		commonCount, platformCount, commonPlatformCount, otherSpeciesCount, commonSpeciesCount := 0.0, 0.0, 0.0, 0.0, 0.0
		targetWeight, otherWeight := float64(len(target.Tag)), float64(len(other.Tag))
		w := 1.0
		xpCount := 0

		for _, idx := range other.Tag {
			if idx2tag[idx] >= 3000 && idx2tag[idx] < 4000 {
				platformCount++
			}
			if idx2tag[idx] >= 2000 && idx2tag[idx] < 3000 {
				otherSpeciesCount++
			}
			if _, exists := targetSet[idx]; exists {
				w = 1.0 // 归一
				switch {
				case idx2tag[idx] >= 1000 && idx2tag[idx] < 2000:
					w = 1.0 // 游戏类型
				case idx2tag[idx] >= 2000 && idx2tag[idx] < 3000:
					commonSpeciesCount++ // 物种
					continue
				case idx2tag[idx] >= 3000 && idx2tag[idx] < 4000:
					commonPlatformCount++ // 平台
					continue
				case idx2tag[idx] >= 9000 && idx2tag[idx] < 10000:
					xpCount++
					if xpCount <= 2 {
						w = 2.0 // XP 强语义低数量标签参加语义竞争
					} else {
						w = 1.0 // 降低权重
					}
				default:
					w = 1.0
				}
				// 特殊加权
				switch {
				case idx2tag[idx] == targetRecord.PrimaryTag:
					w = 2.0 // 唯一主标签加权
				case idx2tag[idx] == targetRecord.SecondaryTag:
					w = 1.5 // 唯一次标签加权
				}

				commonCount += w
				targetWeight += w - 1.0
				otherWeight += w - 1.0
			}
		}
		speciesRatio := 0.0
		if targetSpeciesCount > 0 {
			speciesRatio = commonSpeciesCount / math.Max(targetSpeciesCount, otherSpeciesCount) // Jaccard
		}
		commonCount += math.Min(speciesRatio*2.4, 2.4) // 强语义高数量标签 比较特征值边界
		if platformCount > 0 {
			commonCount += (commonPlatformCount / platformCount) * 1.0 // 弱语义高数量标签归一化计算相似度, 不参加语义竞争
		}

		// 计算余弦相似度 commonCount / (sqrt(len(target)) * sqrt(len(other)))
		sim := 0.0
		magTarget := math.Sqrt(targetWeight)
		magOther := math.Sqrt(otherWeight)
		if magTarget > 0 && magOther > 0 {
			sim = commonCount / (magTarget * magOther)
		}

		// 边界处理
		sim = clamp01(sim)

		if sim > 0 {
			similarities = append(similarities, models.ContentSimilarities{
				ID:         other.ID,
				Similarity: sim,
			})
		}
	}

	// 相似度排序
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})

	// 指数拉伸
	for idx := range similarities {
		similarities[idx].Similarity = displayScoreExp(similarities[idx].Similarity)
	}

	return similarities
}

// 边界处理
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// 线性拉伸
func displayScore(sim float64) float64 {
	return clamp01(0.6 + 0.3*sim) // [0.6, 0.9]
}

// 指数拉伸
func displayScoreExp(sim float64) float64 {
	k := 2.2 // 1.8 ~ 2.8
	return clamp01(1.0 - math.Exp(-k*sim))
}

// S 型指数拉伸
func displayScoreExpLogistic(sim float64) float64 {
	mid := 0.5 // 拐点
	k := 6.0   // 陡峭度
	return clamp01(1.0 / (1.0 + math.Exp(-k*(sim-mid))))
}

// 获取标签映射
func getTagToMap() (tagMapping map[int64][]int64, tagIDs []int64, err common.GFError) {
	// Redis 读缓存
	tagMapping, tagIDs, err = loadFromRedis()
	if err == nil && tagMapping != nil && len(tagIDs) > 0 {
		// 缓存命中，直接返回
		return tagMapping, tagIDs, nil
	}

	// 缓存未命中
	mappingRecords, err := dao.GetRecommendDao().GetTagMappingList()
	if err != nil {
		return nil, nil, common.NewServiceError("获取标签映射记录失败: " + err.GetMsg())
	}
	tagMapping = make(map[int64][]int64)
	for _, rec := range mappingRecords {
		gameID := rec.GameID
		tagID := rec.TagID

		tags := tagMapping[gameID]
		exists := false
		for _, t := range tags {
			if t == tagID {
				exists = true
				break
			}
		}
		if !exists {
			tagMapping[gameID] = append(tags, tagID)
		}
	}

	tagRecords, err := dao.GetRecommendDao().GetTagList()
	if err != nil {
		return nil, nil, common.NewServiceError("获取标签记录失败: " + err.GetMsg())
	}
	tagIDs = make([]int64, 0, len(tagRecords))
	for idx := range tagRecords {
		tagIDs = append(tagIDs, tagRecords[idx].ID)
	}

	// 异步写入Redis缓存
	go saveToRedis(tagMapping, tagIDs)

	return tagMapping, tagIDs, nil
}

// 从Redis加载缓存
func loadFromRedis() (tagMapping map[int64][]int64, tagIDs []int64, err common.GFError) {
	// 读取tagMapping
	mappingStr, err := cs.GetString(redisTagMappingKey)
	if err != nil || mappingStr == "" {
		// 缓存不存在或读取失败
		return nil, nil, nil
	}
	// 反序列化map[int64][]int64
	tagMapping = make(map[int64][]int64)
	if err := sonic.Unmarshal([]byte(mappingStr), &tagMapping); err != nil {
		log.Error("tagMapping反序列化失败: " + err.Error())
		return nil, nil, common.NewServiceError("缓存数据格式错误")
	}

	// 读取tagIDs
	idsStr, err := cs.GetString(redisTagIDsKey)
	if err != nil || idsStr == "" {
		return nil, nil, nil
	}
	// 反序列化[]int64
	if err := sonic.Unmarshal([]byte(idsStr), &tagIDs); err != nil {
		log.Error("tagIDs反序列化失败: " + err.Error())
		return nil, nil, common.NewServiceError("缓存数据格式错误")
	}

	return tagMapping, tagIDs, nil
}

// 保存数据到Redis
func saveToRedis(tagMapping map[int64][]int64, tagIDs []int64) {
	// 序列化tagMapping并保存
	mappingBytes, err := sonic.Marshal(tagMapping)
	if err != nil {
		log.Error("tagMapping序列化失败: " + err.Error())
		return
	}
	if err := cs.SetExpire(redisTagMappingKey, mappingBytes, cacheExpireTime); err != nil {
		log.Error("tagMapping缓存写入失败: " + err.GetMsg())
	}

	// 序列化tagIDs并保存
	idsBytes, err := sonic.Marshal(tagIDs)
	if err != nil {
		log.Error("tagIDs序列化失败: " + err.Error())
		return
	}
	if err := cs.SetExpire(redisTagIDsKey, idsBytes, cacheExpireTime); err != nil {
		log.Error("tagIDs缓存写入失败: " + err.GetMsg())
	}
}
