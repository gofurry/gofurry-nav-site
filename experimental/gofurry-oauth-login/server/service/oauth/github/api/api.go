package api

import (
	"context"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gofurry/gofurry-oauth-login/api/proto/githuboauth"
	"github.com/gofurry/gofurry-oauth-login/common"
	"github.com/gofurry/gofurry-oauth-login/common/log"
	cs "github.com/gofurry/gofurry-oauth-login/common/service"
	"github.com/gofurry/gofurry-oauth-login/common/util"
	"github.com/gofurry/gofurry-oauth-login/env"
	"github.com/gofurry/gofurry-oauth-login/middleware"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var Api *api

type api struct{}

func NewApi() *api {
	return &api{}
}

func init() {
	Api = NewApi()
}

var once = sync.Once{}

var grpcServer *grpc.Server

// 实例的 ip:port
var ip, port = env.GetServerConfig().Server.IPAddress, env.GetServerConfig().Server.Port

//const (
//	ip   = "127.0.0.1"
//	port = "50056"
//)

func (api *api) Init() {
	once.Do(func() {
		// 初始化限流器
		middleware.InitRateLimiter(middleware.RateLimitConfig{QPS: 10, Burst: 20})

		// 初始化熔断器
		serviceName := "github-oauth"
		middleware.InitHystrix(serviceName, "GetAccessToken")
		middleware.InitHystrix(
			serviceName,
			"GetUserInfo",
			middleware.HystrixConfig{Timeout: 12000}, // 超时为12秒
		)
	})

	// GitHub 密钥
	cfg := env.GetServerConfig()
	if cfg.Github.ClientId == "" || cfg.Github.ClientSecret == "" {
		log.Error("GitHub Client ID或Secret未配置")
	}

	// 监听端口
	lis, err := net.Listen("tcp", ":"+port) // gRPC服务端口
	if err != nil {
		log.Error("监听失败: %v", err)
	}

	// 读取TLS证书
	creds, err := credentials.NewServerTLSFromFile(
		env.GetServerConfig().Key.TlsCrt,
		env.GetServerConfig().Key.TlsKey,
	)
	if err != nil {
		log.Error("加载TLS证书失败: %v", err)
		os.Exit(0)
	}

	// 创建gRPC服务器
	grpcServer = grpc.NewServer(
		grpc.Creds(creds), // 启用TLS
		grpc.UnaryInterceptor(middleware.RateLimitInterceptor), // 全局限流
	)

	githuboauth.RegisterGithubOAuthServiceServer(grpcServer, &githubOAuthServer{
		clientID:     cfg.Github.ClientId,
		clientSecret: cfg.Github.ClientSecret,
	})

	// 注册到etcd
	if err := cs.RegisterToEtcd(env.GetServerConfig().Etcd.EtcdKey, ip+":"+port); err != nil {
		log.Error("注册到etcd失败: %v", err)
	}

	log.Info("GitHub OAuth gRPC服务启动，端口" + port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Error("服务启动失败: %v", err)
	}
}

func (api *api) Stop() {
	if grpcServer != nil {
		grpcServer.GracefulStop()
		log.Info("gRPC服务已关闭")
	}
}

// =========================== 实现gRPC服务接口 ===========================
// 请求头
const (
	USER_AGENT  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	APPLICATION = "application/json"
)

type githubOAuthServer struct {
	githuboauth.UnimplementedGithubOAuthServiceServer
	clientID     string
	clientSecret string
}

// GetAccessToken 获取安全令牌
func (s *githubOAuthServer) GetAccessToken(ctx context.Context, req *githuboauth.GetAccessTokenRequest) (*githuboauth.GetAccessTokenResponse, error) {
	if req.Code == "" {
		log.Warn("GetAccessToken: code参数为空")
		return nil, status.Error(codes.InvalidArgument, "code不能为空")
	}

	// 熔断器配置
	serviceName := "github-oauth"
	commandKey := middleware.GetHystrixCommandKey(serviceName, "GetAccessToken") // 生成唯一标识
	// 三方请求配置
	url := "https://github.com/login/oauth/access_token"
	params := map[string]string{
		"client_id":     s.clientID,
		"client_secret": s.clientSecret,
		"code":          req.Code,
	}
	headers := map[string]string{
		"User-Agent": common.USER_AGENT,
		"Accept":     "application/json",
	}
	proxy := env.GetServerConfig().Proxy.Url
	var respStr string

	// 执行熔断器逻辑
	err := middleware.HystrixDo(
		commandKey,
		// 正常业务逻辑
		func() error {
			var httpErr error
			respStr, httpErr = util.GetByHttpWithParams(url, headers, params, 10*time.Second, &proxy)
			return httpErr
		},
		// 熔断降级逻辑
		func(err error) error {
			log.Error("GetAccessToken熔断器触发，错误: ", err)
			return status.Error(codes.Unavailable, "第三方服务暂时不可用，请稍后重试")
		},
	)

	// 熔断器执行失败
	if err != nil {
		return nil, err
	}

	// 解析响应
	accessToken := gjson.Get(respStr, "access_token").String()
	if accessToken == "" {
		errMsg := gjson.Get(respStr, "error").String()
		if errMsg == "" {
			errMsg = "未返回有效access_token"
		}
		log.Warn("GetAccessToken解析失败: ", errMsg)
		return nil, status.Error(codes.Aborted, "获取令牌失败: "+errMsg)
	}

	log.Info("获取access_token成功：%s", accessToken[:10]+"...")
	return &githuboauth.GetAccessTokenResponse{AccessToken: accessToken}, nil
}

// GetUserInfo 获取用户信息
func (s *githubOAuthServer) GetUserInfo(ctx context.Context, req *githuboauth.GetUserInfoRequest) (*githuboauth.GetUserInfoResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token不能为空")
	}

	// 熔断器配置
	serviceName := "github-oauth"
	commandKey := middleware.GetHystrixCommandKey(serviceName, "GetUserInfo")
	// 三方请求配置
	url := "https://api.github.com/user"
	headers := map[string]string{
		"User-Agent":    USER_AGENT,
		"Accept":        APPLICATION,
		"Authorization": "token " + req.AccessToken,
	}
	proxy := env.GetServerConfig().Proxy.Url
	var respStr string

	// 执行熔断器逻辑
	err := middleware.HystrixDo(
		commandKey,
		// 正常业务逻辑
		func() error {
			var httpErr error
			respStr, httpErr = util.GetByHttpWithParams(url, headers, map[string]string{}, 10*time.Second, &proxy)
			return httpErr
		},
		// 熔断降级逻辑
		func(err error) error {
			log.Error("GetUserInfo熔断器触发，错误: ", err)
			return status.Error(codes.Unavailable, "第三方服务暂时不可用，请稍后重试")
		},
	)

	if err != nil {
		return nil, err
	}

	// 解析用户信息
	userInfo := &githuboauth.UserInfo{
		Login:     gjson.Get(respStr, "login").String(),
		AvatarUrl: gjson.Get(respStr, "avatar_url").String(),
		Email:     gjson.Get(respStr, "email").String(),
		Name:      gjson.Get(respStr, "name").String(),
	}
	if userInfo.Login == "" {
		log.Warn("GetUserInfo: 未解析到有效用户信息，响应: ", respStr[:100]+"...")
		return nil, status.Error(codes.Aborted, "获取用户信息失败")
	}

	return &githuboauth.GetUserInfoResponse{UserInfo: userInfo}, nil
}
