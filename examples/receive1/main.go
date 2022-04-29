package main

import (
	"fmt"
	"github.com/zhangdapeng520/zdpgo_email/message/mail"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/zhangdapeng520/zdpgo_email"
	"github.com/zhangdapeng520/zdpgo_email/id"
	"github.com/zhangdapeng520/zdpgo_email/imap"
	_ "github.com/zhangdapeng520/zdpgo_email/message/charset"
)

// 登录函数
func loginEmail(Eserver, UserName, Password string) (*zdpgo_email.EmailImap, error) {
	dial := new(net.Dialer)
	dial.Timeout = time.Duration(3) * time.Second
	c, err := zdpgo_email.DialWithDialerTLS(dial, Eserver, nil)
	if err != nil {
		c, err = zdpgo_email.DialWithDialer(dial, Eserver) // 非加密登录
	}
	if err != nil {
		return nil, err
	}
	//登陆
	if err = c.Login(UserName, Password); err != nil {
		return nil, err
	}
	return c, nil
}

func parseEmail(mr *mail.Reader) (body []byte, fileMap map[string][]byte) {
	fileMap = make(map[string][]byte)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			break
		}
		if p != nil {
			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				body, err = ioutil.ReadAll(p.Body)
				if err != nil {
					fmt.Println("read body err:", err.Error())
				}

			case *mail.AttachmentHeader:
				fileName, _ := h.Filename()
				fileContent, _ := ioutil.ReadAll(p.Body)
				fileMap[fileName] = fileContent
			}
		}
	}
	return
}

func emailListByUid(Eserver, UserName, Password string) (err error) {
	c, err := loginEmail(Eserver, UserName, Password)
	if err != nil {
		fmt.Println("login:", err)
		return
	}
	idClient := id.NewClient(c)
	idClient.ID(
		id.ID{
			id.FieldName:    "IMAPClient",
			id.FieldVersion: "2.1.0",
		},
	)

	defer c.Close()
	mailboxes := make(chan *imap.MailboxInfo, 10)
	mailboxeDone := make(chan error, 1)
	go func() {
		mailboxeDone <- c.List("", "*", mailboxes)
	}()

	// 遍历所有的目录
	for box := range mailboxes {
		// 切换目录
		fmt.Println("切换目录:", box.Name)
		mbox, err := c.Select(box.Name, false)

		// 选择收件箱
		if err != nil {
			fmt.Println("select inbox err: ", err)
			continue
		}
		if mbox.Messages == 0 {
			continue
		}

		// 选择收取邮件的时间段
		criteria := imap.NewSearchCriteria()

		// 收取7天之内的邮件
		criteria.Since = time.Now().Add(-7 * time.Hour * 24)

		// 按条件查询邮件
		ids, err := c.UidSearch(criteria)
		fmt.Println(len(ids))
		if err != nil {
			continue
		}
		if len(ids) == 0 {
			continue
		}
		seqset := new(imap.SeqSet)
		seqset.AddNum(ids...)
		sect := &imap.BodySectionName{Peek: true}

		messages := make(chan *imap.Message, 100)
		messageDone := make(chan error, 1)

		// 查找所有的邮件
		go func() {
			messageDone <- c.UidFetch(seqset, []imap.FetchItem{sect.FetchItem()}, messages)
		}()
		for msg := range messages {
			r := msg.GetBody(sect)
			mr, err := mail.CreateReader(r)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// 提取邮件标题
			header := mr.Header
			fmt.Println(header.Subject())

			// 提取邮件附件
			_, fileName := parseEmail(mr)
			for k, _ := range fileName {
				fmt.Println("收取到附件:", k)
			}
		}
	}
	return
}

func main() {
	emailListByUid(server, email, password)
}
