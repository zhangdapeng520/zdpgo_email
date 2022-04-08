package zdpgo_email

import (
	"errors"
	"fmt"
	email2 "github.com/zhangdapeng520/zdpgo_email/core/email"
	"net/smtp"
)

// Email 核心的邮件对象
type Email struct {
	config *Config       // 配置
	email  *email2.Email // 发送邮件的对象
}

// New 创建邮件对象的实例
func New(config Config) *Email {
	e := Email{}

	// 校验配置
	if err := validateConfig(config); err != nil {
		panic(err)
	}

	// 初始化配置
	e.config = &config

	// 初始化email
	e.email = email2.NewEmail()
	return &e
}

// SendEmail 封装发送文本邮件的方法
// @param title 邮件标题
// @param content 邮件内容
// @param attach 附件
// @param isHtml 是否为HTML内容
// @param emails 普通收件人地址列表
// @param ccEmails 抄送人邮箱地址列表
// @param bccEmails 密送人邮箱地址列表
func (e *Email) SendEmail(title, content, attach string, isHtml bool, emails []string, ccEmails []string, bccEmails []string) error {
	// 校验邮箱
	if emails == nil {
		return errors.New("收件人邮箱不能为空")
	}

	// 设置 sender 发送方 的邮箱 ， 此处可以填写自己的邮箱
	e.email.From = fmt.Sprintf("%s <%s>", e.config.Username, e.config.Email)

	// 设置 receiver 接收方 的邮箱  此处也可以填写自己的邮箱， 就是自己发邮件给自己
	// 普通接收者
	e.email.To = emails

	// 抄送
	if ccEmails != nil {
		e.email.Cc = ccEmails
	}

	// 密送
	if bccEmails != nil {
		e.email.Bcc = bccEmails
	}

	// 设置主题
	e.email.Subject = title

	// 邮件内容，是HTML或者纯文本
	if isHtml {
		e.email.HTML = []byte(content)
	} else {
		e.email.Text = []byte(content)
	}

	// 附件
	if attach != "" {
		_, err := e.email.AttachFile(attach)
		if err != nil {
			return err
		}
	}

	//设置服务器相关的配置
	addr := fmt.Sprintf("%s:%d", e.config.SmtpHost, e.config.SmtpPort)
	err := e.email.Send(addr, smtp.PlainAuth(
		e.config.Id,
		e.config.Email,
		e.config.Password,
		e.config.SmtpHost))
	if err != nil {
		return err
	}
	return nil
}

// SendText 发送文本邮件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
func (e *Email) SendText(title, content string, emails ...string) {
	e.SendEmail(title, content, "", false, emails, nil, nil)
}

// SendHtml 发送HTML邮件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
func (e *Email) SendHtml(title, content string, emails ...string) {
	e.SendEmail(title, content, "", true, emails, nil, nil)
}

// SendHtmlAndAttach 发送HTML邮件且能够携带附件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
// @param attach 附件
func (e *Email) SendHtmlAndAttach(title, content, attach string, emails ...string) {
	e.SendEmail(title, content, attach, true, emails, nil, nil)
}

// SendTextAndAttach 发送文本文件，且能够携带附件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
// @param attach 附件
func (e *Email) SendTextAndAttach(title, content, attach string, emails ...string) {
	e.SendEmail(title, content, attach, false, emails, nil, nil)
}
