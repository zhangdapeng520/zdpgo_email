package zdpgo_email

import (
	"bytes"
	"errors"
	"github.com/zhangdapeng520/zdpgo_email/imap"
	"github.com/zhangdapeng520/zdpgo_email/message"
	"github.com/zhangdapeng520/zdpgo_email/message/mail"
	"io"
	"io/ioutil"
	"net/textproto"
	"strings"
	"time"
)

type SendConfig struct {
	Uid         string   //自定义的邮件头
	From        string   //发件人邮箱
	To          []string //收件人邮箱，可发给多人
	Subject     string   //邮件主题
	Body        string   //邮件内容
	Attachments []string //file path
}

type PreFilter struct {
	Seen           interface{} //true,false,nil
	Subject        string
	From           string
	To             string
	HeaderTagName  string
	HeaderTagValue string
	SentSince      string //日期格式字符串 "2006-01-02"
	Body           []string
}

// PostFilter
// qq等邮箱使用中文过滤时会报错：imap: cannot send literal: no continuation request received
// 先用临时解决办法吧，PreFilter过滤条件不要输入中文，获取结果后再次过滤
type PostFilter struct {
	Subject        string
	From           string
	To             string
	HeaderTagName  string
	HeaderTagValue string
	SentSince      string //日期格式字符串 "2006-01-02"
	Body           []string
	Attachments    []string //filenames
}

// MailMessage 邮件消息
type MailMessage struct {
	Subject        string   // 邮件标题
	From           string   // 发件人
	To             string   // 收件人
	HeaderTagName  string   // 请求头中的标识
	HeaderTagValue string   // 请求头中的标识
	Body           string   // 内容
	Attachments    []string // 附件
}

// 重新处理编码
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	charset = strings.ToLower(charset)
	if charset == "utf-8" || charset == "us-ascii" {
		return input, nil
	} else if charset == "gb2312" || charset == "gbk" {
		buf := &bytes.Buffer{}
		buf.ReadFrom(input)
		in := buf.Bytes()
		b, err := ConvertToUTF8(in, charset)
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(b)
		return r, nil
	}
	return nil, CommonError{Cause: "不支持的编码处理"}
}

// MailDecodeHeader 自行处理strHeader，类似=?GB2312?B?1tDOxLi9vP6y4srU?=
func MailDecodeHeader(s string) (string, error) {
	wordDecoder := WordDecoder{CharsetReader: charsetReader}
	dec, err := wordDecoder.DecodeHeader(s)
	if err != nil {
		return s, err
	}
	return dec, nil
}

