package main

import (
	"github.com/zhangdapeng520/zdpgo_email"
	"log"
	"net/smtp"
)

func main() {
	// 简单设置 log 参数
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	em := zdpgo_email.NewEmail()

	// 设置 sender 发送方 的邮箱 ， 此处可以填写自己的邮箱
	em.From = "张大鹏 <1156956636@qq.com>"

	// 设置 receiver 接收方 的邮箱  此处也可以填写自己的邮箱， 就是自己发邮件给自己
	em.To = []string{"lxgzhw@163.com"}

	// 设置主题
	em.Subject = "小魔童给你发邮件了"

	// 简单设置文件发送的内容，暂时设置成纯文本
	em.Text = []byte("hello world， 咱们用 golang 发个邮件xxxxxxxxxxxxxxxxxxxx！！")

	//设置服务器相关的配置
	err := em.Send("smtp.qq.com:25", smtp.PlainAuth("", "1156956636@qq.com", "oxhcacwebqllhiaf", "smtp.qq.com"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("send successfully ... ")
}
