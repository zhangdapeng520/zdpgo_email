package main

import (
	"fmt"
	"mail-go/mail"
	//mail "mail-go/mailTest"
)

const (
	//MAIL_USER	= "test@xmirror.org"
	//MAIL_USER	= "test"
	//MAIL_PWD	= "test123"

	MAIL_USER	= "zhangsan"
	MAIL_PWD	= "zhangsan123"

	MAIL_SMTP_HOST	= "10.1.3.208"
	MAIL_SMTP_PORT	= 25

	MAIL_IMAP_HOST  = "10.1.3.208"
	MAIL_IMAP_PORT	= 143

	//MAIL_POP3_HOST  = "10.1.3.208"
	//MAIL_POP3_PORT	= 110
)

//const (
//	MAIL_USER	= "3103557991@qq.com"
//	MAIL_PWD	= ""
//
//
//	MAIL_SMTP_HOST	= "smtp.qq.com"
//	// SSL端口：465 非SSL端口：25 exchange SSL端口：587
//	MAIL_SMTP_PORT	= 25
//
//	MAIL_IMAP_HOST  = "imap.qq.com"
//	// // SSL端口：993 非SSL端口：143
//	MAIL_IMAP_PORT	= 143
//)

func main() {
	//flag.Parse()
	//if (to != nil) && (file != nil) {
	//	//字符串分割, 使用字符分割出to,file
	//	tos := strings.Split(*to, ";")
	//	files := strings.Split(*file, ";")
	//	err := SendMail(tos, *sub, *bd, files)
	//	if err != nil {
	//		fmt.Printf("请求异常，请检查请求参数！")
	//		return
	//	}
	//}

	//err := mail.SendGoMail([]string{"zhangsan@xmirror.org"}, "你好", "这是golang测试邮件", nil)
	////err := mail.SendGoMail([]string{"3103557991@qq.com"}, "你好", "test send", nil)
	//if err != nil {
	//	fmt.Printf("发送异常")
	//}

	sc,err := mail.SMTPLogin(MAIL_SMTP_HOST,MAIL_SMTP_PORT,MAIL_USER,MAIL_PWD,false)
	if err == nil{
		defer sc.Close()

		nickname := "xmirror"
		from := fmt.Sprintf("%s<%s>", nickname, "test@xmirror.org")
		//from := fmt.Sprintf("%s<%s>", nickname, "3103557991@qq.com")
		// qq邮箱有校验，不能随意伪造
		//fakefrom := "xxx@qq.com"
		//m.SetHeader("From",nickname + "<" + fakefrom + ">")
		sendConfig := mail.SendConfig{Uid:"localhost-test",From: from,To: []string{"zhangsan@xmirror.org"},Subject:"你好",Body:"这是golang测试邮件"}
		//sendConfig := mail.SendConfig{Uid:"test2",From: from,To: []string{"3103557991@qq.com"},Subject:"你好",Body:"测试邮件"}
		err =  mail.SendMail(sc,&sendConfig)
		//err =  mail.SendMail(c,from,[]string{"zhangsan@xmirror.org"}, "附件测试", "这是附件测试邮件", []string{"D:\\Tool\\GoProjects\\fscan-main.zip","D:/Tool/GoProjects/gobyexample-master.zip"})

		if err != nil{
			fmt.Print(err)
		}else{
			fmt.Println("send success")
		}
	}else{
		fmt.Print(err)
	}

	//opts := map[string]interface{}{"ssl": false}

	c,err := mail.ImapLogin(MAIL_IMAP_HOST,MAIL_IMAP_PORT,MAIL_USER,MAIL_PWD,false)
	if err == nil{
		defer c.Logout()

		//qq邮箱 preFilter不能有中文 有中文的字段放到postFilter
		//qq邮箱 发送成功后需要等一会儿再查询接收到的邮件
		//preFilter := mail.PreFilter{From: "3103557991@qq.com", SentSince: "2022-03-26"}
		//postFilter := mail.PostFilter{Subject: "测试",Body: []string{"邮件"}}
		//preFilter := mail.PreFilter{From: "3103557991@qq.com", SentSince: "2022-04-05"}
		//postFilter := mail.PostFilter{Subject: "附件测试",Body: []string{"中文邮件"},Attachments: []string{"中文附件"}}
		//preFilter := mail.PreFilter{Uid:"test2", From: "3103557991@qq.com", SentSince: "2022-04-07"}
		//postFilter := mail.PostFilter{}
		//searchResults,err := mail.RecvSearch(c,&preFilter,&postFilter)

		//localhost
		//preFilter := mail.PreFilter{Subject:"中文附件", From: "zhangsan@xmirror.org", SentSince: "2022-03-26"}
		//postFilter := mail.PostFilter{Subject: "中文附件",From: "zhangsan@xmirror.org",Body: []string{"邮件"}}
		preFilter := mail.PreFilter{Uid:"localhost-test", From: "test@xmirror.org", SentSince: "2022-04-07"}
		postFilter := mail.PostFilter{}
		searchResults,err :=mail.RecvSearch(c,&preFilter,&postFilter)

		if err != nil{
			fmt.Println(err)
		}else if len(searchResults) > 0{
			for _,msg := range searchResults{
				fmt.Println("Subject: ",msg.Subject)
				fmt.Println("From: ",msg.From)
				fmt.Println("To: ",msg.To)
				fmt.Println("X-Mirror-Id: ",msg.Uid)
			}

		}

		fmt.Println("recv end")
	}

	////s := "=?utf-8?B?UVHpgq7nrrHlm6LpmJ8==?= <10000@qq.com>"
	//r, err := base64.StdEncoding.DecodeString("UVHpgq7nrrHlm6LpmJ8==")
	////r,err:= message.DecodeHeader(s)
	//fmt.Println(string(r),err)
}
