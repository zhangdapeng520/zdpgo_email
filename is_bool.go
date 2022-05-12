package zdpgo_email

/*
@Time : 2022/5/12 9:23
@Author : 张大鹏
@File : is_bool.go
@Software: Goland2021.3.1
@Description: is类型判断方法
*/

// IsSendSuccessByKeyValue 根据携带的Key和Value判断是否发送成功
func (e *Email) IsSendSuccessByKeyValue(from string, startTime string, tagKey, tagValue string) bool {
	// 构造查询条件
	if tagKey == "" {
		tagKey = e.Config.HeaderTagName
	}
	if tagValue == "" {
		tagValue = e.Config.HeaderTagValue
	}
	bf := PreFilter{
		From:           from,
		SentSince:      startTime,
		HeaderTagName:  tagKey,
		HeaderTagValue: tagValue,
	}
	af := PostFilter{}

	// 搜索
	results, err := e.Receive.SearchBF(&bf, &af)
	if err != nil {
		e.Log.Error("搜索邮件失败", "error", err, "filter", bf)
		return false
	}

	// 返回查询结果
	return results != nil && len(results) > 0
}
