package zdpgo_email

/*
@Time : 2022/4/28 23:32
@Author : 张大鹏
@File : email
@Software: Goland2021.3.1
@Description: 核心邮件对象，包含收邮件和发送邮件的功能
*/
import (
	"github.com/zhangdapeng520/zdpgo_random"
	"github.com/zhangdapeng520/zdpgo_yaml"
)

type Email struct {
	Send    *EmailSmtp
	Receive *EmailImap
	Random  *zdpgo_random.Random
	Yaml    *zdpgo_yaml.Yaml
}

func New() *Email {
	return NewWithConfig(Config{})
}

func NewWithConfig(config Config) *Email {
	e := Email{}
	e.Random = zdpgo_random.New()
	e.Yaml = zdpgo_yaml.New()

	// 邮件发送对象
	if config.SmtpConfigs != nil && len(config.SmtpConfigs) > 0 {
		var configSmtp ConfigSmtp
		for _, cfgFile := range config.SmtpConfigs {
			_ = e.Yaml.ReadConfig(cfgFile, &configSmtp)
		}
		e.Send = NewEmailSmtpWithConfig(configSmtp)
	}

	// 邮件接收对象
	if config.ImapConfigs != nil && len(config.ImapConfigs) > 0 {
		var configImap ConfigImap
		for _, cfgFile := range config.ImapConfigs {
			_ = e.Yaml.ReadConfig(cfgFile, &configImap)
		}
		e.Receive, _ = NewEmailImapWithConfig(configImap)
	}

	return &e
}
