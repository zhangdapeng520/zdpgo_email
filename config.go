package zdpgo_email

import (
	"embed"
	"errors"
)

// Config 配置类
type Config struct {
	Debug            bool        `yaml:"debug" json:"debug" env:"debug"`               // 是否为debug模式                    // 是否为debug模式
	LogFilePath      string      `yaml:"log_file_path" json:"log_file_path"`           // 日志文件路径
	HeaderTagName    string      `yaml:"header_tag_name" json:"header_tag_name"`       // 请求头标记名
	HeaderTagValue   string      `yaml:"header_tag_value" json:"header_tag_value"`     // 请求头标记值
	CommonTitle      string      `yaml:"common_title" json:"common_title"`             // 通用邮件标题
	Id               string      `yaml:"id" json:"id"`                                 // 用于SMTP登录
	SendSleepSeconds int         `yaml:"send_sleep_seconds" json:"send_sleep_seconds"` // 发送邮件的间隔休眠时间
	Timeout          int         `yaml:"timeout" json:"timeout"`                       // 超时时间
	Smtp             ConfigEmail `yaml:"smtp" json:"smtp"`                             // SMTP的配置
	Imap             ConfigEmail `yaml:"imap" json:"imap"`                             // IMAP的配置
}

// ConfigEmail email服务器的相关配置
type ConfigEmail struct {
	Email    string `yaml:"email" json:"email"`         // 发送者的邮箱
	Username string `yaml:"username" json:"username"`   // 发送者的名字
	Password string `yaml:"password" json:"password"`   // 发送者的邮箱的校验密码（不一定是登陆密码）
	Host     string `yaml:"smtp_host" json:"smtp_host"` // 邮箱服务器的主机地址（域名）
	Port     int    `yaml:"smtp_port" json:"smtp_port"` // 端口
	IsSSL    bool   `yaml:"is_ssl" json:"is_ssl"`       // 是否为SSL模式
}

type ConfigSmtp struct {
	Debug          bool      `yaml:"debug" json:"debug" env:"debug"`                         // 是否为debug模式
	LogFilePath    string    `yaml:"log_file_path" json:"log_file_path" env:"log_file_path"` // 日志文件路径
	Username       string    `yaml:"username" json:"username"`                               // 发送者的名字
	Email          string    `yaml:"email" json:"email"`                                     // 发送者的邮箱
	Password       string    `yaml:"password" json:"password"`                               // 发送者的邮箱的校验密码（不一定是登陆密码）
	SmtpHost       string    `yaml:"smtp_host" json:"smtp_host"`                             // 邮箱服务器的主机地址（域名）
	SmtpPort       int       `yaml:"smtp_port" json:"smtp_port"`                             // 端口
	Id             string    `yaml:"id" json:"id"`                                           // 权限ID，可以不填
	IsSSL          bool      `yaml:"is_ssl" json:"is_ssl"`                                   // 是否为SSL模式
	HeaderTagName  string    `yaml:"header_tag_name" json:"header_tag_name"`                 // 请求头标记名
	HeaderTagValue string    `yaml:"header_tag_value" json:"header_tag_value"`               // 请求头标记值
	Fs             *embed.FS // 嵌入文件系统
}

// ConfigImap EmailImap的相关配置
type ConfigImap struct {
	Server         string `yaml:"server" json:"server" env:"server"`        // 服务器地址
	Timeout        int    `yaml:"timeout" json:"timeout" env:"timeout"`     // 连接超时时间，默认30秒
	Username       string `yaml:"username" json:"username" env:"username"`  // 用户名
	Email          string `yaml:"email" json:"email" env:"email"`           // 邮箱
	Password       string `yaml:"password" json:"password" env:"password"`  // 密码
	HeaderTagName  string `yaml:"header_tag_name" json:"header_tag_name"`   // 请求头标记名
	HeaderTagValue string `yaml:"header_tag_value" json:"header_tag_value"` // 请求头标记默认值
}

// 校验配置
func validateConfig(config ConfigSmtp) error {
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
