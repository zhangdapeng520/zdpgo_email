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
	Key            string   `json:"key" yaml:"key"`                         // 唯一key
	Title          string   `json:"title" yaml:"title"`                     // 标题
	Body           string   `json:"body" yaml:"body"`                       // 邮件内容
	From           string   `json:"from" yaml:"from"`                       // 发件人
	To             []string `json:"to" yaml:"to"`                           // 收件人
	AttachmentName string   `json:"attachment_name" yaml:"attachment_name"` // 附件名称
	AttachmentPath string   `json:"attachment_path" yaml:"attachment_path"` // 附件路径
	Status         bool     `json:"status" yaml:"status"`                   // 发送状态，成功还是失败
}
