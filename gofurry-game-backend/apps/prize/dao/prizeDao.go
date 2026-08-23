package dao

import (
	"context"
	"errors"

	"github.com/gofurry/gofurry-game-backend/apps/prize/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PrizeDAO struct {
	queries *gamesqlc.Queries
}

func New(pool *pgxpool.Pool) *PrizeDAO {
	if pool == nil {
		return &PrizeDAO{}
	}
	return &PrizeDAO{queries: gamesqlc.New(pool)}
}

func (dao *PrizeDAO) GetById(id int64, dest *models.GfgPrize) common.GFError {
	if dao == nil || dao.queries == nil {
		return common.NewDaoError("database is not initialized")
	}
	row, err := dao.queries.GetPrizeByID(context.Background(), id)
	if err != nil {
		return daoError(err)
	}
	*dest = mapPrize(row)
	return nil
}

func (dao *PrizeDAO) GetMemberById(id int64, email string) (models.GfgPrizeMember, common.GFError) {
	if dao == nil || dao.queries == nil {
		return models.GfgPrizeMember{}, common.NewDaoError("database is not initialized")
	}
	row, err := dao.queries.GetPrizeMemberByEmail(context.Background(), gamesqlc.GetPrizeMemberByEmailParams{PrizeID: id, Email: email})
	if err != nil {
		return models.GfgPrizeMember{}, daoError(err)
	}
	return mapMember(row), nil
}

func (dao *PrizeDAO) Add(record *models.GfgPrizeMember) common.GFError {
	if dao == nil || dao.queries == nil {
		return common.NewDaoError("database is not initialized")
	}
	err := dao.queries.InsertPrizeMember(context.Background(), gamesqlc.InsertPrizeMemberParams{
		ID: record.ID, PrizeID: record.PrizeID, Name: record.Name, Email: record.Email,
		Ip: record.IP, Agent: record.Agent, IsWinner: record.IsWinner, PrizeKey: record.PrizeKey,
		CreateTime: timestamp(record.CreateTime),
	})
	return daoError(err)
}

func (dao *PrizeDAO) GetActivePrizeList() ([]models.GfgPrize, common.GFError) {
	rows, err := dao.queries.ListActivePrizes(context.Background())
	if err != nil {
		return nil, daoError(err)
	}
	result := make([]models.GfgPrize, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapPrize(row))
	}
	return result, nil
}

func (dao *PrizeDAO) GetMembers(id int64) ([]models.GfgPrizeMember, common.GFError) {
	rows, err := dao.queries.ListPrizeMembers(context.Background(), id)
	if err != nil {
		return nil, daoError(err)
	}
	return mapMembers(rows), nil
}

func (dao *PrizeDAO) GetLotteryHistory() ([]models.PrizeCacheModel, common.GFError) {
	rows, err := dao.queries.ListPrizeHistory(context.Background())
	if err != nil {
		return nil, daoError(err)
	}
	result := make([]models.PrizeCacheModel, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.PrizeCacheModel{ID: row.ID, Title: row.Title, Desc: row.Desc, EndTime: localTime(row.EndTime), Prize: string(row.Prize)})
	}
	return result, nil
}

func (dao *PrizeDAO) GetMemberCount(prizeID int64) (int64, common.GFError) {
	count, err := dao.queries.CountPrizeMembers(context.Background(), prizeID)
	return count, daoError(err)
}

func (dao *PrizeDAO) GetWinners(prizeID int64) ([]models.GfgPrizeMember, common.GFError) {
	rows, err := dao.queries.ListPrizeWinners(context.Background(), prizeID)
	if err != nil {
		return nil, daoError(err)
	}
	return mapMembers(rows), nil
}

func (dao *PrizeDAO) GetLotteryActive() ([]models.ActiveLotteryVo, common.GFError) {
	rows, err := dao.queries.ListActiveLotteries(context.Background())
	if err != nil {
		return nil, daoError(err)
	}
	result := make([]models.ActiveLotteryVo, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.ActiveLotteryVo{ID: row.ID, Title: row.Title, Desc: row.Desc, StartTime: localTime(row.StartTime), EndTime: localTime(row.EndTime), Prize: string(row.Prize)})
	}
	return result, nil
}

func (dao *PrizeDAO) Update(id int64, record models.GfgPrizeMember) (int64, common.GFError) {
	count, err := dao.queries.UpdatePrizeMemberWinner(context.Background(), gamesqlc.UpdatePrizeMemberWinnerParams{ID: id, IsWinner: record.IsWinner, PrizeKey: record.PrizeKey})
	return count, daoError(err)
}

// Save intentionally persists false status values; this preserves the old full-update lottery close behavior.
func (dao *PrizeDAO) Save(id int64, record models.GfgPrize) (int64, common.GFError) {
	count, err := dao.queries.SavePrize(context.Background(), gamesqlc.SavePrizeParams{
		ID: id, Title: record.Title, Description: record.Desc, Prize: []byte(record.Prize), Key: record.Key,
		StartTime: timestamp(record.StartTime), EndTime: timestamp(record.EndTime), Status: record.Status,
	})
	return count, daoError(err)
}

func mapPrize(row gamesqlc.GfgPrize) models.GfgPrize {
	return models.GfgPrize{ID: row.ID, Title: row.Title, Desc: row.Desc, Prize: string(row.Prize), Key: row.Key, StartTime: localTime(row.StartTime), EndTime: localTime(row.EndTime), CreateTime: localTime(row.CreateTime), Status: row.Status}
}

func mapMember(row gamesqlc.GfgPrizeMember) models.GfgPrizeMember {
	return models.GfgPrizeMember{ID: row.ID, PrizeID: row.PrizeID, Name: row.Name, Email: row.Email, IP: row.Ip, Agent: row.Agent, IsWinner: row.IsWinner, PrizeKey: row.PrizeKey, CreateTime: localTime(row.CreateTime)}
}

func mapMembers(rows []gamesqlc.GfgPrizeMember) []models.GfgPrizeMember {
	result := make([]models.GfgPrizeMember, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapMember(row))
	}
	return result
}

func timestamp(value cm.LocalTime) pgtype.Timestamp {
	return pgtype.Timestamp{Time: value.Time(), Valid: !value.Time().IsZero()}
}

func localTime(value pgtype.Timestamp) cm.LocalTime {
	if !value.Valid {
		return cm.LocalTime{}
	}
	return cm.LocalTime(value.Time)
}

func daoError(err error) common.GFError {
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
