package main

import (
	"encoding/json"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"log"
)

func main() {
	e := zdpgo_email.NewWithConfig(zdpgo_email.Config{
		SmtpConfigs: nil,
		ImapConfigs: []string{"config/config_imap.yaml", "config/secret/config_imap.yaml"},
	})

	log.Println("登录成功")
	defer e.Receive.Logout()

	// 要测试的内容
	fileters := []zdpgo_email.PreFilter{
		{From: "1156956636@qq.com", SentSince: "2022-04-29"},
		{From: "1156956636@qq.com"},
	}

	// 判断是否发送成功
	fmt.Println("是否发送成功：", e.Receive.IsSendSuccessByKey("1156956636@qq.com", "fMWyiYXfxoTtCCXdfLvvdcYIzXrtesbX"))

	// 进行测试
	for _, preFilter := range fileters {
		fmt.Println("开始测试：", preFilter)
		searchResults, err := e.Receive.SearchByKeyToday(preFilter.From, "fMWyiYXfxoTtCCXdfLvvdcYIzXrtesbX")
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
