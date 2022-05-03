package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
)

func main() {
	e := zdpgo_email.NewWithConfig(zdpgo_email.Config{
		SmtpConfigs: []string{"config/config_smtp.yaml", "config/secret/config_smtp.yaml"},
		ImapConfigs: nil,
	})

	attachments := []string{
		"README.md",
	}
	key, err := e.Send.SendWithKey(
		e.Random.Str.Str(16),
		e.Random.Str.Str(128),
		attachments,
		"1156956636@qq.com",
	)

	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Println("发送邮件成功，唯一标识是：", key)
	}

	//xcgWktkHQaANKtzbUCLRprQEcLWIAekj
}
