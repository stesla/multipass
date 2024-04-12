package main

import (
	"io"
	"log"
	"strings"

	"net"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stesla/telnet"
)

func main() {
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
		log.Fatalln("fatal error reading confi1g file:", err)
	}

	viper.SetEnvPrefix("multipass")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	listener, err := net.Listen("tcp", viper.GetString("address"))
	if err != nil {
		log.Fatalln("fatal error binding to address:", err)
	}
	for {
		tcpconn, err := listener.Accept()
		if err != nil {
			log.Println("error accepting connection:", err)
		}
		conn := telnet.Server(tcpconn)
		go io.Copy(conn, conn)
	}
}
