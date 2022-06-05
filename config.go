package main

import (
	"flag"
	"fmt"
	"github.com/ipfs/go-log/v2"
	"strings"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	maddr "github.com/multiformats/go-multiaddr"
)

// A new type we need for writing a custom flag parser
type addrList []maddr.Multiaddr

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := maddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

type Config struct {
	RendezvousString string
	BootstrapPeers   addrList
	ProtocolID       string
	Address          string
	Port             int
	LogLevel         string
}

func ParseFlags() (Config, error) {
	config := Config{}

	flag.StringVar(&config.RendezvousString, "rendezvous", "follow taffy miao",
		"Unique string to identify group of nodes. Share this with your friends to let them connect with you")
	flag.Var(&config.BootstrapPeers, "peer", "Adds a peer multiaddress to the bootstrap list")
	flag.StringVar(&config.ProtocolID, "pid", "/chat/1.1.0", "Sets a protocol id for stream headers")
	flag.StringVar(&config.Address, "addr", "0.0.0.0", "The address bind to")
	flag.IntVar(&config.Port, "port", 12345, "The port bind to")
	flag.StringVar(&config.LogLevel, "loglevel", "debug", "Log Level")

	flag.Parse()

	if len(config.BootstrapPeers) == 0 {
		config.BootstrapPeers = dht.DefaultBootstrapPeers
	}

	return config, nil
}

func Init() Config {
	help := flag.Bool("h", false, "Display Help")
	config, err := ParseFlags()
	if err != nil {
		panic(err)
	}

	if *help {
		fmt.Println("This program demonstrates a simple p2p chat application using libp2p")
		fmt.Println()
		fmt.Println("Usage: Run './chat in two different terminals. Let them connect to the bootstrap nodes, announce themselves and connect to the peers")
		flag.PrintDefaults()
	}

	log.SetAllLoggers(log.LevelWarn)
	err = log.SetLogLevel(AppName, config.LogLevel)
	if err != nil {
		panic(err)
	}

	return config
}
