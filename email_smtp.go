package zdpgo_email

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/zhangdapeng520/zdpgo_log"
	"github.com/zhangdapeng520/zdpgo_random"
	"github.com/zhangdapeng520/zdpgo_yaml"
	"io"
	"math"
	"math/big"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

const (
	MaxLineLength      = 76                             // MaxLineLength RFC 2045的最大线长是多少
	defaultContentType = "text/plain; charset=us-ascii" // defaultContentType 是默认的内容类型根据RFC 2045，章节5.2
)

// ErrMissingBoundary 当多个部分的实体没有给定边界时返回
var ErrMissingBoundary = errors.New("没有为多部分实体找到边界")

// ErrMissingContentType 是返回时，没有“内容类型”的头MIME实体
var ErrMissingContentType = errors.New("没有找到MIME实体的内容类型")

// EmailSmtp 是用于电子邮件消息的类型吗
type EmailSmtp struct {
	ReplyTo     []string             `json:"reply_to" yaml:"reply_to" env:"reply_to"`
	From        string               `json:"from" yaml:"from" env:"from"`          // 发送者
	To          []string             `json:"to" yaml:"to" env:"to"`                // 接收者
	Bcc         []string             `json:"bcc" yaml:"bcc" env:"bcc"`             // 密送
	Cc          []string             `json:"cc" yaml:"cc" env:"cc"`                // 抄送
	Subject     string               `json:"subject" yaml:"subject" env:"subject"` // 主题,邮件标题
	Text        []byte               `json:"text" yaml:"text" env:"text"`
	HTML        []byte               `json:"html" yaml:"html" env:"html"`
	Sender      string               `json:"sender" yaml:"sender" env:"sender"`
	Headers     textproto.MIMEHeader `json:"headers" yaml:"headers" env:"headers"`             // 协议头
	Attachments []*Attachment        `json:"attachments" yaml:"attachments" env:"attachments"` // 附件
	ReadReceipt []string             `json:"read_receipt" yaml:"read_receipt" env:"read_receipt"`
	Config      *Config              `json:"config" yaml:"config" env:"config"` // 配置对象
	random      *zdpgo_random.Random
	Fs          *embed.FS      // 嵌入文件系统
	Log         *zdpgo_log.Log // 日志对象
}

// part multipart.Part 的副本
type part struct {
	header textproto.MIMEHeader
	body   []byte
}

// NewEmailSmtp 创建邮件对象
func NewEmailSmtp() (email *EmailSmtp, err error) {
	var config Config
	yaml := zdpgo_yaml.New()
	err = yaml.ReadDefaultConfig(&config)
	if err != nil {
		return
	}
	return NewEmailSmtpWithConfig(config)
}

func NewEmailSmtpWithConfig(config Config) (email *EmailSmtp, err error) {
	email = &EmailSmtp{Headers: textproto.MIMEHeader{}}

	// 初始化配置
	if config.HeaderTagName == "" {
		config.HeaderTagName = "X-ZdpgoEmail-Auther"
	}
	if config.HeaderTagValue == "" {
		config.HeaderTagValue = "zhangdapeng520"
	}
	email.Config = &config
	email.random = zdpgo_random.New()

	// 日志
	logConfig := zdpgo_log.Config{
		Debug:       config.Debug,
		OpenJsonLog: true,
		LogFilePath: "logs/zdpgo/zdpgo_email.log",
	}
	if config.Debug {
		logConfig.IsShowConsole = true
	}
	email.Log = zdpgo_log.NewWithConfig(logConfig)
	return
}

// trimReader 是一个自定义的 io.Reader，这将削减任何前导空格，因为这可能导致电子邮件导入失败。
type trimReader struct {
	rd      io.Reader
	trimmed bool
}

// Read 从原始读取器中删除任何unicode空格
func (tr *trimReader) Read(buf []byte) (int, error) {
	n, err := tr.rd.Read(buf)
	if err != nil {
		return n, err
	}
	if !tr.trimmed {
		t := bytes.TrimLeftFunc(buf[:n], unicode.IsSpace)
		tr.trimmed = true
		n = copy(buf, t)
	}
	return n, err
}

func handleAddressList(v []string) []string {
	res := []string{}
	for _, a := range v {
		w := strings.Split(a, ",")
		for _, addr := range w {
			decodedAddr, err := (&mime.WordDecoder{}).DecodeHeader(strings.TrimSpace(addr))
			if err == nil {
				res = append(res, decodedAddr)
			} else {
				res = append(res, addr)
			}
		}
	}
	return res
}

