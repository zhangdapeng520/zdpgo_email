package zdpgo_email

import "time"

/*
@Time : 2022/5/12 10:44
@Author : 张大鹏
@File : check.go
@Software: Goland2021.3.1
@Description: 检查相关的方法
*/

// CheckResults 检查邮件发送结果列表
// @param results 原本的结果列表
// @param newResults 检查后的结果列表，主要修改了status确认是否发送成功
func (e *Email) CheckResults(results []EmailResult) (newResults []EmailResult) {
	nowTime := time.Now().Format("2006-01-02")
	e.Log.Debug("当前时间", "nowTime", nowTime)

	for _, result := range results {
		preFilter := PreFilter{
			From:           result.From,
			SentSince:      nowTime,
			HeaderTagName:  e.Config.HeaderTagName,
			HeaderTagValue: result.Key,
		}
		status := e.IsSendSuccessByKeyValue(preFilter.From, preFilter.SentSince, preFilter.HeaderTagName, preFilter.HeaderTagValue)
		if status {
			e.Log.Debug("邮件发送成功", "key", preFilter.HeaderTagValue)
		} else {
			e.Log.Debug("邮件发送失败", "key", preFilter.HeaderTagValue)
		}
		result.ReceiveStatus = status
		newResults = append(newResults, result)
		time.Sleep(time.Minute) // 一分钟一次，防止太快
	}
	e.Log.Debug("结果校验成功", "results", newResults)
	return
}
