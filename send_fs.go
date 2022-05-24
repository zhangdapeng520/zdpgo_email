package zdpgo_email

import (
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
func (e *Email) SendFsAttachmentsMany(reqList []EmailRequest) (results []EmailResult, err error) {

	// 遍历附件
	for _, req := range reqList {
		if req.Title == "" {
			req.Title = e.Config.CommonTitle
		}
		if req.Body == "" {
			req.Body = "<h1>测试用的随机字符串</h1><br/>" + e.Random.Str(128)
		}
		result := EmailResult{
			Attachments: req.Attachments,
			Title:       req.Title,
			Body:        req.Body,
			From:        e.Config.Smtp.Email,
			To:          req.ToEmails,
		}
		e.Log.Debug("正在发送邮件", "file", req.Attachments)

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
		err = e.Send.SendWithTagAndFs(e.Config.HeaderTagName, key, req)

		// 校验是否成功
		if err != nil {
			e.Log.Error("发送邮件失败", "error", err)
			return
		} else {
			e.Log.Debug("发送邮件成功")
		}
		results = append(results, result)

		time.Sleep(time.Minute) // 一分钟一次，防止太快
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
func (e *Email) SendFsAttachmentsManyAndCheckResult(reqList []EmailRequest) (results []EmailResult, err error) {

	// 批量发送邮件
	sendFsAttachmentsMany, err := e.SendFsAttachmentsMany(reqList)
	if err != nil {
		e.Log.Error("批量发送邮件失败", "error", err)
		return
	}
	e.Log.Debug("批量发送邮件成功", "results", sendFsAttachmentsMany)

	// 验证是否发送成功
	var newResults = e.CheckResults(sendFsAttachmentsMany)
	return newResults, nil
}
