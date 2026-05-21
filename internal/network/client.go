package network

import "net"

func Dial(addr string) (*MsgConn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewMsgConn(conn), nil
}
