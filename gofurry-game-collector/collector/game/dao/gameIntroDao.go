package dao

import (
	"context"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/common"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GameIntroDao 游戏简介DAO层
type GameIntroDao struct{}

// NewGameIntroDao 实例化
func NewGameIntroDao() *GameIntroDao {
	return &GameIntroDao{}
}

// SaveOrUpdate 保存或更新游戏简介
func (d *GameIntroDao) SaveOrUpdate(ctx context.Context, intro *models.GameIntro) common.GFError {
	if intro.GameID == 0 || intro.Content == "" || intro.Lang == "" {
		return common.NewDaoError("game_id、content、lang不能为空")
	}

	// 填充时间字段
	now := time.Now()
	if intro.CreateTime.IsZero() {
		intro.CreateTime = now
	}
	intro.UpdateTime = now

	// 更新条件 game_id + lang 唯一
	filter := bson.M{
		"game_id": intro.GameID,
		"lang":    intro.Lang,
	}

	// 构建更新内容
	update := bson.M{
		"$set": intro, // 全量更新
	}

	// 执行更新
	opts := options.Update().SetUpsert(true)
	_, err := cs.Mongo.Collection(models.GameIntro{}.TableName()).UpdateOne(
		ctx,
		filter,
		update,
		opts,
	)
	if err != nil {
		return common.NewDaoError("保存游戏简介失败: " + err.Error())
	}
	return nil
}

// GetByGameIDAndLang 根据游戏ID+语言查询简介
func (d *GameIntroDao) GetByGameIDAndLang(ctx context.Context, gameID int64, lang string) (res models.GameIntro, err common.GFError) {
	if gameID == 0 || lang == "" {
		return res, common.NewDaoError("game_id、lang不能为空")
	}

	// 构建查询条件
	filter := bson.M{
		"game_id": gameID,
		"lang":    lang,
	}

	// 执行查询
	var intro models.GameIntro
	dbErr := cs.Mongo.Collection(models.GameIntro{}.TableName()).FindOne(ctx, filter).Decode(&intro)
	if dbErr != nil {
		if dbErr == mongo.ErrNoDocuments {
			return res, nil
		}
		return res, common.NewDaoError("查询游戏简介失败: " + dbErr.Error())
	}
	return intro, nil
}

// BatchGetByGameIDs 批量查询多个游戏的简介
func (d *GameIntroDao) BatchGetByGameIDs(ctx context.Context, gameIDs []int64, lang string) (map[int64]*models.GameIntro, common.GFError) {
	if len(gameIDs) == 0 || lang == "" {
		return nil, common.NewDaoError("game_ids、lang不能为空")
	}

	// 构建查询条件
	filter := bson.M{
		"game_id": bson.M{"$in": gameIDs},
		"lang":    lang,
	}

	// 执行查询
	cursor, err := cs.Mongo.Collection(models.GameIntro{}.TableName()).Find(ctx, filter)
	if err != nil {
		return nil, common.NewDaoError("批量查询游戏简介失败: " + err.Error())
	}
	defer cursor.Close(ctx)

	// 解析结果为map（game_id -> intro）
	introMap := make(map[int64]*models.GameIntro)
	for cursor.Next(ctx) {
		var intro models.GameIntro
		if err := cursor.Decode(&intro); err != nil {
			return nil, common.NewDaoError("解析游戏简介失败: " + err.Error())
		}
		introMap[intro.GameID] = &intro
	}

	if err := cursor.Err(); err != nil {
		return nil, common.NewDaoError("游标遍历失败: " + err.Error())
	}
	return introMap, nil
}

// DeleteByGameID 根据游戏ID删除所有语言的简介
func (d *GameIntroDao) DeleteByGameID(ctx context.Context, gameID int64) common.GFError {
	if gameID == 0 {
		return common.NewDaoError("game_id不能为空")
	}

	filter := bson.M{"game_id": gameID}
	_, err := cs.Mongo.Collection(models.GameIntro{}.TableName()).DeleteMany(ctx, filter)
	if err != nil {
		return common.NewDaoError("删除游戏简介失败: " + err.Error())
	}
	return nil
}
