package dao

import (
	"errors"

	"github.com/gofurry/gofurry-user/apps/oauth/models"
	"github.com/gofurry/gofurry-user/common"
	"github.com/gofurry/gofurry-user/common/abstract"
	"gorm.io/gorm"
)

var newOauthDao = new(oauthDao)

func init() {
	newOauthDao.Init()
	newOauthDao.Mode = models.GfUserOauth{}
}

type oauthDao struct{ abstract.Dao }

func GetOauthDao() *oauthDao { return newOauthDao }

func (dao oauthDao) FindOneByName(name string, provider string) (record models.GfUserOauth, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfUserOauth).Where("open_id = ? AND provider = ?", name, provider).Take(&record)
	if err := db.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return record, common.NewDaoError(common.RETURN_RECORD_NOT_FOUND)
		} else {
			return record, common.NewDaoError(err.Error())
		}
	}
	return
}
