package network

import "net"

type Listener struct {
	ln net.Listener
}

func Listen(addr string) (*Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Listener{ln: ln}, nil
}

func (l *Listener) Accept() (*MsgConn, error) {
	conn, err := l.ln.Accept()
	if err != nil {
		return nil, err
	}
	return NewMsgConn(conn), nil
}

func (l *Listener) Close() error {
	return l.ln.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.ln.Addr()
}
