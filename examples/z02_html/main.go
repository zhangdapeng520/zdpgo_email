package main

import (
	"github.com/zhangdapeng520/zdpgo_email"
	"log"
	"net/smtp"
)

func main() {
	// 简单设置l og 参数
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	em := zdpgo_email.NewEmail()
	//设置发送方的邮箱
	em.From = "张大鹏 <1156956636@qq.com>"

	// 设置接收方的邮箱
	em.To = []string{"lxgzhw@163.com"}

	// 抄送
	//em.Cc = []string{"xxx@qq.com"}

	// 密送
	//em.Bcc = []string{"xxx@qq.com"}

	//设置主题
	em.Subject = "小魔童给你发邮件了"
	//设置文件发送的内容
	em.HTML = []byte(`
   <h3><a href="https://juejin.cn/user/3465271329953806">欢迎来到小魔童哪吒的主页</a></h3>
<br/>
<table border="0" style="font-family: '微软雅黑 Light';" cellpadding="3px">
    <tr>
        <td> 1 </td>
        <td><a href="https://juejin.cn/post/6975686540601245709">GO 中 defer的实现原理</a></td>
    </tr>
    <tr>
        <td> 2 </td>
        <td><a href="https://juejin.cn/post/6975280009082568740">GO 中 Chan 实现原理分享</a></td>
    </tr>
    <tr>
        <td> 3 </td>
        <td><a href="https://juejin.cn/post/6974908615232585764">GO 中 map 的实现原理</a></td>
    </tr>
    <tr>
        <td> 4 </td>
        <td><a href="https://juejin.cn/post/6974539862800072718">GO 中 slice 的实现原理</a></td>
    </tr>
    <tr>
        <td> 5 </td>
        <td> <a href="https://juejin.cn/post/6974169270624190495">GO 中 string 的实现原理</a></td>
    </tr>
    <tr>
        <td> 6 </td>
        <td><a href="https://juejin.cn/post/6973793593987170317">GO 中 ETCD 的编码案例分享</a></td>
    </tr>
    <tr>
        <td> 7 </td>
        <td><a href="https://juejin.cn/post/6973455825905909797">服务注册与发现之ETCD</a></td>
    </tr>
    <tr>
        <td> 8 </td>
        <td> <a href="https://juejin.cn/post/6973108979777929230">GO通道和 sync 包的分享</a></td>
    </tr>
    <tr>
        <td> 9 </td>
        <td> <a href="https://juejin.cn/post/6972846349968474142">GO的锁和原子操作分享</a></td>
    </tr>
</table>
`)
	// 添加附件
	//em.AttachFile("./README.md")
	//em.AttachFile("./test.html")
	//em.AttachFile("./test.txt")
	em.AttachFile("./test.1txt")

	// 设置服务器相关的配置
	err := em.Send("smtp.qq.com:25", smtp.PlainAuth("", "1156956636@qq.com", "oxhcacwebqllhiaf", "smtp.qq.com"))

	if err != nil {
		log.Fatal(err)
	}
	log.Println("send successfully ... ")
}
