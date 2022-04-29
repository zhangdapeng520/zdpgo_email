package zdpgo_email

import (
	"bytes"
	"github.com/zhangdapeng520/zdpgo_email/imap"
	"github.com/zhangdapeng520/zdpgo_email/message"
	"github.com/zhangdapeng520/zdpgo_email/message/mail"
	"io"
	"io/ioutil"
	"log"
	"net/textproto"
	"strings"
	"time"
)

const UID_NAME = "X-Mirror-Uid"

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
	Uid       string
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
	Uid         string
	SentSince   string //日期格式字符串 "2006-01-02"
	Body        []string
	Attachments []string //filenames
}

type MailMessage struct {
	//Uid				int
	Subject     string
	From        string
	To          string
	Uid         string
	Body        string
	Attachments []string //filenames
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
	var searchResults []MailMessage

	// Select INBOX
	mbox, err := e.Select("INBOX", false)
	if err != nil {
		return searchResults, err
	}

	if mbox.Messages == 0 {
		return searchResults, CommonError{Cause: "No message in mailbox"}
	}

	criteria := imap.NewSearchCriteria()
	if bf.Seen == true {
		criteria.WithFlags = []string{imap.SeenFlag}
	} else if bf.Seen == false {
		criteria.WithoutFlags = []string{imap.SeenFlag}
	}

	if bf.Subject != "" || bf.From != "" || bf.Uid != "" {
		header := make(textproto.MIMEHeader)
		if bf.Subject != "" {
			header.Add("SUBJECT", bf.Subject)
			af.Subject = bf.Subject
		}
		if bf.From != "" {
			header.Add("FROM", bf.From)
			af.From = bf.From
		}
		if bf.Uid != "" {
			//自定义的header头，不区分大小写
			header.Add(UID_NAME, bf.Uid)
			af.Uid = bf.Uid
		}
		criteria.Header = header
	}

	if len(bf.Body) > 0 {
		criteria.Body = bf.Body
		af.Body = bf.Body
	}

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
		return searchResults, CommonError{Cause: "uids is null"}
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	// Get the whole message body
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message)
	done_search := make(chan error, 1)
	go func() {
		done_search <- e.Fetch(seqset, items, messages)
	}()

outLoop:
	for msg := range messages {
		if msg == nil {
			return searchResults, CommonError{Cause: "Server didn't returned message"}
		}

		r := msg.GetBody(&section)
		if r == nil {
			return searchResults, CommonError{Cause: "Server didn't returned message body"}
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
		if realFrom, err = MailDecodeHeader(header.Get("From")); err != nil {
			return searchResults, err
		}
		if !strings.Contains(realFrom, af.From) {
			continue
		}

		// 根据收件人邮箱过滤
		if realTo, err = MailDecodeHeader(header.Get("To")); err != nil {
			return searchResults, err
		}
		if !strings.Contains(realTo, af.To) {
			continue
		}

		// 根据邮件标题过滤
		if realSubject, err = MailDecodeHeader(header.Get("Subject")); err != nil {
			return searchResults, err
		}
		if !strings.Contains(realSubject, af.Subject) {
			continue
		}

		// 根据UID过滤
		if realUid, err = MailDecodeHeader(header.Get(UID_NAME)); err != nil {
			return searchResults, err
		}
		if realUid != af.Uid {
			continue
		}

		// Process each message's part
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				if !message.IsUnknownCharset(err) {
					return searchResults, err
				}
			}

			bodyBytes, _ := ioutil.ReadAll(p.Body)
			// 尝试编码转换
			if IsGBK(bodyBytes) {
				bodyBytes, err = ConvertToUTF8(bodyBytes, "gbk")
				if err != nil {
					log.Println(err)
				}
			}
			bodyText := string(bodyBytes)

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				// This is the message's text (can be plain-text or HTML)
				if len(af.Body) > 0 {
					for _, body := range af.Body {
						if !strings.Contains(bodyText, body) {
							continue outLoop
						}
					}
				}
				//log.Printf("Got text: %v", bodyText)
			case *mail.AttachmentHeader:
				// This is an attachment
				filename, _ := h.Filename()
				// 尝试编码转换
				if strings.HasPrefix(filename, "=?") {
					filename, err = MailDecodeHeader(filename)
					if err != nil {
						log.Println(err)
					}
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

		m := MailMessage{Subject: realSubject, From: realFrom, To: realTo, Uid: realUid}
		searchResults = append(searchResults, m)
	}

	if err := <-done_search; err != nil {
		return searchResults, err
	}

	return searchResults, nil
}
