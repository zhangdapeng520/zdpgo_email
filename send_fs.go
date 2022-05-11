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
	emailTitle string,
	emailBody string,
	toEmails ...string,
) (results []EmailResult, err error) {

	// 遍历附件
	for _, file := range attachments {
		result := EmailResult{AttachmentName: filepath.Base(file)}
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
		result.Title = emailTitle
		result.From = e.Send.From
		result.To = toEmails
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
