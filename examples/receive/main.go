package main

import (
	"github.com/zhangdapeng520/zdpgo_email"
	"github.com/zhangdapeng520/zdpgo_email/imap"
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

	// 创建20个收件箱
	mailboxes := make(chan *imap.MailboxInfo, 30)
	go func() {
		c.List("", "*", mailboxes)
	}()

	// 列取邮件夹
	for m := range mailboxes {
		mbox, err := c.Select(m.Name, false)
		if err != nil {
			log.Fatal(err)
		}
		to := mbox.Messages
		log.Printf("%s : %d", m.Name, to)
	}
}
