package gomail

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Message 准备邮件内容
type Message struct {
	header      header
	parts       []*part
	attachments []*AttachmentFile // 附件列表
	embedded    []*AttachmentFile
	charset     string
	encoding    Encoding
	hEncoder    mimeEncoder
	buf         bytes.Buffer
	Fs          *embed.FS           // 嵌入文件系统
	Readers     map[string]*os.File // 读取器列表
}

type header map[string][]string

type part struct {
	contentType string
	copier      func(io.Writer) error
	encoding    Encoding
}

// NewMessage 创建一条消息，默认使用UTF-8编码
func NewMessage(settings ...MessageSetting) *Message {
	m := &Message{
		header:   make(header),
		charset:  "UTF-8",
		encoding: QuotedPrintable,
	}

	// 应用设置
	m.applySettings(settings)

	// 编码
	if m.encoding == Base64 {
		m.hEncoder = bEncoding
	} else {
		m.hEncoder = qEncoding
	}
	return m
}

// NewMessageWithFs 使用嵌入文件系统创建消息
func NewMessageWithFs(fs *embed.FS, settings ...MessageSetting) *Message {
	m := NewMessage(settings...)
	m.Fs = fs
	return m
}

// NewMessageWithReaders 使用读取器列表创建消息
func NewMessageWithReaders(readers map[string]*os.File, settings ...MessageSetting) *Message {
	m := NewMessage(settings...)
	m.Readers = readers
	return m
}

// Reset resets the message so it can be reused. The message keeps its previous
// settings so it is in the same state that after a call to NewMessage.
func (m *Message) Reset() {
	for k := range m.header {
		delete(m.header, k)
	}
	m.parts = nil
	m.attachments = nil
	m.embedded = nil
}

func (m *Message) applySettings(settings []MessageSetting) {
	for _, s := range settings {
		s(m)
	}
}

// A MessageSetting can be used as an argument in NewMessage to configure an
// email.
type MessageSetting func(m *Message)

// SetCharset is a message setting to set the charset of the email.
func SetCharset(charset string) MessageSetting {
	return func(m *Message) {
		m.charset = charset
	}
}

// SetEncoding is a message setting to set the encoding of the email.
func SetEncoding(enc Encoding) MessageSetting {
	return func(m *Message) {
		m.encoding = enc
	}
}

// Encoding represents a MIME encoding scheme like quoted-printable or base64.
type Encoding string

const (
	// QuotedPrintable represents the quoted-printable encoding as defined in
	// RFC 2045.
	QuotedPrintable Encoding = "quoted-printable"
	// Base64 represents the base64 encoding as defined in RFC 2045.
	Base64 Encoding = "base64"
	// Unencoded can be used to avoid encoding the body of an email. The headers
	// will still be encoded using quoted-printable encoding.
	Unencoded Encoding = "8bit"
)

// SetHeader sets a value to the given header field.
func (m *Message) SetHeader(field string, value ...string) {
	m.encodeHeader(value)
	m.header[field] = value
}

func (m *Message) encodeHeader(values []string) {
	for i := range values {
		values[i] = m.encodeString(values[i])
	}
}

func (m *Message) encodeString(value string) string {
	return m.hEncoder.Encode(m.charset, value)
}

// SetHeaders sets the message headers.
func (m *Message) SetHeaders(h map[string][]string) {
	for k, v := range h {
		m.SetHeader(k, v...)
	}
}

// SetAddressHeader sets an address to the given header field.
func (m *Message) SetAddressHeader(field, address, name string) {
	m.header[field] = []string{m.FormatAddress(address, name)}
}

// FormatAddress formats an address and a name as a valid RFC 5322 address.
func (m *Message) FormatAddress(address, name string) string {
	if name == "" {
		return address
	}

	enc := m.encodeString(name)
	if enc == name {
		m.buf.WriteByte('"')
		for i := 0; i < len(name); i++ {
			b := name[i]
			if b == '\\' || b == '"' {
				m.buf.WriteByte('\\')
			}
			m.buf.WriteByte(b)
		}
		m.buf.WriteByte('"')
	} else if hasSpecials(name) {
		m.buf.WriteString(bEncoding.Encode(m.charset, name))
	} else {
		m.buf.WriteString(enc)
	}
	m.buf.WriteString(" <")
	m.buf.WriteString(address)
	m.buf.WriteByte('>')

	addr := m.buf.String()
	m.buf.Reset()
	return addr
}

