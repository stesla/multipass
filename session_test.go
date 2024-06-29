package main

import (
	"testing"

	"github.com/stesla/telnet"
	"github.com/stretchr/testify/mock"
)

func TestNegotiateOptions(t *testing.T) {
	conn := telnet.NewMockConn(t)
	session := newSession(conn)

	conn.EXPECT().AddListener("update-option", mock.Anything).Once()
	conn.EXPECT().BindOption(mock.Anything).Times(3)

	options := []byte{
		telnet.TransmitBinary,
		telnet.Charset,
		telnet.SuppressGoAhead,
	}
	for _, option := range options {
		conn.EXPECT().EnableOptionForThem(option, true).Return(nil).Once()
		conn.EXPECT().EnableOptionForUs(option, true).Return(nil).Once()
	}

	session.negotiateOptions()
}
