package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
)

func main() {
	email := zdpgo_email.NewEmailSmtp()

	err := email.SendWithTag(
		"X-Mirror-Uid",
		"hhh",
		"测试自定义标签发送邮件1112X-Mirror-Uid",
		"<h1>邮件内容X-Mirror-Uid</h1>",
		nil,
		"1156956636@qq.com",
	)

	err = email.SendWithDefaultTag(
		"测试自定义标签发送邮件1112X-ZdpgoEmail-Auther",
		"<h1>邮件内容X-ZdpgoEmail-Auther</h1>",
		nil,
		"1156956636@qq.com",
	)

	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Println("发送邮件成功")
	}
}
