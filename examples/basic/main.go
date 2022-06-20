package main

import (
	"github.com/zhangdapeng520/zdpgo_email"
	"github.com/zhangdapeng520/zdpgo_email/examples/basic/send"
	"github.com/zhangdapeng520/zdpgo_log"
)

func main() {
	e, _ := zdpgo_email.NewWithConfig(&zdpgo_email.Config{
		Email:    "1156956636@qq.com",
		Username: "1156956636@qq.com",
		Password: "",
		Host:     "smtp.qq.com",
		Port:     465,
		IsSSL:    true,
	}, zdpgo_log.Tmp)
	send.Send(e)     // 发送单个HTML邮件
	send.SendMany(e) // 批量发送HTML邮件
}
