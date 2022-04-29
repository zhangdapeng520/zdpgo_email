package main

import (
	"github.com/zhangdapeng520/zdpgo_email"
)

func main() {
	// 创建邮件对象
	email := zdpgo_email.NewEmailSmtp()

	// 简单设置 log 参数
	email.SendText("这是一封测试邮件22222", "我在用Golang发邮件。。。", "1156956636@qq.com")
}
