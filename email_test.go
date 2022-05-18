package zdpgo_email

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_yaml"
	"testing"
)

/*
@Time : 2022/5/11 11:08
@Author : 张大鹏
@File : email_test.go
@Software: Goland2021.3.1
@Description: 测试email的相关功能
*/

func getEmail() *Email {
	Yaml := zdpgo_yaml.New()
	var smtp ConfigSmtp
	Yaml.ReadConfig("config/config_smtp.yaml", &smtp)
	Yaml.ReadConfig("config/secret/config_smtp.yaml", &smtp)

	// 正常情况
	e, err := NewWithConfig(Config{
		Debug: true,
		Smtp: ConfigEmail{
			Email:    smtp.Email,
			Username: smtp.Username,
			Password: smtp.Password,
			Host:     smtp.SmtpHost,
			Port:     smtp.SmtpPort,
			IsSSL:    smtp.IsSSL,
		},
	})

	if err != nil {
		fmt.Println("获取邮件失败", "error", err)
		return nil
	}

	return e
}

// 测试获取发送对象功能
func TestEmail_GetSender(t *testing.T) {
	// 创建邮件对象
	e, err := New()
	if err != nil {
		fmt.Println(err)
	}

	// 获取发送器
	// 异常情况，配置都为空，无法正常拨号
	sender, err := e.GetSender()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("发送器1：", sender)

	// 正常情况
	e = getEmail()
	sender, err = e.GetSender()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("发送器2：", sender)
}

// 测试健康情况
func TestEmail_IsHealth(t *testing.T) {
	// 异常情况：空对象
	e, _ := New()
	fmt.Println(e.IsHealth())

	// 读取配置
	e = getEmail()
	fmt.Println(e.IsHealth())
}
