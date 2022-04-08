package main

import (
	"github.com/zhangdapeng520/zdpgo_email"
	"github.com/zhangdapeng520/zdpgo_yaml"
)

func main() {
	// 读取配置
	y := zdpgo_yaml.New()
	var config zdpgo_email.Config
	err := y.ReadDefaultConfig(&config)
	if err != nil {
		panic(err)
	}

	// 创建邮件对象
	email := zdpgo_email.New(config)

	//设置文件发送的内容
	content := `
   <h3><a href="https://www.baidu.com">欢迎来到张大鹏的主页</a></h3>
<br/>
<table border="0" style="font-family: '微软雅黑 Light';" cellpadding="3px">
    <tr>
        <td> 1 </td>
        <td><a href="https://www.baidu.com">GO 中 defer的实现原理</a></td>
    </tr>
    <tr>
        <td> 2 </td>
        <td><a href="https://www.baidu.com">GO 中 Chan 实现原理分享</a></td>
    </tr>
    <tr>
        <td> 3 </td>
        <td><a href="https://www.baidu.com">GO 中 map 的实现原理</a></td>
    </tr>
    <tr>
        <td> 4 </td>
        <td><a href="https://www.baidu.com">GO 中 slice 的实现原理</a></td>
    </tr>
</table>
`

	// 发送HTML邮件
	email.SendHtml("这是一封测试邮件", content, "lxgzhw@163.com")
}
