package main

import (
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
	c, err = zdpgo_email.NewEmailImapWithServer(server, nil)

	// 连接失败报错
	if err != nil {
		log.Fatal(err)
	}
	log.Println("连接到邮件服务器成功")

	// 登陆
	if err = c.Login(username, password); err != nil {
		log.Fatal(err)
	}
	log.Println("登录成功")
	defer c.Logout()

	//qq邮箱 preFilter不能有中文 有中文的字段放到postFilter
	//qq邮箱 发送成功后需要等一会儿再查询接收到的邮件
	//preFilter := mail1.PreFilter{From: "3103557991@qq.com", SentSince: "2022-03-26"}
	//postFilter := mail1.PostFilter{Subject: "测试",Body: []string{"邮件"}}
	//preFilter := mail1.PreFilter{From: "3103557991@qq.com", SentSince: "2022-04-05"}
	//postFilter := mail1.PostFilter{Subject: "附件测试",Body: []string{"中文邮件"},Attachments: []string{"中文附件"}}
	//preFilter := mail1.PreFilter{Uid:"test2", From: "3103557991@qq.com", SentSince: "2022-04-07"}
	//postFilter := mail1.PostFilter{}
	//searchResults,err := mail1.RecvSearch(c,&preFilter,&postFilter)

	//localhost
	//preFilter := mail1.PreFilter{Subject:"中文附件", From: "zhangsan@xmirror.org", SentSince: "2022-03-26"}
	//postFilter := mail1.PostFilter{Subject: "中文附件",From: "zhangsan@xmirror.org",Body: []string{"邮件"}}
	//preFilter := zdpgo_email.PreFilter{Uid: "localhost-test", From: "1156956636@qq.com", SentSince: "2022-04-07"}
	preFilter := zdpgo_email.PreFilter{Uid: "localhost-test", From: "1156956636@qq.com", SentSince: "2022-04-07"}
	postFilter := zdpgo_email.PostFilter{}
	searchResults, err := c.SearchBF(&preFilter, &postFilter)

	if err != nil {
		fmt.Println(err)
	} else if len(searchResults) > 0 {
		for _, msg := range searchResults {
			fmt.Println("Subject: ", msg.Subject)
			fmt.Println("From: ", msg.From)
			fmt.Println("To: ", msg.To)
			fmt.Println("X-Mirror-Id: ", msg.Uid)
		}

	}

	fmt.Println("recv end")
}
