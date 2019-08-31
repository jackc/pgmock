package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgmock/pgmockproxy/proxy"
)

var options struct {
	listenAddress string
	remoteAddress string
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage:  %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&options.listenAddress, "listen", "127.0.0.1:15432", "Proxy listen address")
	flag.StringVar(&options.remoteAddress, "remote", "127.0.0.1:5432", "Remote PostgreSQL server address")
	flag.Parse()

	ln, err := net.Listen("tcp", options.listenAddress)
	if err != nil {
		log.Fatal(err)
	}

	clientConn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}

	listenNetwork := "tcp"
	if _, err := os.Stat(options.remoteAddress); err == nil {
		listenNetwork = "unix"
	}

	serverConn, err := net.Dial(listenNetwork, options.remoteAddress)
	if err != nil {
		log.Fatal(err)
	}

	proxy := proxy.NewProxy(clientConn, serverConn)
	err = proxy.Run()
	if err != nil {
		log.Fatal(err)
	}
}
