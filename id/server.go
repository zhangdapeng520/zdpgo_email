package id

import (
	"github.com/zhangdapeng520/zdpgo_email/imap/server"
)

type Conn interface {
	ID() ID

	setID(id ID)
}

type conn struct {
	server.Conn

	id ID
}

func (conn *conn) ID() ID {
	return conn.id
}

func (conn *conn) setID(id ID) {
	conn.id = id
}

type Handler struct {
	Command

	ext *extension
}

func (hdlr *Handler) Handle(conn server.Conn) error {
	if conn, ok := conn.(Conn); ok {
		conn.setID(hdlr.Command.ID)
	}

	return conn.WriteResp(&Response{hdlr.ext.serverID})
}

type extension struct {
	serverID ID
}

func (ext *extension) Capabilities(c server.Conn) []string {
	return []string{Capability}
}

func (ext *extension) Command(name string) server.HandlerFactory {
	if name != commandName {
		return nil
	}

	return func() server.Handler {
		return &Handler{ext: ext}
	}
}

func (ext *extension) NewConn(c server.Conn) server.Conn {
	return &conn{Conn: c}
}

func NewExtension(serverID ID) server.Extension {
	return &extension{serverID}
}
