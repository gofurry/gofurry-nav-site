package dao

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/ping/models"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/abstract"
	"github.com/gofurry/gofurry-nav-collector/common/log"
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

	var totalDeleted int64 = 0
	batchSize := 1000 // 每批删除 1000 条
	db := dao.Gm.Table(models.TableNameGfnCollectorLogPing)

	for {
		// 单批删除 SQL
		sql := `
			DELETE FROM ` + models.TableNameGfnCollectorLogPing + ` t1
			USING (
				SELECT 
				  id,
				  ROW_NUMBER() OVER (
					PARTITION BY name 
					ORDER BY create_time DESC
				  ) AS rn
				FROM ` + models.TableNameGfnCollectorLogPing + `
				LIMIT ?  -- 限制单批扫描数据量
			) t2
			WHERE t1.id = t2.id AND t2.rn > ?;`

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		// 执行单批删除
		result := db.WithContext(ctx).Exec(sql, batchSize, keepCount)
		if err := db.Error; err != nil {
			return totalDeleted, common.NewDaoError("分批删除失败: " + err.Error())
		}

		// 累计删除条数, 判断是否结束
		deleted := result.RowsAffected
		totalDeleted += deleted
		if deleted < int64(batchSize) {
			break // 没有更多数据可删, 退出循环
		}

		// 每批删除后休眠 1 秒, 降低数据库压力
		time.Sleep(1 * time.Second)
		log.Info(fmt.Sprintf("Ping日志清理：单批删除 %d 条，累计删除 %d 条", deleted, totalDeleted))
	}

	return totalDeleted, nil
}
