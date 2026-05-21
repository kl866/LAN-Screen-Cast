package network

import (
	"encoding/binary"
	"io"
	"net"

	"lan-screen-cast/internal/protocol"
)

type MsgConn struct {
	conn net.Conn
}

func NewMsgConn(conn net.Conn) *MsgConn {
	return &MsgConn{conn: conn}
}

func (m *MsgConn) Conn() net.Conn { return m.conn }

func (m *MsgConn) Send(typ byte, payload []byte) error {
	frame := protocol.EncodeMessage(typ, payload)
	_, err := m.conn.Write(frame)
	return err
}

func (m *MsgConn) Read() (typ byte, payload []byte, err error) {
	header := make([]byte, protocol.HeaderSize)
	if _, err := io.ReadFull(m.conn, header); err != nil {
		return 0, nil, err
	}
	if string(header[:4]) != string(protocol.Magic[:]) {
		return 0, nil, protocol.ErrBadMagic
	}
	typ = header[4]
	payLen := binary.BigEndian.Uint32(header[5:9])
	if payLen > 100*1024*1024 {
		return 0, nil, protocol.ErrPayloadTooLarge
	}
	payload = make([]byte, payLen)
	if _, err := io.ReadFull(m.conn, payload); err != nil {
		return 0, nil, err
	}
	return typ, payload, nil
}

func (m *MsgConn) Close() error {
	return m.conn.Close()
}
