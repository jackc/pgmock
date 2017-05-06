package cmd

import (
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/jackc/pgmock"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Proxy a PostgreSQL client connection to a server",

	Run: func(cmd *cobra.Command, args []string) {
		ln, err := net.Listen("tcp", viper.GetString("listenAddress"))
		if err != nil {
			log.Fatal(err)
		}

		for {
			clientConn, err := ln.Accept()
			if err != nil {
				log.Fatal(err)
			}

			serverConn, err := net.Dial("tcp", viper.GetString("remoteAddress"))
			if err != nil {
				log.Fatal(err)
			}

			proxy, err := pgmock.NewProxy(clientConn, serverConn)
			if err != nil {
				log.Fatal(err)
			}

			err = proxy.Run()
			log.Error(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(proxyCmd)

	proxyCmd.Flags().StringP("listen-address", "l", "127.0.0.1:15432", "Proxy listen address")
	viper.BindPFlag("listenAddress", proxyCmd.Flags().Lookup("listen-address"))

	proxyCmd.Flags().StringP("remote-address", "r", "127.0.0.1:5432", "Remote PostgreSQL server address")
	viper.BindPFlag("remoteAddress", proxyCmd.Flags().Lookup("remote-address"))
}
