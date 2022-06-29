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

func Send(e *zdpgo_email.Email) {
	req := zdpgo_email.EmailRequest{
		Title:    "单个HTML测试",
		Body:     "https://www.baidu.com",
		ToEmails: []string{"lxgzhw@163.com", "1156956636@qq.com"},
	}
	result, err := e.Send(req)
	if err != nil {
		fmt.Println("发送邮件失败", "error", err)
	}
	fmt.Println(result.Title, result.SendStatus, result.Key, result.StartTime, result.EndTime)
}

func SendMany(e *zdpgo_email.Email) {
	reqList := []zdpgo_email.EmailRequest{
		{Body: "https://www.baidu.com", ToEmails: []string{"lxgzhw@163.com", "1156956636@qq.com"},
			CcEmails: []string{"zhangdp@anpro-tech.com"}},
		{Body: "https://www.sogo.com", ToEmails: []string{"lxgzhw@163.com", "1156956636@qq.com"}, CcEmails: []string{"zhangdp@anpro-tech.com"}},
		{Body: "https://www.google.com", ToEmails: []string{"lxgzhw@163.com", "1156956636@qq.com"}, CcEmails: []string{"zhangdp@anpro-tech.com"}},
		{Attachments: []string{"uploads/1.txt"}, ToEmails: []string{"lxgzhw@163.com", "1156956636@qq.com"}, CcEmails: []string{"zhangdp@anpro-tech.com"}},
		{Attachments: []string{"uploads/2.txt"}, ToEmails: []string{"lxgzhw@163.com", "1156956636@qq.com"}, CcEmails: []string{"zhangdp@anpro-tech.com"}},
		{Attachments: []string{"uploads/3.txt"}, ToEmails: []string{"lxgzhw@163.com", "1156956636@qq.com"}, CcEmails: []string{"zhangdp@anpro-tech.com"}},
		{Attachments: []string{"uploads/1.txt", "uploads/2.txt", "uploads/3.txt"}, ToEmails: []string{"lxgzhw@163.com", "1156956636@qq.com"}, CcEmails: []string{"zhangdp@anpro-tech.com"}},
	}
	// 批量发送邮件
	sendFsAttachmentsMany, err := e.SendMany(reqList)
	if err != nil {
		fmt.Println("批量发送邮件失败")
		return
	}
	fmt.Println("批量发送邮件成功", sendFsAttachmentsMany)
	for _, result := range sendFsAttachmentsMany {
		fmt.Println(result.Title, result.SendStatus, result.Key, result.StartTime, result.EndTime)
	}
}
