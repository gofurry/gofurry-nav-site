package dao

import (
	"errors"

	"github.com/gofurry/gofurry-user/apps/user/models"
	"github.com/gofurry/gofurry-user/common"
	"github.com/gofurry/gofurry-user/common/abstract"
	"gorm.io/gorm"
)

var newUserDao = new(userDao)

func init() {
	newUserDao.Init()
	newUserDao.Mode = models.GfUser{}
}

type userDao struct{ abstract.Dao }

func GetUserDao() *userDao { return newUserDao }

func (dao *userDao) FindOneByName(name string) (record models.GfUser, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfUser).Where("name = ?", name).Take(&record)
	if err := db.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return record, common.NewDaoError(common.RETURN_RECORD_NOT_FOUND)
		} else {
			return record, common.NewDaoError(err.Error())
		}
	}
	return
}

func (dao *userDao) FindOneByEmail(email string) (record models.GfUser, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfUser).Where("email = ?", email).Take(&record)
	if err := db.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return record, common.NewDaoError(common.RETURN_RECORD_NOT_FOUND)
		} else {
			return record, common.NewDaoError(err.Error())
		}
	}
	return
}
