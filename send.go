package zdpgo_email

import (
	"errors"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email/gomail"
	"mime"
	"net/smtp"
	"os"
	"strings"
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
func (e *EmailSmtp) SendEmail(title, content, attach string, isHtml bool, emails []string, ccEmails []string,
	bccEmails []string) error {
	// 校验邮箱
	if emails == nil {
		return errors.New("收件人邮箱不能为空")
	}

	// 设置 sender 发送方 的邮箱 ， 此处可以填写自己的邮箱
	e.From = fmt.Sprintf("%s <%s>", e.Config.Username, e.Config.Email)

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
	if attach != "" {
		_, err := e.AttachFile(attach)
		if err != nil {
			return err
		}
	}

	//设置服务器相关的配置
	addr := fmt.Sprintf("%s:%d", e.Config.SmtpHost, e.Config.SmtpPort)
	err := e.Send(addr, smtp.PlainAuth(
		e.Config.Id,
		e.Config.Email,
		e.Config.Password,
		e.Config.SmtpHost))
	if err != nil {
		return err
	}
	return nil
}

// SendText 发送文本邮件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
func (e *EmailSmtp) SendText(title, content string, emails ...string) {
	e.SendEmail(title, content, "", false, emails, nil, nil)
}

// SendHtml 发送HTML邮件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
func (e *EmailSmtp) SendHtml(title, content string, emails ...string) {
	e.SendEmail(title, content, "", true, emails, nil, nil)
}

// SendHtmlAndAttach 发送HTML邮件且能够携带附件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
// @param attach 附件
func (e *EmailSmtp) SendHtmlAndAttach(title, content, attach string, emails ...string) {
	e.SendEmail(title, content, attach, true, emails, nil, nil)
}

// SendTextAndAttach 发送文本文件，且能够携带附件
// @param toEmail 发送给哪个邮箱，也就是收件人
// @param title 邮件标题
// @param content 邮件内容
// @param attach 附件
func (e *EmailSmtp) SendTextAndAttach(title, content, attach string, emails ...string) {
	e.SendEmail(title, content, attach, false, emails, nil, nil)
}

// SendWithTag 通过标签发送邮件
func (e *EmailSmtp) SendWithTag(tagKey, tagValue, emailTitle string, emailBody string, emailAttachments []string,
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
	err := e.SendGoMail(emailTitle, emailBody, emailAttachments, toEmails...)
	if err != nil {
		return err
	}
	return nil
}

func (e *EmailSmtp) SendWithDefaultTag(emailTitle string, emailBody string, emailAttachments []string,
	toEmails ...string) error {
	err := e.SendWithTag("", "", emailTitle, emailBody, emailAttachments, toEmails...)
	if err != nil {
		return err
	}
	return nil
}

// SendWithKey 生成一个随机的key作为邮件的标识进行发送
func (e *EmailSmtp) SendWithKey(emailTitle string, emailBody string, emailAttachments []string,
	toEmails ...string) (string, error) {
	key := e.random.Str.Str(32)
	err := e.SendWithTag("", key, emailTitle, emailBody, emailAttachments, toEmails...)
	if err != nil {
		return "", err
	}
	return key, nil
}

// SendGoMail 使用gomail发送邮件
// @param emailTitle 邮件标题
// @param emailBody 邮件内容
// @param emailAttachments 邮件附件
// @param toEmails 收件人邮箱
// @return err 异常信息
func (e *EmailSmtp) SendGoMail(emailTitle string, emailBody string, emailAttachments []string,
	toEmails ...string) (err error) {
	m := gomail.NewMessage()

	// 设置邮件内容
	m.SetHeader(e.Config.HeaderTagName, e.Config.HeaderTagValue)
	m.SetHeader("From", e.Config.Email)
	m.SetHeader("To", toEmails...)
	m.SetHeader("Subject", emailTitle)
	m.SetBody("text/html", emailBody)
	for _, file := range emailAttachments {
		_, err = os.Stat(file)
		if err != nil {
			return
		} else {
			m.Attach(file)
		}
	}

	// 发送邮件
	c, err := e.GetGoMailSendCloser()
	defer c.Close()
	if err != nil {
		return
	}
	err = gomail.Send(c, m)
	return
}

func (e *EmailSmtp) sendGoMail1(mailTo []string, subject string, body string) error {
	// 设置邮箱主体
	mailConn := map[string]string{
		"user": e.Config.Email,    //发送人邮箱（邮箱以自己的为准）
		"pass": e.Config.Password, //发送人邮箱的密码，现在可能会需要邮箱 开启授权密码后在pass填写授权码
		"host": e.Config.SmtpHost, //邮箱服务器（此时用的是qq邮箱）
	}
	fmt.Println("密码", "pfqwwxvltpcshehh", e.Config.Password, "mailConn", mailConn)

	m := gomail.NewMessage(
		// 发送文本时设置编码，防止乱码。 如果txt文本设置了之后还是乱码，那可以将原txt文本在保存时就选择utf-8格式保存
		gomail.SetEncoding(gomail.Base64),
	)
	m.SetHeader("From", m.FormatAddress(mailConn["user"], "zdpgo_email")) // 添加别名
	m.SetHeader("To", mailTo...)                                          // 发送给用户(可以多个)
	m.SetHeader("Subject", subject)                                       // 设置邮件主题
	m.SetBody("text/html", body)                                          // 设置邮件正文

	// 一个文件（加入发送一个 txt 文件）：/tmp/foo.txt，我需要将这个文件以邮件附件的方式进行发送，同时指定附件名为：附件.txt
	//同时解决了文件名乱码问题
	name := "go.mod"
	m.Attach(name,
		gomail.Rename(name), //重命名
		gomail.SetHeader(map[string][]string{
			"Content-Disposition": []string{
				fmt.Sprintf(`attachment; filename="%s"`, mime.QEncoding.Encode("UTF-8", name)),
			},
		}),
	)

	/*
	   创建SMTP客户端，连接到远程的邮件服务器，需要指定服务器地址、端口号、用户名、密码，如果端口号为465的话，
	   自动开启SSL，这个时候需要指定TLSConfig
	*/
	d := gomail.NewDialer(mailConn["host"], 465, mailConn["user"], mailConn["pass"]) // 设置邮件正文
	//d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	err := d.DialAndSend(m)
	return err
}
