package service

import (
	"time"

	"github.com/gofurry/gofurry-user/common"
	ca "github.com/gofurry/gofurry-user/common/abstract"
	"github.com/gofurry/gofurry-user/common/log"
	cs "github.com/gofurry/gofurry-user/common/service"
	"github.com/gofurry/gofurry-user/common/util"
	"github.com/gofurry/gofurry-user/roof/env"
)

type emailService struct{}

var emailSingleton = new(emailService)

func GetEmailService() *emailService { return emailSingleton }

// 发送邮箱验证码
func (svc *emailService) SendEmail(email string) common.GFError {
	// 入参校验
	req := struct {
		Email string `validate:"required,email,min=1,max=100" label:"邮箱" json:"email"`
	}{Email: email}

	errorResults := ca.ValidateServiceApi.Validate(req)
	if len(errorResults) > 0 {
		log.Warn("(svc *emailService) SendEmail 入参有误")
		return common.NewServiceError(errorResults[0].ErrMsg)
	}

	// 发送邮箱验证码
	code, err := cs.EmailSendCode(email)
	if err != nil {
		return err
	}
	// 邮件验证码存redis
	code = util.CreateMD5(code + env.GetServerConfig().Auth.AuthSalt)
	_ = cs.SetExpire("email:"+email, code, 300*time.Second)
	return nil
}
