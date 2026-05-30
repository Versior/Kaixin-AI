package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/basketikun/infinite-canvas/repository"
)

// emailCode 内存验证码存储
type emailCode struct {
	code      string
	email     string
	createdAt time.Time
}

var (
	codeStore   = make(map[string]*emailCode) // key: email
	codeStoreMu sync.RWMutex
)

const codeTTL = 5 * time.Minute

// SendVerificationCode 发送邮箱验证码
func SendVerificationCode(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return safeMessageError{message: "邮箱不能为空"}
	}

	settings, err := repository.GetSettings()
	if err != nil {
		return err
	}
	s := normalizeSettings(settings)
	smtpCfg := s.Private.Auth.SMTP

	if smtpCfg.Enabled == nil || !*smtpCfg.Enabled {
		return safeMessageError{message: "邮箱验证未开启"}
	}
	if smtpCfg.Host == "" || smtpCfg.Port == 0 {
		return safeMessageError{message: "SMTP 服务器未配置"}
	}

	// 生成 6 位验证码
	code := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)

	// 保存验证码
	codeStoreMu.Lock()
	codeStore[email] = &emailCode{
		code:      code,
		email:     email,
		createdAt: time.Now(),
	}
	codeStoreMu.Unlock()

	// 构建邮件
	from := smtpCfg.From
	if from == "" {
		from = smtpCfg.Username
	}
	subject := "【无限画布】注册验证码"
	body := fmt.Sprintf("您的验证码是：%s\n\n有效期 5 分钟，请勿泄露给他人。", code)

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, email, subject, body))

	// 发送
	addr := fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port)

	var auth smtp.Auth
	if smtpCfg.Username != "" {
		auth = smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)
	}

	if smtpCfg.UseTLS != nil && *smtpCfg.UseTLS {
		// TLS（端口 465）
		tlsConfig := &tls.Config{
			ServerName: smtpCfg.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return safeMessageError{message: "SMTP 连接失败: " + err.Error()}
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, smtpCfg.Host)
		if err != nil {
			return safeMessageError{message: "SMTP 客户端创建失败: " + err.Error()}
		}
		defer c.Close()

		if auth != nil {
			if err = c.Auth(auth); err != nil {
				return safeMessageError{message: "SMTP 认证失败: " + err.Error()}
			}
		}
		if err = c.Mail(from); err != nil {
			return err
		}
		if err = c.Rcpt(email); err != nil {
			return err
		}
		w, err := c.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(msg)
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
		return c.Quit()
	}

	// 非 TLS（端口 587 / 25）
	if err := smtp.SendMail(addr, auth, from, []string{email}, msg); err != nil {
		return safeMessageError{message: "邮件发送失败: " + err.Error()}
	}
	return nil
}

// VerifyCode 验证邮箱验证码
func VerifyCode(email, code string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	code = strings.TrimSpace(code)

	codeStoreMu.Lock()
	defer codeStoreMu.Unlock()

	stored, ok := codeStore[email]
	if !ok {
		return safeMessageError{message: "请先获取验证码"}
	}

	if time.Since(stored.createdAt) > codeTTL {
		delete(codeStore, email)
		return safeMessageError{message: "验证码已过期，请重新获取"}
	}

	if stored.code != code {
		return safeMessageError{message: "验证码错误"}
	}

	// 验证成功，删除
	delete(codeStore, email)
	return nil
}

// CleanupExpiredCodes 清理过期验证码（可定时调用）
func CleanupExpiredCodes() {
	codeStoreMu.Lock()
	defer codeStoreMu.Unlock()
	now := time.Now()
	for k, v := range codeStore {
		if now.Sub(v.createdAt) > codeTTL {
			delete(codeStore, k)
		}
	}
}
