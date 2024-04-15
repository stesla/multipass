package main

import (
	"net"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stesla/telnet"
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
		log.Info().Str("address", tcpconn.RemoteAddr().String()).Msg("connected")

		conn := telnet.Server(tcpconn)
		session := newSession(conn)
		if err := session.negotiateOptions(); err != nil {
			log.Debug().Err(err).
				Str("address", tcpconn.RemoteAddr().String()).
				Msg("error negotiating telnet options")
		} else {
			go func() {
				err := session.runForever()
				log.Info().Err(err).
					Str("address", tcpconn.RemoteAddr().String()).
					Msg("disconnected")
			}()
		}
	}
}
