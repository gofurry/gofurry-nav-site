package service

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	netmail "net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/roof/env"
	mail "github.com/wneessen/go-mail"
)

func captureEmail() (*EmailService, <-chan *mail.Msg) {
	messages := make(chan *mail.Msg, 4)
	return &EmailService{sender: "sender@example.com", senderName: "GoFurry 邮件服务", dialAndSend: func(m *mail.Msg) error { messages <- m; return nil }}, messages
}

func renderedEmail(t *testing.T, m *mail.Msg) *netmail.Message {
	t.Helper()
	var wire bytes.Buffer
	if _, err := m.WriteTo(&wire); err != nil {
		t.Fatal(err)
	}
	message, err := netmail.ReadMessage(&wire)
	if err != nil {
		t.Fatal(err)
	}
	from, err := netmail.ParseAddress(message.Header.Get("From"))
	if err != nil || from.Name != "GoFurry 邮件服务" || from.Address != "sender@example.com" {
		t.Fatalf("sender changed: %v %v", from, err)
	}
	return message
}

func decodedBody(t *testing.T, body io.Reader, encoding string) string {
	t.Helper()
	switch strings.ToLower(encoding) {
	case "quoted-printable":
		body = quotedprintable.NewReader(body)
	case "base64":
		body = base64.NewDecoder(base64.StdEncoding, body)
	}
	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

func TestEmailTemplatesAndMIMECompatibility(t *testing.T) {
	// Golden files captured from the pre-migration template expressions. Only the
	// existing substitutions and MIME CRLF normalization are applied. The final
	// raw-string indentation tab is restored after loading the text fixture.
	for _, tc := range []struct {
		name, subject string
		send          func(*EmailService) common.GFError
		vars          []string
	}{
		{"SendActivationEmail", "参与抽奖", func(es *EmailService) common.GFError {
			return es.SendActivationEmail("to@example.com", "参与抽奖", "https://example.com/activate?code=abc", "确认参与", "抽奖参与确认", "30分钟")
		}, []string{"{{activateLink}}", "https://example.com/activate?code=abc", "{{activateText}}", "确认参与", "{{text}}", "抽奖参与确认", "{{duration}}", "30分钟"}},
		{"SendLotteryEmail", "中奖通知", func(es *EmailService) common.GFError {
			return es.SendLotteryEmail("to@example.com", "中奖通知", "PRIZE-123", "您获得游戏礼品")
		}, []string{"{{code}}", "PRIZE-123", "{{text}}", "您获得游戏礼品"}},
		{"SendPasswordResetEmail", "GoFurry 密码重置", func(es *EmailService) common.GFError {
			return es.SendPasswordResetEmail("to@example.com", "https://example.com/reset")
		}, []string{"{{resetLink}}", "https://example.com/reset"}},
		{"SendAccountNoticeEmail", "GoFurry 账号安全提醒", func(es *EmailService) common.GFError {
			return es.SendAccountNoticeEmail("to@example.com", "登录", "192.0.2.1", "2026-09-13")
		}, []string{"{{operation}}", "登录", "{{ip}}", "192.0.2.1", "{{timeStr}}", "2026-09-13"}},
		{"buildCodeEmailContent", "GoFurry 邮箱验证码", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			es, messages := captureEmail()
			vars := append([]string{"{{year}}", strconv.Itoa(time.Now().Year())}, tc.vars...)
			if tc.send == nil {
				code, err := es.SendCode("to@example.com")
				if err != nil {
					t.Fatal(err)
				}
				if len(code) != common.EMAIL_CODE_LENGTH {
					t.Fatal("code length changed")
				}
				vars = append(vars, "{{code}}", code)
			} else if err := tc.send(es); err != nil {
				t.Fatal(err)
			}
			message := renderedEmail(t, <-messages)
			subject, err := new(mime.WordDecoder).DecodeHeader(message.Header.Get("Subject"))
			if err != nil || subject != tc.subject {
				t.Fatalf("subject=%q error=%v", subject, err)
			}
			kind, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
			if err != nil || kind != "text/html" || !strings.EqualFold(params["charset"], "utf-8") {
				t.Fatalf("invalid HTML content type: %s", message.Header.Get("Content-Type"))
			}
			want, err := os.ReadFile(filepath.Join("testdata", "email", tc.name+".html"))
			if err != nil {
				t.Fatal(err)
			}
			expected := strings.NewReplacer(vars...).Replace(strings.ReplaceAll(string(want), "\r\n", "\n")) + "\t"
			got := decodedBody(t, message.Body, message.Header.Get("Content-Transfer-Encoding"))
			if got != expected {
				t.Fatalf("template HTML/CSS/copy changed for %s", tc.name)
			}
		})
	}
}

