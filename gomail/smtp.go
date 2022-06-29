package gomail

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Dialer SMTP的邮件拨号器
type Dialer struct {
	Host      string      // SMTP服务地址
	Port      int         // SMTP端口号
	Username  string      // 用户名
	Password  string      // 密码
	Auth      smtp.Auth   // 权限
	SSL       bool        // 是否开启SSL
	TLSConfig *tls.Config // SSL配置
	LocalName string      // 本地服务名称，默认“localhost”
}

// NewDialer 创建拨号器
func NewDialer(host string, port int, username, password string) *Dialer {
	return &Dialer{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		SSL:      port == 465,
	}
}

// Dial 使用拨号器进行拨号
func (d *Dialer) Dial() (sender SendCloser, err error) {
	var (
		conn net.Conn
		c    smtpClient
	)

	// 拨号并进行超时控制
	conn, err = netDialTimeout("tcp", addr(d.Host, d.Port), 10*time.Second)
	if err != nil {
		return
	}

	// 如果开启了SSL，进行https连接
	if d.SSL {
		conn = tlsClient(conn, d.tlsConfig())
	}

	// 创建SMTP连接
	c, err = smtpNewClient(conn, d.Host)
	if err != nil {
		return
	}

	// 如果定义了本地名称
	if d.LocalName != "" {
		// 使用SMTP连接打招呼，查看是否能通信
		if err = c.Hello(d.LocalName); err != nil {
			return
		}
	}

	// 如果没有启用SSL
	if !d.SSL {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err = c.StartTLS(d.tlsConfig()); err != nil {
				c.Close()
				return nil, err
			}
		}
	}

	if d.Auth == nil && d.Username != "" {
		if ok, auths := c.Extension("AUTH"); ok {
			if strings.Contains(auths, "CRAM-MD5") {
				d.Auth = smtp.CRAMMD5Auth(d.Username, d.Password)
			} else if strings.Contains(auths, "LOGIN") &&
				!strings.Contains(auths, "PLAIN") {
				d.Auth = &loginAuth{
					username: d.Username,
					password: d.Password,
					host:     d.Host,
				}
			} else {
				d.Auth = smtp.PlainAuth("", d.Username, d.Password, d.Host)
			}
		}
	}

	if d.Auth != nil {
		if err = c.Auth(d.Auth); err != nil {
			c.Close()
			return nil, err
		}
	}

	return &smtpSender{c, d}, nil
}

func (d *Dialer) tlsConfig() *tls.Config {
	if d.TLSConfig == nil {
		return &tls.Config{ServerName: d.Host}
	}
	return d.TLSConfig
}

// 拼接地址
func addr(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

// DialAndSend opens a connection to the SMTP server, sends the given emails and
// closes the connection.
func (d *Dialer) DialAndSend(m ...*Message) error {
	s, err := d.Dial()
	if err != nil {
		return err
	}
	defer s.Close()

	return Send(s, m...)
}

type smtpSender struct {
	smtpClient
	d *Dialer
}

func (c *smtpSender) Send(from string, to []string, msg io.WriterTo) error {
	if err := c.Mail(from); err != nil {
		if err == io.EOF {
			// This is probably due to a timeout, so reconnect and try again.
			sc, derr := c.d.Dial()
			if derr == nil {
				if s, ok := sc.(*smtpSender); ok {
					*c = *s
					return c.Send(from, to, msg)
				}
			}
		}
		return err
	}

	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	if _, err = msg.WriteTo(w); err != nil {
		w.Close()
		return err
	}

	return w.Close()
}

func (c *smtpSender) Close() error {
	return c.Quit()
}

// Stubbed out for tests.
var (
	netDialTimeout = net.DialTimeout
	tlsClient      = tls.Client
	smtpNewClient  = func(conn net.Conn, host string) (smtpClient, error) {
		return smtp.NewClient(conn, host)
	}
)

// SMTP连接对象
type smtpClient interface {
	Hello(string) error // 打招呼
	Extension(string) (bool, string)
	StartTLS(*tls.Config) error
	Auth(smtp.Auth) error
	Mail(string) error
	Rcpt(string) error
	Data() (io.WriteCloser, error)
	Quit() error
	Close() error
}
