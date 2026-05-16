package abstract

import (
	"errors"

	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	database "github.com/gofurry/gofurry-game-backend/roof/db"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

/*
 * @Desc: 统一增删改查接口
 * @author: 福狼
 * @version: v1.0.0
 */

type Dao struct {
	Gm   *gorm.DB
	Mode any
}

func (dao *Dao) Init() {
	dao.Gm = database.Orm.DB()
}

func (dao *Dao) Add(record any) common.GFError {
	db := dao.Gm.Create(record)
	if err := db.Error; err != nil {
		log.Error(err)
		pe, ok := err.(*pgconn.PgError)
		if ok {
			if pe.Code == "23502" {
				return common.NewDaoError("必要数据为空，入库失败")
			}
			if pe.Code == "23505" {
				return common.NewDaoError("数据重复，入库失败")
			}
		}
		return common.NewDaoError(err.Error())
	}
	return nil
}

func (dao *Dao) Update(id int64, record any) (int64, common.GFError) {
	db := dao.Gm.Omit("create_time", "node").Where("id = ?", id).Updates(record)
	if err := db.Error; err != nil {
		log.Error(err)
		pe, ok := err.(*pgconn.PgError)
		if ok {
			if pe.Code == "23502" {
				return 0, common.NewDaoError("必要数据为空，入库失败")
			}
			if pe.Code == "23505" {
				return 0, common.NewDaoError("数据重复，入库失败")
			}
		}
		return 0, common.NewDaoError(err.Error())
	}
	return db.RowsAffected, nil
}

func (dao *Dao) Save(id int64, record any) (int64, common.GFError) {
	db := dao.Gm.Omit("create_time", "node").Where("id = ?", id).Save(record)
	if err := db.Error; err != nil {
		log.Error(err)
		pe, ok := err.(*pgconn.PgError)
		if ok {
			if pe.Code == "23502" {
				return 0, common.NewDaoError("必要数据为空，入库失败")
			}
			if pe.Code == "23505" {
				return 0, common.NewDaoError("数据重复，入库失败")
			}
		}
		return 0, common.NewDaoError(err.Error())
	}
	return db.RowsAffected, nil
}

func (dao *Dao) Delete(idList []int64, tableMode any) (int64, common.GFError) {
	db := dao.Gm.Where("id in ?", idList).Delete(tableMode)
	if err := db.Error; err != nil {
		log.Error(err)
		return 0, common.NewDaoError(err.Error())
	}
	return db.RowsAffected, nil
}

func (dao *Dao) GetById(id int64, record any) common.GFError {
	db := dao.Gm.Where("id = ?", id).Take(record)
	if err := db.Error; err != nil {
		log.Error(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewDaoError("404")
		}
		return common.NewDaoError(err.Error())
	}
	return nil
}

func (dao *Dao) Count(tableMode any) (int64, common.GFError) {
	var count int64
	db := dao.Gm.Model(tableMode).Count(&count)
	if err := db.Error; err != nil {
		log.Error(err)
		return 0, common.NewDaoError("统计数量失败.")
	}
	return count, nil
}