func TestEmailCCBCCAndAttachment(t *testing.T) {
	es, messages := captureEmail()
	if err := es.SendEmail("to@example.com", []string{"抄送 <cc@example.com>"}, []string{"bcc@example.com"}, "邮件标题", "<p>正文</p>"); err != nil {
		t.Fatal(err)
	}
	m := <-messages
	rcpts, err := m.GetRecipients()
	if err != nil || !reflect.DeepEqual(rcpts, []string{"<to@example.com>", "<cc@example.com>", "<bcc@example.com>"}) {
		t.Fatalf("recipients=%v err=%v", rcpts, err)
	}
	message := renderedEmail(t, m)
	if message.Header.Get("Cc") == "" || message.Header.Get("Bcc") != "" {
		t.Fatal("CC missing or BCC leaked")
	}
	file := filepath.Join(t.TempDir(), "original.txt")
	if err := os.WriteFile(file, []byte("礼品内容"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := es.SendEmailWithAttachment("to@example.com", "带附件", "<p>正文</p>", map[string]string{file: "礼品.txt"}); err != nil {
		t.Fatal(err)
	}
	message = renderedEmail(t, <-messages)
	kind, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
	if err != nil || kind != "multipart/mixed" {
		t.Fatal("attachment MIME structure changed")
	}
	reader := multipart.NewReader(message.Body, params["boundary"])
	first, err := reader.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if decodedBody(t, first, first.Header.Get("Content-Transfer-Encoding")) != "<p>正文</p>" {
		t.Fatal("attachment HTML changed")
	}
	attachment, err := reader.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	filename, nameErr := new(mime.WordDecoder).DecodeHeader(attachment.FileName())
	content := decodedBody(t, attachment, attachment.Header.Get("Content-Transfer-Encoding"))
	if nameErr != nil || filename != "礼品.txt" || content != "礼品内容" {
		t.Fatalf("attachment filename/content changed: filename=%q content=%q error=%v", filename, content, nameErr)
	}
	if err := es.SendEmailWithAttachment("to@example.com", "附件", "正文", map[string]string{file + "missing": "missing.txt"}); err == nil {
		t.Fatal("missing attachment silently sent")
	}
}

func TestEmailRetryPreservesMessage(t *testing.T) {
	es, _ := captureEmail()
	calls := 0
	var original *mail.Msg
	failure := errors.New("SMTP unavailable")
	es.dialAndSend = func(m *mail.Msg) error {
		calls++
		if original == nil {
			original = m
		}
		if original != m {
			t.Error("retry dropped message headers")
		}
		return failure
	}
	start := time.Now()
	err := es.SendEmail("to@example.com", []string{"cc@example.com"}, []string{"bcc@example.com"}, "重试", "正文")
	if err == nil || err.GetMsg() != "邮件发送失败, 请稍后重试" || calls != 3 || time.Since(start) < 3*time.Second {
		t.Fatalf("retry/error semantics changed: calls=%d error=%v", calls, err)
	}
	calls = 0
	es.dialAndSend = func(*mail.Msg) error {
		calls++
		if calls == 1 {
			return failure
		}
		return nil
	}
	if err := es.sendEmailWithRetry("to@example.com", "重试成功", "正文"); err != nil || calls != 2 {
		t.Fatalf("retry recovery failed: %v", err)
	}
}

func TestEmailServiceTimeouts(t *testing.T) {
	for _, attachment := range []bool{false, true} {
		t.Run(strconv.FormatBool(attachment), func(t *testing.T) {
			t.Parallel()
			es, _ := captureEmail()
			release := make(chan struct{})
			defer close(release)
			es.dialAndSend = func(*mail.Msg) error { <-release; return nil }
			start := time.Now()
			var err common.GFError
			want := 12 * time.Second
			if attachment {
				file := filepath.Join(t.TempDir(), "file.txt")
				os.WriteFile(file, []byte("content"), 0600)
				err = es.SendEmailWithAttachment("to@example.com", "附件", "正文", map[string]string{file: "file.txt"})
				want = 15 * time.Second
			} else {
				_, err = es.SendCode("to@example.com")
			}
			if err == nil || !strings.Contains(err.GetMsg(), "超时") || time.Since(start) < want || time.Since(start) > want+3*time.Second {
				t.Fatalf("timeout changed: %v elapsed=%s", err, time.Since(start))
			}
		})
	}
}

func TestEmailTransportLocalSMTP(t *testing.T) {
	// An isolated loopback SMTP double; never contacts the configured real relay.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	received := make(chan string, 1)
	smtpErrors := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			smtpErrors <- err
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		r := textproto.NewReader(bufio.NewReader(conn))
		w := textproto.NewWriter(bufio.NewWriter(conn))
		w.PrintfLine("220 localhost test SMTP")
		var recipients []string
		for {
			line, err := r.ReadLine()
			if err != nil {
				smtpErrors <- err
				return
			}
			switch {
			case strings.HasPrefix(line, "EHLO"):
				w.PrintfLine("250-localhost")
				w.PrintfLine("250 AUTH PLAIN")
			case strings.HasPrefix(line, "AUTH PLAIN "):
				if strings.TrimPrefix(line, "AUTH PLAIN ") != base64.StdEncoding.EncodeToString([]byte("\x00sender@example.com\x00test-password")) {
					smtpErrors <- errors.New("SMTP credentials changed")
					return
				}
				w.PrintfLine("235 authenticated")
			case strings.HasPrefix(line, "RCPT TO:"):
				recipients = append(recipients, line)
				w.PrintfLine("250 ok")
			case line == "DATA":
				w.PrintfLine("354 send mail")
				body, err := r.ReadDotBytes()
				if err != nil {
					smtpErrors <- err
					return
				}
				received <- strings.Join(recipients, "\n") + "\n" + string(body)
				w.PrintfLine("250 queued")
			case line == "QUIT":
				w.PrintfLine("221 bye")
				smtpErrors <- nil
				return
			default:
				w.PrintfLine("250 ok")
			}
		}
	}()
	cfg := env.GetServerConfig()
	previous := cfg.Email
	defer func() { cfg.Email = previous }()
	cfg.Email.EmailHost = "127.0.0.1"
	cfg.Email.EmailPort = listener.Addr().(*net.TCPAddr).Port
	cfg.Email.EmailUser = "sender@example.com"
	cfg.Email.EmailPassword = "test-password"
	es, err := newEmailService()
	if err != nil {
		t.Fatal(err)
	}
	if err := es.SendEmail("to@example.com", []string{"cc@example.com"}, []string{"bcc@example.com"}, "SMTP 测试", "<p>测试</p>"); err != nil {
		t.Fatal(err)
	}
	if err := <-smtpErrors; err != nil {
		t.Fatal(err)
	}
	wire := <-received
	for _, address := range []string{"to@example.com", "cc@example.com", "bcc@example.com"} {
		if !strings.Contains(wire, "RCPT TO:<"+address+">") {
			t.Errorf("missing envelope recipient %s", address)
		}
	}
	if strings.Contains(wire, "Bcc:") {
		t.Fatal("BCC leaked into MIME headers")
	}
}
