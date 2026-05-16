package util

/*
 * @Desc: 工具类
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"fmt"
	"strconv"

	"github.com/gofurry/gofurry-game-collector/roof/env"
	"github.com/bwmarrin/snowflake"
)

var clusterId, _ = snowflake.NewNode(int64(env.GetServerConfig().ClusterId))

// 雪花算法生成新 ID
func GenerateId() int64 {
	id := clusterId.Generate()
	return id.Int64()
}

// int64 转字符串
func Int642String(i64 int64) string { return strconv.FormatInt(i64, 10) }

// float64 转字符串
func Float642String(f64 float64) string { return fmt.Sprintf("%.0f", f64) }

// int 转字符串
func Int2String(i int) string { return fmt.Sprintf("%d", i) }
