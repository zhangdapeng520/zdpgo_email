# zdpgo_email

使用Golang操作Email

项目地址：https://github.com/zhangdapeng520/zdpgo_email

## 版本历史

- v0.1.0 2022年4月8日 常用功能
- v1.0.1 2022/5/3 新增：基于key的发送和校验
- v1.0.2 2022/5/10 优化：新增邮箱如果失败，则返回错误而非抛出错误

## 使用示例

### 发送普通邮件

```go
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

	// 简单设置 log 参数
	email.SendText("这是一封测试邮件", "我在用Golang发邮件。。。", "lxgzhw@163.com")
}
```

### 发送HTML邮件

```go
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
```