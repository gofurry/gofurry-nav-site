package dao

import (
	"errors"
	"sync"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/abstract"
	"gorm.io/gorm"
)

var newNavPageDao = new(navPageDao)
var navPageDaoMu sync.Mutex

type navPageDao struct{ abstract.Dao }

func GetNavPageDao() *navPageDao {
	navPageDaoMu.Lock()
	defer navPageDaoMu.Unlock()
	if newNavPageDao.Gm == nil {
		newNavPageDao.Init()
	}
	return newNavPageDao
}

func (dao *navPageDao) GetSiteList() (res []models.GfnSite, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfnSite + " AS s").
		Select(`
			s.id,
			s.name,
			s.name_en,
			json_build_object('domain', COALESCE(cd.domains, ARRAY[]::text[]))::text AS domain,
			s.info,
			s.info_en,
			s.create_time,
			s.update_time,
			s.country,
			s.nsfw,
			s.welfare,
			s.view_count,
			s.icon,
			s.deleted
		`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT array_agg(TRIM(COALESCE(prefix, '') || name) ORDER BY id ASC) AS domains
				FROM ` + models.TableNameGfnCollectorDomain + `
				WHERE site_id = s.id
					AND site_id > 0
					AND deleted IS NOT TRUE
					AND TRIM(COALESCE(prefix, '') || name) <> ''
			) cd ON TRUE
		`).
		Where("s.deleted IS NOT TRUE")
	db.Order("s.update_time DESC, s.id DESC")
	db.Find(&res)
	if dbErr := db.Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}

func (dao *navPageDao) GetSiteIndexList() (res []models.GfnSiteIndex, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfnSite + " AS s").
		Select(`
			s.id,
			json_build_object('domain', COALESCE(cd.domains, ARRAY[]::text[]))::text AS domain,
			s.update_time
		`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT array_agg(TRIM(COALESCE(prefix, '') || name) ORDER BY id ASC) AS domains
				FROM ` + models.TableNameGfnCollectorDomain + `
				WHERE site_id = s.id
					AND site_id > 0
					AND deleted IS NOT TRUE
					AND TRIM(COALESCE(prefix, '') || name) <> ''
			) cd ON TRUE
		`).
		Where("s.deleted IS NOT TRUE")
	db.Order("s.id ASC")
	db.Find(&res)
	if dbErr := db.Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}

func (dao *navPageDao) GetGroupList() (res []models.GfnSiteGroup, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfnSiteGroup)
	db.Order("priority ASC")
	db.Find(&res)
	if dbErr := db.Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}

func (dao *navPageDao) GetGroupMapList() (res []models.GfnSiteGroupMap, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfnSiteGroupMap)
	db.Order("group_id ASC, weight DESC, update_time DESC, id DESC, site_id ASC")
	db.Find(&res)
	if dbErr := db.Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}

func (dao *navPageDao) GetFeaturedSiteList() (res []models.GfnFeaturedSite, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfnFeaturedSite + " AS f").
		Joins("INNER JOIN " + models.TableNameGfnSite + " AS s ON s.id = f.site_id AND s.deleted IS NOT TRUE").
		Order("f.weight DESC, f.id DESC")
	db.Find(&res)
	if dbErr := db.Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return
}

// 按语言随机返回一条金句。
func (dao navPageDao) GetSayingByRandom(lang string) (*models.GfnSaying, common.GFError) {
	var res models.GfnSaying
	db := dao.Gm.Table(models.TableNameGfnSaying).
		Where("language = ?", lang).
		Order("random()")
	db.Limit(1).First(&res)
	if err := db.Error; err != nil {
		if lang != "zh" && errors.Is(err, gorm.ErrRecordNotFound) {
			return dao.GetSayingByRandom("zh")
		}
		return nil, common.NewDaoError(err.Error())
	}
	return &res, nil
}
