package net

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sec-bit/zkPoD-node/net/rlpx"
)

func genKey(t *testing.T) *ecdsa.PrivateKey {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v\n", err)
	}
	return key
}

func startServer(
	t *testing.T,
	wg *sync.WaitGroup,
	addr *rlpx.Addr,
	key *ecdsa.PrivateKey,
	addrChan chan *net.TCPAddr,
	nodeHandler func(*Node, *ecdsa.PrivateKey) error,
) {
	defer wg.Done()

	l, err := rlpx.Listen(addr)
	if err != nil {
		t.Fatalf("failed to listen on %s: %v", addr, err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			t.Fatalf("failed to close listener on %s: %v",
				addr, err)
		}
	}()

	addrChan <- l.Addr().TCPAddr

	conn, err := l.Accept()
	if err != nil {
		t.Fatalf("failed to accept connection on %s: %v",
			addr, err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("failed to close connection on server side: %v",
				err)
		}
	}()

	rkey, err := conn.Handshake(key, false)
	if err != nil {
		t.Fatalf("server-side handshake failed: %v", err)
	}

	node, err := NewNode(conn, key, rkey)
	if err != nil {
		t.Fatalf("failed to create server node: %v", err)
	}
	defer func() {
		if err := node.Close(); err != nil {
			t.Fatalf("failed to close server node: %v", err)
		}
		if node.state != stateClosed {
			t.Fatalf("server node not in Closed state")
		}
	}()

	if nodeHandler == nil {
		return
	}

	if err := nodeHandler(node, key); err != nil {
		t.Fatalf("failed to handle server node: %v", err)
	}
}

func startClient(
	t *testing.T,
	wg *sync.WaitGroup,
	addr *rlpx.Addr,
	key *ecdsa.PrivateKey,
	nodeWorker func(*Node, *ecdsa.PrivateKey) error,
) {
	defer wg.Done()

	conn, err := rlpx.Dial(addr)
	if err != nil {
		t.Fatalf("failed to dial %s: %v", addr, err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("failed to close connection on client side: %v",
				err)
		}
	}()

	rkey, err := conn.Handshake(key, true)
	if err != nil {
		t.Fatalf("client-side handshake failed: %v", err)
	}

	node, err := NewNode(conn, key, rkey)
	if err != nil {
		t.Fatalf("failed to create client node: %v", err)
	}
	defer func() {
		if err := node.Close(); err != nil {
			t.Fatalf("failed to close client node: %v", err)
		}
		if node.state != stateClosed {
			t.Fatalf("client node not in Closed state")
		}
	}()

	if nodeWorker == nil {
		return
	}

	if err := nodeWorker(node, key); err != nil {
		t.Fatalf("failed to handle the client connection: %v", err)
	}
}

func TestNodes(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := rlpx.NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if node.state != stateConnected {
				return fmt.Errorf("server node not in Connected state")
			}
			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if node.state != stateConnected {
				return fmt.Errorf("client node not in Connected state")
			}
			return nil
		},
	)

	wg.Wait()
}

func TestSendRecvMsg(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := rlpx.NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	ping := []byte("ping")
	pong := []byte("pong")

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(node *Node, key *ecdsa.PrivateKey) error {
			typ, length, err := node.recvMsgHeader()
			if err != nil {
				return fmt.Errorf(
					"failed to receive ping header: %v", err)
			}
			if typ != msgSessionRequest {
				return fmt.Errorf(
					"incorrect ping msg type %d, expect %d",
					typ, msgSessionRequest)
			}
			if length != uint64(len(ping)) {
				return fmt.Errorf(
					"incorrect ping msg length %d, expect %d",
					length, len(ping))
			}

			buf := new(bytes.Buffer)
			if err := node.recvMsgPayload(buf, length); err != nil {
				return fmt.Errorf(
					"failed to receive ping payload: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), ping) {
				return fmt.Errorf(
					"incorrect ping msg payload %v, expect %v",
					buf.Bytes(), ping)
			}

			if err := node.sendMsg(
				msgSessionAck, bytes.NewReader(pong), uint64(len(pong)),
			); err != nil {
				return fmt.Errorf("failed to send pong: %v", err)
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if err := node.sendMsg(
				msgSessionRequest, bytes.NewReader(ping), uint64(len(ping)),
			); err != nil {
				return fmt.Errorf("failed to send ping: %v", err)
			}

			typ, length, err := node.recvMsgHeader()
			if err != nil {
				return fmt.Errorf(
					"failed to receive pong header: %v", err)
			}
			if typ != msgSessionAck {
				return fmt.Errorf(
					"incorrect pong msg type %d, expect %d",
					typ, msgSessionAck)
			}
			if length != uint64(len(pong)) {
				return fmt.Errorf(
					"incorrect pong msg length %d, expect %d",
					length, len(pong))
			}

			buf := new(bytes.Buffer)
			if err := node.recvMsgPayload(buf, length); err != nil {
				return fmt.Errorf(
					"failed to receive pong payload: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), pong) {
				return fmt.Errorf(
					"incorrect pong msg payload %v, expect %v",
					buf, pong)
			}

			return nil
		},
	)

	wg.Wait()
}
