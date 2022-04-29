package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
)

func main() {
	email := zdpgo_email.NewEmailSmtp()
	closer, err := email.GetGoMailSendCloser()
	if err != nil {
		fmt.Print(err)
	}
	defer closer.Close()

	sendConfig := zdpgo_email.SendConfig{Uid: "localhost-test", From: "1156956636@qq.com", To: []string{"1156956636@qq.com"},
		Subject: "你好111", Body: "这是golang测试邮件"}
	err = email.SendGoMail1(closer, &sendConfig)

	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Println("send success")
	}
}
