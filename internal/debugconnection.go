package internal

import (
	"fmt"
	"log/slog"
	"net"
)

type DebugConnection struct {
	Name string
	net.Conn
}

func WrapWithDebugConnection(name string, conn net.Conn) net.Conn {
	return &DebugConnection{
		Name: name,
		Conn: conn,
	}
}

func (l *DebugConnection) Read(b []byte) (int, error) {
	n, err := l.Conn.Read(b)
	if err != nil {
		return n, err
	}
	slog.Debug("Read "+fmt.Sprint(n)+" bytes from "+l.Name, "length", n)
	return n, nil
}

func (l *DebugConnection) Write(b []byte) (int, error) {
	n := len(b)
	slog.Debug("Write "+fmt.Sprint(n)+" bytes to "+l.Name, "length", n)
	return l.Conn.Write(b)
}
