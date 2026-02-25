package email

import (
	"log"
	"time"

	"github.com/go-mail/mail/v2"
)

// EmailConfig 邮件配置（建议从配置文件读取，这里先硬编码适配指定邮箱）
var EmailConfig = struct {
	SMTPHost string
	SMTPPort int
	Username string // 发件人邮箱 sunzhaoc_2011@163.com
	Password string // 163邮箱SMTP授权码（需自行在邮箱设置中开启）
	From     string // 发件人名称
	Timeout  time.Duration
}{
	SMTPHost: "smtp.163.com",
	SMTPPort: 465,
	Username: "sunzhaoc_2011@163.com",
	Password: "XKRhX397xxR6UZ3x",
	From:     "Plant系统验证码",
	Timeout:  60 * time.Second,
}

// SendVerificationCode 发送验证码到指定邮箱
func SendVerificationCode(toEmail, code string) error {
	m := mail.NewMessage()
	m.SetHeader("From", m.FormatAddress(EmailConfig.Username, EmailConfig.From))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "ANTPLANT-STORE - 密码重置验证码")
	m.SetBody("text/plain", "您好，您的密码重置验证码为："+code+"，有效期5分钟，请及时使用。")

	d := mail.NewDialer(
		EmailConfig.SMTPHost,
		EmailConfig.SMTPPort,
		EmailConfig.Username,
		EmailConfig.Password,
	)
	d.Timeout = EmailConfig.Timeout
	d.SSL = true

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		log.Printf("发送验证码邮件失败：%v", err)
		return err
	}
	return nil
}
