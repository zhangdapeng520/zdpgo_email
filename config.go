package zdpgo_email

// Config 配置类
type Config struct {
	HeaderTagName        string `yaml:"header_tag_name" json:"header_tag_name"`               // 请求头标记名
	HeaderTagValue       string `yaml:"header_tag_value" json:"header_tag_value"`             // 请求头标记值
	CommonTitle          string `yaml:"common_title" json:"common_title"`                     // 通用邮件标题
	Id                   string `yaml:"id" json:"id"`                                         // 用于SMTP登录
	SendSleepMillisecond int    `yaml:"send_sleep_millisecond" json:"send_sleep_millisecond"` // 发送邮件的间隔休眠时间
	Timeout              int    `yaml:"timeout" json:"timeout"`                               // 超时时间
	Email                string `yaml:"email" json:"email"`                                   // 发送者的邮箱
	Username             string `yaml:"username" json:"username"`                             // 发送者的名字
	Password             string `yaml:"password" json:"password"`                             // 发送者的邮箱的校验密码（不一定是登陆密码）
	Host                 string `yaml:"smtp_host" json:"smtp_host"`                           // 邮箱服务器的主机地址（域名）
	Port                 int    `yaml:"smtp_port" json:"smtp_port"`                           // 端口
	IsSSL                bool   `yaml:"is_ssl" json:"is_ssl"`                                 // 是否为SSL模式
}
