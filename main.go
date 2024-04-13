package main

import (
	"io"
	"net"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stesla/telnet"
	"golang.org/x/text/encoding/unicode"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	pflag.String("address", ":2300", "the address to listen on")
	viper.BindPFlag("address", pflag.Lookup("address"))

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.multipass/")
	viper.AddConfigPath(".")
	switch err := viper.ReadInConfig().(type) {
	case viper.ConfigFileNotFoundError:
	case nil:
	default:
		log.Fatal().Err(err).Msg("fatal error reading config file")
	}

	viper.SetEnvPrefix("multipass")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	listener, err := net.Listen("tcp", viper.GetString("address"))
	if err != nil {
		log.Fatal().Err(err).Msg("fatal error binding to address")
	}
	for {
		tcpconn, err := listener.Accept()
		if err != nil {
			log.Info().Err(err).Msg("error accepting connection")
		}
		log.Info().Str("address", tcpconn.RemoteAddr().String()).Msg("incoming connection")

		conn := telnet.Server(tcpconn)

		conn.AddListener("update-option", telnet.FuncListener{
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
			conn.BindOption(opt)
			conn.EnableOptionForThem(opt.Byte(), true)
			conn.EnableOptionForUs(opt.Byte(), true)
		}

		go io.Copy(conn, conn)
	}
}
