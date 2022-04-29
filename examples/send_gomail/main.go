package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
)

func main() {
	email := zdpgo_email.NewEmailSmtp()

	err := email.SendGoMail(
		"这是一个测试邮件222",
		"不要理我，我在自己玩",
		nil,
		"1156956636@qq.com",
	)

	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Println("发送邮件成功")
	}
}
