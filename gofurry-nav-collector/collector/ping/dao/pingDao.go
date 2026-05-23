package dao

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/ping/models"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/abstract"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	"github.com/gofurry/gofurry-nav-collector/common/retention"
)

var newPingDao = new(pingDao)

func init() {
	newPingDao.Init()
}

type pingDao struct{ abstract.Dao }

func GetPingDao() *pingDao { return newPingDao }

// 获取站点列表
func (dao pingDao) GetList() ([]models.Domain, common.GFError) {
	var res []models.Domain
	db := dao.Gm.Table(models.TableNameGfnSite).Select("domain").Where("deleted IS NOT TRUE")
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
	log.Info(fmt.Sprintf("Ping日志清理完成：累计删除 %d 条", totalDeleted))

	return totalDeleted, nil
}