func hasSpecials(text string) bool {
	for i := 0; i < len(text); i++ {
		switch c := text[i]; c {
		case '(', ')', '<', '>', '[', ']', ':', ';', '@', '\\', ',', '.', '"':
			return true
		}
	}

	return false
}

// SetDateHeader sets a date to the given header field.
func (m *Message) SetDateHeader(field string, date time.Time) {
	m.header[field] = []string{m.FormatDate(date)}
}

// FormatDate formats a date as a valid RFC 5322 date.
func (m *Message) FormatDate(date time.Time) string {
	return date.Format(time.RFC1123Z)
}

// GetHeader gets a header field.
func (m *Message) GetHeader(field string) []string {
	return m.header[field]
}

// SetBody sets the body of the message. It replaces any content previously set
// by SetBody, AddAlternative or AddAlternativeWriter.
func (m *Message) SetBody(contentType, body string, settings ...PartSetting) {
	m.parts = []*part{m.newPart(contentType, newCopier(body), settings)}
}

// AddAlternative adds an alternative part to the message.
//
// It is commonly used to send HTML emails that default to the plain text
// version for backward compatibility. AddAlternative appends the new part to
// the end of the message. So the plain text part should be added before the
// HTML part. See http://en.wikipedia.org/wiki/MIME#Alternative
func (m *Message) AddAlternative(contentType, body string, settings ...PartSetting) {
	m.AddAlternativeWriter(contentType, newCopier(body), settings...)
}

func newCopier(s string) func(io.Writer) error {
	return func(w io.Writer) error {
		_, err := io.WriteString(w, s)
		return err
	}
}

// AddAlternativeWriter adds an alternative part to the message. It can be
// useful with the text/template or html/template packages.
func (m *Message) AddAlternativeWriter(contentType string, f func(io.Writer) error, settings ...PartSetting) {
	m.parts = append(m.parts, m.newPart(contentType, f, settings))
}

func (m *Message) newPart(contentType string, f func(io.Writer) error, settings []PartSetting) *part {
	p := &part{
		contentType: contentType,
		copier:      f,
		encoding:    m.encoding,
	}

	for _, s := range settings {
		s(p)
	}

	return p
}

// A PartSetting can be used as an argument in Message.SetBody,
// Message.AddAlternative or Message.AddAlternativeWriter to configure the part
// added to a message.
type PartSetting func(*part)

// SetPartEncoding sets the encoding of the part added to the message. By
// default, parts use the same encoding than the message.
func SetPartEncoding(e Encoding) PartSetting {
	return PartSetting(func(p *part) {
		p.encoding = e
	})
}

// 文件对象
type AttachmentFile struct {
	Name     string                  // 名称
	Header   map[string][]string     // 请求头
	CopyFunc func(w io.Writer) error // 复制函数
}

func (f *AttachmentFile) setHeader(field, value string) {
	f.Header[field] = []string{value}
}

// A FileSetting can be used as an argument in Message.Attach or Message.Embed.
type FileSetting func(*AttachmentFile)

// SetHeader is a file setting to set the MIME header of the message part that
// contains the file content.
//
// Mandatory headers are automatically added if they are not set when sending
// the email.
func SetHeader(h map[string][]string) FileSetting {
	return func(f *AttachmentFile) {
		for k, v := range h {
			f.Header[k] = v
		}
	}
}

// Rename is a file setting to set the name of the attachment if the name is
// different than the filename on disk.
func Rename(name string) FileSetting {
	return func(f *AttachmentFile) {
		f.Name = name
	}
}

// SetCopyFunc is a file setting to replace the function that runs when the
// message is sent. It should copy the content of the file to the io.Writer.
//
// The default copy function opens the file with the given filename, and copy
// its content to the io.Writer.
func SetCopyFunc(f func(io.Writer) error) FileSetting {
	return func(fi *AttachmentFile) {
		fi.CopyFunc = f
	}
}

// appendFile 追加文件
func (m *Message) appendFile(list []*AttachmentFile, name string, settings []FileSetting) []*AttachmentFile {
	f := &AttachmentFile{
		Name:   filepath.Base(name),
		Header: make(map[string][]string),
		CopyFunc: func(w io.Writer) error {
			h, err := os.Open(name)
			if err != nil {
				return err
			}
			if _, err := io.Copy(w, h); err != nil {
				h.Close()
				return err
			}
			return h.Close()
		},
	}

	for _, s := range settings {
		s(f)
	}

	if list == nil {
		return []*AttachmentFile{f}
	}

	return append(list, f)
}

