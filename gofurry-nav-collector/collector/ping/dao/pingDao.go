package dao

import (
	"strconv"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/ping/models"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/abstract"
	"github.com/gofurry/gofurry-nav-collector/common/retention"
)

var newPingDao = new(pingDao)

func init() {
	newPingDao.Init()
}

type pingDao struct{ abstract.Dao }

func GetPingDao() *pingDao { return newPingDao }

// 获取站点列表
func (dao pingDao) GetList() ([]models.GfnCollectorDomain, common.GFError) {
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

// 保留 count 条ping历史记录
func (dao pingDao) DeleteByNum(count string) (int64, common.GFError) {
	keepCount, err := strconv.Atoi(count)
	if err != nil {
		return 0, common.NewDaoError("count 格式错误: " + err.Error())
	}

	db := dao.Gm.Table(models.TableNameGfnCollectorLogPing)
	totalDeleted, deleteErr := retention.DeleteByNameLimit(db, models.TableNameGfnCollectorLogPing, keepCount, retention.DefaultBatchSize, 2*time.Minute, time.Second)
	if deleteErr != nil {
		return totalDeleted, common.NewDaoError("Ping日志分批删除失败: " + deleteErr.Error())
	}

	return totalDeleted, nil
}
