package zdpgo_email

import (
	"fmt"
	"testing"
)

func prepareZdpEmail() *ZdpEmail {
	emailConfig := ZdpEmailConfig{
		Debug:             true,
		SendName:          "张大鹏",
		SendEmail:         "1156956636@qq.com",
		SendEmailPassword: "oxhcacwebqllhiaf",
		EmailSmtpHost:     "smtp.qq.com",
		EmailSmtpPort:     25,
	}
	email := New(emailConfig)
	return email
}

// 测试新建邮件
func TestZdpEmail_NewEmail(t *testing.T) {
	email := prepareZdpEmail()
	fmt.Println(email)
	fmt.Println(email.config)
	fmt.Println(email.config.SendName)
}

// 测试是否为debug模式
func TestZdpEmail_IsDebug(t *testing.T) {
	email := prepareZdpEmail()
	fmt.Println(email.IsDebug())
}

// 测试设置debug
func TestZdpEmail_SetDebug(t *testing.T) {
	email := prepareZdpEmail()
	email.SetDebug(false)
	fmt.Println(email.IsDebug())
}

// 测试发送邮件
func TestZdpEmail_SendText(t *testing.T) {
	email := prepareZdpEmail()
	email.SendText("这是一封测试邮件", "我在用Golang发邮件。。。", "lxgzhw@163.com")
}

// 测试发送HTML邮件
func TestZdpEmail_SendHtml(t *testing.T) {
	email := prepareZdpEmail()
	email.SendHtml("这是一封测试邮件", "<h1>这是一封HTML邮件</h1><div style='color:red'>疯狂的内容</div>", "lxgzhw@163.com")
}

// 测试发送HTML邮件携带附件
func TestZdpEmail_SendHtmlAndAttach(t *testing.T) {
	email := prepareZdpEmail()
	email.SendHtmlAndAttach("这是一封测试邮件", "<h1>这是一封HTML邮件</h1><div style='color:red'>疯狂的内容</div>", "zdpgo_email.log", "lxgzhw@163.com")
}

// 测试发送文本邮件携带附件
func TestZdpEmail_SendTextAndAttach(t *testing.T) {
	email := prepareZdpEmail()
	email.SendTextAndAttach("这是一封测试邮件", "<h1>这是一封HTML邮件</h1><div style='color:red'>疯狂的内容</div>", "zdpgo_email.log", "lxgzhw@163.com")
}