// NewEmailFromReader 从一个io.Reade读取字节流, 并返回包含已解析数据的电子邮件结构体。
// 该函数需要RFC 5322格式的数据。
func NewEmailFromReader(r io.Reader) (email *EmailSmtp, err error) {
	var (
		ct     string
		params map[string]string
	)

	email, err = NewEmailSmtp()
	s := &trimReader{rd: r}
	tp := textproto.NewReader(bufio.NewReader(s))
	hdrs, err := tp.ReadMIMEHeader()
	if err != nil {
		return
	}
	// Set the subject, to, cc, bcc, and from
	for h, v := range hdrs {
		switch h {
		case "Subject":
			email.Subject = v[0]
			subj, err := (&mime.WordDecoder{}).DecodeHeader(email.Subject)
			if err == nil && len(subj) > 0 {
				email.Subject = subj
			}
			delete(hdrs, h)
		case "To":
			email.To = handleAddressList(v)
			delete(hdrs, h)
		case "Cc":
			email.Cc = handleAddressList(v)
			delete(hdrs, h)
		case "Bcc":
			email.Bcc = handleAddressList(v)
			delete(hdrs, h)
		case "Reply-To":
			email.ReplyTo = handleAddressList(v)
			delete(hdrs, h)
		case "From":
			email.From = v[0]
			fr, err := (&mime.WordDecoder{}).DecodeHeader(email.From)
			if err == nil && len(fr) > 0 {
				email.From = fr
			}
			delete(hdrs, h)
		}
	}
	email.Headers = hdrs
	body := tp.R
	// Recursively parse the MIME parts
	ps, err := parseMIMEParts(email.Headers, body)
	if err != nil {
		return
	}
	for _, p := range ps {
		if ct := p.header.Get("Content-Type"); ct == "" {
			return email, ErrMissingContentType
		}
		ct, _, err = mime.ParseMediaType(p.header.Get("Content-Type"))
		if err != nil {
			return
		}
		// Check if part is an attachment based on the existence of the Content-Disposition header with a value of "attachment".
		if cd := p.header.Get("Content-Disposition"); cd != "" {
			cd, params, err = mime.ParseMediaType(p.header.Get("Content-Disposition"))
			if err != nil {
				return
			}
			filename, filenameDefined := params["filename"]
			if cd == "attachment" || (cd == "inline" && filenameDefined) {
				_, err = email.Attach(bytes.NewReader(p.body), filename, ct)
				if err != nil {
					return
				}
				continue
			}
		}
		switch {
		case ct == "text/plain":
			email.Text = p.body
		case ct == "text/html":
			email.HTML = p.body
		}
	}
	return
}

// parseMIMEParts 将递归遍历一个MIME实体并返回一个[]MIME。包含每个(扁平)mime的部分。
// 需要注意的是递归的次数是没有限制的，所以在解析未知的MIME结构时要小心!
func parseMIMEParts(hs textproto.MIMEHeader, b io.Reader) ([]*part, error) {
	var ps []*part
	// If no content type is given, set it to the default
	if _, ok := hs["Content-Type"]; !ok {
		hs.Set("Content-Type", defaultContentType)
	}
	ct, params, err := mime.ParseMediaType(hs.Get("Content-Type"))
	if err != nil {
		return ps, err
	}
	// If it's a multipart email, recursively parse the parts
	if strings.HasPrefix(ct, "multipart/") {
		if _, ok := params["boundary"]; !ok {
			return ps, ErrMissingBoundary
		}
		mr := multipart.NewReader(b, params["boundary"])
		for {
			var buf bytes.Buffer
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return ps, err
			}
			if _, ok := p.Header["Content-Type"]; !ok {
				p.Header.Set("Content-Type", defaultContentType)
			}
			subct, _, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
			if err != nil {
				return ps, err
			}
			if strings.HasPrefix(subct, "multipart/") {
				sps, err := parseMIMEParts(p.Header, p)
				if err != nil {
					return ps, err
				}
				ps = append(ps, sps...)
			} else {
				var reader io.Reader
				reader = p
				const cte = "Content-Transfer-Encoding"
				if p.Header.Get(cte) == "base64" {
					reader = base64.NewDecoder(base64.StdEncoding, reader)
				}
				// Otherwise, just append the part to the list
				// Copy the part data into the buffer
				if _, err := io.Copy(&buf, reader); err != nil {
					return ps, err
				}
				ps = append(ps, &part{body: buf.Bytes(), header: p.Header})
			}
		}
	} else {
		// If it is not a multipart email, parse the body content as a single "part"
		switch hs.Get("Content-Transfer-Encoding") {
		case "quoted-printable":
			b = quotedprintable.NewReader(b)
		case "base64":
			b = base64.NewDecoder(base64.StdEncoding, b)
		}
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, b); err != nil {
			return ps, err
		}
		ps = append(ps, &part{body: buf.Bytes(), header: hs})
	}
	return ps, nil
}

