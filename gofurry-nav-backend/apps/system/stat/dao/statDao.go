package dao

import (
	navModels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	logModels "github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/models"
	"github.com/gofurry/gofurry-nav-backend/apps/system/stat/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/abstract"
)

var newStatDao = new(statDao)

func init() {
	newStatDao.Init()
}

type statDao struct{ abstract.Dao }

func GetStatDao() *statDao { return newStatDao }

// 获取内容最多的分组
func (dao *statDao) GetGroupCount(lang string) (res []models.GroupCountVo, err common.GFError) {
	subQuery := dao.Gm.Table(navModels.TableNameGfnSiteGroupMap).
		Select("group_id, COUNT(*) AS count").
		Group("group_id")

	db := dao.Gm.Table("(?) AS gsgm", subQuery)
	db.Joins("JOIN " + navModels.TableNameGfnSiteGroup + " AS gsg ON gsgm.group_id = gsg.id")

	switch lang {
	case "zh":
		db.Select("gsg.name, gsgm.count")
	case "en":
		db.Select("gsg.name_en AS name, gsgm.count")
	default:
		db.Select("gsg.name, gsgm.count")
	}

	db.Order("gsgm.count DESC").Limit(4)

	if dbErr := db.Find(&res).Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return res, nil
}

func (dao *statDao) GetSiteList(lang string) (res []models.SiteListVo, err common.GFError) {
	db := dao.Gm.Table(navModels.TableNameGfnSite)
	switch lang {
	case "zh":
		db.Select("name, country, create_time")
	case "en":
		db.Select("name_en AS name, country, create_time")
	default:
		db.Select("name, country, create_time")
	}

	db.Order("create_time DESC").Limit(10)

	if dbErr := db.Find(&res).Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}

func (dao *statDao) GetCommonCount() (res models.CommonCountModel, err common.GFError) {
	sql := `
        SELECT
            (SELECT COUNT(*) FROM gfn_site) AS site_count,
            (SELECT COUNT(*) FROM gfn_collector_domain) AS domain_count,
            (SELECT COUNT(*) FROM gfn_collector_log_dns) AS dns_count,
            (SELECT COUNT(*) FROM gfn_collector_log_http) AS http_count,
            (SELECT COUNT(*) FROM gfn_collector_log_ping) AS ping_count
    `
	result := dao.Gm.Raw(sql).Scan(&res)
	if result.Error != nil {
		return res, common.NewDaoError(result.Error.Error())
	}
	return res, nil
}

func (dao *statDao) GetSiteCommon() (res []models.SiteTypeModel, err common.GFError) {
	db := dao.Gm.Table(navModels.TableNameGfnSite)
	db.Select("nsfw, welfare")
	if dbErr := db.Find(&res).Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}

func (dao *statDao) GetLatestPingStatus() (res []models.PingStatusModel, err common.GFError) {
	db := dao.Gm.Table(logModels.TableNameGfnCollectorLogPing)
	db.Select("DISTINCT ON (name) name, status, create_time")
	db.Order("name, create_time DESC")
	if dbErr := db.Find(&res).Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}

func (dao *statDao) GetLatestPingLog() (res []logModels.GfnCollectorLogPing, err common.GFError) {
	db := dao.Gm.Table(logModels.TableNameGfnCollectorLogPing)
	db.Select("DISTINCT ON (name) name, id, status, create_time, delay, loss")
	db.Order("name, create_time DESC")
	if dbErr := db.Find(&res).Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}
