package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"log"
	"net/smtp"
	"sync"
	"time"
)

func main() {
	// 简单设置l og 参数
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	// 创建有5个缓冲的通道，数据类型是  *email.Email
	ch := make(chan *zdpgo_email.Email, 5)

	// 连接池
	p, err := zdpgo_email.NewPool(
		"smtp.qq.com:25",
		3, // 数量设置成 3 个
		smtp.PlainAuth("", "1156956636@qq.com", "oxhcacwebqllhiaf", "smtp.qq.com"),
	)

	if err != nil {
		log.Fatal("email.NewPool error : ", err)
	}

	// sync 包，控制同步
	var wg sync.WaitGroup

	// 3个协程同时发送邮件
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()

			// 若 ch 无数据，则阻塞， 若 ch 关闭，则退出循环
			for e := range ch {

				// 超时时间 10 秒
				// 调用连接池的发送邮件方法
				// 参数1：*zdpgo_email.Email
				// 参数2：10*time.Second
				err := p.Send(e, 10*time.Second)

				if err != nil {
					log.Printf("p.Send error : %v , e = %v , i = %d\n", err, e, i)
				}
			}
		}()
	}

	// 准备邮件内容
	for i := 0; i < 5; i++ {
		e := zdpgo_email.NewEmail()
		// 设置发送邮件的基本信息
		e.From = "张大鹏 <1156956636@qq.com>"
		e.To = []string{"lxgzhw@163.com"}
		e.Subject = "test email.NewPool " + fmt.Sprintf("  the %d email", i)
		e.Text = []byte(fmt.Sprintf("test email.NewPool , the %d email !", i))
		ch <- e
	}

	// 关闭通道
	close(ch)

	// 等待子协程退出
	wg.Wait()
	log.Println("send successfully ... ")
}
