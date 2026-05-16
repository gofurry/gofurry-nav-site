package service

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/gofurry/gofurry-user/apps/user/dao"
	"github.com/gofurry/gofurry-user/apps/user/models"
	"github.com/gofurry/gofurry-user/common"
	ca "github.com/gofurry/gofurry-user/common/abstract"
	"github.com/gofurry/gofurry-user/common/log"
	cm "github.com/gofurry/gofurry-user/common/models"
	cs "github.com/gofurry/gofurry-user/common/service"
	"github.com/gofurry/gofurry-user/common/util"
	"github.com/gofurry/gofurry-user/roof/env"
	"github.com/gofiber/fiber/v2"
)

type userService struct{}

var userSingleton = new(userService)

func GetUserService() *userService { return userSingleton }

// 头像
var Avatars = []string{"龙", "虎", "狼"}

// Login 用户登录
func (svc *userService) Login(c *fiber.Ctx, req models.UserLoginRequest) (tokenStr string, err common.GFError) {
	// 检验入参合法性
	errorResults := ca.ValidateServiceApi.Validate(req)
	if errorResults != nil {
		return tokenStr, common.NewServiceError("传入参数有误")
	}
	// 查找是否有该用户,支持账户名和邮箱登录
	userRecord, err := dao.GetUserDao().FindOneByEmail(req.Name)
	if err != nil {
		return "", common.NewServiceError("未找到该邮箱的账户记录.")
	}

	// 用户是否被封禁
	if userRecord.Status == "banned" {
		return "", common.NewServiceError("该用户已被封禁")
	}

	// 解密前端密码
	//decryptPassword, decryptErr := util.DecryptPassword(req.Password, env.GetServerConfig().Key.LoginPrivate)
	//if decryptErr != nil {
	//	return "", common.NewServiceError(decryptErr.Error())
	//}
	decryptPassword := req.Password

	// 校验密码
	password := util.CreateMD5(decryptPassword + env.GetServerConfig().Auth.AuthSalt)
	if password != userRecord.Password {
		err = common.NewServiceError("密码错误.")
		return
	}
	// 生成 token 存 redis
	tokenStr, tokenErr := util.NewToken(strconv.FormatInt(userRecord.ID, 10), userRecord.Name)
	if tokenErr != nil {
		log.Error(tokenErr)
		err = common.NewServiceError("创建Token错误.")
		return
	}

	// 登录记录
	newLoginLog := &models.GfLoginLog{
		UserID:     userRecord.ID,
		IP:         util.GetIP(c),
		Agent:      c.Get("User-Agent"),
		CreateTime: cm.LocalTime(time.Now()),
		LoginType:  "login",
	}
	newLoginLog.SetNewId()

	err = dao.GetUserLogDao().Add(newLoginLog)
	if err != nil {
		log.Error("登录记录入库失败: ", err)
	}

	// token 存 redis
	cs.SetExpire("jwt:"+tokenStr, tokenStr, common.JWT_RELET_NUM*time.Hour)
	currentUser := models.CurrentUser{
		ID:   userRecord.ID,
		Name: userRecord.Name,
	}
	c.Locals(common.COMMON_AUTH_CURRENT, currentUser)

	return tokenStr, nil
}

// Register 用户注册
func (svc *userService) Register(req models.UserRegisterRequest) (err common.GFError) {
	// 入参校验
	reqErr := ca.ValidateServiceApi.Validate(req)
	if reqErr != nil {
		return common.NewServiceError("入参有误: " + reqErr[0].ErrMsg)
	}
	// 注册查重
	_, err = dao.GetUserDao().FindOneByEmail(req.Email)
	if err == nil {
		return common.NewServiceError("邮箱已被注册")
	}
	// 校对验证码
	code, err := cs.GetString("email:" + req.Email)
	if code != util.CreateMD5(req.Code+env.GetServerConfig().Auth.AuthSalt) || err != nil {
		return common.NewServiceError("邮箱验证码错误")
	}

	// 解密前端密码
	//decryptNewPassword, decryptErr := util.DecryptPassword(req.Password, env.GetServerConfig().Key.LoginPrivate)
	//if decryptErr != nil {
	//	return common.NewServiceError(decryptErr.Error())
	//}
	decryptNewPassword := req.Password

	// 插入用户信息
	userTab := &models.GfUser{
		Password: util.CreateMD5(decryptNewPassword + env.GetServerConfig().Auth.AuthSalt),
		Nickname: req.Name,
		Oauth:    false,
		Role:     req.Role,
		Status:   "normal",
		Avatar:   Avatars[rand.Intn(len(Avatars))],
	}
	userTab.SetNewId()
	userTab.SetName("UID:" + util.Int642String(userTab.ID))
	userTab.CreateTime = cm.LocalTime(time.Now())
	userTab.UpdateTime = userTab.CreateTime
	userTab.Email = &req.Email
	defaultInfo := "暂无个人简介."
	userTab.Info = &defaultInfo

	err = dao.GetUserDao().Add(userTab)
	if err != nil {
		return common.NewServiceError("注册记录入库失败.")
	}
	return nil
}