// appendFileWithFs 使用嵌入文件系统中的文件追加
func (m *Message) appendFileWithFs(fs *embed.FS, list []*AttachmentFile, name string, settings []FileSetting) []*AttachmentFile {
	f := &AttachmentFile{
		Name:   filepath.Base(name),
		Header: make(map[string][]string),
		CopyFunc: func(w io.Writer) error {
			h, err := fs.Open(name)
			if err != nil {
				return err
			}
			if _, err := io.Copy(w, h); err != nil {
				h.Close()
				return err
			}
			return h.Close()
		},
	}

	for _, s := range settings {
		s(f)
	}

	if list == nil {
		return []*AttachmentFile{f}
	}

	return append(list, f)
}

// AddAttachmentFile 添加附件
func (m *Message) AddAttachmentFile(fileName string, settings []FileSetting) {
	f := &AttachmentFile{
		Name:   filepath.Base(fileName),
		Header: make(map[string][]string),
		CopyFunc: func(w io.Writer) error {
			h, err := os.Open(fileName)
			if err != nil {
				return err
			}
			if _, err = io.Copy(w, h); err != nil {
				h.Close()
				return err
			}
			return h.Close()
		},
	}

	for _, s := range settings {
		s(f)
	}
	m.attachments = append(m.attachments, f)
}

// AddAttachmentFileObj 添加文件对象作为附件
func (m *Message) AddAttachmentFileObj(fileName string, fileObj *os.File) {
	f := &AttachmentFile{
		Name:   filepath.Base(fileName),
		Header: make(map[string][]string),
		CopyFunc: func(w io.Writer) error {
			if _, err := io.Copy(w, fileObj); err != nil {
				fileObj.Close()
				return err
			}
			return fileObj.Close()
		},
	}
	m.attachments = append(m.attachments, f)
}

// appendFileWithReaders 使用读取器列表中的文件追加
func (m *Message) appendFileWithReaders(readers map[string]*os.File, list []*AttachmentFile, name string,
	settings []FileSetting) []*AttachmentFile {
	f := &AttachmentFile{
		Name:   filepath.Base(name),
		Header: make(map[string][]string),
		CopyFunc: func(w io.Writer) error {
			if h, ok := readers[name]; ok { // 获取指定文件的reader
				if _, err := io.Copy(w, h); err != nil {
					h.Close()
					return err
				}
				return h.Close()
			}
			return errors.New(fmt.Sprintf("文件不存在：%s", name))
		},
	}

	for _, s := range settings {
		s(f)
	}

	if list == nil {
		return []*AttachmentFile{f}
	}

	return append(list, f)
}

// appendFileWithFiles 使用文件列表追加附件列表
func (m *Message) appendFileWithFiles(files map[string]*os.File, list []*AttachmentFile, name string,
	settings []FileSetting) []*AttachmentFile {

	// 创建文件
	f := &AttachmentFile{
		Name:   filepath.Base(name),       // 文件名
		Header: make(map[string][]string), // 请求头
		CopyFunc: func(w io.Writer) error { // 复制方法
			if h, ok := files[name]; ok { // 获取指定文件的reader
				if _, err := io.Copy(w, h); err != nil {
					h.Close()
					return err
				}
				return h.Close()
			}
			return errors.New(fmt.Sprintf("文件不存在：%s", name))
		},
	}

	// 设置文件
	for _, s := range settings {
		s(f)
	}

	// 文件列表为空
	if list == nil {
		return []*AttachmentFile{f}
	}

	// 追加文件到文件列表
	return append(list, f)
}

// Attach 添加邮件附件
func (m *Message) Attach(filename string, settings ...FileSetting) {
	m.attachments = m.appendFile(m.attachments, filename, settings)
}

// AttachWithFs 使用嵌入文件系统作为附件
func (m *Message) AttachWithFs(fs *embed.FS, filename string, settings ...FileSetting) {
	m.attachments = m.appendFileWithFs(fs, m.attachments, filename, settings)
}

// AttachWithReaders 使用读取器列表作为附件
func (m *Message) AttachWithReaders(readers map[string]*os.File, filename string, settings ...FileSetting) {
	m.attachments = m.appendFileWithReaders(readers, m.attachments, filename, settings)
}

// AttachWithFiles 使用文件列表追加附件
func (m *Message) AttachWithFiles(files map[string]*os.File, filename string, settings ...FileSetting) {
	m.attachments = m.appendFileWithFiles(files, m.attachments, filename, settings)
}

// Embed embeds the images to the email.
func (m *Message) Embed(filename string, settings ...FileSetting) {
	m.embedded = m.appendFile(m.embedded, filename, settings)
}