// Attach 从io.Reader读取内容作为邮件的附件
// 要求参数包含一个io.Reader输入流, 文件名, Content-Type
// @param r：输入流
// @param filename：文件名，不包含后缀
// @param c：Content-Type
func (e *EmailSmtp) Attach(r io.Reader, filename string, c string) (a *Attachment, err error) {
	var buffer bytes.Buffer
	if _, err = io.Copy(&buffer, r); err != nil {
		return
	}
	at := &Attachment{
		Filename:    filename,
		ContentType: c,
		Header:      textproto.MIMEHeader{},
		Content:     buffer.Bytes(),
	}
	e.Attachments = append(e.Attachments, at)
	return at, nil
}

// AttachFile 用来给邮件内容添加附件
// @param filename 文件名，一般是相对路径
// 如果文件能够引用成功，就创建一个Attachment附件对象
// This Attachment is then appended to the slice of Email.Attachments.
// The function will then return the Attachment for reference, as well as nil for the error, if successful.
func (e *EmailSmtp) AttachFile(filename string) (a *Attachment, err error) {
	// 打开文件
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	// 根据文件后缀获取文件类型的后缀
	ct := mime.TypeByExtension(filepath.Ext(filename))

	// 获取文件名，不包含后缀
	basename := filepath.Base(filename)

	// 邮箱，添加附件
	return e.Attach(f, basename, ct)
}

// msgHeaders 合并邮件的各领域的标准兼容的方式和自定义标题一起创建一个MIMEHeader用于生成的消息。它不会改变e.Headers。
// "e"'s fields To, Cc, From, Subject 将被使用，除非它们出现在e.Headers中。除非在e.Headers中设置，
// "Date" 将会填充为当前时间
func (e *EmailSmtp) msgHeaders() (textproto.MIMEHeader, error) {
	res := make(textproto.MIMEHeader, len(e.Headers)+6)
	if e.Headers != nil {
		for _, h := range []string{"Reply-To", "To", "Cc", "From", "Subject", "Date", "Message-Id", "MIME-Version"} {
			if v, ok := e.Headers[h]; ok {
				res[h] = v
			}
		}
	}
	// Set headers if there are values.
	if _, ok := res["Reply-To"]; !ok && len(e.ReplyTo) > 0 {
		res.Set("Reply-To", strings.Join(e.ReplyTo, ", "))
	}
	if _, ok := res["To"]; !ok && len(e.To) > 0 {
		res.Set("To", strings.Join(e.To, ", "))
	}
	if _, ok := res["Cc"]; !ok && len(e.Cc) > 0 {
		res.Set("Cc", strings.Join(e.Cc, ", "))
	}
	if _, ok := res["Subject"]; !ok && e.Subject != "" {
		res.Set("Subject", e.Subject)
	}
	if _, ok := res["Message-Id"]; !ok {
		id, err := generateMessageID()
		if err != nil {
			return nil, err
		}
		res.Set("Message-Id", id)
	}
	// Date and From are required headers.
	if _, ok := res["From"]; !ok {
		res.Set("From", e.From)
	}
	if _, ok := res["Date"]; !ok {
		res.Set("Date", time.Now().Format(time.RFC1123Z))
	}
	if _, ok := res["MIME-Version"]; !ok {
		res.Set("MIME-Version", "1.0")
	}
	for field, vals := range e.Headers {
		if _, ok := res[field]; !ok {
			res[field] = vals
		}
	}
	return res, nil
}

