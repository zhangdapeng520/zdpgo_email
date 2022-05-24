package zdpgo_email

import (
	"errors"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email/gomail"
	"net/smtp"
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
func (e *EmailSmtp) SendWithTag(tagKey, tagValue string, req EmailRequest) error {

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
	err := e.SendGoMail(req)
	if err != nil {
		e.Log.Error("发送邮件失败", "error", err)
		return err
	}
	return nil
}

// SendWithTagAndFs 使用标签和嵌入文件系统发送邮件
func (e *EmailSmtp) SendWithTagAndFs(tagKey, tagValue string, req EmailRequest) error {
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
	err := e.SendGoMailWithFs(req)
	if err != nil {
		return err
	}
	return nil
}

// SendWithTagAndReaders 使用标签和读取器列表发送邮件
func (e *EmailSmtp) SendWithTagAndReaders(tagKey, tagValue string, req EmailRequest) error {
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
	err := e.SendGoMailWithReaders(req)
	if err != nil {
		e.Log.Error("SendWithTagAndReaders 使用标签和读取器列表发送邮件失败", "error", err)
		return err
	}
	return nil
}

// SendWithTagAndFiles 使用标签和文件列表发送邮件
func (e *EmailSmtp) SendWithTagAndFiles(tagKey, tagValue string, req EmailRequest) error {
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
	err := e.SendGoMailWithFiles(req)
	if err != nil {
		e.Log.Error("使用标签和文件列表发送邮件失败", "error", err)
		return err
	}
	return nil
}

// SendWithDefaultTag 使用默认的tag和value发送邮件
func (e *EmailSmtp) SendWithDefaultTag(req EmailRequest) error {
	err := e.SendWithTag("", "", req)
	if err != nil {
		return err
	}
	return nil
}

// SendWithDefaultTagWithFs 使用默认标签和嵌入文件系统发送邮件
func (e *EmailSmtp) SendWithDefaultTagWithFs(req EmailRequest) error {
	err := e.SendWithTagAndFs("", "", req)
	if err != nil {
		e.Log.Error("SendWithDefaultTagWithFs 使用默认标签和嵌入文件系统发送邮件失败", "error", err)
		return err
	}
	return nil
}

// SendWithDefaultTagWithReaders 使用默认标签和读取器发送邮件
func (e *EmailSmtp) SendWithDefaultTagWithReaders(req EmailRequest) error {
	err := e.SendWithTagAndReaders("", "", req)
	if err != nil {
		return err
	}
	return nil
}

// SendWithDefaultTagWithFiles 使用默认标签和文件列表发送文件
func (e *EmailSmtp) SendWithDefaultTagWithFiles(req EmailRequest) error {
	e.Log.Debug("使用默认标签和文件列表发送文件")
	err := e.SendWithTagAndFiles("", "", req)
	if err != nil {
		e.Log.Error("使用默认标签和文件列表发送文件失败", "error", err)
		return err
	}
	return nil
}

// SendWithKey 生成一个随机的key作为邮件的标识进行发送
func (e *EmailSmtp) SendWithKey(req EmailRequest) (string, error) {
	key := e.random.Str(32)
	err := e.SendWithTag("", key, req)
	if err != nil {
		return "", err
	}
	return key, nil
}

// GetMessage 获取HTML类型的消息对象
// @param req 邮件请求对象
func (e *EmailSmtp) GetMessage(req EmailRequest) (*gomail.Message, error) {
	// 创建消息对象
	m := gomail.NewMessage()

	// 设置请求头
	m.SetHeader(e.Config.HeaderTagName, e.Config.HeaderTagValue)

	// 发件人
	m.SetHeader("From", e.Config.Smtp.Email)

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

	// 设置请求体
	if req.Type == "text" {
		m.SetBody("text/plain", req.Body)
	} else {
		m.SetBody("text/html", req.Body)
	}

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
					Log.Error("添加附件失败", "error", err, "file", file)
				} else {
					m.Attach(file)
				}
			}
		}
	}

	// 返回消息对象
	return m, nil
}

// SendGoMail 使用gomail发送邮件
// @param req 发送邮件请求对象
// @return 错误信息
func (e *EmailSmtp) SendGoMail(req EmailRequest) error {

	// 创建消息对象
	m, err := e.GetMessage(req)
	if err != nil {
		e.Log.Error("获取发送消息对象失败", "error", err)
		return err
	}

	// 发送邮件
	err = e.GetSenderAndSendEmail(m)
	if err != nil {
		Log.Error("发送邮件失败", "error", err)
	}

	return nil
}

