package util

/*
 * @Desc: 工具类
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/bwmarrin/snowflake"
	"github.com/gofurry/gofurry-game-collector/roof/env"
)

var (
	clusterIDOnce sync.Once
	clusterIDNode *snowflake.Node
)

// 雪花算法生成新 ID
func GenerateId() int64 {
	clusterIDOnce.Do(func() {
		clusterIDNode, _ = snowflake.NewNode(int64(env.GetServerConfig().ClusterId))
	})
	id := clusterIDNode.Generate()
	return id.Int64()
}

// int64 转字符串
func Int642String(i64 int64) string { return strconv.FormatInt(i64, 10) }

// float64 转字符串
func Float642String(f64 float64) string { return fmt.Sprintf("%.0f", f64) }

// int 转字符串
func Int2String(i int) string { return fmt.Sprintf("%d", i) }
