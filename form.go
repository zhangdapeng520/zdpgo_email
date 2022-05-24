package zdpgo_email

import (
	"embed"
	"os"
)

/*
@Time : 2022/5/11 20:23
@Author : 张大鹏
@File : form.go
@Software: Goland2021.3.1
@Description: email表单相关数据
*/

// EmailResult 邮件结果
type EmailResult struct {
	Key           string   `json:"key"`            // 唯一key
	Title         string   `json:"title"`          // 标题
	Body          string   `json:"body"`           // 邮件内容
	Type          string   `json:"type" `          // 邮件类型：html、text文本
	From          string   `json:"from"`           // 发件人
	ToEmails      []string `json:"to_emails"`      // 收件人
	CcEmails      []string `json:"cc_emails"`      // 抄送人邮箱地址列表
	BccEmails     []string `json:"bcc_emails"`     // 密送人邮箱地址列表
	Attachments   []string `json:"attachments"`    // 附件列表
	StartTime     int      `json:"start_time"`     // 发送邮件的开始时间
	EndTime       int      `json:"end_time"`       // 发送邮件的结束时间
	SendStatus    bool     `json:"send_status"`    // 发送状态，成功还是失败
	ReceiveStatus bool     `json:"receive_status"` // 接收状态，成功还是失败
}

// EmailRequest 邮件请求信息
type EmailRequest struct {
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Type        string    `json:"type"`  // 类型：text,html
	IsFs        bool      `json:"is_fs"` // 是否为嵌入文件系统
	Fs          *embed.FS // 嵌入文件系统对象
	IsFiles     bool      `json:"is_readers"` // 是否为自定义的文件输入流
	Files       map[string]*os.File
	Attachments []string `json:"attachments"`
	ToEmails    []string `json:"to_emails"`  // 收件人
	CcEmails    []string `json:"cc_emails"`  // 抄送人
	BccEmails   []string `json:"bcc_emails"` // 密送人
}
