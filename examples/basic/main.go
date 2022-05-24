package main

import (
	"basic/search"
	"basic/secret"
	"basic/send"
	"github.com/zhangdapeng520/zdpgo_email"
)

func main() {
	e, _ := zdpgo_email.NewWithConfig(&zdpgo_email.Config{
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
	send.Send(e)     // 发送单个HTML邮件
	send.SendMany(e) // 批量发送HTML邮件
	search.Search(e) // 搜索
}
