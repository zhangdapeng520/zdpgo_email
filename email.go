package zdpgo_email

/*
@Time : 2022/4/28 23:32
@Author : 张大鹏
@File : email
@Software: Goland2021.3.1
@Description: 核心邮件对象，包含收邮件和发送邮件的功能
*/
import (
	"embed"
	"github.com/zhangdapeng520/zdpgo_email/gomail"
	"github.com/zhangdapeng520/zdpgo_log"
	"github.com/zhangdapeng520/zdpgo_random"
	"github.com/zhangdapeng520/zdpgo_yaml"
)

type Email struct {
	Send    *EmailSmtp
	Receive *EmailImap
	Fs      *embed.FS // 嵌入的文件系统
	Random  *zdpgo_random.Random
	Yaml    *zdpgo_yaml.Yaml
	Log     *zdpgo_log.Log // 日志对象
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
	email.Log = zdpgo_log.NewWithConfig(zdpgo_log.Config{
		Debug:       config.Debug,
		OpenJsonLog: true,
		LogFilePath: config.LogFilePath,
	})
	if config.Debug {
		email.Log.Debug("创建email日志对象成功", "config", config)
	}

	// 邮件发送对象
	if config.SmtpConfigs != nil && len(config.SmtpConfigs) > 0 {
		var configSmtp ConfigSmtp
		for _, cfgFile := range config.SmtpConfigs {
			err = email.Yaml.ReadConfig(cfgFile, &configSmtp)
			if err != nil {
				email.Log.Error("读取邮件发送配置失败", "error", err, "cfgFile", cfgFile)
				return
			}
		}
		email.Send, err = NewEmailSmtpWithConfig(configSmtp)
		if err != nil {
			email.Log.Error("创建邮件发送对象失败", "error", err, "configSmtp", configSmtp)
			return
		}
	}

	// 嵌入文件系统
	if config.IsUseFs {
		email.Fs = config.Fs
		email.Send.Fs = config.Fs
	}

	// 邮件接收对象
	if config.ImapConfigs != nil && len(config.ImapConfigs) > 0 {
		var configImap ConfigImap
		for _, cfgFile := range config.ImapConfigs {
			_ = email.Yaml.ReadConfig(cfgFile, &configImap)
		}
		email.Receive, err = NewEmailImapWithConfig(configImap)
		if err != nil {
			email.Log.Error("创建邮件接收对象失败", "error", err, "configImap", configImap)
			return
		}
	}

	return
}

// NewWithSmtpAndImapConfig 使用smtp配置和imap配置创建邮件对象
func NewWithSmtpAndImapConfig(smtp ConfigSmtp, imap ConfigImap) (email *Email, err error) {
	email = &Email{}
	email.Random = zdpgo_random.New()
	email.Yaml = zdpgo_yaml.New()

	// 邮件发送对象
	email.Send, err = NewEmailSmtpWithConfig(smtp)
	if err != nil {
		return
	}

	// 邮件接收对象
	email.Receive, err = NewEmailImapWithConfig(imap)
	if err != nil {
		return
	}

	return
}

// IsHealth 检测是否健康，能否正常连接
func (e *Email) IsHealth() bool {
	// 没有发送对象
	if e.Send == nil {
		e.Log.Debug("邮件发送对象为空")
		return false
	}

	// 获取发送器
	sender, err := e.GetSender(*e.Send.Config)
	if err != nil {
		e.Log.Error("获取邮件发送器失败", "error", err, "config", e.Send.Config)
		return false
	}

	if sender == nil {
		e.Log.Error("邮件发送器为空", "sender", sender)
		return false
	}

	return true
}

// GetSender 获取发送对象
func (e *Email) GetSender(config ConfigSmtp) (sender gomail.SendCloser, err error) {
	// 创建拨号器
	d := &gomail.Dialer{
		Host:     config.SmtpHost,
		Port:     config.SmtpPort,
		Username: config.Email,
		Password: config.Password,
		SSL:      config.IsSSL,
	}

	// 拨号
	sender, err = d.Dial()
	return
}
