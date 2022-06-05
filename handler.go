package main

import (
	"bufio"
	"fmt"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"os"
)

type Node struct {
	ID     peer.ID
	Stream *network.Stream
	Socket *bufio.ReadWriter
	Send   chan []byte
}

type NodeManager struct {
	Peers      map[string]*Node
	Connect    chan *Node
	Disconnect chan *Node
	ReadData   chan []byte
	WriteData  chan []byte // broadcast
}

var Manager = NodeManager{
	Connect:    make(chan *Node),
	Disconnect: make(chan *Node),
	Peers:      make(map[string]*Node),
	ReadData:   make(chan []byte, 1024),
	WriteData:  make(chan []byte, 1024),
}

func (node *Node) Read() {
	defer node.Close()
	for {
		str, err := node.Socket.ReadString('\n')
		if err != nil || str == "" {
			logger.Warn("Read error ", err)
			return
		}
		if str != "\n" {
			Manager.ReadData <- []byte(str)
		}
	}
}

func (node *Node) Write() {
	defer node.Close()
	for {
		data := <-node.Send
		_, err := node.Socket.WriteString(string(data))
		if err != nil {
			logger.Warn("Write error ", err)
			return
		}
		err = node.Socket.Flush()
		if err != nil {
			logger.Warn("Error flushing buffer")
			return
		}
	}
}

func (node *Node) Close() {
	stream := *node.Stream
	err := stream.Reset()
	if err != nil {
		logger.Error("Reset stream failed")
	}
	Manager.Disconnect <- node
}

func (manager *NodeManager) Start() {
	for {
		select {
		case node := <-manager.Connect:
			logger.Info("Got a new stream!")
			manager.Peers[string(node.ID)] = node
		case node := <-manager.Disconnect:
			delete(manager.Peers, string(node.ID))
		case data := <-manager.WriteData:
			for _, node := range manager.Peers {
				node.Send <- data
			}
		}
	}
}
func handleStream(stream network.Stream) {
	node := Node{
		ID:     stream.Conn().RemotePeer(),
		Socket: bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)),
		Stream: &stream,
		Send:   make(chan []byte),
	}
	Manager.Connect <- &node
	go node.Read()
	go node.Write()
}

func onRead() {
	const greenColor = "\x1b[32m"
	const restColor = "\x1b[0m"
	for {
		data := <-Manager.ReadData
		fmt.Printf("%s%s%s> ", greenColor, data, restColor)
	}
}

func onWrite() {
	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		str, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin")
			panic(err)
		}
		Manager.WriteData <- []byte(str)
	}
}
