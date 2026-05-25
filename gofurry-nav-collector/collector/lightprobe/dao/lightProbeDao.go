package dao

import (
	"github.com/gofurry/gofurry-nav-collector/collector/lightprobe/models"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/abstract"
)

var newLightProbeDao = new(lightProbeDao)

func init() {
	newLightProbeDao.Init()
}

type lightProbeDao struct{ abstract.Dao }

func GetLightProbeDao() *lightProbeDao { return newLightProbeDao }

func (dao lightProbeDao) GetList() ([]models.GfnCollectorDomain, common.GFError) {
	var res []models.GfnCollectorDomain
	db := dao.Gm.Table(models.TableNameGfnCollectorDomain + " AS cd").
		Select("cd.*").
		Joins("JOIN " + models.TableNameGfnSite + " AS s ON s.id = cd.site_id").
		Where("cd.deleted IS NOT TRUE AND cd.site_id > 0 AND s.deleted IS NOT TRUE").
		Order("cd.site_id ASC, cd.id ASC")
	db.Find(&res)
	if err := db.Error; err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return res, nil
}