// SearchByTag 根据标签进行搜索
func (e *EmailImap) SearchByTag(from string, startTime string, tagKey, tagValue string) ([]MailMessage, error) {
	e.Config.HeaderTagName = tagKey
	bf := PreFilter{
		From:           from,
		SentSince:      startTime,
		HeaderTagName:  tagKey,
		HeaderTagValue: tagValue,
	}
	af := PostFilter{}
	results, err := e.SearchBF(&bf, &af)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// SearchByDefaultTag 根据默认的tag进行搜索
func (e *EmailImap) SearchByDefaultTag(from string, startTime string) ([]MailMessage, error) {
	mailMessages, err := e.SearchByTag(from, startTime, e.Config.HeaderTagName, e.Config.HeaderTagValue)
	if err != nil {
		return nil, err
	}
	return mailMessages, nil
}

// SearchByKey 根据时间和key搜索邮件
func (e *EmailImap) SearchByKey(from string, startTime string, key string) ([]MailMessage, error) {
	mailMessages, err := e.SearchByTag(from, startTime, e.Config.HeaderTagName, key)
	if err != nil {
		return nil, err
	}
	return mailMessages, nil
}

// SearchByKeyToday 根据key搜索今天的邮件
func (e *EmailImap) SearchByKeyToday(from string, key string) ([]MailMessage, error) {
	startTime := time.Now().Format("2006-01-02") // 今天
	mailMessages, err := e.SearchByTag(from, startTime, e.Config.HeaderTagName, key)
	if err != nil {
		return nil, err
	}
	return mailMessages, nil
}

// IsSendSuccessByKey 根据key判断邮件是否发送成功
func (e *EmailImap) IsSendSuccessByKey(from string, key string) bool {
	result, err := e.SearchByKeyToday(from, key)
	return err == nil && result != nil && len(result) > 0
}

func (e *EmailImap) SearchBF(bf *PreFilter, af *PostFilter) ([]MailMessage, error) {
	if bf.From == "" {
		return nil, errors.New("bf.From 发件人不能为空")
	}

	var searchResults []MailMessage

	// 选择收件箱
	mbox, err := e.Select("INBOX", false)
	if err != nil {
		return searchResults, err
	}
	if mbox.Messages == 0 {
		err = errors.New("收件箱中没有内容")
		return searchResults, err
	}

	// 构建查询条件
	criteria := imap.NewSearchCriteria()
	if bf.Seen == true {
		criteria.WithFlags = []string{imap.SeenFlag}
	} else if bf.Seen == false {
		criteria.WithoutFlags = []string{imap.SeenFlag}
	}

	// 执行查询
	if bf.Subject != "" || bf.From != "" || bf.HeaderTagValue != "" {
		header := make(textproto.MIMEHeader)
		if bf.Subject != "" {
			header.Add("SUBJECT", bf.Subject)
			af.Subject = bf.Subject
		}
		if bf.From != "" {
			header.Add("FROM", bf.From)
			af.From = bf.From
		}
		if bf.HeaderTagValue != "" {
			af.HeaderTagValue = bf.HeaderTagValue
			if bf.HeaderTagName != "" {
				header.Add(bf.HeaderTagName, bf.HeaderTagValue) // 自定义的header头，不区分大小写
				af.HeaderTagName = bf.HeaderTagName
			} else {
				header.Add(e.Config.HeaderTagName, bf.HeaderTagValue) // 自定义的header头，不区分大小写
				af.HeaderTagName = e.Config.HeaderTagName
			}
		}
		criteria.Header = header
	}

	// 查询内容
	if len(bf.Body) > 0 {
		criteria.Body = bf.Body
		af.Body = bf.Body
	}

	// 查询时间
	var t time.Time
	if bf.SentSince != "" {
		//只支持了日期且不区分时区
		t, err = time.Parse("2006-01-02", bf.SentSince)
		if err != nil {
			return searchResults, err
		}
		criteria.SentSince = t
		af.SentSince = bf.SentSince
	}

	// 执行搜索
	uids, err := e.Search(criteria)
	if err != nil {
		return searchResults, err
	}

	if len(uids) == 0 {
		return searchResults, errors.New("查询结果为空")
	}
	if err != nil {
		return searchResults, err
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	// 或者整个消息的内容
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message)
	doneSearch := make(chan error, 1)
	go func() {
		doneSearch <- e.Fetch(seqset, items, messages)
	}()
outLoop:
	for msg := range messages {
		if msg == nil {
			err = errors.New("服务器没有返回消息")
			return searchResults, err
		}

		r := msg.GetBody(&section)
		if r == nil {
			err = errors.New("服务器没有返回内容")
			return searchResults, err
		}

		mr, err := mail.CreateReader(r)
		if err != nil {
			return searchResults, err
		}

		var (
			realDate    time.Time
			realFrom    string
			realTo      string
			realSubject string
			realUid     string
		)

		header := mr.Header

		// 根据时间过滤
		var t time.Time
		if realDate, err = header.Date(); err == nil {
			if af.SentSince != "" && realDate.Before(t) {
				continue
			}
		} else {
			return searchResults, err
		}

		// 根据发件人邮箱过滤
		fromHeader := header.Get("From")
		if fromHeader != "" {
			if realFrom, err = MailDecodeHeader(header.Get("From")); err != nil {
				return searchResults, err
			}
			if !strings.Contains(realFrom, af.From) {
				continue
			}
		}

		// 根据收件人邮箱过滤
		toHeader := header.Get("To")
		if toHeader != "" {
			if realTo, err = MailDecodeHeader(header.Get("To")); err != nil {
				return searchResults, err
			}
			if !strings.Contains(realTo, af.To) {
				continue
			}
		}

		// 根据邮件标题过滤
		subjectHeader := header.Get("Subject")
		if subjectHeader != "" {
			if realSubject, err = MailDecodeHeader(header.Get("Subject")); err != nil {
				return searchResults, err
			}
			if !strings.Contains(realSubject, af.Subject) {
				continue
			}
		}

		// 根据UID过滤
		HeaderTagNameHeader := header.Get(af.HeaderTagName)
		if HeaderTagNameHeader != "" {
			if realUid, err = MailDecodeHeader(HeaderTagNameHeader); err != nil {
				return searchResults, err
			}
			if realUid != af.HeaderTagValue {
				continue
			}
		}

		// 处理每一条消息
		bodyText := ""
		for {
			p, err := mr.NextPart()
			if err != nil {
				if err == io.EOF || err.Error() == "multipart: NextPart: EOF" {
					break
				}
				if !message.IsUnknownCharset(err) {
					return searchResults, err
				}
			}
			bodyBytes, _ := ioutil.ReadAll(p.Body)
			if len(bodyBytes) > 0 {
				// 尝试编码转换
				if IsGBK(bodyBytes) {
					bodyBytes, _ = ConvertToUTF8(bodyBytes, "gbk")
				}
				bodyText = string(bodyBytes)
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader: // 获取邮件内容
				if len(af.Body) > 0 {
					for _, body := range af.Body {
						if !strings.Contains(bodyText, body) {
							continue outLoop
						}
					}
				}
			case *mail.AttachmentHeader: // 获取附件
				filename, _ := h.Filename()
				// 尝试编码转换
				if strings.HasPrefix(filename, "=?") {
					filename, _ = MailDecodeHeader(filename)
				}

				if len(af.Attachments) > 0 {
					for _, attachName := range af.Attachments {
						if !strings.Contains(filename, attachName) {
							continue outLoop
						}
					}
				}

			}
		}

		// 再过滤一遍
		if af.HeaderTagValue != "" && af.HeaderTagValue != realUid {
			continue
		}

		// 将满足条件的邮件封装，之后统一返回
		m := MailMessage{Subject: realSubject, From: realFrom, To: realTo, HeaderTagName: af.HeaderTagName,
			HeaderTagValue: realUid, Body: bodyText}
		searchResults = append(searchResults, m)

	}
	if err := <-doneSearch; err != nil {
		return searchResults, err
	}

	return searchResults, nil
}
