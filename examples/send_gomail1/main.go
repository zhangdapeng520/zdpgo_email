package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
)

func main() {
	email := zdpgo_email.NewEmailSmtp()

	// 邮件接收方
	mailTo := []string{
		//可以是多个接收人
		"lxgzhw@163.com",
		"1156956636@qq.com",
	}

	subject := "Hello World!" // 邮件主题
	body := "测试发送邮件"          // 邮件正文

	err := email.SendGoMail(mailTo, subject, body)
	if err != nil {
		fmt.Println("Send fail! - ", err)
		return
	}
	fmt.Println("Send successfully!")
}
