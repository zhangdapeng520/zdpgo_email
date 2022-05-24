package search

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
)

/*
@Time : 2022/5/23 15:00
@Author : 张大鹏
@File : search.go
@Software: Goland2021.3.1
@Description:
*/

func Search(e *zdpgo_email.Email) {

	// 要测试的内容
	fileters := []zdpgo_email.PreFilter{
		{From: "1156956636@qq.com", SentSince: "2022-04-29", HeaderTagName: "X-ZdpgoEmail-Auther", HeaderTagValue: "zhangdapeng520"},
	}

	// 进行测试
	for _, preFilter := range fileters {
		fmt.Println("开始测试：", preFilter)
		searchResults, err := e.ReceiveObj.SearchByTag(preFilter.From, preFilter.SentSince, preFilter.HeaderTagName,
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
