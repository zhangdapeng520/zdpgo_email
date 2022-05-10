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

// New 新建邮件对象，支持发送邮件和接收邮件
func New() (email *Email, err error) {
	return NewWithConfig(Config{})
}

// NewWithConfig 根据配置文件，创建邮件对象
func NewWithConfig(config Config) (email *Email, err error) {
	email = &Email{}
	email.Random = zdpgo_random.New()
	email.Yaml = zdpgo_yaml.New()

	// 邮件发送对象
	if config.SmtpConfigs != nil && len(config.SmtpConfigs) > 0 {
		var configSmtp ConfigSmtp
		for _, cfgFile := range config.SmtpConfigs {
			err = email.Yaml.ReadConfig(cfgFile, &configSmtp)
			if err != nil {
				return
			}
		}
		email.Send, err = NewEmailSmtpWithConfig(configSmtp)
		if err != nil {
			return
		}
	}

	// 邮件接收对象
	if config.ImapConfigs != nil && len(config.ImapConfigs) > 0 {
		var configImap ConfigImap
		for _, cfgFile := range config.ImapConfigs {
			_ = email.Yaml.ReadConfig(cfgFile, &configImap)
		}
		email.Receive, err = NewEmailImapWithConfig(configImap)
		if err != nil {
			return
		}
	}

	return
}
