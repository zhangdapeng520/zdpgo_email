package main

import (
	"embed"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"send/secret"
)

//go:embed upload/*
var fsObj embed.FS

func main() {
	fmt.Println("===============", fsObj)

	smtp := zdpgo_email.ConfigSmtp{
		Username: "1156956636@qq.com",
		Email:    "1156956636@qq.com",
		Password: secret.SmtpPassword,
		SmtpHost: "smtp.qq.com",
		SmtpPort: 465,
		IsSSL:    true,
		Fs:       &fsObj,
	}
	imap := zdpgo_email.ConfigImap{
		Server:   "imap.qq.com:993",
		Username: "1156956636@qq.com",
		Email:    "1156956636@qq.com",
		Password: secret.ImapPassword,
	}
	e, _ := zdpgo_email.NewWithSmtpAndImapConfig(smtp, imap)

	attachments := []string{
		"upload/test.txt",
	}
	err := e.Send.SendWithDefaultTagWithFs(
		&fsObj,
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
