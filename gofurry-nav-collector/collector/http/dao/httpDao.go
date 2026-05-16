package dao

import (
	"context"
	"strconv"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/http/models"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/abstract"
)

var newHTTPDao = new(httpDao)

func init() {
	newHTTPDao.Init()
}

type httpDao struct{ abstract.Dao }

func GetHTTPDao() *httpDao { return newHTTPDao }

func (dao httpDao) GetList() ([]models.GfnCollectorDomain, common.GFError) {
	var res []models.GfnCollectorDomain
	db := dao.Gm.Table(models.TableNameGfnCollectorDomain)
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

	var totalDeleted int64 = 0
	batchSize := 1000 // 每批删除1000条
	db := dao.Gm.Table(models.TableNameGfnCollectorLogHTTP)

	// 分批删除循环
	for {
		// 高性能 DELETE ... USING 写法
		sql := `
			DELETE FROM ` + models.TableNameGfnCollectorLogHTTP + ` t1
			USING (
				SELECT 
				  id,
				  ROW_NUMBER() OVER (
					PARTITION BY name 
					ORDER BY create_time DESC
				  ) AS rn
				FROM ` + models.TableNameGfnCollectorLogHTTP + `
				LIMIT ?  -- 限制单批扫描数据量，降低负载
			) t2
			WHERE t1.id = t2.id AND t2.rn > ?;`

		// 设置单批执行超时
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		// 执行单批删除
		result := db.WithContext(ctx).Exec(sql, batchSize, keepCount)
		if err := db.Error; err != nil {
			return totalDeleted, common.NewDaoError("HTTP日志分批删除失败: " + err.Error())
		}

		// 累计删除条数
		deleted := result.RowsAffected
		totalDeleted += deleted
		if deleted < int64(batchSize) {
			break
		}

		// 每批删除后休眠1秒
		time.Sleep(1 * time.Second)
	}

	return totalDeleted, nil
}
