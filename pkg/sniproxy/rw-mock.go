package sniproxy

import (
	"io"
	"net"
	"time"
)

// writeMockingConn is a fake wrapper for an io.ReadWriter that fakes all writes to keep a connectionread only
type writeMockingConn struct {
	reader io.Reader
}

func (conn writeMockingConn) Read(p []byte) (int, error)         { return conn.reader.Read(p) }
func (conn writeMockingConn) Write(p []byte) (int, error)        { return len(p), nil }
func (conn writeMockingConn) Close() error                       { return nil }
func (conn writeMockingConn) LocalAddr() net.Addr                { return nil }
func (conn writeMockingConn) RemoteAddr() net.Addr               { return nil }
func (conn writeMockingConn) SetDeadline(t time.Time) error      { return nil }
func (conn writeMockingConn) SetReadDeadline(t time.Time) error  { return nil }
func (conn writeMockingConn) SetWriteDeadline(t time.Time) error { return nil }
