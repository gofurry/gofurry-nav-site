package dao

import (
	"strconv"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/dns/models"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/abstract"
	"github.com/gofurry/gofurry-nav-collector/common/retention"
)

var newDNSDao = new(dnsDao)

func init() {
	newDNSDao.Init()
}

type dnsDao struct{ abstract.Dao }

func GetDNSDao() *dnsDao { return newDNSDao }

func (dao dnsDao) GetList() ([]models.GfnCollectorDomain, common.GFError) {
	var res []models.GfnCollectorDomain
	db := dao.Gm.Table(models.TableNameGfnCollectorDomain).Where("deleted IS NOT TRUE")
	db.Find(&res)
	if err := db.Error; err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return res, nil
}

// 保留 count 条request历史记录
func (dao dnsDao) DeleteByNum(count string) (int64, common.GFError) {
	// 转换保留条数为整数
	keepCount, err := strconv.Atoi(count)
	if err != nil {
		return 0, common.NewDaoError("count 格式错误: " + err.Error())
	}

	db := dao.Gm.Table(models.TableNameGfnCollectorLogDn)
	totalDeleted, deleteErr := retention.DeleteByNameLimit(db, models.TableNameGfnCollectorLogDn, keepCount, retention.DefaultBatchSize, 2*time.Minute, time.Second)
	if deleteErr != nil {
		return totalDeleted, common.NewDaoError("DNS日志分批删除失败: " + deleteErr.Error())
	}

	return totalDeleted, nil
}
