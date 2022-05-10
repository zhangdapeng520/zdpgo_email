package main

import (
	"embed"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
)

//go:embed upload/* email_file/*
var fsObj embed.FS

func main() {
	e, _ := zdpgo_email.NewWithConfig(zdpgo_email.Config{
		SmtpConfigs: []string{"config/config_smtp.yaml", "config/secret/config_smtp.yaml"},
		ImapConfigs: nil,
		Fs:          fsObj, // 嵌入文件系统
		IsUseFs:     true,
	})

	attachments := []string{
		"email_file/95557a29de4b70a25ce62a03472be684",
	}
	err := e.Send.SendWithDefaultTagWithFs(
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
