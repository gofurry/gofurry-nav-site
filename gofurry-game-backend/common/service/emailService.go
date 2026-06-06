package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	"github.com/gofurry/gofurry-game-backend/common/util"
	"github.com/gofurry/gofurry-game-backend/roof/env"
	"gopkg.in/gomail.v2"
)

/*
 * @Desc: email服务
 * @author: 福狼
 * @version: v1.0.2
 */

// TODO:发件人池, 免费账号单日仅可发送1000条邮件, 大需求量可使用云服务

type EmailService struct {
	dialer     *gomail.Dialer // SMTP 拨号器
	sender     string         // 发件人邮箱
	senderName string         // 发件人名称
}

// 全局邮箱服务实例
var (
	emailService *EmailService
	emailOnce    sync.Once
)

// GetEmailService 获取邮箱服务单例
func GetEmailService() *EmailService {
	emailOnce.Do(func() {
		var err error
		emailService, err = newEmailService()
		if err != nil {
			log.Fatal("邮箱服务初始化失败", "error", err)
		}
	})
	return emailService
}

// newEmailService 初始化邮箱服务
func newEmailService() (*EmailService, error) {
	cfg := env.GetServerConfig()
	if cfg == nil {
		return nil, errors.New("邮箱配置为空")
	}

	// 校验核心配置
	if cfg.Email.EmailHost == "" {
		return nil, errors.New("SMTP 服务器地址不能为空")
	}
	if cfg.Email.EmailPort <= 0 {
		return nil, errors.New("SMTP 端口必须为正整数")
	}
	if cfg.Email.EmailUser == "" || cfg.Email.EmailPassword == "" {
		return nil, errors.New("邮箱账号/密码不能为空")
	}

	// 创建 SMTP 拨号器
	dialer := gomail.NewDialer(
		cfg.Email.EmailHost,
		cfg.Email.EmailPort,
		cfg.Email.EmailUser,
		cfg.Email.EmailPassword,
	)

	return &EmailService{
		dialer:     dialer,
		sender:     cfg.Email.EmailUser,
		senderName: "GoFurry 邮件服务",
	}, nil
}

// SendCode 发送邮箱验证码
func (es *EmailService) SendCode(email string) (string, common.GFError) {
	// 参数校验
	if email == "" {
		slog.Warn("[EmailService] 发送验证码失败", "reason", "邮箱为空")
		return "", common.NewServiceError("邮箱地址不能为空")
	}
	if !es.IsEmailValid(email) {
		slog.Warn("[EmailService] 发送验证码失败", "reason", "邮箱格式错误", "email", email)
		return "", common.NewServiceError("邮箱格式不正确")
	}

	// 生成验证码
	code := util.GenerateRandomCode(common.EMAIL_CODE_LENGTH)
	if code == "" {
		slog.Error("[EmailService] 生成验证码失败", "email", email)
		return "", common.NewServiceError("验证码生成失败, 请重试")
	}

	// 构建邮件内容
	msg, err := es.buildCodeEmailContent(code)
	if err != nil {
		slog.Error("[EmailService] 构建邮件内容失败", "email", email, "error", err)
		return "", common.NewServiceError("邮件内容构建失败")
	}

	// 异步发送邮件
	var sendErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		sendErr = es.sendEmailWithRetry(email, "GoFurry 邮箱验证码", msg)
	}()

	// 等待发送结果(12秒超时, 避免业务阻塞)
	select {
	case <-done:
	case <-time.After(12 * time.Second):
		slog.Warn("[EmailService] 邮件发送超时", "email", email)
		return "", common.NewServiceError("邮件发送超时, 请稍后查看邮箱")
	}

	// 处理发送结果
	if sendErr != nil {
		slog.Error("[EmailService] 邮件发送失败", "email", email, "error", sendErr)
		return "", common.NewServiceError("邮件发送失败, 请稍后重试")
	}

	slog.Info("[EmailService] 验证码发送成功", "email", email, "code_length", len(code))
	return code, nil
}

