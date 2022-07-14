package main

import (
	"fmt"

	"github.com/zhangdapeng520/zdpgo_email"
)

func main() {
	e, _ := zdpgo_email.NewWithConfig(&zdpgo_email.Config{
		Email:    email,
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
		IsSSL:    true,
	})

	req := zdpgo_email.EmailRequest{
		Title:    "单个HTML测试",
		Body:     "https://www.baidu.com",
		ToEmails: []string{"lxgzhw@163.com", "1156956636@qq.com"},
	}
	result, err := e.Send(req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result.Title, result.SendStatus, result.Key, result.StartTime, result.EndTime)
}