// SendGoMailWithFs 使用嵌入文件系统发送邮件
func (e *EmailSmtp) SendGoMailWithFs(req EmailRequest) error {
	m, err := e.GetMessage(req)
	if err != nil {
		e.Log.Error("获取邮件发送消息失败", "error", err)
		return err
	}

	// 发送邮件
	err = e.GetSenderAndSendEmail(m)
	if err != nil {
		Log.Error("发送邮件失败", "error", err)
		return err
	}

	return nil
}

// SendGoMailWithReaders 使用读取器列表发送邮件
func (e *EmailSmtp) SendGoMailWithReaders(req EmailRequest) error {
	m, err := e.GetMessage(req)
	if err != nil {
		e.Log.Error("获取发送消息对象失败", "error", err)
		return err
	}

	// 发送邮件
	err = e.GetSenderAndSendEmail(m)
	if err != nil {
		Log.Error("发送邮件失败", "error", err)
		return err
	}

	return nil
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
func (e *EmailSmtp) SendGoMailWithFiles(req EmailRequest) error {
	e.Log.Debug("SendGoMailWithFiles 使用gmail和文件列表发送文件")

	// 创建消息对象
	m, err := e.GetMessage(req)
	if err != nil {
		e.Log.Error("获取发送消息对象失败", "error", err)
		return err
	}

	// 发送邮件
	err = e.GetSenderAndSendEmail(m)
	if err != nil {
		Log.Error("发送邮件失败", "error", err)
		return err
	}

	return nil
}

// SendHtmlManyAndCheckResult 批量发送HTML模板邮件并校验结果
// @param contents 内容列表
// @param internalSeconds 发送没封邮件的间隔时间，防止发送过快
// @return results 校验结果列表
// @return err 错误信息
func (e *Email) SendHtmlManyAndCheckResult(reqList []EmailRequest) (results []EmailResult, err error) {

	// 批量发送邮件
	sendFsAttachmentsMany, err := e.SendMany(reqList)
	if err != nil {
		e.Log.Error("批量发送邮件失败", "error", err)
		return
	}
	e.Log.Debug("批量发送邮件成功", "results", sendFsAttachmentsMany)

	// 验证是否发送成功
	var newResults = e.CheckResults(sendFsAttachmentsMany)
	return newResults, nil
}

// SendMany 批量发送HTML文本文件
// @param reqList 邮件发送请求对象列表
// @return 发送结果列表，错误信息
func (e *Email) SendMany(reqList []EmailRequest) ([]EmailResult, error) {
	var results []EmailResult
	for _, req := range reqList {
		result, err := e.SendHtml(req)
		if err != nil {
			e.Log.Error("发送HTML邮件失败", "error", err, "req", req)
			return nil, err
		}
		results = append(results, result)
		if e.Config.SendSleepMillisecond > 0 {
			time.Sleep(time.Duration(e.Config.SendSleepMillisecond) * time.Millisecond)
		}
	}
	return results, nil
}

// GetEmailResult 获取邮件响应结果对象
func (e *Email) GetEmailResult(req EmailRequest) EmailResult {
	if req.Title == "" {
		req.Title = e.Config.CommonTitle
	}
	if req.Body == "" {
		req.Body = "<h1>测试用的随机字符串</h1><br/>" + e.Random.Str(32)
	}
	var result = EmailResult{
		Key:           "",
		Title:         req.Title,
		Body:          req.Body,
		From:          e.Config.Smtp.Email,
		To:            req.ToEmails,
		Attachments:   req.Attachments,
		StartTime:     int(time.Now().Unix()),
		EndTime:       0,
		SendStatus:    false,
		ReceiveStatus: false,
	}
	return result
}

// SendHtml 发送HTML类型的邮件
// @param req 请求对象
// @return 发送结果，错误信息
func (e *Email) SendHtml(req EmailRequest) (EmailResult, error) {
	req.Type = "html"
	result := e.GetEmailResult(req)

	// 无法连接邮件服务器
	if !e.IsHealth() {
		return result, errors.New("邮件服务器无法正常连接")
	}

	// 发送邮件
	key := e.Random.Str(16)
	result.Key = key
	err := e.Send.SendWithTag(e.Config.HeaderTagName, key, req)
	result.EndTime = int(time.Now().Unix())

	// 校验是否成功
	if err != nil {
		e.Log.Error("发送邮件失败", "error", err)
	} else {
		result.SendStatus = true
		e.Log.Debug("发送邮件成功")
	}

	// 返回结果
	return result, nil
}