// SendEmail 通用邮件发送 抄送/密送
func (es *EmailService) SendEmail(to string, cc []string, bcc []string, subject, htmlBody string) common.GFError {
	// 参数校验
	if to == "" {
		slog.Warn("[EmailService] 发送邮件失败", "reason", "收件人邮箱为空")
		return common.NewServiceError("收件人邮箱不能为空")
	}
	if !es.IsEmailValid(to) {
		slog.Warn("[EmailService] 发送邮件失败", "reason", "收件人邮箱格式错误", "email", to)
		return common.NewServiceError("收件人邮箱格式不正确")
	}
	if subject == "" {
		slog.Warn("[EmailService] 发送邮件失败", "reason", "邮件标题为空", "to", to)
		return common.NewServiceError("邮件标题不能为空")
	}

	// 构建邮件消息
	m := gomail.NewMessage()
	from := es.mimeEncode(es.senderName) + " <" + es.sender + ">"
	m.SetHeader("From", from)
	m.SetHeader("To", to)

	// 处理抄送/密送
	if len(cc) > 0 {
		m.SetHeader("Cc", cc...)
	}
	if len(bcc) > 0 {
		m.SetHeader("Bcc", bcc...)
	}

	m.SetHeader("Subject", es.mimeEncode(subject))
	m.SetBody("text/html; charset=UTF-8", htmlBody) // 防XSS

	// 异步发送邮件
	var sendErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		sendErr = es.sendEmailWithRetry(to, subject, htmlBody)
	}()

	// 等待发送结果
	select {
	case <-done:
	case <-time.After(12 * time.Second):
		slog.Warn("[EmailService] 邮件发送超时", "to", to, "subject", subject)
		return common.NewServiceError("邮件发送超时, 请稍后查看邮箱")
	}

	if sendErr != nil {
		slog.Error("[EmailService] 邮件发送失败", "to", to, "subject", subject, "error", sendErr)
		return common.NewServiceError("邮件发送失败, 请稍后重试")
	}

	slog.Info("[EmailService] 邮件发送成功", "to", to, "subject", subject)
	return nil
}

// SendEmailWithAttachment 发送带附件的邮件
func (es *EmailService) SendEmailWithAttachment(to, subject, htmlBody string, attachments map[string]string) common.GFError {
	// 参数校验
	if to == "" || !es.IsEmailValid(to) {
		slog.Warn("[EmailService] 带附件邮件发送失败", "reason", "收件人邮箱错误", "email", to)
		return common.NewServiceError("收件人邮箱格式不正确")
	}
	if len(attachments) == 0 {
		slog.Warn("[EmailService] 带附件邮件发送失败", "reason", "附件列表为空", "to", to)
		return common.NewServiceError("附件列表不能为空")
	}

	// 构建邮件
	m := gomail.NewMessage()
	from := es.mimeEncode(es.senderName) + " <" + es.sender + ">"
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", es.mimeEncode(subject))
	m.SetBody("text/html; charset=UTF-8", htmlBody)

	// 添加附件
	for filePath, fileName := range attachments {
		m.Attach(filePath, gomail.Rename(fileName))
	}

	// 异步发送
	var sendErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		maxRetries := 2
		var err error
		for i := 0; i <= maxRetries; i++ {
			err = es.dialer.DialAndSend(m)
			if err == nil {
				sendErr = nil
				return
			}
			log.Warn("[EmailService] 带附件邮件发送重试", "retry", i+1, "to", to, "error", err)
			if i < maxRetries {
				time.Sleep(time.Duration(i+1) * time.Second)
			}
		}
		sendErr = fmt.Errorf("重试%d次后仍发送失败: %w", maxRetries, err)
	}()

	// 超时控制
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		slog.Warn("[EmailService] 带附件邮件发送超时", "to", to)
		return common.NewServiceError("带附件邮件发送超时")
	}

	if sendErr != nil {
		slog.Error("[EmailService] 带附件邮件发送失败", "to", to, "error", sendErr)
		return common.NewServiceError("带附件邮件发送失败")
	}

	slog.Info("[EmailService] 带附件邮件发送成功", "to", to, "attachment_count", len(attachments))
	return nil
}

