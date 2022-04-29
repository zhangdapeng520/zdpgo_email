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
		Server:   "imap.qq.com:993",
		Username: "1156956636@qq.com",
		Password: "gluhsjysrosbbadc",
	})

	// 连接失败报错
	if err != nil {
		log.Fatal(err)
	}
	log.Println("登录成功")
	defer c.Logout()

	// 要测试的内容
	fileters := []zdpgo_email.PreFilter{
		{From: "1156956636@qq.com", SentSince: "2022-04-07", HeaderTag: "zhangdapeng520"},
		{From: "1156956636@qq.com", SentSince: "2022-04-07"},
		{From: "1156956636@qq.com"},
		{From: "1156956636@qq.com", HeaderTag: "zhangdapeng520"},
	}

	// 进行测试
	postFilter := zdpgo_email.PostFilter{}
	for _, preFilter := range fileters {
		fmt.Println("开始测试：", preFilter)
		searchResults, err := c.SearchBF(&preFilter, &postFilter)
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
