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
	Key            string   `json:"key"`             // 唯一key
	Title          string   `json:"title"`           // 标题
	Body           string   `json:"body"`            // 邮件内容
	From           string   `json:"from"`            // 发件人
	To             []string `json:"to"`              // 收件人
	AttachmentName string   `json:"attachment_name"` // 附件名称
	AttachmentPath string   `json:"attachment_path"` // 附件路径
	StartTime      int      `json:"start_time"`      // 发送邮件的开始时间
	EndTime        int      `json:"end_time"`        // 发送邮件的结束时间
	SendStatus     bool     `json:"send_status"`     // 发送状态，成功还是失败
	ReceiveStatus  bool     `json:"receive_status"`  // 接收状态，成功还是失败
}
