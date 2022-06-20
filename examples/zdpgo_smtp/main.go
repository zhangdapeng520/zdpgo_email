package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"github.com/zhangdapeng520/zdpgo_log"
)

/*
@Time : 2022/6/6 11:29
@Author : 张大鹏
@File : main.go
@Software: Goland2021.3.1
@Description:
*/

func main() {
	e, _ := zdpgo_email.NewWithConfig(&zdpgo_email.Config{
		Email:    "1156956636@qq.com",
		Username: "zhangdapeng520",
		Password: "zhangdapeng520",
		Host:     "localhost",
		Port:     3333,
		IsSSL:    false,
	}, zdpgo_log.Tmp)
	req := zdpgo_email.EmailRequest{
		Title:    "单个HTML测试",
		Body:     "https://www.baidu.com",
		ToEmails: []string{"lxgzhw@163.com", "1156956636@qq.com"},
	}
	result, err := e.Send(req)
	if err != nil {
		e.Log.Error("发送邮件失败", "error", err)
	}
	fmt.Println(result.Title, result.SendStatus, result.Key, result.StartTime, result.EndTime)
}
