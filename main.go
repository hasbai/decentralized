package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"

	"github.com/ipfs/go-log/v2"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

const AppName = "dim"

var logger = log.Logger(AppName)

func main() {
	config := Init()

	go Manager.Start()
	go onRead()
	go onWrite()

	// Construct a new p2p Host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/%s/tcp/%d", config.Address, config.Port),
		),
		libp2p.EnableHolePunching(),
		libp2p.Identity(GetPrivateKey()),
	)
	if err != nil {
		panic(err)
	}
	logger.Info("Host created. We are:", host.ID())
	logger.Info(host.Addrs())

	// Set a function as stream handler. This function is called when a peerNode
	// initiates a connection and starts a stream with this peerNode.
	host.SetStreamHandler(protocol.ID(config.ProtocolID), handleStream)

	// Start a DHT, for use in peerNode discovery. We can't just make a new DHT
	// client because we want each peerNode to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peerNode discovery.
	ctx := context.Background()
	kademliaDHT, err := dht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peerNode table every five minutes.
	logger.Info("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	for _, peerAddr := range config.BootstrapPeers {
		peerInfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerInfo); err != nil {
				logger.Warn(err)
			} else {
				logger.Info("Connection established with bootstrap node:", *peerInfo)
			}
		}()
	}
	wg.Wait()
	logger.Info("Bootstrap finished")

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	logger.Info("Announcing ourselves...")
	discovery := routing.NewRoutingDiscovery(kademliaDHT)
	_, err = discovery.Advertise(ctx, config.RendezvousString)
	if err != nil {
		panic(err)
	}
	logger.Info("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	logger.Info("Searching for other peers...")
	peerChan, err := discovery.FindPeers(ctx, config.RendezvousString)
	if err != nil {
		panic(err)
	}

	for peerNode := range peerChan {
		if peerNode.ID == host.ID() || len(peerNode.Addrs) == 0 {
			continue
		}

		logger.Debug("Found peerNode, connecting:", peerNode)
		stream, err := host.NewStream(ctx, peerNode.ID, protocol.ID(config.ProtocolID))
		if err != nil {
			logger.Warn("Connection failed:", err)
			continue
		}

		logger.Info("Connected to:", peerNode)
		handleStream(stream)
	}

	select {}
}
