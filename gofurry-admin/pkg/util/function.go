package util

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
	env "github.com/gofurry/awesome-fiber-template/v3/medium/config"
)

// Snowflake node state derived from the configured cluster id.
var (
	clusterIDOnce sync.Once
	clusterNode   *snowflake.Node
)

// GenerateId creates a distributed snowflake id.
func GenerateId() int64 {
	clusterIDOnce.Do(func() {
		clusterNode, _ = snowflake.NewNode(int64(env.GetServerConfig().ClusterId))
	})
	return clusterNode.Generate().Int64()
}

// GenerateRandomCode creates a random alphanumeric string.
func GenerateRandomCode(length int) string {
	letters := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	randCode := ""
	for i := 1; i <= length; i++ {
		randCode += string(letters[rand.Intn(len(letters))])
	}
	return randCode
}

// GetDateFormatStr formats a time value using the given layout.
func GetDateFormatStr(format string, date time.Time) string {
	return date.Format(format)
}

// String2Int converts a string into an int.
func String2Int(numString string) (int, error) {
	if strings.TrimSpace(numString) == "" {
		return 0, errors.New("string value cannot be empty")
	}
	id, err := strconv.Atoi(numString)
	return id, err
}

// Int642String converts an int64 to a string.
func Int642String(i64 int64) string { return strconv.FormatInt(i64, 10) }

// Int2String converts an int to a string.
func Int2String(i int) string { return fmt.Sprintf("%d", i) }

// String2Int64 converts a string into an int64.
func String2Int64(numString string) (int64, error) {
	if strings.TrimSpace(numString) == "" {
		return 0, errors.New("string value cannot be empty")
	}
	id, parseErr := strconv.ParseInt(strings.TrimSpace(numString), 10, 64)
	return id, parseErr
}
