package zdpgo_email

import (
	"bytes"
	"errors"
	"fmt"
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
	Seen      interface{} //true,false,nil
	Subject   string
	From      string
	To        string
	HeaderTag string
	SentSince string //日期格式字符串 "2006-01-02"
	Body      []string
}

// PostFilter
// qq等邮箱使用中文过滤时会报错：imap: cannot send literal: no continuation request received
// 先用临时解决办法吧，PreFilter过滤条件不要输入中文，获取结果后再次过滤
type PostFilter struct {
	Subject     string
	From        string
	To          string
	HeaderTag   string
	SentSince   string //日期格式字符串 "2006-01-02"
	Body        []string
	Attachments []string //filenames
}

// MailMessage 邮件消息
type MailMessage struct {
	Subject     string   // 邮件标题
	From        string   // 发件人
	To          string   // 收件人
	HeaderTag   string   // 请求头中的标识
	Body        string   // 内容
	Attachments []string // 附件
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

func (e *EmailImap) SearchBF(bf *PreFilter, af *PostFilter) ([]MailMessage, error) {
	if bf.From == "" {
		return nil, errors.New("bf.From 发件人不能同时为空")
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
	if bf.Subject != "" || bf.From != "" || bf.HeaderTag != "" {
		header := make(textproto.MIMEHeader)
		if bf.Subject != "" {
			header.Add("SUBJECT", bf.Subject)
			af.Subject = bf.Subject
		}
		if bf.From != "" {
			header.Add("FROM", bf.From)
			af.From = bf.From
		}
		if bf.HeaderTag != "" {
			header.Add(e.Config.HeaderTagName, bf.HeaderTag) // 自定义的header头，不区分大小写
			af.HeaderTag = bf.HeaderTag
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
	for {
		if msg, ok := <-messages; ok { //每一层遍历都创建 v ok,同第一种方式
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
			HeaderTagNameHeader := header.Get(e.Config.HeaderTagName)
			if HeaderTagNameHeader != "" {
				if realUid, err = MailDecodeHeader(HeaderTagNameHeader); err != nil {
					return searchResults, err
				}
				if realUid != af.HeaderTag {
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

			m := MailMessage{Subject: realSubject, From: realFrom, To: realTo, HeaderTag: realUid, Body: bodyText}
			searchResults = append(searchResults, m)
		} else {
			break
		}
	}
	if err := <-doneSearch; err != nil {
		return searchResults, err
	}

	return searchResults, nil
}

func (e *EmailImap) SearchBF1(bf *PreFilter, af *PostFilter) ([]MailMessage, error) {
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
	if bf.Subject != "" || bf.From != "" || bf.HeaderTag != "" {
		header := make(textproto.MIMEHeader)
		if bf.Subject != "" {
			header.Add("SUBJECT", bf.Subject)
			af.Subject = bf.Subject
		}
		if bf.From != "" {
			header.Add("FROM", bf.From)
			af.From = bf.From
		}
		if bf.HeaderTag != "" {
			header.Add(e.Config.HeaderTagName, bf.HeaderTag) // 自定义的header头，不区分大小写
			af.HeaderTag = bf.HeaderTag
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

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	// 或者整个消息的内容
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message)
	doneSearch := make(chan error, 1)
	go func() {
		result := e.Fetch(seqset, items, messages)
		fmt.Println("result", result)
		doneSearch <- result
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
			fmt.Println("发件人请求头", fromHeader)
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
			fmt.Println("收件人请求头", toHeader)
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
			fmt.Println("邮件标题请求头", subjectHeader)
			if realSubject, err = MailDecodeHeader(header.Get("Subject")); err != nil {
				fmt.Println("2222222出错了", err)
				return searchResults, err
			}
			if !strings.Contains(realSubject, af.Subject) {
				continue
			}
		}

		// 根据UID过滤
		HeaderTagNameHeader := header.Get(e.Config.HeaderTagName)
		if HeaderTagNameHeader != "" {
			fmt.Println("自定义请求头", HeaderTagNameHeader)
			if realUid, err = MailDecodeHeader(HeaderTagNameHeader); err != nil {
				fmt.Println("111111111111出错了", err)
				return searchResults, err
			}
			if realUid != af.HeaderTag {
				continue
			}
		}
		fmt.Println("33333333333333333333")

		// 处理每一条消息
		bodyText := ""
		for {
			fmt.Println("========================1111")
			p, err := mr.NextPart()
			if err != nil {
				fmt.Println("xxxxxxxxxxxxxxxx", err)
				if err == io.EOF || err.Error() == "multipart: NextPart: EOF" {
					break
				}
				if !message.IsUnknownCharset(err) {
					fmt.Println("xxxxxxxxxxxxxxxx11111", err)
					return searchResults, err
				}
			}
			fmt.Println("========================")
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
				fmt.Println("xxxxxxxxxxxxxxxx22222", err)
				if len(af.Body) > 0 {
					for _, body := range af.Body {
						if !strings.Contains(bodyText, body) {
							continue outLoop
						}
					}
				}
			case *mail.AttachmentHeader: // 获取附件
				fmt.Println("xxxxxxxxxxxxxxxx333333", err)
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

		fmt.Println(66666666666)
		m := MailMessage{Subject: realSubject, From: realFrom, To: realTo, HeaderTag: realUid, Body: bodyText}
		searchResults = append(searchResults, m)
		fmt.Println(77777777777777)
	}
	fmt.Println(555555555555)
	if err := <-doneSearch; err != nil {
		fmt.Println("333333333333333333334444444444444")
		return searchResults, err
	}

	fmt.Println("走到了最终的地方。。。。")
	return searchResults, nil
}
