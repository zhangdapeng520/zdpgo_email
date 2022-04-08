package zdpgo_email

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_yaml"
	"testing"
)

func prepareEmail() *Email {
	y := zdpgo_yaml.New()
	var config Config
	err := y.ReadDefaultConfig(&config)
	if err != nil {
		panic(err)
	}
	fmt.Println("读取到的配置是：", config)
	email := New(config)
	return email
}

// 测试新建邮件
func TestEmail_NewEmail(t *testing.T) {
	email := prepareEmail()
	fmt.Println(email)
	fmt.Println(email.config)
	fmt.Println(email.config.Username)
}

// 测试发送邮件
func TestEmail_SendText(t *testing.T) {
	email := prepareEmail()
	email.SendText("这是一封测试邮件", "我在用Golang发邮件。。。", "lxgzhw@163.com")
}

// 测试发送HTML邮件
func TestEmail_SendHtml(t *testing.T) {
	email := prepareEmail()
	email.SendHtml("这是一封测试邮件", "<h1>这是一封HTML邮件</h1><div style='color:red'>疯狂的内容</div>", "lxgzhw@163.com")
}

// 测试发送HTML邮件携带附件
func TestEmail_SendHtmlAndAttach(t *testing.T) {
	email := prepareEmail()
	email.SendHtmlAndAttach("这是一封测试邮件", "<h1>这是一封HTML邮件</h1><div style='color:red'>疯狂的内容</div>", "zdpgo_email.log", "lxgzhw@163.com")
}

// 测试发送文本邮件携带附件
func TestEmail_SendTextAndAttach(t *testing.T) {
	email := prepareEmail()
	email.SendTextAndAttach("这是一封测试邮件", "<h1>这是一封HTML邮件</h1><div style='color:red'>疯狂的内容</div>", "zdpgo_email.log", "lxgzhw@163.com")
}
