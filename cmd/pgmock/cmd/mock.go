package cmd

import (
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/jackc/pgmock"
	"github.com/jackc/pgx/pgproto3"
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

			err = mock.Send(&pgproto3.AuthenticationOk{})
			if err != nil {
				log.Fatal(err)
			}

			err = mock.Send(&pgproto3.BackendKeyData{ProcessID: 0, SecretKey: 0})
			if err != nil {
				log.Fatal(err)
			}

			err = mock.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
			if err != nil {
				log.Fatal(err)
			}

			for {
				msg, err := mock.Receive()
				if err != nil {
					log.Fatal(err)
				}

				if _, ok := msg.(*pgproto3.Query); ok {
					err = mock.Send(&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
						{
							Name:         "Hey Jack",
							DataTypeOID:  23,
							DataTypeSize: 4,
							TypeModifier: 4294967295,
							Format:       pgproto3.TextFormat,
						},
					}})
					if err != nil {
						log.Fatal(err)
					}

					err = mock.Send(&pgproto3.DataRow{Values: [][]byte{
						[]byte("5"),
					}})
					if err != nil {
						log.Fatal(err)
					}

					err = mock.Send(&pgproto3.CommandComplete{CommandTag: "SELECT 2"})
					if err != nil {
						log.Fatal(err)
					}

					err = mock.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
					if err != nil {
						log.Fatal(err)
					}
				} else {
					log.Fatal("unexpected message")
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(mockCmd)

	mockCmd.Flags().StringP("listen-address", "l", "127.0.0.1:15432", "Proxy listen address")
	viper.BindPFlag("listenAddress", mockCmd.Flags().Lookup("listen-address"))

}
