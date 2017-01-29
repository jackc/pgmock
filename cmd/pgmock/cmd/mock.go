package cmd

import (
	"net"

	"github.com/jackc/pgmock"
	"github.com/jackc/pgmock/pgmsg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mockCmd represents the mock command
var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "Run a mock PostgreSQL server",

	Run: func(cmd *cobra.Command, args []string) {
		ln, err := net.Listen("tcp", viper.GetString("listenAddress"))
		if err != nil {
			log.Fatal(err)
		}

		for {
			frontendConn, err := ln.Accept()
			if err != nil {
				log.Fatal(err)
			}

			mock, err := pgmock.NewMock(frontendConn)
			if err != nil {
				log.Fatal(err)
			}

			err = mock.Send(&pgmsg.AuthenticationOk{})
			if err != nil {
				log.Fatal(err)
			}

			err = mock.Send(&pgmsg.BackendKeyData{ProcessID: 0, SecretKey: 0})
			if err != nil {
				log.Fatal(err)
			}

			err = mock.Send(&pgmsg.ReadyForQuery{TxStatus: 'I'})
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(mockCmd)

	mockCmd.Flags().StringP("listen-address", "l", "127.0.0.1:15432", "Proxy listen address")
	viper.BindPFlag("listenAddress", mockCmd.Flags().Lookup("listen-address"))

}
