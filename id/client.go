package id

import (
	"errors"
	"github.com/zhangdapeng520/zdpgo_email"
	"github.com/zhangdapeng520/zdpgo_email/imap"
)

// Client is an ID client.
type Client struct {
	c *zdpgo_email.EmailImap
}

// NewClient creates a new client.
func NewClient(c *zdpgo_email.EmailImap) *Client {
	return &Client{c: c}
}

// SupportID checks if the server supports the ID extension.
func (c *Client) SupportID() (bool, error) {
	return c.c.Support(Capability)
}

// ID sends an ID command to the server and returns the server's ID.
func (c *Client) ID(clientID ID) (serverID ID, err error) {
	if state := c.c.State(); imap.ConnectedState&state != state {
		return nil, errors.New("Not connected")
	}

	var cmd imap.Commander = &Command{ID: clientID}

	res := &Response{}
	status, err := c.c.Execute(cmd, res)
	if err != nil {
		return
	}
	if err = status.Err(); err != nil {
		return
	}

	serverID = res.ID

	return
}