// SendActivationEmail 发送激活链接邮件
func (es *EmailService) SendActivationEmail(to, title string, activateLink string, activateText string, text string, duration string) common.GFError {
	year := strconv.Itoa(time.Now().Year())
	// 构建密码重置邮件内容
	htmlBody := `
	<html>
	<head>
		<meta charset="UTF-8">
		<title>GoFurry 激活邮件</title>
		<style>
			body { font-family: "Microsoft YaHei", "Helvetica Neue", sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
			.container { background-color: #f9f9f9; border-radius: 8px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
			.logo { color: #2c3e50; font-size: 24px; font-weight: bold; margin-bottom: 20px; display: flex; align-items: center; }
			.logo span { color: #3498db; margin-right: 8px; }
			.greeting { font-size: 18px; margin-bottom: 20px; }
			.link-box { background-color: #fff; border: 1px dashed #ddd; border-radius: 4px; padding: 20px; text-align: center; margin: 25px 0; }
			.reset-link { font-size: 16px; color: #3498db; text-decoration: none; padding: 10px 20px; background-color: #e8f4fd; border-radius: 4px; }
			.note { color: #666; font-size: 14px; margin: 20px 0; }
			.warning { color: #e74c3c; font-size: 13px; padding: 10px; background-color: #fef0f0; border-radius: 4px; margin-top: 15px; }
			.footer { margin-top: 30px; color: #999; font-size: 12px; text-align: center; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="logo">
				<span>🐺</span> GoFurry
			</div>
			<div class="greeting">您好！</div>
			<p>` + text + `</p>
			<div class="link-box">
				<a href="` + activateLink + `" class="reset-link">` + activateText + `</a>
			</div>
			<p class="note">
				• 该链接有效期为 <strong>` + duration + `</strong>，请在有效期内完成操作<br>
				• 如果非您本人操作，请忽略本邮件，账号安全不会受影响
			</p>
			<div class="footer">
				<p>GoFurry 邮箱服务 © ` + year + `</p>
			</div>
		</div>
	</body>
	</html>
	`
	// 调用通用发送方法
	return es.SendEmail(to, nil, nil, title, htmlBody)
}

// SendLotteryEmail 发送中奖邮件
func (es *EmailService) SendLotteryEmail(to, title string, code string, text string) common.GFError {
	year := strconv.Itoa(time.Now().Year())
	// 构建中奖邮件内容
	htmlBody := `
	<html>
	<head>
		<meta charset="UTF-8">
		<title>GoFurry 抽奖服务-获奖</title>
		<style>
			body { font-family: "Microsoft YaHei", "Helvetica Neue", sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
			.container { background-color: #f9f9f9; border-radius: 8px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
			.logo { color: #2c3e50; font-size: 24px; font-weight: bold; margin-bottom: 20px; display: flex; align-items: center; }
			.logo span { color: #3498db; margin-right: 8px; }
			.greeting { font-size: 18px; margin-bottom: 20px; }
			.code-box { background-color: #fff; border: 1px dashed #ddd; border-radius: 4px; padding: 20px; text-align: center; margin: 25px 0; }
			.code { font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #2c3e50; margin: 0; }
			.note { color: #666; font-size: 14px; margin: 20px 0; }
			.warning { color: #e74c3c; font-size: 13px; padding: 10px; background-color: #fef0f0; border-radius: 4px; margin-top: 15px; }
			.footer { margin-top: 30px; color: #999; font-size: 12px; text-align: center; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="logo">
				<span>🐺</span> GoFurry
			</div>
			<div class="greeting">您好！</div>
			<p>` + text + `</p>
			<div class="code-box">
				<p class="code">[ ` + code + ` ]</p>
			</div>
			<p class="note">
				• 祝您能够享用本站的礼品
			</p>
			<div class="warning">
				如果您未发起此操作, 请忽略本邮件, 您的账号安全不会受到影响。
			</div>
			<div class="footer">
				<p>GoFurry 邮箱服务 © ` + year + `</p>
			</div>
		</div>
	</body>
	</html>
	`
	// 调用通用发送方法
	return es.SendEmail(to, nil, nil, title, htmlBody)
}

