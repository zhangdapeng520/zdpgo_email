package main

import (
	"github.com/zhangdapeng520/zdpgo_email"
	"github.com/zhangdapeng520/zdpgo_yaml"
)

func main() {
	// 读取配置
	y := zdpgo_yaml.New()
	var config zdpgo_email.Config
	err := y.ReadDefaultConfig(&config)
	if err != nil {
		panic(err)
	}

	// 创建邮件对象
	email := zdpgo_email.New(config)

	// 简单设置 log 参数
	email.SendText("这是一封测试邮件", "我在用Golang发邮件。。。", "lxgzhw@163.com")
}
