package mail

import (
	"bytes"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email/message/mail"
	"github.com/zhangdapeng520/zdpgo_email/imap"
	"github.com/zhangdapeng520/zdpgo_email/imap/client"
	"io"
	"io/ioutil"
	"log"
	"mail-go/utils"
	"net/textproto"
	"os"
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

/*
title smtp登录
@param string host SMTP主机
@param int port SMTP端口
@param string username
@param string password
@param bool ssl 是否使用SSL
@return gomail.SendCloser,error 登录成功后的SendCloser，错误信息
*/

func SMTPLogin(host string, port int, username string, password string, ssl bool) (gomail.SendCloser, error) {
	d := &gomail.Dialer{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		SSL:      ssl,
	}
	c, err := d.Dial()
	if err != nil {
		return nil, err
	}
	return c, nil
}

/*
title 使用gomail发送邮件
@param gomail.SendCloser c 登录成功后的SendCloser
@param SendConfig sendConfig
@return error
*/

func SendMail(c gomail.SendCloser, sendConfig *SendConfig) error {
	m := gomail.NewMessage()

	//自定义邮件头
	if sendConfig.Uid != "" {
		m.SetHeader(UID_NAME, sendConfig.Uid)
	}
	if sendConfig.From != "" {
		m.SetHeader("From", sendConfig.From)
	} else {
		return utils.CommonError{Cause: "From not set"}
	}
	if sendConfig.To != nil {
		m.SetHeader("To", sendConfig.To...)
	} else {
		return utils.CommonError{Cause: "To not set"}
	}
	if sendConfig.Subject != "" {
		m.SetHeader("Subject", sendConfig.Subject)
	} else {
		return utils.CommonError{Cause: "Subject not set"}
	}
	if sendConfig.Body != "" {
		m.SetBody("text/html", sendConfig.Body)
	} else {
		return utils.CommonError{Cause: "Body not set"}
	}

	if sendConfig.Attachments != nil {

		for _, file := range sendConfig.Attachments {
			_, err := os.Stat(file)
			if err != nil {
				fmt.Println("Error:", file, "does not exist")
				return err
			} else {
				fmt.Println("uploading", file, "...")
				m.Attach(file)
			}
		}
	}

	err := gomail.Send(c, m)
	return err
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
		b, err := utils.ConvertToUTF8(in, charset)
		if err != nil {
			return nil, err
		}
		r := bytes.NewReader(b)
		return r, nil
	}
	return nil, utils.CommonError{Cause: "不支持的编码处理"}
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

func ImapLogin(host string, port int, username string, password string, ssl bool) (*client.Client, error) {
	log.Println("Connecting to server...")

	var (
		c   *client.Client
		err error
	)

	// Connect to server
	addr := fmt.Sprintf("%s:%d", host, port)
	// SSL
	if ssl == true {
		c, err = client.DialTLS(addr, nil)
	} else {
		c, err = client.Dial(addr)
	}

	if err != nil {
		return nil, err
	}
	log.Println("Connected")

	// Login
	if err := c.Login(username, password); err != nil {
		return nil, err
	}
	log.Println("Logged in")

	return c, nil
}

func RecvSearch(c *client.Client, bf *PreFilter, af *PostFilter) ([]MailMessage, error) {
	var searchResults []MailMessage

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return searchResults, err
	}

	if mbox.Messages == 0 {
		log.Println("No message in mailbox")
		return searchResults, utils.CommonError{Cause: "No message in mailbox"}
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
		log.Println("set test header:", header)
		criteria.Header = header
	}

	if len(bf.Body) > 0 {
		criteria.Body = bf.Body
		af.Body = bf.Body
	}

	var t time.Time
	if bf.SentSince != "" {
		//只支持了日期且不区分时区
		//t, err := time.ParseInLocation("2006-01-02", "2022-03-27", time.Local)
		t, err = time.Parse("2006-01-02", bf.SentSince)
		if err != nil {
			return searchResults, err
		}
		log.Println("set test time: ", t)
		criteria.SentSince = t
		af.SentSince = bf.SentSince
	}

	//log.Println("set criteria: ", criteria)
	//这里测试qq邮箱 subject为中文时报错 cannot send literal: no continuation request received
	uids, err := c.Search(criteria)
	if err != nil {
		return searchResults, err
	}

	log.Println("Search complete, found uids:", uids)

	if len(uids) == 0 {
		log.Println("uids is null")
		return searchResults, utils.CommonError{Cause: "uids is null"}
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	// Get the whole message body
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message)
	done_search := make(chan error, 1)
	go func() {
		done_search <- c.Fetch(seqset, items, messages)
		log.Println("Fetch complete")
	}()

outLoop:
	for msg := range messages {
		if msg == nil {
			log.Println("Server didn't returned message")
			return searchResults, utils.CommonError{Cause: "Server didn't returned message"}
		}

		r := msg.GetBody(&section)
		if r == nil {
			log.Println("Server didn't returned message body")
			return searchResults, utils.CommonError{Cause: "Server didn't returned message body"}
		}

		// Create a new mailTest reader
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

		// Print some info about the message
		header := mr.Header

		if realDate, err = header.Date(); err == nil {
			log.Println("Date:", realDate)
			if af.SentSince != "" && realDate.Before(t) {
				log.Println("invalid Date:", realDate)
				continue
			}
		} else {
			return searchResults, err
		}

		////if from, err := header.AddressList("From"); err == nil {
		//if from, err := header.Text("From"); err == nil {
		//	log.Println("From:", from)
		//	r,_:= MailDecodeHeader(from)
		//	log.Println(r)
		//}else {
		//	log.Fatal(err)
		//}

		if realFrom, err = MailDecodeHeader(header.Get("From")); err != nil {
			return searchResults, err
		}
		log.Println("From:", realFrom)
		if !strings.Contains(realFrom, af.From) {
			log.Println("invalid from:", realFrom)
			continue
		}

		////if to, err := header.AddressList("To"); err == nil {
		//if to, err := header.Text("To"); err == nil {
		//	log.Println("To:", to)
		//}else {
		//	log.Fatal(err)
		//}

		if realTo, err = MailDecodeHeader(header.Get("To")); err != nil {
			return searchResults, err
		}
		log.Println("To:", realTo)
		if !strings.Contains(realTo, af.To) {
			log.Println("invalid to:", realTo)
			continue
		}

		//if realSubject, err = header.Subject(); err == nil {
		//}else {
		//	// 尝试自行编码转换
		//	realSubject, err = MailDecodeHeader(realSubject)
		//	if err != nil {
		//		log.Fatal(err)
		//	}
		//}
		if realSubject, err = MailDecodeHeader(header.Get("Subject")); err != nil {
			return searchResults, err
		}
		log.Println("Subject:", realSubject)
		if !strings.Contains(realSubject, af.Subject) {
			log.Println("invalid subject:", realSubject)
			continue
		}

		if realUid, err = MailDecodeHeader(header.Get(UID_NAME)); err != nil {
			return searchResults, err
		}
		log.Printf("%s:%s", UID_NAME, realUid)
		if !strings.Contains(realUid, af.Uid) {
			log.Printf("invalid %s:%s", UID_NAME, realUid)
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
			if utils.IsGBK(bodyBytes) {
				bodyBytes, err = utils.ConvertToUTF8(bodyBytes, "gbk")
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
							log.Println("invalid body")
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

				log.Printf("Got attachment: %v", filename)
				//writeToFile(filename, bodyText)

				if len(af.Attachments) > 0 {
					for _, attachName := range af.Attachments {
						if !strings.Contains(filename, attachName) {
							log.Println("invalid attachment")
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

func writeToFile(filename string, content string) error {
	var f *os.File
	var err error
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		f, err = os.Create(filename) //创建文件
		log.Println("文件不存在，已创建")
	} else {
		f, err = os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
		log.Println("文件存在")
	}

	if err != nil {
		return err
	}

	_, err = io.WriteString(f, content)

	if err != nil {
		return err
	}

	return nil
}