// SendPasswordResetEmail 发送密码重置邮件
func (es *EmailService) SendPasswordResetEmail(to, resetLink string) common.GFError {
	year := strconv.Itoa(time.Now().Year())
	// 构建密码重置邮件内容
	htmlBody := `
	<html>
	<head>
		<meta charset="UTF-8">
		<title>GoFurry 密码重置</title>
		<style>
			body { font-family: "Microsoft YaHei", "Helvetica Neue", sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
			.container { background-color: #f9f9f9; border-radius: 8px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
			.logo { color: #2c3e50; font-size: 24px; font-weight: bold; margin-bottom: 20px; display: flex; align-items: center; }
			.logo span { color: #3498db; margin-right: 8px; }
			.greeting { font-size: 18px; margin-bottom: 20px; }
			.link-box { background-color: #fff; border: 1px dashed #ddd; border-radius: 4px; padding: 20px; text-align: center; margin: 25px 0; }
			.reset-link { font-size: 16px; color: #3498db; text-decoration: none; padding: 10px 20px; background-color: #e8f4fd; border-radius: 4px; }
			.note { color: #666; font-size: 14px; margin: 20px 0; }
			.warning { color: #e74c3c; font-size: 13px; padding: 10px; background-color: #fef0f0; border-radius: 4px; margin-top: 15px; }
			.footer { margin-top: 30px; color: #999; font-size: 12px; text-align: center; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="logo">
				<span>🐺</span> GoFurry
			</div>
			<div class="greeting">您好！</div>
			<p>您正在申请重置 GoFurry 账号密码，点击下方链接完成重置：</p>
			<div class="link-box">
				<a href="` + resetLink + `" class="reset-link">点击重置密码</a>
			</div>
			<p class="note">
				• 该链接有效期为 <strong>15分钟</strong>，请在有效期内完成操作<br>
				• 如果非您本人操作，请忽略本邮件，账号安全不会受影响
			</p>
			<div class="footer">
				<p>GoFurry 邮箱服务 © ` + year + `</p>
			</div>
		</div>
	</body>
	</html>
	`
	// 调用通用发送方法
	return es.SendEmail(to, nil, nil, "GoFurry 密码重置", htmlBody)
}

// SendAccountNoticeEmail 发送账号异常操作提醒
func (es *EmailService) SendAccountNoticeEmail(to, operation, ip, timeStr string) common.GFError {
	year := strconv.Itoa(time.Now().Year())
	// 构建账号提醒邮件内容
	htmlBody := `
	<html>
	<head>
		<meta charset="UTF-8">
		<title>GoFurry 账号安全提醒</title>
		<style>
			body { font-family: "Microsoft YaHei", "Helvetica Neue", sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
			.container { background-color: #f9f9f9; border-radius: 8px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
			.logo { color: #2c3e50; font-size: 24px; font-weight: bold; margin-bottom: 20px; display: flex; align-items: center; }
			.logo span { color: #3498db; margin-right: 8px; }
			.greeting { font-size: 18px; margin-bottom: 20px; }
			.info-box { background-color: #fff; border: 1px dashed #ddd; border-radius: 4px; padding: 20px; margin: 25px 0; }
			.info-item { line-height: 2; font-size: 14px; }
			.warning { color: #e74c3c; font-size: 13px; padding: 10px; background-color: #fef0f0; border-radius: 4px; margin-top: 15px; }
			.footer { margin-top: 30px; color: #999; font-size: 12px; text-align: center; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="logo">
				<span>🐺</span> GoFurry
			</div>
			<div class="greeting">您好！</div>
			<p>检测到您的账号有异常操作，详情如下：</p>
			<div class="info-box">
				<div class="info-item">操作类型：` + operation + `</div>
				<div class="info-item">操作IP：` + ip + `</div>
				<div class="info-item">操作时间：` + timeStr + `</div>
			</div>
			<div class="warning">
				如果非您本人操作，请立即修改密码并检查账号安全！
			</div>
			<div class="footer">
				<p>GoFurry 邮箱服务 © ` + year + `</p>
			</div>
		</div>
	</body>
	</html>
	`
	// 调用通用发送方法
	return es.SendEmail(to, nil, nil, "GoFurry 账号安全提醒", htmlBody)
}

// SendBatchEmail 批量发送邮件（异步）
func (es *EmailService) SendBatchEmail(toList []string, subject, htmlBody string) common.GFError {
	if len(toList) == 0 {
		slog.Warn("[EmailService] 批量邮件发送失败", "reason", "收件人列表为空")
		return common.NewServiceError("收件人列表不能为空")
	}

	// 异步批量发送
	go func() {
		successCount := 0
		failCount := 0
		failEmails := make([]string, 0)

		for _, to := range toList {
			if !es.IsEmailValid(to) {
				failCount++
				failEmails = append(failEmails, to)
				continue
			}
			if err := es.SendEmail(to, nil, nil, subject, htmlBody); err == nil {
				successCount++
			} else {
				failCount++
				failEmails = append(failEmails, to)
			}
			// 批量发送间隔，避免限流
			time.Sleep(200 * time.Millisecond)
		}

		// 记录批量发送结果
		slog.Info("[EmailService] 批量邮件发送完成",
			"total", len(toList),
			"success", successCount,
			"fail", failCount,
			"fail_emails", failEmails,
		)
	}()

	return nil
}

