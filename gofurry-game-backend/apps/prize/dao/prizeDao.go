package dao

import (
	"errors"

	"github.com/gofurry/gofurry-game-backend/apps/prize/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/abstract"
	"gorm.io/gorm"
)

var newPrizeDao = new(prizeDao)

func init() {
	newPrizeDao.Init()
}

type prizeDao struct{ abstract.Dao }

func GetPrizeDao() *prizeDao { return newPrizeDao }

func (dao prizeDao) GetMemberById(id int64, email string) (res models.GfgPrizeMember, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfgPrizeMember).Where("prize_id = ? AND email = ?", id, email)
	if dbErr := db.Take(&res).Error; dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return res, common.NewDaoError(common.RETURN_RECORD_NOT_FOUND)
		}
		return res, common.NewDaoError(dbErr.Error())
	}
	return res, nil
}

func (dao prizeDao) GetActivePrizeList() (res []models.GfgPrize, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfgPrize).Where("status IS TRUE")

	if dbErr := db.Find(&res).Error; dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return res, common.NewDaoError(common.RETURN_RECORD_NOT_FOUND)
		}
		return res, common.NewDaoError(dbErr.Error())
	}
	return res, nil
}

func (dao prizeDao) GetMembers(id int64) (res []models.GfgPrizeMember, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfgPrizeMember).Where("prize_id = ?", id)
	if dbErr := db.Find(&res).Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return res, nil
}

func (dao prizeDao) GetLotteryHistory() (res []models.PrizeCacheModel, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfgPrize).
		Where("status IS FALSE").
		Order("end_time DESC")

	if dbErr := db.Find(&res).Error; dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return res, common.NewDaoError(common.RETURN_RECORD_NOT_FOUND)
		}
		return res, common.NewDaoError(dbErr.Error())
	}

	return res, nil
}

func (dao prizeDao) GetMemberCount(prizeID int64) (count int64, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfgPrizeMember).
		Where("prize_id = ?", prizeID).
		Count(&count)

	if db.Error != nil {
		return 0, common.NewDaoError(db.Error.Error())
	}

	return count, nil
}

func (dao prizeDao) GetWinners(prizeID int64) (res []models.GfgPrizeMember, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfgPrizeMember).
		Where("prize_id = ? AND is_winner = TRUE", prizeID)

	if dbErr := db.Find(&res).Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}

	return res, nil
}

func (dao prizeDao) GetLotteryActive() (res []models.ActiveLotteryVo, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfgPrize).
		Where("status IS TRUE").
		Order("end_time DESC")

	if dbErr := db.Find(&res).Error; dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return res, common.NewDaoError(common.RETURN_RECORD_NOT_FOUND)
		}
		return res, common.NewDaoError(dbErr.Error())
	}

	return res, nil
}
