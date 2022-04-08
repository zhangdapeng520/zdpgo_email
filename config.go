package zdpgo_email

import "errors"

// Config 配置类
type Config struct {
	Username string `yaml:"username" json:"username"`   // 发送者的名字
	Email    string `yaml:"email" json:"email"`         // 发送者的邮箱
	Password string `yaml:"password" json:"password"`   // 发送者的邮箱的校验密码（不一定是登陆密码）
	SmtpHost string `yaml:"smtp_host" json:"smtp_host"` // 邮箱服务器的主机地址（域名）
	SmtpPort int    `yaml:"smtp_port" json:"smtp_port"` // 端口
	Id       string `yaml:"id" json:"id"`               // 权限ID，可以不填
}

// 校验配置
func validateConfig(config Config) error {
	// 校验配置参数
	//SendName          string // 发送者的名字
	if config.Username == "" {
		return errors.New("用户名不能为空！")
	}
	//SendEmail         string // 发送者的邮箱
	if config.Email == "" {
		return errors.New("邮箱不能为空！")
	}
	//SendEmailPassword string // 发送者的邮箱的校验密码（不一定是登陆密码）
	if config.Password == "" {
		return errors.New("密码不能为空！")
	}
	//EmailSmtpHost     string // 邮箱服务器的主机地址（域名）
	if config.SmtpHost == "" {
		return errors.New("邮箱服务器地址不能为空！")
	}
	//EmailSmtpPort     uint16 // 端口
	if config.SmtpPort == 0 {
		return errors.New("邮箱服务器端口号不能为空！")
	}
	return nil
}
