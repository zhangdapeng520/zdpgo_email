package gomail

import "github.com/zhangdapeng520/zdpgo_log"

/*
@Time : 2022/5/18 9:24
@Author : 张大鹏
@File : logger.go
@Software: Goland2021.3.1
@Description: logger日志相关
*/

var Log *zdpgo_log.Log // 日志对象

func init() {
	if Log == nil {
		Log = zdpgo_log.NewWithDebug(true, "logs/zdpgo/zdpgo_email.log")
	}
}
