package zdpgo_email

import (
	"embed"
	"path/filepath"
	"time"
)

/*
@Time : 2022/5/11 20:12
@Author : 张大鹏
@File : send_fs.go
@Software: Goland2021.3.1
@Description: 使用embed嵌入系统作为附件发送邮件相关代码
*/

// SendFsAttachmentsMany 使用默认标签和嵌入文件系统批量发送邮件
// @param fsObj 嵌入文件系统
// @param attachments 附件列表
// @param emailTitle 邮件标题
// @param emailBody 邮件内容
// @param toEmails 收件人邮箱
// @return results 发送结果
// @return err 错误信息
func (e *Email) SendFsAttachmentsMany(
	fsObj *embed.FS,
	attachments []string,
	toEmails ...string,
) (results []EmailResult, err error) {

	// 遍历附件
	for _, file := range attachments {
		emailTitle := e.Config.CommonTitle
		emailBody := "<h1>测试用的随机字符串</h1><br/>" + e.Random.Str.Str(128)
		result := EmailResult{
			AttachmentName: filepath.Base(file),
			AttachmentPath: file,
			Title:          emailTitle,
			Body:           emailBody,
			From:           e.Config.Smtp.Email,
			To:             toEmails,
		}
		e.Log.Debug("正在发送邮件", "file", file)

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
		key := e.Random.Str.Str(16)
		result.Key = key
		err = e.Send.SendWithTagAndFs(
			fsObj,
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
			return
		} else {
			e.Log.Debug("发送邮件成功")
		}
		results = append(results, result)

		// 休息10秒钟，防止频繁发送邮件被对方邮件服务器限制
		time.Sleep(10 * time.Second)
	}
	return
}

// SendFsAttachmentsManyAndCheckResult 使用默认标签和嵌入文件系统批量发送邮件并校验结果
// @param fsObj 嵌入文件系统
// @param attachments 附件列表
// @param emailTitle 邮件标题
// @param emailBody 邮件内容
// @param toEmails 收件人邮箱
// @return results 发送结果
// @return err 错误信息
func (e *Email) SendFsAttachmentsManyAndCheckResult(
	fsObj *embed.FS,
	attachments []string,
	toEmails ...string,
) (results []EmailResult, err error) {

	// 批量发送邮件
	sendFsAttachmentsMany, err := e.SendFsAttachmentsMany(fsObj, attachments, toEmails...)
	if err != nil {
		e.Log.Error("批量发送邮件失败", "error", err)
		return
	}
	e.Log.Debug("批量发送邮件成功", "results", sendFsAttachmentsMany)

	// 验证是否发送成功
	time.Sleep(time.Second * 60) // 一分钟以后校验是否发送成功
	var newResults []EmailResult
	nowTime := time.Now().Format("2006-01-02")
	e.Log.Debug("当前时间", "nowTime", nowTime)

	for _, result := range sendFsAttachmentsMany {
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
		result.Status = status
		newResults = append(newResults, result)
	}
	e.Log.Debug("结果校验成功", "results", newResults)

	return newResults, nil
}
