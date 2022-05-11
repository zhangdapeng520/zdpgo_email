package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"os"
	"send/secret"
)

func main() {
	smtp := zdpgo_email.ConfigSmtp{
		Debug:    true,
		Username: "1156956636@qq.com",
		Email:    "1156956636@qq.com",
		Password: secret.SmtpPassword,
		SmtpHost: "smtp.qq.com",
		SmtpPort: 465,
		IsSSL:    true, // QQ邮箱必须为true
	}
	e, err := zdpgo_email.NewWithSmtpConfig(smtp)
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
