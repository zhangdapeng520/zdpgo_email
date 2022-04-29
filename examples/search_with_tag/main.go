package main

import (
	"encoding/json"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"log"
)

func main() {
	var (
		c   *zdpgo_email.EmailImap // email邮件客户端
		err error
	)

	// 建立连接
	c, err = zdpgo_email.NewEmailImapWithConfig(zdpgo_email.ConfigImap{
		Server:   server,
		Username: username,
		Password: password,
	})

	// 连接失败报错
	if err != nil {
		log.Fatal(err)
	}
	log.Println("登录成功")
	defer c.Logout()

	// 要测试的内容
	fileters := []zdpgo_email.PreFilter{
		//{From: "1156956636@qq.com", SentSince: "2022-04-29", HeaderTagName: "abc123", HeaderTagValue: "hhh"},
		{From: "1156956636@qq.com", HeaderTagName: "X-ZdpgoEmail-Auther", HeaderTagValue: "zhangdapeng520"},
	}

	// 进行测试
	for _, preFilter := range fileters {
		fmt.Println("开始测试：", preFilter)
		searchResults, err := c.SearchByTag(preFilter.From, preFilter.SentSince, preFilter.HeaderTagName, preFilter.HeaderTagValue)
		if err != nil {
			fmt.Println(err)
		} else if len(searchResults) > 0 {
			for _, msg := range searchResults {
				jsonMsg, err := json.Marshal(msg)
				if err != nil {
					return
				}
				fmt.Println(string(jsonMsg))
			}
		}
		fmt.Println("结束测试：", preFilter)
	}
}
