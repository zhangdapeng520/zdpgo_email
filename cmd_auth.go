package zdpgo_email

import (
	"errors"
	"time"

	"github.com/zhangdapeng520/zdpgo_email/imap"
	"github.com/zhangdapeng520/zdpgo_email/imap/commands"
	"github.com/zhangdapeng520/zdpgo_email/imap/responses"
)

// ErrNotLoggedIn is returned if a function that requires the client to be
// logged in is called then the client isn't.
var ErrNotLoggedIn = errors.New("Not logged in")

func (c *EmailImap) ensureAuthenticated() error {
	state := c.State()
	if state != imap.AuthenticatedState && state != imap.SelectedState {
		return ErrNotLoggedIn
	}
	return nil
}

// Select 选择一个邮箱，以便可以访问该邮箱中的消息。
// 在尝试新的选择之前，将取消当前选中的任何邮箱。
// 即使readOnly参数设置为false，服务器也可以决定以只读模式打开邮箱。
func (c *EmailImap) Select(name string, readOnly bool) (*imap.MailboxStatus, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	cmd := &commands.Select{
		Mailbox:  name,
		ReadOnly: readOnly,
	}

	mbox := &imap.MailboxStatus{Name: name, Items: make(map[imap.StatusItem]interface{})}
	res := &responses.Select{
		Mailbox: mbox,
	}
	c.locker.Lock()
	c.mailbox = mbox
	c.locker.Unlock()

	status, err := c.execute(cmd, res)
	if err != nil {
		c.locker.Lock()
		c.mailbox = nil
		c.locker.Unlock()
		return nil, err
	}
	if err := status.Err(); err != nil {
		c.locker.Lock()
		c.mailbox = nil
		c.locker.Unlock()
		return nil, err
	}

	c.locker.Lock()
	mbox.ReadOnly = (status.Code == imap.CodeReadOnly)
	c.state = imap.SelectedState
	c.locker.Unlock()
	return mbox, nil
}

// Create 创建具有给定名称的邮箱。
func (c *EmailImap) Create(name string) error {
	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	cmd := &commands.Create{
		Mailbox: name,
	}

	status, err := c.execute(cmd, nil)
	if err != nil {
		return err
	}
	return status.Err()
}

// Delete 永久删除具有给定名称的邮箱。
func (c *EmailImap) Delete(name string) error {
	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	cmd := &commands.Delete{
		Mailbox: name,
	}

	status, err := c.execute(cmd, nil)
	if err != nil {
		return err
	}
	return status.Err()
}

// Rename 更改邮箱的名称。
func (c *EmailImap) Rename(existingName, newName string) error {
	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	cmd := &commands.Rename{
		Existing: existingName,
		New:      newName,
	}

	status, err := c.execute(cmd, nil)
	if err != nil {
		return err
	}
	return status.Err()
}

// Subscribe 将指定的邮箱名称添加到服务器的“活动”或“订阅”邮箱集。
func (c *EmailImap) Subscribe(name string) error {
	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	cmd := &commands.Subscribe{
		Mailbox: name,
	}

	status, err := c.execute(cmd, nil)
	if err != nil {
		return err
	}
	return status.Err()
}

// Unsubscribe 从服务器的“活动”或“订阅”邮箱集中删除指定的邮箱名称。
func (c *EmailImap) Unsubscribe(name string) error {
	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	cmd := &commands.Unsubscribe{
		Mailbox: name,
	}

	status, err := c.execute(cmd, nil)
	if err != nil {
		return err
	}
	return status.Err()
}

// List 从客户端可用的所有名称的完整集合中返回名称的子集。
// 空的name参数是一个特殊的请求，要求返回层次结构分隔符和引用中给出的名称的根名称。
// 字符“*”是一个通配符，在这个位置匹配零个或多个字符。字符“%”类似于“*”，但它不匹配层次分隔符。
func (c *EmailImap) List(ref, name string, ch chan *imap.MailboxInfo) error {
	defer close(ch)

	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	cmd := &commands.List{
		Reference: ref,
		Mailbox:   name,
	}
	res := &responses.List{Mailboxes: ch}

	status, err := c.execute(cmd, res)
	if err != nil {
		return err
	}
	return status.Err()
}

// Lsub 从用户声明为“活动”或“订阅”的名称集中返回名称的子集。
func (c *EmailImap) Lsub(ref, name string, ch chan *imap.MailboxInfo) error {
	defer close(ch)

	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	cmd := &commands.List{
		Reference:  ref,
		Mailbox:    name,
		Subscribed: true,
	}
	res := &responses.List{
		Mailboxes:  ch,
		Subscribed: true,
	}

	status, err := c.execute(cmd, res)
	if err != nil {
		return err
	}
	return status.Err()
}

