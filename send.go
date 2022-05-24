package zdpgo_email

import (
	"errors"
	"github.com/zhangdapeng520/zdpgo_email/gomail"
	"os"
	"strings"
	"time"
)

/*
@Time : 2022/4/29 9:54
@Author : 张大鹏
@File : send.go
@Software: Goland2021.3.1
@Description:发送邮件相关的方法
*/

// SendWithTag 通过标签发送邮件
func (e *Email) SendWithTag(tagKey, tagValue string, req EmailRequest) error {
	// 检查key是否符合规范
	if tagKey != "" {
		tagArr := strings.Split(tagKey, "-")
		if len(tagArr) != 3 {
			msg := "key必须由两个“-”分割的字符组成"
			e.Log.Error("key必须由两个“-”分割的字符组成")
			return errors.New(msg)
		} else if tagArr[0] != "X" {
			msg := "Key的第一个字符必须是大写的X"
			e.Log.Error(msg)
			return errors.New(msg)
		}
		e.Config.HeaderTagName = tagKey
	}

	// 检查value
	if tagValue != "" {
		e.Config.HeaderTagValue = tagValue
	}

	// 发送邮件
	e.Result.Key = e.Config.HeaderTagValue
	err := e.SendGoMail(req)
	if err != nil {
		e.Log.Error("发送邮件失败", "error", err)
		return err
	}
	return nil
}

// GetMessage 获取HTML类型的消息对象
// @param req 邮件请求对象
func (e *Email) GetMessage(req EmailRequest) (*gomail.Message, error) {
	// 创建消息对象
	m := gomail.NewMessage()

	// 设置请求头
	m.SetHeader(e.Config.HeaderTagName, e.Config.HeaderTagValue)

	// 发件人
	m.SetHeader("From", e.Config.Email)

	// 收件人
	if req.ToEmails == nil || len(req.ToEmails) == 0 {
		return nil, errors.New("ToEmails 收件人不能为空")
	}
	m.SetHeader("To", req.ToEmails...)

	// 抄送
	if req.CcEmails != nil && len(req.CcEmails) > 0 {
		m.SetHeader("Cc", req.CcEmails...)
	}

	// 密送
	if req.BccEmails != nil && len(req.BccEmails) > 0 {
		m.SetHeader("Bcc", req.BccEmails...)
	}

	// 设置标题
	if req.Title == "" {
		req.Title = e.Config.CommonTitle
	}
	m.SetHeader("Subject", req.Title)
	e.Result.Title = req.Title

	// 设置请求体
	if req.Body == "" {
		req.Body = "<h1>测试邮件内容</h1><div>验证码：<span style='color:red;'>" + e.Result.Key + "</span></div>"
	}
	if req.Type == "text" {
		m.SetBody("text/plain", req.Body)
	} else {
		m.SetBody("text/html", req.Body)
	}
	e.Result.Body = req.Body

	// 设置附件
	if req.Attachments != nil && len(req.Attachments) > 0 {
		for _, file := range req.Attachments {
			if req.IsFs {
				m.AttachWithFs(req.Fs, file)
			} else if req.IsFiles {
				m.AttachWithReaders(req.Files, file)
			} else {
				_, err := os.Stat(file) // 判断文件是否存在
				if err != nil {
					e.Log.Error("添加附件失败", "error", err, "file", file)
				} else {
					m.Attach(file)
				}
			}
		}
	}

	// 返回消息对象
	return m, nil
}

// SendGoMail 使用发送邮件
// @param req 发送邮件请求对象
// @return 错误信息
func (e *Email) SendGoMail(req EmailRequest) error {
	// 创建消息对象
	m, err := e.GetMessage(req)
	if err != nil {
		e.Log.Error("获取发送消息对象失败", "error", err)
		return err
	}

	// 发送邮件
	err = e.SendEmailWithMessage(m)
	e.Result.EndTime = int(time.Now().Unix())
	if err != nil {
		e.Log.Error("发送邮件失败", "error", err)
	}
	e.Result.SendStatus = true

	return nil
}

// SendEmailWithMessage 使用消息对象发送邮件
// @param message 消息对象
func (e *Email) SendEmailWithMessage(message *gomail.Message) error {
	// 获取邮件发送器
	sender, err := e.GetSender()
	if err != nil {
		e.Log.Error("获取邮件发送器失败", "error", err)
		return err
	}
	defer func(sender gomail.SendCloser) {
		err = sender.Close()
		if err != nil {
			e.Log.Error("关闭邮件发送器失败", "error", err)
		}
	}(sender)

	// 发送邮件
	err = gomail.Send(sender, message)
	if err != nil {
		e.Log.Error("发送邮件失败", "error", err)
	}

	return nil
}

// SendMany 批量发送HTML文本文件
// @param reqList 邮件发送请求对象列表
// @return 发送结果列表，错误信息
func (e *Email) SendMany(reqList []EmailRequest) ([]EmailResult, error) {
	var results []EmailResult
	for _, req := range reqList {
		result, err := e.Send(req)
		if err != nil {
			e.Log.Error("发送邮件失败", "error", err, "req", req)
			return nil, err
		}
		results = append(results, result)
		if e.Config.SendSleepMillisecond > 0 {
			time.Sleep(time.Duration(e.Config.SendSleepMillisecond) * time.Millisecond)
		}
	}
	return results, nil
}

// Send 发送邮件
// @param req 请求对象
// @return 发送结果，错误信息
func (e *Email) Send(req EmailRequest) (EmailResult, error) {
	// 邮件类型
	if req.Type == "" {
		req.Type = "html"
	}

	// 创建响应对象
	e.Result = &EmailResult{
		Type:        req.Type,
		From:        e.Config.Email,
		ToEmails:    req.ToEmails,
		CcEmails:    req.CcEmails,
		BccEmails:   req.BccEmails,
		Attachments: req.Attachments,
		StartTime:   int(time.Now().Unix()),
	}
	if e.Result.Key == "" {
		e.Result.Key = e.Random.Str(32)
	}

	// 无法连接邮件服务器
	if !e.IsHealth() {
		return *e.Result, errors.New("邮件服务器无法正常连接")
	}

	// 发送邮件
	key := e.Random.Str(16)
	e.Result.Key = key
	err := e.SendWithTag(e.Config.HeaderTagName, key, req)
	e.Result.EndTime = int(time.Now().Unix())

	// 校验是否成功
	if err != nil {
		e.Log.Error("发送邮件失败", "error", err)
	} else {
		e.Result.SendStatus = true
		e.Log.Debug("发送邮件成功")
	}

	// 返回结果
	return *e.Result, nil
}
