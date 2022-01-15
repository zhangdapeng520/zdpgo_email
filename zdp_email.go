package zdpgo_email

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_log"
	"net/smtp"
)

// 核心的邮件对象
type ZdpEmail struct {
	log    *zdpgo_log.Log  // 记录日志
	config *ZdpEmailConfig // 配置
	email  *Email          // 发送邮件的对象
}

// 配置类
type ZdpEmailConfig struct {
	Debug             bool   // 是否为debug模式
	LogFilePath       string // 日志文件的存放路径
	SendName          string // 发送者的名字
	SendEmail         string // 发送者的邮箱
	SendEmailPassword string // 发送者的邮箱的校验密码（不一定是登陆密码）
	EmailSmtpHost     string // 邮箱服务器的主机地址（域名）
	EmailSmtpPort     uint16 // 端口
	Identity          string // 权限ID，可以不填
}

// 创建邮件对象的实例
func New(config ZdpEmailConfig) *ZdpEmail {
	e := ZdpEmail{}

	// 日志初始化
	if config.LogFilePath == "" {
		config.LogFilePath = "zdpgo_email.log"
	}
	logConfig := zdpgo_log.LogConfig{
		Debug: config.Debug,
		Path:  config.LogFilePath,
	}
	zlog := zdpgo_log.New(logConfig)
	e.log = zlog

	// 校验配置参数
	//SendName          string // 发送者的名字
	if config.SendName == "" {
		e.log.Panic("发送者的名字不能为空！")
	}
	//SendEmail         string // 发送者的邮箱
	if config.SendEmail == "" {
		e.log.Panic("发送者的邮箱不能为空！")
	}
	//SendEmailPassword string // 发送者的邮箱的校验密码（不一定是登陆密码）
	if config.SendEmail == "" {
		e.log.Panic("发送者的邮箱不能为空！")
	}
	//EmailSmtpHost     string // 邮箱服务器的主机地址（域名）
	if config.EmailSmtpHost == "" {
		e.log.Panic("发送者的邮箱服务器地址不能为空！")
	}
	//EmailSmtpPort     uint16 // 端口
	if config.EmailSmtpPort == 0 {
		e.log.Panic("发送者的邮箱服务器端口号不能为0！")
	}

	// 初始化配置
	e.config = &config

	// 初始化email
	e.email = NewEmail()
	return &e
}

// 修改debug模式
func (e *ZdpEmail) SetDebug(debug bool) {
	// 修改配置的debug
	e.config.Debug = debug

	// 日志的debug
	e.log.SetDebug(debug)
}

// 判断是否为debug模式
func (e *ZdpEmail) IsDebug() bool {
	return e.config.Debug
}

// SendEmail 封装发送文本邮件的方法
// @param title 邮件标题
// @param content 邮件内容
// @param attach 附件
// @param isHtml 是否为HTML内容
// @param emails 普通收件人地址列表
// @param ccEmails 抄送人邮箱地址列表
// @param bccEmails 密送人邮箱地址列表
func (e *ZdpEmail) SendEmail(title, content, attach string, isHtml bool, emails []string, ccEmails []string, bccEmails []string) {
	// 校验邮箱
	if emails == nil {
		e.log.Panic("邮箱不能为空")
		return
	}

	// 设置 sender 发送方 的邮箱 ， 此处可以填写自己的邮箱
	e.email.From = fmt.Sprintf("%s <%s>", e.config.SendName, e.config.SendEmail)

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
			e.log.Error("添加附件失败：", err)
		}
	}

	//设置服务器相关的配置
	addr := fmt.Sprintf("%s:%d", e.config.EmailSmtpHost, e.config.EmailSmtpPort)
	err := e.email.Send(addr, smtp.PlainAuth(
		e.config.Identity,
		e.config.SendEmail,
		e.config.SendEmailPassword,
		e.config.EmailSmtpHost))
	if err != nil {
		e.log.Error("发送邮件失败：", err)
	}
	e.log.Info("发送邮件成功！")
}

// SendText 发送文本邮件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
func (e *ZdpEmail) SendText(title, content string, emails ...string) {
	e.SendEmail(title, content, "", false, emails, nil, nil)
}

// SendHtml 发送HTML邮件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
func (e *ZdpEmail) SendHtml(title, content string, emails ...string) {
	e.SendEmail(title, content, "", true, emails, nil, nil)
}

// SendHtmlAndAttach 发送HTML邮件且能够携带附件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
// @param attach 附件
func (e *ZdpEmail) SendHtmlAndAttach(title, content, attach string, emails ...string) {
	e.SendEmail(title, content, attach, true, emails, nil, nil)
}

// SendTextAndAttach 发送文本文件，且能够携带附件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
// @param attach 附件
func (e *ZdpEmail) SendTextAndAttach(title, content, attach string, emails ...string) {
	e.SendEmail(title, content, attach, false, emails, nil, nil)
}