func writeMessage(buff io.Writer, msg []byte, multipart bool, mediaType string, w *multipart.Writer) error {
	if multipart {
		header := textproto.MIMEHeader{
			"Content-Type":              {mediaType + "; charset=UTF-8"},
			"Content-Transfer-Encoding": {"quoted-printable"},
		}
		if _, err := w.CreatePart(header); err != nil {
			return err
		}
	}

	qp := quotedprintable.NewWriter(buff)
	// Write the text
	if _, err := qp.Write(msg); err != nil {
		return err
	}
	return qp.Close()
}

func (e *EmailSmtp) categorizeAttachments() (htmlRelated, others []*Attachment) {
	for _, a := range e.Attachments {
		if a.HTMLRelated {
			htmlRelated = append(htmlRelated, a)
		} else {
			others = append(others, a)
		}
	}
	return
}

// Bytes 将Email对象转换为[]字节表示，包括所有需要的MIMEHeaders、边界等。
func (e *EmailSmtp) Bytes() ([]byte, error) {
	// TODO: better guess buffer size
	buff := bytes.NewBuffer(make([]byte, 0, 4096))

	headers, err := e.msgHeaders()
	if err != nil {
		return nil, err
	}

	htmlAttachments, otherAttachments := e.categorizeAttachments()
	if len(e.HTML) == 0 && len(htmlAttachments) > 0 {
		return nil, errors.New("there are HTML attachments, but no HTML body")
	}

	var (
		isMixed       = len(otherAttachments) > 0
		isAlternative = len(e.Text) > 0 && len(e.HTML) > 0
		isRelated     = len(e.HTML) > 0 && len(htmlAttachments) > 0
	)

	var w *multipart.Writer
	if isMixed || isAlternative || isRelated {
		w = multipart.NewWriter(buff)
	}
	switch {
	case isMixed:
		headers.Set("Content-Type", "multipart/mixed;\r\n boundary="+w.Boundary())
	case isAlternative:
		headers.Set("Content-Type", "multipart/alternative;\r\n boundary="+w.Boundary())
	case isRelated:
		headers.Set("Content-Type", "multipart/related;\r\n boundary="+w.Boundary())
	case len(e.HTML) > 0:
		headers.Set("Content-Type", "text/html; charset=UTF-8")
		headers.Set("Content-Transfer-Encoding", "quoted-printable")
	default:
		headers.Set("Content-Type", "text/plain; charset=UTF-8")
		headers.Set("Content-Transfer-Encoding", "quoted-printable")
	}
	headerToBytes(buff, headers)
	_, err = io.WriteString(buff, "\r\n")
	if err != nil {
		return nil, err
	}

	// Check to see if there is a Text or HTML field
	if len(e.Text) > 0 || len(e.HTML) > 0 {
		var subWriter *multipart.Writer

		if isMixed && isAlternative {
			// Create the multipart alternative part
			subWriter = multipart.NewWriter(buff)
			header := textproto.MIMEHeader{
				"Content-Type": {"multipart/alternative;\r\n boundary=" + subWriter.Boundary()},
			}
			if _, err := w.CreatePart(header); err != nil {
				return nil, err
			}
		} else {
			subWriter = w
		}
		// Create the body sections
		if len(e.Text) > 0 {
			// Write the text
			if err := writeMessage(buff, e.Text, isMixed || isAlternative, "text/plain", subWriter); err != nil {
				return nil, err
			}
		}
		if len(e.HTML) > 0 {
			messageWriter := subWriter
			var relatedWriter *multipart.Writer
			if (isMixed || isAlternative) && len(htmlAttachments) > 0 {
				relatedWriter = multipart.NewWriter(buff)
				header := textproto.MIMEHeader{
					"Content-Type": {"multipart/related;\r\n boundary=" + relatedWriter.Boundary()},
				}
				if _, err := subWriter.CreatePart(header); err != nil {
					return nil, err
				}

				messageWriter = relatedWriter
			} else if isRelated && len(htmlAttachments) > 0 {
				relatedWriter = w
				messageWriter = w
			}
			// Write the HTML
			if err := writeMessage(buff, e.HTML, isMixed || isAlternative || isRelated, "text/html", messageWriter); err != nil {
				return nil, err
			}
			if len(htmlAttachments) > 0 {
				for _, a := range htmlAttachments {
					a.setDefaultHeaders()
					ap, err := relatedWriter.CreatePart(a.Header)
					if err != nil {
						return nil, err
					}
					// Write the base64Wrapped content to the part
					base64Wrap(ap, a.Content)
				}

				if isMixed || isAlternative {
					relatedWriter.Close()
				}
			}
		}
		if isMixed && isAlternative {
			if err := subWriter.Close(); err != nil {
				return nil, err
			}
		}
	}
	// Create attachment part, if necessary
	for _, a := range otherAttachments {
		a.setDefaultHeaders()
		ap, err := w.CreatePart(a.Header)
		if err != nil {
			return nil, err
		}
		// Write the base64Wrapped content to the part
		base64Wrap(ap, a.Content)
	}
	if isMixed || isAlternative || isRelated {
		if err := w.Close(); err != nil {
			return nil, err
		}
	}
	return buff.Bytes(), nil
}

