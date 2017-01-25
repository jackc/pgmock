package main

import (
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgmock"
)

func main() {
	ln, err := net.Listen("tcp", ":5433")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}

	mockConn := pgmock.NewMockConn(conn)

	err = mockConn.Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("seemed to work...")
}
