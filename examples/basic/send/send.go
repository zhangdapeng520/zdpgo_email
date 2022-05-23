package send

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
)

/*
@Time : 2022/5/23 14:42
@Author : 张大鹏
@File : send.go
@Software: Goland2021.3.1
@Description:
*/

func SendHtmlMany(e *zdpgo_email.Email) {
	contents := []string{
		"https://www.baidu.com",
		"https://www.sogo.com",
		"https://www.google.com",
	}

	// 批量发送邮件
	sendFsAttachmentsMany, err := e.SendHtmlMany(contents, "lxgzhw@163.com")
	if err != nil {
		fmt.Println("批量发送邮件失败")
		return
	}
	fmt.Println("批量发送邮件成功", sendFsAttachmentsMany)
	for _, result := range sendFsAttachmentsMany {
		fmt.Println(result.Title, result.SendStatus, result.Key, result.StartTime, result.EndTime)
	}
}

func SendAttachmentMany(e *zdpgo_email.Email) {
	// 准备数据
	data := []struct {
		Attachments []string
		ToEmail     string
	}{
		{Attachments: []string{"uploads/1.txt", "uploads/2.doc", "uploads/3.exe"}, ToEmail: "lxgzhw@163.com"},
	}

	// 测试数据
	for _, d := range data {
		emailResults, err := e.SendAttachmentMany(d.Attachments, d.ToEmail)
		if err != nil {
			panic(err)
		}
		for _, result := range emailResults {
			fmt.Println(result.Title, result.SendStatus)
		}
	}
}

func SendAttachments(e *zdpgo_email.Email) {
	attachments := []string{"uploads/1.txt", "uploads/2.doc", "uploads/3.exe"}
	e.SendAttachments("", "", attachments, "lxgzhw@163.com")
}
