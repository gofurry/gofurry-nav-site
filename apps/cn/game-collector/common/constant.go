package common

/*
 * @Desc: 静态定义
 * @author: 福狼
 * @version: v1.0.0
 */

// 项目
const (
	COMMON_PROJECT_NAME = "gf-game-collector" // 项目名
)

// 状态标识
const (
	RETURN_FAILED           = 0                  //失败
	RETURN_SUCCESS          = 1                  //成功
	RETURN_RECORD_NOT_FOUND = "record-not-found" //未找到
)

// 时间
const (
	TIME_FORMAT_DATE = "2006-01-02 15:04:05"
)

// 请求头
const (
	USER_AGENT         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	ACCEPT_LANGUAGE_CN = "zh-CN,zh"
	ACCEPT_LANGUAGE_EN = "en"
	APPLICATION        = "application/json"
)
