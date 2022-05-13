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
	})
	contents := []string{
		"https://www.baidu.com",
		"https://www.sogo.com",
		"https://www.google.com",
	}

	// 批量发送邮件
	sendFsAttachmentsMany, err := e.SendHtmlMany(contents, 1, "1156956636@qq.com")
	if err != nil {
		fmt.Println("批量发送邮件失败")
		return
	}
	fmt.Println("批量发送邮件成功", sendFsAttachmentsMany)
	for _, result := range sendFsAttachmentsMany {
		fmt.Println(result.Title, result.SendStatus, result.Key, result.StartTime, result.EndTime)
	}
}
