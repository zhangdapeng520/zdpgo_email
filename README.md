# zdpgo_email

使用Golang操作Email

项目地址：https://github.com/zhangdapeng520/zdpgo_email

## 版本历史

- v0.1.0 2022/4/8 常用功能
- v1.0.1 2022/5/3 新增：基于key的发送和校验
- v1.0.2 2022/5/10 优化：新增邮箱如果失败，则返回错误而非抛出错误
- v1.0.3 2022/5/10 新增：支持使用嵌入文件系统发送邮件附件
- v1.0.4 2022/5/10 BUG修复：修复嵌入文件无法正常读取的BUG
- v1.0.5 2022/5/11 BUG修复：邮箱连接失败导致异常退出
- v1.0.6 2022/5/12 新增：发送邮件并校验结果
- v1.0.7 2022/5/12 优化：代码优化
- v1.0.8 2022/5/13 升级：升级random组件为v1.1.5

## 使用示例

### 发送邮件

```go
package main

import (
	"embed"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"send/secret"
)

//go:embed upload/*
var fsObj embed.FS

func main() {
	fmt.Println("===============", fsObj)

	smtp := zdpgo_email.ConfigSmtp{
		Username: "1156956636@qq.com",
		Email:    "1156956636@qq.com",
		Password: secret.SmtpPassword,
		SmtpHost: "smtp.qq.com",
		SmtpPort: 465,
		IsSSL:    true,
		Fs:       &fsObj,
	}
	imap := zdpgo_email.ConfigImap{
		Server:   "imap.qq.com:993",
		Username: "1156956636@qq.com",
		Email:    "1156956636@qq.com",
		Password: secret.ImapPassword,
	}
	e, _ := zdpgo_email.NewWithSmtpAndImapConfig(smtp, imap)

	attachments := []string{
		"upload/test.txt",
	}
	err := e.Send.SendWithDefaultTagWithFs(
		&fsObj,
		e.Random.Str.Str(16),
		e.Random.Str.Str(128),
		attachments,
		"1156956636@qq.com",
	)

	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Println("发送邮件成功")
	}
}
```