// IsEmailValid 校验邮箱是否合法
func (es *EmailService) IsEmailValid(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// sendEmailWithRetry 发送邮件
func (es *EmailService) sendEmailWithRetry(to, subject, body string) error {
	// 构建邮件消息
	m := gomail.NewMessage()
	from := es.mimeEncode(es.senderName) + " <" + es.sender + ">"
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", es.mimeEncode(subject))
	m.SetBody("text/html; charset=UTF-8", body) // 防XSS

	// 重试
	maxRetries := 2
	var err error
	for i := 0; i <= maxRetries; i++ {
		err = es.dialer.DialAndSend(m)
		if err == nil {
			return nil
		}
		log.Warn("[EmailService] 邮件发送重试", "retry", i+1, "email", to, "error", err)
		if i < maxRetries {
			time.Sleep(time.Duration(i+1) * time.Second) // 指数退避
		}
	}

	return fmt.Errorf("重试%d次后仍发送失败: %w", maxRetries, err)
}

// buildCodeEmailContent 构建验证码邮件内容
func (es *EmailService) buildCodeEmailContent(code string) (string, error) {
	year := strconv.Itoa(time.Now().Year())
	tpl := `
	<html>
	<head>
		<meta charset="UTF-8">
		<title>GoFurry 验证码</title>
		<style>
			body { font-family: "Microsoft YaHei", "Helvetica Neue", sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
			.container { background-color: #f9f9f9; border-radius: 8px; padding: 30px; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
			.logo { color: #2c3e50; font-size: 24px; font-weight: bold; margin-bottom: 20px; display: flex; align-items: center; }
			.logo span { color: #3498db; margin-right: 8px; }
			.greeting { font-size: 18px; margin-bottom: 20px; }
			.code-box { background-color: #fff; border: 1px dashed #ddd; border-radius: 4px; padding: 20px; text-align: center; margin: 25px 0; }
			.code { font-size: 32px; font-weight: bold; letter-spacing: 8px; color: #2c3e50; margin: 0; }
			.note { color: #666; font-size: 14px; margin: 20px 0; }
			.warning { color: #e74c3c; font-size: 13px; padding: 10px; background-color: #fef0f0; border-radius: 4px; margin-top: 15px; }
			.footer { margin-top: 30px; color: #999; font-size: 12px; text-align: center; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="logo">
				<span>🐺</span> GoFurry
			</div>
			<div class="greeting">您好！</div>
			<p>感谢您使用 GoFurry 服务, 您正在进行邮箱验证操作。</p>
			<div class="code-box">
				<p class="code">[ ` + code + ` ]</p>
			</div>
			<p class="note">
				• 该验证码有效期为 <strong>5分钟</strong>, 请在有效期内完成验证<br>
				• 验证码仅用于本次操作, 请勿向他人泄露
			</p>
			<div class="warning">
				如果您未发起此操作, 请忽略本邮件, 您的账号安全不会受到影响。
			</div>
			<div class="footer">
				<p>GoFurry 邮箱服务 © ` + year + `</p>
			</div>
		</div>
	</body>
	</html>
	`
	return tpl, nil
}

// mimeEncode 对中文进行MIME编码
func (es *EmailService) mimeEncode(s string) string {
	hasNonASCII := false
	for _, r := range s {
		if r > 127 {
			hasNonASCII = true
			break
		}
	}
	if !hasNonASCII {
		return s
	}
	b := []byte(s)
	encoded := base64.StdEncoding.EncodeToString(b)
	return "=?UTF-8?B?" + encoded + "?="
}

// 原有快捷方法
func EmailSendCode(email string) (string, common.GFError) {
	return GetEmailService().SendCode(email)
}

func IsEmailValid(email string) bool {
	return GetEmailService().IsEmailValid(email)
}

func EmailSendPasswordReset(email, resetLink string) common.GFError {
	return GetEmailService().SendPasswordResetEmail(email, resetLink)
}

func EmailSendAccountNotice(email, operation, ip, timeStr string) common.GFError {
	return GetEmailService().SendAccountNoticeEmail(email, operation, ip, timeStr)
}

func EmailSendBatch(toList []string, subject, htmlBody string) common.GFError {
	return GetEmailService().SendBatchEmail(toList, subject, htmlBody)
}

func EmailSendWithAttachment(to, subject, htmlBody string, attachments map[string]string) common.GFError {
	return GetEmailService().SendEmailWithAttachment(to, subject, htmlBody, attachments)
}

func EmailSendCustom(to string, cc []string, bcc []string, subject, htmlBody string) common.GFError {
	return GetEmailService().SendEmail(to, cc, bcc, subject, htmlBody)
}
