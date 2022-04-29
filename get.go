package zdpgo_email

import (
	"github.com/zhangdapeng520/zdpgo_email/gomail"
)

/*
@Time : 2022/4/29 10:06
@Author : 张大鹏
@File : get.go
@Software: Goland2021.3.1
@Description:获取对象相关的方法
*/

// GetGoMailSendCloser 获取gomail对象
// @param host 服务器地址
// @param port 服务器端口
// @param username 用户名
// @param password 密码
func (e *EmailSmtp) GetGoMailSendCloser() (gomail.SendCloser, error) {
	d := &gomail.Dialer{
		Host:     e.Config.SmtpHost,
		Port:     e.Config.SmtpPort,
		Username: e.Config.Email,
		Password: e.Config.Password,
		SSL:      e.Config.IsSSL,
	}
	c, err := d.Dial()
	if err != nil {
		return nil, err
	}
	return c, nil
}
