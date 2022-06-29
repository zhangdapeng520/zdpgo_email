package main

import (
	"github.com/zhangdapeng520/zdpgo_email"
	"github.com/zhangdapeng520/zdpgo_email/examples/basic/send"
)

func main() {
	e, _ := zdpgo_email.NewWithConfig(&zdpgo_email.Config{
		Email:    email,
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
		IsSSL:    true,
	})
	send.Send(e)     // 发送单个HTML邮件
	send.SendMany(e) // 批量发送HTML邮件
}
