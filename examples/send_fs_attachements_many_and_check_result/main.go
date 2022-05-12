package main

import (
	"embed"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"path"
	"send/secret"
)

//go:embed upload/*
var fsObj embed.FS

func main() {
	e, _ := zdpgo_email.NewWithConfig(zdpgo_email.Config{
		Debug: true,
		Smtp: zdpgo_email.ConfigEmail{
			Email:    "1156956636@qq.com",
			Username: "1156956636@qq.com",
			Password: secret.SmtpPassword,
			Host:     "smtp.qq.com",
			Port:     465,
			IsSSL:    true,
		},
		Imap: zdpgo_email.ConfigEmail{
			Email:    "1156956636@qq.com",
			Username: "1156956636@qq.com",
			Password: secret.ImapPassword,
			Host:     "imap.qq.com",
			Port:     993,
			IsSSL:    true,
		},
	})

	dirFiles, err := fsObj.ReadDir("upload")
	if err != nil {
		fmt.Println("读取文件夹失败", err)
		return
	}
	var attachments []string
	for _, f := range dirFiles {
		attachments = append(attachments, path.Join("upload", f.Name()))
	}

	// 批量发送邮件
	sendFsAttachmentsMany, err := e.SendFsAttachmentsManyAndCheckResult(&fsObj, attachments, "1156956636@qq.com")
	if err != nil {
		fmt.Println("批量发送邮件失败")
		return
	}
	fmt.Println("批量发送邮件成功", sendFsAttachmentsMany)
	for _, result := range sendFsAttachmentsMany {
		fmt.Println(result.Status, result.Title, result.Key)
	}
}
