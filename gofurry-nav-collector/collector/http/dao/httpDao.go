package dao

import (
	"strconv"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/http/models"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/abstract"
	"github.com/gofurry/gofurry-nav-collector/common/retention"
)

var newHTTPDao = new(httpDao)

func init() {
	newHTTPDao.Init()
}

type httpDao struct{ abstract.Dao }

func GetHTTPDao() *httpDao { return newHTTPDao }

func (dao httpDao) GetList() ([]models.GfnCollectorDomain, common.GFError) {
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

// 保留 count 条request历史记录
func (dao httpDao) DeleteByNum(count string) (int64, common.GFError) {
	// 转换保留条数为整数
	keepCount, err := strconv.Atoi(count)
	if err != nil {
		return 0, common.NewDaoError("count 格式错误: " + err.Error())
	}

	db := dao.Gm.Table(models.TableNameGfnCollectorLogHTTP)
	totalDeleted, deleteErr := retention.DeleteByNameLimit(db, models.TableNameGfnCollectorLogHTTP, keepCount, retention.DefaultBatchSize, 2*time.Minute, time.Second)
	if deleteErr != nil {
		return totalDeleted, common.NewDaoError("HTTP日志分批删除失败: " + deleteErr.Error())
	}

	return totalDeleted, nil
}
