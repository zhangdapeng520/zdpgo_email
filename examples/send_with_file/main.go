package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"os"
	"send/secret"
)

func main() {
	e, err := zdpgo_email.NewWithConfig(zdpgo_email.Config{
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
	if err != nil {
		fmt.Println("创建邮件对象失败", "error", err)
		return
	}

	testFile, _ := os.Open("upload/test.txt")
	var files = map[string]*os.File{
		"upload/test.txt": testFile,
	}
	err = e.Send.SendWithDefaultTagWithFiles(
		files,
		e.Random.Str.Str(16),
		e.Random.Str.Str(128),
		"1156956636@qq.com",
	)

	if err != nil {
		fmt.Print("发送邮件失败：", err)
	} else {
		fmt.Println("发送邮件成功")
	}
}
