package runtime

import (
	"errors"
	"fmt"
)

type ChannelConn struct {
	ch chan []byte
}

func NewChannelConn(ch chan []byte) *ChannelConn {
	return &ChannelConn{ch: ch}
}

func (conn *ChannelConn) Write(val []byte) (int, error) {
	conn.ch <- val
	return len(val), nil
}

func (conn *ChannelConn) Read(buffer []byte) (int, error) {
	val := <-conn.ch
	if len(val) > len(buffer) {
		return -1, errors.New(fmt.Sprintf("buffer length not long"))
	}
	n := copy(buffer, val)
	return n, nil
}

func (conn *ChannelConn) Close() {
	close(conn.ch)
}
