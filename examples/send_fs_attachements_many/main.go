package main

import (
	"embed"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email"
	"path"
	"send/secret"
	"time"
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
	sendFsAttachmentsMany, err := e.SendFsAttachmentsMany(&fsObj, attachments, "1156956636@qq.com")
	if err != nil {
		fmt.Println("批量发送邮件失败")
		return
	}
	fmt.Println("批量发送邮件成功", sendFsAttachmentsMany)

	// 验证是否发送成功
	time.Sleep(time.Second * 60) // 一分钟以后校验是否发送成功
	var newResults []zdpgo_email.EmailResult
	for _, result := range sendFsAttachmentsMany {
		preFilter := zdpgo_email.PreFilter{
			From:           "1156956636@qq.com",
			SentSince:      "2022-04-29",
			HeaderTagName:  "X-ZdpgoEmail-Auther",
			HeaderTagValue: result.Key,
		}
		fmt.Println("开始测试：", preFilter)
		status := e.IsSendSuccessByKeyValue(preFilter.From, preFilter.SentSince, preFilter.HeaderTagName, preFilter.HeaderTagValue)
		if status {
			fmt.Println("邮件发送成功：", preFilter.HeaderTagValue)
		} else {
			fmt.Println("邮件发送失败：", preFilter.HeaderTagValue)
		}
		result.Status = status
		newResults = append(newResults, result)
	}
	fmt.Println("发送的最终结果：", newResults)
}
