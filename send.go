package zdpgo_email

import (
	"embed"
	"errors"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email/gomail"
	"net/smtp"
	"os"
	"path/filepath"
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

// SendEmail 封装发送文本邮件的方法
// @param title 邮件标题
// @param content 邮件内容
// @param attach 附件
// @param isHtml 是否为HTML内容
// @param emails 普通收件人地址列表
// @param ccEmails 抄送人邮箱地址列表
// @param bccEmails 密送人邮箱地址列表
func (e *EmailSmtp) SendEmail(
	title string,
	content string,
	attachments []string,
	isHtml bool,
	emails []string,
	ccEmails []string,
	bccEmails []string) error {
	// 校验邮箱
	if emails == nil {
		msg := "收件人邮箱不能为空"
		Log.Error(msg)
		return errors.New(msg)
	}

	// 设置 sender 发送方 的邮箱 ， 此处可以填写自己的邮箱
	e.From = fmt.Sprintf("%s <%s>", e.Config.Smtp.Username, e.Config.Smtp.Email)

	// 设置 receiver 接收方 的邮箱  此处也可以填写自己的邮箱， 就是自己发邮件给自己
	e.To = emails

	// 抄送
	if ccEmails != nil {
		e.Cc = ccEmails
	}

	// 密送
	if bccEmails != nil {
		e.Bcc = bccEmails
	}

	// 设置主题,也就是邮件的标题
	e.Subject = title

	// 邮件内容，是HTML或者纯文本
	if isHtml {
		e.HTML = []byte(content)
	} else {
		e.Text = []byte(content)
	}

	// 附件
	if attachments != nil && len(attachments) > 0 {
		for _, attach := range attachments {
			_, err := e.AttachFile(attach)
			if err != nil {
				Log.Error("添加附件失败", "error", err, "file", attach)
				return err
			}
		}
	}

	//设置服务器相关的配置
	addr := fmt.Sprintf("%s:%d", e.Config.Smtp.Host, e.Config.Smtp.Port)
	err := e.Send(addr, smtp.PlainAuth(
		e.Config.Id,
		e.Config.Smtp.Email,
		e.Config.Smtp.Password,
		e.Config.Smtp.Host))
	if err != nil {
		Log.Error("发送邮件失败", "error", err)
		return err
	}
	return nil
}

// SendText 发送文本邮件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
func (e *EmailSmtp) SendText(title, content string, emails ...string) {
	err := e.SendEmail(title, content, nil, false, emails, nil, nil)
	if err != nil {
		Log.Error("发送文本类型的邮件失败", "error", err)
	}
}

// SendHtml 发送HTML邮件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
func (e *EmailSmtp) SendHtml(title, content string, emails ...string) {
	err := e.SendEmail(title, content, nil, true, emails, nil, nil)
	if err != nil {
		Log.Error("发送HTML类型的邮件失败", "error", err)
	}
}

// SendHtmlAndAttach 发送HTML邮件且能够携带附件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
// @param attach 附件
func (e *EmailSmtp) SendHtmlAndAttach(title, content, attach string, emails ...string) {
	err := e.SendEmail(title, content, []string{attach}, true, emails, nil, nil)
	if err != nil {
		Log.Error("发送HTML类型的邮件并添加附件失败", "error", err)
	}
}

// SendTextAndAttach 发送文本文件，且能够携带附件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
// @param attach 附件
func (e *EmailSmtp) SendTextAndAttach(title, content, attach string, emails ...string) {
	err := e.SendEmail(title, content, []string{attach}, false, emails, nil, nil)
	if err != nil {
		Log.Error("发送文本类型的邮件并添加附件失败", "error", err)
	}
}

// SendWithTag 通过标签发送邮件
func (e *EmailSmtp) SendWithTag(tagKey, tagValue, emailTitle string, emailBody string, emailAttachments []string,
	toEmails ...string) error {

	// 检查key是否符合规范
	if tagKey != "" {
		tagArr := strings.Split(tagKey, "-")
		if len(tagArr) != 3 {
			msg := "key必须由两个“-”分割的字符组成"
			Log.Error("key必须由两个“-”分割的字符组成")
			return errors.New(msg)
		} else if tagArr[0] != "X" {
			msg := "Key的第一个字符必须是大写的X"
			Log.Error(msg)
			return errors.New(msg)
		}
		e.Config.HeaderTagName = tagKey
	}

	// 检查value
	if tagValue != "" {
		e.Config.HeaderTagValue = tagValue
	}

	// 发送邮件
	err := e.SendGoMail(emailTitle, emailBody, emailAttachments, toEmails...)
	if err != nil {
		return err
	}
	return nil
}

// SendWithTagAndFs 使用标签和嵌入文件系统发送邮件
func (e *EmailSmtp) SendWithTagAndFs(fs *embed.FS, tagKey, tagValue, emailTitle string, emailBody string,
	emailAttachments []string,
	toEmails ...string) error {
	if tagKey != "" {
		tagArr := strings.Split(tagKey, "-")
		if len(tagArr) != 3 {
			return errors.New("key必须由两个“-”分割的字符组成")
		} else if tagArr[0] != "X" {
			return errors.New("Key的第一个字符必须是大写的X")
		}
		e.Config.HeaderTagName = tagKey
	}
	if tagValue != "" {
		e.Config.HeaderTagValue = tagValue
	}
	err := e.SendGoMailWithFs(fs, emailTitle, emailBody, emailAttachments, toEmails...)
	if err != nil {
		return err
	}
	return nil
}

// SendWithTagAndReaders 使用标签和读取器列表发送邮件
func (e *EmailSmtp) SendWithTagAndReaders(readers map[string]*os.File, tagKey, tagValue, emailTitle string,
	emailBody string,
	emailAttachments []string,
	toEmails ...string) error {
	if tagKey != "" {
		tagArr := strings.Split(tagKey, "-")
		if len(tagArr) != 3 {
			return errors.New("key必须由两个“-”分割的字符组成")
		} else if tagArr[0] != "X" {
			return errors.New("Key的第一个字符必须是大写的X")
		}
		e.Config.HeaderTagName = tagKey
	}
	if tagValue != "" {
		e.Config.HeaderTagValue = tagValue
	}
	err := e.SendGoMailWithReaders(readers, emailTitle, emailBody, emailAttachments, toEmails...)
	if err != nil {
		e.Log.Error("SendWithTagAndReaders 使用标签和读取器列表发送邮件失败", "error", err)
		return err
	}
	return nil
}

// SendWithTagAndFiles 使用标签和文件列表发送邮件
func (e *EmailSmtp) SendWithTagAndFiles(files map[string]*os.File, tagKey, tagValue, emailTitle string,
	emailBody string, toEmails ...string) error {
	e.Log.Debug("SendWithTagAndFiles 使用标签和文件列表发送邮件")

	// 处理标签和校验Key
	if tagKey != "" {
		tagArr := strings.Split(tagKey, "-")
		if len(tagArr) != 3 {
			return errors.New("key必须由两个“-”分割的字符组成")
		} else if tagArr[0] != "X" {
			return errors.New("Key的第一个字符必须是大写的X")
		}
		e.Config.HeaderTagName = tagKey
	}
	if tagValue != "" {
		e.Config.HeaderTagValue = tagValue
	}

	// 发送邮件
	err := e.SendGoMailWithFiles(files, emailTitle, emailBody, toEmails...)
	if err != nil {
		e.Log.Error("使用标签和文件列表发送邮件失败", "error", err)
		return err
	}
	return nil
}

// SendWithDefaultTag 使用默认的tag和value发送邮件
func (e *EmailSmtp) SendWithDefaultTag(emailTitle string, emailBody string, emailAttachments []string,
	toEmails ...string) error {
	err := e.SendWithTag("", "", emailTitle, emailBody, emailAttachments, toEmails...)
	if err != nil {
		return err
	}
	return nil
}

// SendWithDefaultTagWithFs 使用默认标签和嵌入文件系统发送邮件
func (e *EmailSmtp) SendWithDefaultTagWithFs(fs *embed.FS, emailTitle string, emailBody string,
	emailAttachments []string,
	toEmails ...string) error {
	err := e.SendWithTagAndFs(fs, "", "", emailTitle, emailBody, emailAttachments, toEmails...)
	if err != nil {
		e.Log.Error("SendWithDefaultTagWithFs 使用默认标签和嵌入文件系统发送邮件失败", "error", err)
		return err
	}
	return nil
}

// SendWithDefaultTagWithReaders 使用默认标签和读取器发送邮件
func (e *EmailSmtp) SendWithDefaultTagWithReaders(readers map[string]*os.File, emailTitle string, emailBody string,
	emailAttachments []string,
	toEmails ...string) error {
	err := e.SendWithTagAndReaders(readers, "", "", emailTitle, emailBody, emailAttachments, toEmails...)
	if err != nil {
		return err
	}
	return nil
}

// SendWithDefaultTagWithFiles 使用默认标签和文件列表发送文件
func (e *EmailSmtp) SendWithDefaultTagWithFiles(files map[string]*os.File, emailTitle string, emailBody string, toEmails ...string) error {
	e.Log.Debug("使用默认标签和文件列表发送文件")
	err := e.SendWithTagAndFiles(files, "", "", emailTitle, emailBody, toEmails...)
	if err != nil {
		e.Log.Error("使用默认标签和文件列表发送文件失败", "error", err)
		return err
	}
	return nil
}

// SendWithKey 生成一个随机的key作为邮件的标识进行发送
func (e *EmailSmtp) SendWithKey(emailTitle string, emailBody string, emailAttachments []string,
	toEmails ...string) (string, error) {
	key := e.random.Str(32)
	err := e.SendWithTag("", key, emailTitle, emailBody, emailAttachments, toEmails...)
	if err != nil {
		return "", err
	}
	return key, nil
}

// GetHtmlMessage 获取HTML类型的消息对象
// @param emailTitle 邮件标题
// @param emailBody 邮件内容
// @param toEmails 收件人邮箱
func (e *EmailSmtp) GetHtmlMessage(
	emailTitle string,
	emailBody string,
	toEmails ...string) *gomail.Message {
	// 创建消息对象
	m := gomail.NewMessage()

	// 设置请求头
	m.SetHeader(e.Config.HeaderTagName, e.Config.HeaderTagValue)
	m.SetHeader("From", e.Config.Smtp.Email)
	m.SetHeader("To", toEmails...)
	m.SetHeader("Subject", emailTitle)

	// 设置请求体
	m.SetBody("text/html", emailBody)
	return m
}

// SendGoMail 使用gomail发送邮件
// @param emailTitle 邮件标题
// @param emailBody 邮件内容
// @param emailAttachments 邮件附件
// @param toEmails 收件人邮箱
// @return err 异常信息
func (e *EmailSmtp) SendGoMail(
	emailTitle string,
	emailBody string,
	emailAttachments []string,
	toEmails ...string) (err error) {

	// 创建消息对象
	m := e.GetHtmlMessage(emailTitle, emailBody, toEmails...)

	// 设置附件
	for _, file := range emailAttachments {
		_, err = os.Stat(file) // 判断文件是否存在
		if err != nil {
			Log.Error("添加附件失败", "error", err, "file", file)
			return
		} else {
			m.Attach(file)
		}
	}

	// 发送邮件
	err = e.GetSenderAndSendEmail(m)
	if err != nil {
		Log.Error("发送邮件失败", "error", err)
	}
	return
}

// SendGoMailWithFs 使用嵌入文件系统发送邮件
func (e *EmailSmtp) SendGoMailWithFs(fs *embed.FS, emailTitle string, emailBody string, emailAttachments []string,
	toEmails ...string) (err error) {
	m := e.GetHtmlMessage(emailTitle, emailBody, toEmails...)

	for _, file := range emailAttachments {
		m.AttachWithFs(fs, file)
	}

	// 发送邮件
	err = e.GetSenderAndSendEmail(m)
	if err != nil {
		Log.Error("发送邮件失败", "error", err)
	}
	return
}

// SendGoMailWithReaders 使用读取器列表发送邮件
func (e *EmailSmtp) SendGoMailWithReaders(readers map[string]*os.File, emailTitle string, emailBody string,
	emailAttachments []string,
	toEmails ...string) (err error) {
	m := e.GetHtmlMessage(emailTitle, emailBody, toEmails...)

	for _, file := range emailAttachments {
		m.AttachWithReaders(readers, file)
	}

	// 发送邮件
	err = e.GetSenderAndSendEmail(m)
	if err != nil {
		Log.Error("发送邮件失败", "error", err)
	}
	return
}

// GetSenderAndSendEmail 获取gomail邮件发送器然后发送邮件
// @param m 消息对象
func (e *EmailSmtp) GetSenderAndSendEmail(m *gomail.Message) error {
	// 获取邮件发送器
	c, err := e.GetGoMailSendCloser()
	defer func(c gomail.SendCloser) {
		err = c.Close()
		if err != nil {
			Log.Error("关闭邮件发送器失败", "error", err)
		}
	}(c)
	if err != nil {
		Log.Error("获取邮件发送器失败", "error", err)
		return err
	}

	// 发送邮件
	err = gomail.Send(c, m)
	if err != nil {
		Log.Error("发送邮件失败", "error", err)
	}

	return nil
}

// SendGoMailWithFiles 使用gmail和文件列表发送文件
func (e *EmailSmtp) SendGoMailWithFiles(files map[string]*os.File, emailTitle string, emailBody string, toEmails ...string) (err error) {
	e.Log.Debug("SendGoMailWithFiles 使用gmail和文件列表发送文件")

	// 创建消息对象
	m := e.GetHtmlMessage(emailTitle, emailBody, toEmails...)

	// 添加附件
	for fileName, fileObj := range files {
		m.AddAttachmentFileObj(fileName, fileObj)
	}

	// 发送邮件
	err = e.GetSenderAndSendEmail(m)
	if err != nil {
		Log.Error("发送邮件失败", "error", err)
	}
	return
}

// SendHtmlManyAndCheckResult 批量发送HTML模板邮件并校验结果
// @param contents 内容列表
// @param internalSeconds 发送没封邮件的间隔时间，防止发送过快
// @return results 校验结果列表
// @return err 错误信息
func (e *Email) SendHtmlManyAndCheckResult(
	contents []string,
	toEmails ...string,
) (results []EmailResult, err error) {

	// 批量发送邮件
	sendFsAttachmentsMany, err := e.SendHtmlMany(contents, toEmails...)
	if err != nil {
		e.Log.Error("批量发送邮件失败", "error", err)
		return
	}
	e.Log.Debug("批量发送邮件成功", "results", sendFsAttachmentsMany)

	// 验证是否发送成功
	var newResults = e.CheckResults(sendFsAttachmentsMany)
	return newResults, nil
}

// SendHtmlMany 批量发送HTML文本文件
// @param contents 内容列表
// @param toEmails 收件人邮箱
// @return results 发送结果
// @return err 错误信息
func (e *Email) SendHtmlMany(
	contents []string,
	toEmails ...string,
) (results []EmailResult, err error) {

	// 遍历邮件内容
	for _, emailBody := range contents {
		emailTitle := e.Config.CommonTitle
		result := EmailResult{
			Title:     emailTitle,
			Body:      emailBody,
			From:      e.Config.Smtp.Email,
			To:        toEmails,
			StartTime: int(time.Now().Unix()),
		}
		e.Log.Debug("正在发送邮件", "emailBody", emailBody)

		// 重连三次，判断邮件能够正常访问服务器
		for i := 0; i < 3; i++ {
			if e.IsHealth() {
				break
			} else {
				e.Log.Warning("无法正常连接到邮件服务器，正在尝试重连", "num", i+1)
				time.Sleep(time.Second * 3)
			}
		}

		// 发送邮件
		key := e.Random.Str(16)
		result.Key = key
		err = e.Send.SendWithTag(
			e.Config.HeaderTagName,
			key,
			emailTitle,
			emailBody,
			[]string{},
			toEmails...,
		)

		// 校验是否成功
		if err != nil {
			result.SendStatus = false
			e.Log.Error("发送邮件失败", "error", err)
		} else {
			result.SendStatus = true
			e.Log.Debug("发送邮件成功")
		}
		result.EndTime = int(time.Now().Unix())
		results = append(results, result)
		time.Sleep(time.Duration(e.Config.SendSleepSeconds) * time.Second) // 指定间隔时间，防止过快发送
	}
	return
}

// SendAttachmentMany 使用默认标签附件列表批量发送邮件
// @param attachments 附件列表
// @param emailTitle 邮件标题
// @param emailBody 邮件内容
// @param toEmails 收件人邮箱
// @return results 发送结果
// @return err 错误信息
func (e *Email) SendAttachmentMany(
	attachments []string,
	toEmails ...string,
) (results []EmailResult, err error) {
	// 遍历附件
	for _, file := range attachments {
		emailTitle := e.Config.CommonTitle
		emailBody := "<h1>测试用的随机字符串</h1><br/>" + e.Random.Str(128)
		result := EmailResult{
			AttachmentName: filepath.Base(file),
			AttachmentPath: file,
			Title:          emailTitle,
			Body:           emailBody,
			From:           e.Config.Smtp.Email,
			To:             toEmails,
		}
		e.Log.Debug("正在携带附件发送邮件", "title", emailTitle, "file", file)

		// 重连三次，判断邮件能够正常访问服务器
		for i := 0; i < 3; i++ {
			if e.IsHealth() {
				break
			} else {
				e.Log.Warning("无法正常连接到邮件服务器，正在尝试重连", "num", i+1)
				time.Sleep(time.Second * 3)
			}
		}

		// 发送邮件
		key := e.Random.Str(16)
		result.Key = key
		err = e.Send.SendWithTag(
			e.Config.HeaderTagName,
			key,
			emailTitle,
			emailBody,
			[]string{file},
			toEmails...,
		)

		// 校验是否成功
		if err != nil {
			e.Log.Error("发送邮件失败", "error", err)
			result.SendStatus = false
		} else {
			result.SendStatus = true
			e.Log.Debug("发送邮件成功")
		}
		results = append(results, result)
		time.Sleep(time.Duration(e.Config.SendSleepSeconds) * time.Second) // 休眠一下
	}
	return
}

// SendAttachments 发送附件，附件可以有多个
// @param attachments 附件列表
// @param emailTitle 邮件标题
// @param emailBody 邮件内容
// @param toEmails 收件人邮箱
// @return results 发送结果
// @return err 错误信息
func (e *Email) SendAttachments(emailTitle, emailBody string, attachments []string, toEmails ...string) (
	EmailResult, error) {
	result := EmailResult{}

	// 无法正常连接服务器
	if !e.IsHealth() {
		return result, errors.New("无法正常连接邮件服务器")
	}

	// 发送邮件
	if emailTitle == "" {
		emailTitle = e.Config.CommonTitle
	}
	if emailBody == "" {
		emailBody = "<h1>测试用的随机字符串</h1><br/>" + e.Random.Str(128)
	}
	key := e.Random.Str(16)
	err := e.Send.SendWithTag(
		e.Config.HeaderTagName,
		key,
		emailTitle,
		emailBody,
		attachments,
		toEmails...,
	)

	// 准备响应结果
	result = EmailResult{
		Attachments: attachments,
		Title:       emailTitle,
		Body:        emailBody,
		From:        e.Config.Smtp.Email,
		To:          toEmails,
	}

	// 校验是否成功
	if err != nil {
		e.Log.Error("发送邮件失败", "error", err)
		result.SendStatus = false
	} else {
		result.SendStatus = true
		e.Log.Debug("发送邮件成功")
	}

	// 返回发送结果
	return result, nil
}
