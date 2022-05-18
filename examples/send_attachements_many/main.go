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

	// 准备数据
	data := []struct {
		Attachments []string
		ToEmail     string
	}{
		{Attachments: []string{"uploads/1.txt", "uploads/2.doc", "uploads/3.exe"}, ToEmail: "lxgzhw@163.com"},
	}

	// 测试数据
	for _, d := range data {
		emailResults, err := e.SendAttachmentMany(1, d.Attachments, d.ToEmail)
		if err != nil {
			panic(err)
		}
		for _, result := range emailResults {
			fmt.Println(result.Title, result.SendStatus)
		}
	}
}
