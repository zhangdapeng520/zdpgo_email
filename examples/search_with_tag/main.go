package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"log"
)

func main() {
	// 建立连接
	e, err := zdpgo_email.NewWithConfig(zdpgo_email.Config{
		SmtpConfigs: nil,
		ImapConfigs: []string{"config/config_imap.yaml", "config/secret/config_imap.yaml"},
	})

	// 连接失败报错
	if err != nil {
		log.Fatal(err)
	}
	log.Println("登录成功")

	// 要测试的内容
	fileters := []zdpgo_email.PreFilter{
		//{From: "1156956636@qq.com", SentSince: "2022-04-29", HeaderTagName: "abc123", HeaderTagValue: "hhh"},
		{From: "1156956636@qq.com", HeaderTagName: "X-ZdpgoEmail-Auther", HeaderTagValue: "zhangdapeng520"},
	}

	// 进行测试
	for _, preFilter := range fileters {
		fmt.Println("开始测试：", preFilter)
		searchResults, err := e.Receive.SearchByTag(preFilter.From, preFilter.SentSince, preFilter.HeaderTagName,
			preFilter.HeaderTagValue)
		if err != nil {
			fmt.Println(err)
		} else if len(searchResults) > 0 {
			for _, msg := range searchResults {
				fmt.Println("=========================")
				fmt.Println(msg.Subject)
				fmt.Println(msg.From)
				fmt.Println(msg.To)
				fmt.Println(msg.HeaderTagName)
				fmt.Println(msg.HeaderTagValue)
				fmt.Println(msg.Attachments)
				fmt.Println("=========================")
			}
		}
		fmt.Println("结束测试：", preFilter)
	}
}
