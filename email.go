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
	"sync"

	"github.com/zhangdapeng520/zdpgo_email/gomail"
)

type Email struct {
	Fs     *embed.FS    // 嵌入的文件系统
	Config *Config      // 配置对象
	Result *EmailResult // 邮件发送结果
	Lock   sync.Mutex   // 同步锁
}

// New 新建邮件对象，支持发送邮件和接收邮件
func New() (email *Email, err error) {
	return NewWithConfig(&Config{})
}

// NewWithConfig 根据配置文件，创建邮件对象
func NewWithConfig(config *Config) (email *Email, err error) {
	email = &Email{}

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

	// 保存配置
	email.Config = config

	// 返回创建的邮件对象
	return
}

// IsHealth 检测是否健康，能否正常连接
func (e *Email) IsHealth() bool {
	// 获取发送器
	sender, err := e.GetSender()
	if err != nil {
		return false
	}
	defer sender.Close()

	return sender != nil
}

// GetSender 获取发送对象
func (e *Email) GetSender() (gomail.SendCloser, error) {
	// 创建拨号器
	d := &gomail.Dialer{
		Host:     e.Config.Host,
		Port:     e.Config.Port,
		Username: e.Config.Email,
		Password: e.Config.Password,
		SSL:      e.Config.IsSSL,
	}

	// 拨号
	sender, err := d.Dial()
	if err != nil {
		return nil, err
	}

	// 返回发送器
	return sender, nil
}