// Status requests the status of the indicated mailbox. It does not change the
// currently selected mailbox, nor does it affect the state of any messages in
// the queried mailbox.
//
// See RFC 3501 section 6.3.10 for a list of items that can be requested.
func (c *EmailImap) Status(name string, items []imap.StatusItem) (*imap.MailboxStatus, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}

	cmd := &commands.Status{
		Mailbox: name,
		Items:   items,
	}
	res := &responses.Status{
		Mailbox: new(imap.MailboxStatus),
	}

	status, err := c.execute(cmd, res)
	if err != nil {
		return nil, err
	}
	return res.Mailbox, status.Err()
}

// Append appends the literal argument as a new message to the end of the
// specified destination mailbox. This argument SHOULD be in the format of an
// RFC 2822 message. flags and date are optional arguments and can be set to
// nil and the empty struct.
func (c *EmailImap) Append(mbox string, flags []string, date time.Time, msg imap.Literal) error {
	if err := c.ensureAuthenticated(); err != nil {
		return err
	}

	cmd := &commands.Append{
		Mailbox: mbox,
		Flags:   flags,
		Date:    date,
		Message: msg,
	}

	status, err := c.execute(cmd, nil)
	if err != nil {
		return err
	}
	return status.Err()
}

// Enable requests the server to enable the named extensions. The extensions
// which were successfully enabled are returned.
//
// See RFC 5161 section 3.1.
func (c *EmailImap) Enable(caps []string) ([]string, error) {
	if ok, err := c.Support("ENABLE"); !ok || err != nil {
		return nil, ErrExtensionUnsupported
	}

	// ENABLE is invalid if a mailbox has been selected.
	if c.State() != imap.AuthenticatedState {
		return nil, ErrNotLoggedIn
	}

	cmd := &commands.Enable{Caps: caps}
	res := &responses.Enabled{}

	if status, err := c.Execute(cmd, res); err != nil {
		return nil, err
	} else {
		return res.Caps, status.Err()
	}
}

func (c *EmailImap) idle(stop <-chan struct{}) error {
	cmd := &commands.Idle{}

	res := &responses.Idle{
		Stop:      stop,
		RepliesCh: make(chan []byte, 10),
	}

	if status, err := c.Execute(cmd, res); err != nil {
		return err
	} else {
		return status.Err()
	}
}

// IdleOptions holds options for Client.Idle.
type IdleOptions struct {
	// LogoutTimeout is used to avoid being logged out by the server when
	// idling. Each LogoutTimeout, the IDLE command is restarted. If set to
	// zero, a default is used. If negative, this behavior is disabled.
	LogoutTimeout time.Duration
	// Poll interval when the server doesn't support IDLE. If zero, a default
	// is used. If negative, polling is always disabled.
	PollInterval time.Duration
}

// Idle indicates to the server that the client is ready to receive unsolicited
// mailbox update messages. When the client wants to send commands again, it
// must first close stop.
//
// If the server doesn't support IDLE, go-imap falls back to polling.
func (c *EmailImap) Idle(stop <-chan struct{}, opts *IdleOptions) error {
	if ok, err := c.Support("IDLE"); err != nil {
		return err
	} else if !ok {
		return c.idleFallback(stop, opts)
	}

	logoutTimeout := 25 * time.Minute
	if opts != nil {
		if opts.LogoutTimeout > 0 {
			logoutTimeout = opts.LogoutTimeout
		} else if opts.LogoutTimeout < 0 {
			return c.idle(stop)
		}
	}

	t := time.NewTicker(logoutTimeout)
	defer t.Stop()

	for {
		stopOrRestart := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			done <- c.idle(stopOrRestart)
		}()

		select {
		case <-t.C:
			close(stopOrRestart)
			if err := <-done; err != nil {
				return err
			}
		case <-stop:
			close(stopOrRestart)
			return <-done
		case err := <-done:
			close(stopOrRestart)
			if err != nil {
				return err
			}
		}
	}
}

func (c *EmailImap) idleFallback(stop <-chan struct{}, opts *IdleOptions) error {
	pollInterval := time.Minute
	if opts != nil {
		if opts.PollInterval > 0 {
			pollInterval = opts.PollInterval
		} else if opts.PollInterval < 0 {
			return ErrExtensionUnsupported
		}
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := c.Noop(); err != nil {
				return err
			}
		case <-stop:
			return nil
		case <-c.LoggedOut():
			return errors.New("disconnected while idling")
		}
	}
}
