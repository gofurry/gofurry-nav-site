package service

/*
 * @Desc: 三方登录
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"time"

	"github.com/gofurry/gofurry-user/common"
	"github.com/gofurry/gofurry-user/common/abstract"
	"github.com/gofurry/gofurry-user/common/log"
	"github.com/gofurry/gofurry-user/common/util"
	"github.com/gofurry/gofurry-user/roof/env"
	"github.com/tidwall/gjson"
)

var githubConfig = &abstract.Oauth{
	ClientId:     env.GetServerConfig().Github.ClientId,
	ClientSecret: env.GetServerConfig().Github.ClientSecret,
	RedirectUrl:  env.GetServerConfig().Github.RedirectUrl,
}

var giteeConfig = &abstract.Oauth{
	ClientId:     env.GetServerConfig().Gitee.ClientId,
	ClientSecret: env.GetServerConfig().Gitee.ClientSecret,
	RedirectUrl:  env.GetServerConfig().Gitee.RedirectUrl,
}

// 设置请求头，明确指定语言为中文
var headersMap = map[string]string{
	"User-Agent": common.USER_AGENT,
	"Accept":     common.APPLICATION,
}

// 获取 Github accessToken
func GetGithubToken(code string) (string, common.GFError) {
	//请求github
	url := "https://github.com/login/oauth/access_token"
	// 设置参数
	paramsMap := map[string]string{
		"client_id":     githubConfig.ClientId,
		"client_secret": githubConfig.ClientSecret,
		"code":          code,
	}

	// 请求
	respDataStr, httpErr := util.GetByHttpWithParams(url, headersMap, paramsMap, 10000*time.Millisecond, &env.GetServerConfig().Proxy.Url)
	if httpErr != nil {
		log.Warn(httpErr)
		return "", common.NewServiceError(httpErr.Error())
	}
	return gjson.Get(respDataStr, "access_token").String(), nil
}

// 通过 access_token 获取 Github 用户信息
func GetGithubUserInfo(accessToken string) (string, common.GFError) {
	url := `https://api.github.com/user`
	// 设置参数
	paramsMap := map[string]string{}
	// 设置请求头，明确指定语言为中文
	headersMap["Authorization"] = "token " + accessToken
	// 请求
	respDataStr, httpErr := util.GetByHttpWithParams(url, headersMap, paramsMap, 10000*time.Millisecond, &env.GetServerConfig().Proxy.Url)
	if httpErr != nil {
		log.Warn(httpErr)
		return "", common.NewServiceError(httpErr.Error())
	}
	return respDataStr, nil
}

// 获取 Gitee accessToken
func GetGiteeToken(code string) (string, common.GFError) {
	//请求gitee
	url := "https://gitee.com/oauth/token"
	// 设置参数
	paramsMap := map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"client_id":     giteeConfig.ClientId,
		"redirect_uri":  giteeConfig.RedirectUrl,
		"client_secret": giteeConfig.ClientSecret,
	}

	// 请求
	respDataStr, httpErr := util.PostByHttpWithParams(url, headersMap, paramsMap, 10000*time.Millisecond, &env.GetServerConfig().Proxy.Url)
	if httpErr != nil {
		log.Warn(httpErr)
		return "", common.NewServiceError(httpErr.Error())
	}
	return gjson.Get(respDataStr, "access_token").String(), nil
}

// 通过 access_token 获取 Gitee 用户信息
func GetGiteeUserInfo(accessToken string) (string, common.GFError) {
	url := "https://gitee.com/api/v5/user"
	// 设置参数
	paramsMap := map[string]string{
		"access_token": accessToken,
	}
	// 请求
	respDataStr, httpErr := util.GetByHttpWithParams(url, headersMap, paramsMap, 10000*time.Millisecond, &env.GetServerConfig().Proxy.Url)
	if httpErr != nil {
		log.Warn(httpErr)
		return "", common.NewServiceError(httpErr.Error())
	}
	return respDataStr, nil
}
