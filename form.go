package zdpgo_email

/*
@Time : 2022/5/11 20:23
@Author : 张大鹏
@File : form.go
@Software: Goland2021.3.1
@Description: email表单相关数据
*/

// EmailResult 邮件结果
type EmailResult struct {
	Key            string   `json:"key" yaml:"key" env:"key"`                                     // 唯一key
	Title          string   `json:"title" yaml:"title" env:"title"`                               // 标题
	From           string   `json:"from" yaml:"from" env:"from"`                                  // 发件人
	To             []string `json:"to" yaml:"to" env:"to"`                                        // 收件人
	AttachmentName string   `json:"attachment_name" yaml:"attachment_name" env:"attachment_name"` // 附近名称
}
