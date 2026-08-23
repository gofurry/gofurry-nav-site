package dao

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gofurry/gofurry-game-backend/apps/review/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewDAO struct {
	queries *gamesqlc.Queries
}

func New(pool *pgxpool.Pool) *ReviewDAO {
	if pool == nil {
		return &ReviewDAO{}
	}
	return &ReviewDAO{queries: gamesqlc.New(pool)}
}

func (dao *ReviewDAO) GetHotGame(num int) ([]models.AvgScoreResult, common.GFError) {
	rows, err := dao.queries.GetHotGames(context.Background(), int32(num))
	if err != nil {
		return nil, reviewDAOError(err)
	}
	result := make([]models.AvgScoreResult, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.AvgScoreResult{GameID: row.GameID, AvgScore: row.AvgScore, CommentCount: row.CommentCount, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, Header: row.Header})
	}
	return result, nil
}

func (dao *ReviewDAO) GetScoreById(id int64) (models.AvgScoreResult, common.GFError) {
	row, err := dao.queries.GetGameScore(context.Background(), id)
	if err != nil {
		return models.AvgScoreResult{}, common.NewDaoError(err.Error())
	}
	return models.AvgScoreResult{GameID: row.GameID, AvgScore: row.AvgScore, CommentCount: row.CommentCount, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, Header: row.Header}, nil
}

func (dao *ReviewDAO) GetReviewByIPAndName(id string, ip string, name string) (models.GfgGameComment, common.GFError) {
	gameID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return models.GfgGameComment{}, common.NewDaoError(err.Error())
	}
	row, err := dao.queries.GetReviewByIdentity(context.Background(), gamesqlc.GetReviewByIdentityParams{Ip: ip, GameID: gameID, Name: name})
	if err != nil {
		return models.GfgGameComment{}, reviewDAOError(err)
	}
	return models.GfgGameComment{ID: row.ID, Region: row.Region, Content: row.Content, Score: row.Score, CreateTime: reviewLocalTime(row.CreateTime), GameID: row.GameID, IP: row.Ip, Name: row.Name}, nil
}

func (dao *ReviewDAO) GetListByLimit(num int, lang string) ([]models.AnonymousReviewResponse, common.GFError) {
	rows, err := dao.queries.ListAnonymousReviews(context.Background(), gamesqlc.ListAnonymousReviewsParams{Lang: lang, LimitCount: int32(num)})
	if err != nil {
		return nil, reviewDAOError(err)
	}
	result := make([]models.AnonymousReviewResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.AnonymousReviewResponse{Region: row.Region, Score: row.Score, Content: row.Content, IP: row.Ip, Time: reviewLocalTime(row.Time), GameName: row.GameName, GameCover: row.GameCover})
	}
	return result, nil
}

func (dao *ReviewDAO) Add(record *models.GfgGameComment) common.GFError {
	err := dao.queries.InsertReview(context.Background(), gamesqlc.InsertReviewParams{
		ID: record.ID, Region: record.Region, Content: record.Content, Score: record.Score,
		CreateTime: pgtype.Timestamp{Time: time.Time(record.CreateTime), Valid: !time.Time(record.CreateTime).IsZero()},
		GameID:     record.GameID, Ip: record.IP, Name: record.Name,
	})
	return reviewDAOError(err)
}

func reviewLocalTime(value pgtype.Timestamp) cm.LocalTime {
	if !value.Valid {
		return cm.LocalTime{}
	}
	return cm.LocalTime(value.Time)
}

func reviewDAOError(err error) common.GFError {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return common.NewDaoError(common.RETURN_RECORD_NOT_FOUND)
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23502":
			return common.NewDaoError("必要数据为空，入库失败")
		case "23505":
			return common.NewDaoError("数据重复，入库失败")
		}
	}
	log.Error(err)
	return common.NewDaoError(err.Error())
}
