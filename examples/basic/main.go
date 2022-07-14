package main

import (
	"github.com/zhangdapeng520/zdpgo_email"
	"github.com/zhangdapeng520/zdpgo_email/examples/basic/send"
	"github.com/zhangdapeng520/zdpgo_email/examples/secret"
)

func main() {
	e, _ := zdpgo_email.NewWithConfig(&zdpgo_email.Config{
		Email:    secret.Email,
		Username: secret.Username,
		Password: secret.Password,
		Host:     secret.Host,
		Port:     secret.Port,
		IsSSL:    true,
	})
	send.Send(e)     // 发送单个HTML邮件
	send.SendMany(e) // 批量发送HTML邮件
}
