package main

import (
	"bufio"

	"github.com/stesla/telnet"
	"golang.org/x/text/encoding/unicode"
)

type session struct {
	conn telnet.Conn
}

func newSession(conn telnet.Conn) *session {
	return &session{
		conn: conn,
	}
}

func (s *session) handleLine(line string) (err error) {
	line = "ECHO: " + line
	_, err = s.conn.Write([]byte(line))
	return
}

func (s *session) negotiateOptions() error {
	s.conn.AddListener("update-option", telnet.FuncListener{
		Func: func(event any) {
			switch t := event.(type) {
			case telnet.UpdateOptionEvent:
				switch opt := t.Option; opt.Byte() {
				case telnet.Charset:
					if t.WeChanged && opt.EnabledForUs() {
						t.Conn().RequestEncoding(unicode.UTF8)
					}
				}
			}
		},
	})

	for _, opt := range []telnet.Option{
		telnet.NewSuppressGoAheadOption(),
		telnet.NewTransmitBinaryOption(),
		telnet.NewCharsetOption(true),
	} {
		opt.Allow(true, true)
		s.conn.BindOption(opt)
		s.conn.EnableOptionForThem(opt.Byte(), true)
		s.conn.EnableOptionForUs(opt.Byte(), true)
	}

	return nil
}

func (s *session) runForever() (err error) {
	scanner := bufio.NewScanner(s.conn)
	for scanner.Scan() {
		if err = s.handleLine(scanner.Text()); err != nil {
			return
		}
	}
	return scanner.Err()
}
