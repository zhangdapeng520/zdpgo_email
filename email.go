package zdpgo_email

/*
@Time : 2022/4/28 23:32
@Author : 张大鹏
@File : email
@Software: Goland2021.3.1
@Description: 核心邮件对象，包含收邮件和发送邮件的功能
*/
import (
	"crypto/tls"
	"embed"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email/gomail"
	"github.com/zhangdapeng520/zdpgo_log"
	"github.com/zhangdapeng520/zdpgo_random"
	"github.com/zhangdapeng520/zdpgo_yaml"
	"net"
	"net/textproto"
	"time"
)

type Email struct {
	SendObj    *EmailSmtp
	ReceiveObj *EmailImap
	Fs         *embed.FS // 嵌入的文件系统
	Random     *zdpgo_random.Random
	Yaml       *zdpgo_yaml.Yaml
	Log        *zdpgo_log.Log // 日志对象
	Config     *Config        // 配置对象
	Result     *EmailResult   // 邮件发送结果
}

// New 新建邮件对象，支持发送邮件和接收邮件
func New() (email *Email, err error) {
	return NewWithConfig(Config{})
}

// NewWithConfig 根据配置文件，创建邮件对象
func NewWithConfig(config Config) (email *Email, err error) {
	email = &Email{}
	email.Random = zdpgo_random.New()
	email.Yaml = zdpgo_yaml.New()

	// 日志对象
	if config.LogFilePath == "" {
		config.LogFilePath = "logs/zdpgo/zdpgo_email.log"
	}
	logConfig := zdpgo_log.Config{
		Debug:       config.Debug,
		OpenJsonLog: true,
		LogFilePath: config.LogFilePath,
	}
	if config.Debug {
		logConfig.IsShowConsole = true
	}
	email.Log = zdpgo_log.NewWithConfig(logConfig)
	email.Log.Debug("创建email日志对象成功", "config", config)
	gomail.Log = email.Log // 初始化gomail中的日志对象

	// 标识符
	if config.HeaderTagName == "" {
		config.HeaderTagName = "X-ZdpgoEmail-Auther"
	}
	if config.HeaderTagValue == "" {
		config.HeaderTagValue = "zhangdapeng520"
	}
	if config.CommonTitle == "" {
		config.CommonTitle = "【ZDP-Go-Email】邮件发送测试（仅限学习研究，切勿滥用）"
	}

	// 邮件发送对象
	if config.Smtp.Host != "" && config.Smtp.Port != 0 && config.Smtp.Password != "" {
		email.SendObj = &EmailSmtp{Headers: textproto.MIMEHeader{}}
		email.SendObj.Random = zdpgo_random.New()
		email.SendObj.Log = email.Log
	}

	// 邮件接收对象
	if config.Imap.Host != "" && config.Imap.Port != 0 && config.Imap.Password != "" {
		dialer := new(net.Dialer)
		var (
			conn net.Conn
		)
		conn, err = dialer.Dial("tcp", fmt.Sprintf("%s:%d", config.Imap.Host, config.Imap.Port))
		if err != nil {
			email.Log.Error("创建邮件接收对象失败")
			return
		}
		// 连接到邮件服务器
		tlsConfig := &tls.Config{
			ServerName: config.Imap.Host,
		}
		tlsConn := tls.Client(conn, tlsConfig)
		if config.Timeout == 0 {
			config.Timeout = 30
		}
		err = tlsConn.SetDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
		if err != nil {
			email.Log.Error("连接到收件邮件服务器失败", "error", err)
			return
		}

		// 创建邮件接收对象
		email.ReceiveObj, err = NewEmailImap(tlsConn)
		if err != nil {
			email.Log.Error("创建邮件接收对象失败", "error", err)
			return
		}

		// 设置tls
		email.ReceiveObj.isTLS = true
		email.ReceiveObj.serverName = config.Imap.Host

		// 登录
		if config.Imap.Email == "" {
			config.Imap.Email = config.Imap.Username
		}
		err = email.ReceiveObj.Login(config.Imap.Email, config.Imap.Password)
		if err != nil {
			email.Log.Error("登录邮件收件服务器失败", "error", err)
		}
	}

	// 保存配置
	email.Config = &config
	if email.SendObj != nil {
		email.SendObj.Config = &config
	}
	if email.ReceiveObj != nil {
		email.ReceiveObj.Config = &config
	}

	return
}

// IsHealth 检测是否健康，能否正常连接
func (e *Email) IsHealth() bool {
	// 没有发送对象
	if e.SendObj == nil {
		e.Log.Debug("邮件发送对象为空")
		return false
	}

	// 获取发送器
	sender, err := e.GetSender()
	if err != nil {
		e.Log.Error("获取邮件发送器失败", "error", err, "config", e.SendObj.Config)
		return false
	}
	defer sender.Close()

	if sender == nil {
		e.Log.Error("邮件发送器为空", "sender", sender)
		return false
	}

	return true
}

// GetSender 获取发送对象
func (e *Email) GetSender() (gomail.SendCloser, error) {
	// 创建拨号器
	d := &gomail.Dialer{
		Host:     e.Config.Smtp.Host,
		Port:     e.Config.Smtp.Port,
		Username: e.Config.Smtp.Email,
		Password: e.Config.Smtp.Password,
		SSL:      e.Config.Smtp.IsSSL,
	}

	// 拨号
	sender, err := d.Dial()
	if err != nil {
		e.Log.Error("获取发送对象失败", "error", err)
		return nil, err
	}

	// 返回发送器
	return sender, nil
}