// Send 使用给定主机和SMTP身份验证(可选)的电子邮件，返回SMTP抛出的任何错误
// 这个函数合并 To, Cc, and Bcc 字段并调用 smtp.SendMail 方法，使用 Email.Bytes() 输出消息
func (e *EmailSmtp) Send(addr string, a smtp.Auth) error {
	// Merge the To, Cc, and Bcc fields
	to := make([]string, 0, len(e.To)+len(e.Cc)+len(e.Bcc))
	to = append(append(append(to, e.To...), e.Cc...), e.Bcc...)
	for i := 0; i < len(to); i++ {
		addr, err := mail.ParseAddress(to[i])
		if err != nil {
			return err
		}
		to[i] = addr.Address
	}
	// Check to make sure there is at least one recipient and one "From" address
	if e.From == "" || len(to) == 0 {
		return errors.New("必须至少指定一个“From地址”和一个“To地址")
	}
	sender, err := e.parseSender()
	if err != nil {
		return err
	}
	raw, err := e.Bytes()
	if err != nil {
		return err
	}
	return smtp.SendMail(addr, a, sender, to, raw)
}

// 选择并解析SMTP信封发件人地址。选择电子邮件。发件人(如果设置)，或回退到电子邮件。
func (e *EmailSmtp) parseSender() (string, error) {
	if e.Sender != "" {
		sender, err := mail.ParseAddress(e.Sender)
		if err != nil {
			return "", err
		}
		return sender.Address, nil
	} else {
		from, err := mail.ParseAddress(e.From)
		if err != nil {
			return "", err
		}
		return from.Address, nil
	}
}

// SendWithTLS 通过可选的tls密钥发送电子邮件。
// 如果您需要连接到使用不受信任证书的主机，TLS配置将非常有用。
func (e *EmailSmtp) SendWithTLS(addr string, a smtp.Auth, t *tls.Config) error {
	// Merge the To, Cc, and Bcc fields
	to := make([]string, 0, len(e.To)+len(e.Cc)+len(e.Bcc))
	to = append(append(append(to, e.To...), e.Cc...), e.Bcc...)
	for i := 0; i < len(to); i++ {
		addr, err := mail.ParseAddress(to[i])
		if err != nil {
			return err
		}
		to[i] = addr.Address
	}
	// Check to make sure there is at least one recipient and one "From" address
	if e.From == "" || len(to) == 0 {
		return errors.New("Must specify at least one From address and one To address")
	}
	sender, err := e.parseSender()
	if err != nil {
		return err
	}
	raw, err := e.Bytes()
	if err != nil {
		return err
	}

	conn, err := tls.Dial("tcp", addr, t)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, t.ServerName)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Hello("localhost"); err != nil {
		return err
	}

	if a != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(sender); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(raw)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

// SendWithStartTLS 使用带有可选TLS秘密的STARTTLS通过TLS发送电子邮件。
// 如果您需要连接到使用不受信任证书的主机，TLS配置将非常有用。
func (e *EmailSmtp) SendWithStartTLS(addr string, a smtp.Auth, t *tls.Config) error {
	// Merge the To, Cc, and Bcc fields
	to := make([]string, 0, len(e.To)+len(e.Cc)+len(e.Bcc))
	to = append(append(append(to, e.To...), e.Cc...), e.Bcc...)
	for i := 0; i < len(to); i++ {
		addr, err := mail.ParseAddress(to[i])
		if err != nil {
			return err
		}
		to[i] = addr.Address
	}
	// Check to make sure there is at least one recipient and one "From" address
	if e.From == "" || len(to) == 0 {
		return errors.New("Must specify at least one From address and one To address")
	}
	sender, err := e.parseSender()
	if err != nil {
		return err
	}
	raw, err := e.Bytes()
	if err != nil {
		return err
	}

	// Taken from the standard library
	// https://github.com/golang/go/blob/master/src/net/smtp/smtp.go#L328
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Hello("localhost"); err != nil {
		return err
	}
	// Use TLS if available
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err = c.StartTLS(t); err != nil {
			return err
		}
	}

	if a != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(sender); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(raw)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

