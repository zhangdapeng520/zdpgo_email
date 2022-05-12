package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"send/secret"
)

func main() {
	e, _ := zdpgo_email.NewWithConfig(zdpgo_email.Config{
		Debug: true,
		Smtp: zdpgo_email.ConfigEmail{
			Email:    "1156956636@qq.com",
			Username: "1156956636@qq.com",
			Password: secret.SmtpPassword,
			Host:     "smtp.qq.com",
			Port:     465,
			IsSSL:    true,
		},
		Imap: zdpgo_email.ConfigEmail{
			Email:    "1156956636@qq.com",
			Username: "1156956636@qq.com",
			Password: secret.ImapPassword,
			Host:     "imap.qq.com",
			Port:     993,
			IsSSL:    true,
		},
	})
	contents := []string{
		"https://www.baidu.com",
		"https://www.sogo.com",
		"https://www.google.com",
	}
	// 批量发送邮件
	sendFsAttachmentsMany, err := e.SendHtmlManyAndCheckResult(contents, "1156956636@qq.com")
	if err != nil {
		fmt.Println("批量发送邮件失败")
		return
	}
	fmt.Println("批量发送邮件成功", sendFsAttachmentsMany)
	for _, result := range sendFsAttachmentsMany {
		fmt.Println(result.Status, result.Title, result.Key)
	}
}
