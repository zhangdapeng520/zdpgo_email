package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
)

func main() {
	e, _ := zdpgo_email.NewWithConfig(zdpgo_email.Config{
		SmtpConfigs: []string{"config/config_smtp.yaml", "config/secret/config_smtp.yaml"},
		ImapConfigs: nil,
	})

	attachments := []string{
		"README.md",
	}
	err := e.Send.SendWithDefaultTag(
		e.Random.Str.Str(16),
		e.Random.Str.Str(128),
		attachments,
		"1156956636@qq.com",
	)

	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Println("发送邮件成功")
	}
}