// Attachment 是表示电子邮件附件的结构体。
type Attachment struct {
	Filename    string
	ContentType string
	Header      textproto.MIMEHeader
	Content     []byte
	HTMLRelated bool
}

// 设置邮件头
func (at *Attachment) setDefaultHeaders() {
	// 默认MIME类型
	contentType := "application/octet-stream"
	if len(at.ContentType) > 0 {
		contentType = at.ContentType
	}
	at.Header.Set("Content-Type", contentType)

	if len(at.Header.Get("Content-Disposition")) == 0 {
		disposition := "attachment"
		if at.HTMLRelated {
			disposition = "inline"
		}
		at.Header.Set("Content-Disposition", fmt.Sprintf("%s;\r\n filename=\"%s\"", disposition, at.Filename))
	}
	if len(at.Header.Get("Content-ID")) == 0 {
		at.Header.Set("Content-ID", fmt.Sprintf("<%s>", at.Filename))
	}
	if len(at.Header.Get("Content-Transfer-Encoding")) == 0 {
		at.Header.Set("Content-Transfer-Encoding", "base64")
	}
}

// base64Wrap 对附件内容进行编码，并根据RFC 2045标准对其进行包装(每76个字符)
// 结果将被输出到 io.Writer
func base64Wrap(w io.Writer, b []byte) {
	// 57 raw bytes per 76-byte base64 line.
	const maxRaw = 57
	// Buffer for each line, including trailing CRLF.
	buffer := make([]byte, MaxLineLength+len("\r\n"))
	copy(buffer[MaxLineLength:], "\r\n")
	// Process raw chunks until there's no longer enough to fill a line.
	for len(b) >= maxRaw {
		base64.StdEncoding.Encode(buffer, b[:maxRaw])
		w.Write(buffer)
		b = b[maxRaw:]
	}
	// Handle the last chunk of bytes.
	if len(b) > 0 {
		out := buffer[:base64.StdEncoding.EncodedLen(len(b))]
		base64.StdEncoding.Encode(out, b)
		out = append(out, "\r\n"...)
		w.Write(out)
	}
}

// headerToBytes 渲染 "header" to "buff".
// 如果一个字段有多个值，则会发出多个“field: value\r\n”行。
func headerToBytes(buff io.Writer, header textproto.MIMEHeader) {
	for field, vals := range header {
		for _, subval := range vals {
			// bytes.Buffer.Write() never returns an error.
			io.WriteString(buff, field)
			io.WriteString(buff, ": ")
			// Write the encoded header if needed
			switch {
			case field == "Content-Type" || field == "Content-Disposition":
				buff.Write([]byte(subval))
			case field == "From" || field == "To" || field == "Cc" || field == "Bcc":
				participants := strings.Split(subval, ",")
				for i, v := range participants {
					addr, err := mail.ParseAddress(v)
					if err != nil {
						continue
					}
					participants[i] = addr.String()
				}
				buff.Write([]byte(strings.Join(participants, ", ")))
			default:
				buff.Write([]byte(mime.QEncoding.Encode("UTF-8", subval)))
			}
			io.WriteString(buff, "\r\n")
		}
	}
}

var maxBigInt = big.NewInt(math.MaxInt64)

// generateMessageID 生成消息ID RFC 2822
// 以下参数用于生成 Message-ID:
// - 纪元后的纳秒
// - 进程PID
// - 随机数
// - 发送者主机名称
func generateMessageID() (string, error) {
	t := time.Now().UnixNano()
	pid := os.Getpid()
	rint, err := rand.Int(rand.Reader, maxBigInt)
	if err != nil {
		return "", err
	}
	h, err := os.Hostname()
	// If we can't get the hostname, we'll use localhost
	if err != nil {
		h = "localhost.localdomain"
	}
	msgid := fmt.Sprintf("<%d.%d.%d@%s>", t, pid, rint, h)
	return msgid, nil
}
