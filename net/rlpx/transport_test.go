package rlpx

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sec-bit/zkPoD-node/net/utils"
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
	addr *Addr,
	key *ecdsa.PrivateKey,
	addrChan chan *net.TCPAddr,
	connHandler func(*Connection, *ecdsa.PrivateKey) error,
) {
	defer wg.Done()

	l, err := Listen(addr)
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

	if connHandler == nil {
		return
	}

	if err := connHandler(conn, key); err != nil {
		t.Fatalf("failed to handle connection: %v", err)
	}
}

func startClient(
	t *testing.T,
	wg *sync.WaitGroup,
	addr *Addr,
	key *ecdsa.PrivateKey,
	connWorker func(*Connection, *ecdsa.PrivateKey) error,
) {
	defer wg.Done()

	conn, err := Dial(addr)
	if err != nil {
		t.Fatalf("failed to dial %s: %v", addr, err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatalf("failed to close connection on client side: %v",
				err)
		}
	}()

	if connWorker == nil {
		return
	}

	if err := connWorker(conn, key); err != nil {
		t.Fatalf("failed to handle job on the connection: %v", err)
	}
}

func TestListen(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	serverAddr, _ := NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan, nil)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, nil, nil)

	wg.Wait()
}

func TestPreAuthHandshake(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(conn *Connection, key *ecdsa.PrivateKey) error {
			return conn.receiverPreAuthHandshake(key)
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(conn *Connection, key *ecdsa.PrivateKey) error {
			rPubkey, err := conn.initiatorPreAuthHandshake()
			if err != nil {
				return err
			}
			_ = rPubkey

			return nil
		},
	)

	wg.Wait()
}

func TestHandshake(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(conn *Connection, key *ecdsa.PrivateKey) error {
			rpubkey, err := conn.Handshake(key, false)
			if err != nil {
				return err
			}

			if !isPubKeyEqual(rpubkey, &clientKey.PublicKey) {
				return fmt.Errorf(
					"received public key %s not match with client public key %s",
					pubkeyToString(rpubkey),
					pubkeyToString(&clientKey.PublicKey))
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(conn *Connection, key *ecdsa.PrivateKey) error {
			rpubkey, err := conn.Handshake(key, true)
			if err != nil {
				return err
			}

			if !isPubKeyEqual(rpubkey, &serverKey.PublicKey) {
				return fmt.Errorf(
					"received public key %s not match with server public key %s",
					pubkeyToString(rpubkey),
					pubkeyToString(&serverKey.PublicKey))
			}

			return nil
		},
	)

	wg.Wait()
}

func TestReadWriteMsg(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	ping := []byte("ping")
	pong := []byte("pong")

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(conn *Connection, key *ecdsa.PrivateKey) error {
			rpubkey, err := conn.Handshake(key, false)
			if err != nil {
				return err
			}

			if !isPubKeyEqual(rpubkey, &clientKey.PublicKey) {
				return fmt.Errorf(
					"received public key %s not match with client public key %s",
					pubkeyToString(rpubkey),
					pubkeyToString(&clientKey.PublicKey))
			}

			buf := new(bytes.Buffer)
			if n, err := conn.Read(buf); err != nil {
				return fmt.Errorf(
					"failed to receive ping: %v", err)
			} else if n != uint64(len(ping)) {
				return fmt.Errorf(
					"mismatch ping size, get %d, expect %d",
					n, len(ping))
			}
			bufBytes := buf.Bytes()
			if !bytes.Equal(ping, bufBytes) {
				return fmt.Errorf(
					"received message %s is not ping",
					hex.EncodeToString(bufBytes))
			}
			t.Logf("ping received\n")

			if err := conn.Write(
				bytes.NewReader(pong), uint64(len(pong)),
			); err != nil {
				return fmt.Errorf("failed to pong: %v", err)
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(conn *Connection, key *ecdsa.PrivateKey) error {
			rpubkey, err := conn.Handshake(key, true)
			if err != nil {
				return err
			}

			if !isPubKeyEqual(rpubkey, &serverKey.PublicKey) {
				return fmt.Errorf(
					"received public key %s not match with server public key %s",
					pubkeyToString(rpubkey),
					pubkeyToString(&serverKey.PublicKey))
			}

			if err := conn.Write(
				bytes.NewReader(ping), uint64(len(ping)),
			); err != nil {
				return fmt.Errorf("failed to ping: %v", err)
			}

			buf := new(bytes.Buffer)
			if n, err := conn.Read(buf); err != nil {
				return fmt.Errorf(
					"failed to receive pong: %v", err)
			} else if n != uint64(len(pong)) {
				return fmt.Errorf(
					"mismatch pong size, get %d, expect %d",
					n, len(pong))
			}
			bufBytes := buf.Bytes()
			if !bytes.Equal(pong, bufBytes) {
				return fmt.Errorf(
					"received message %s is not pong",
					hex.EncodeToString(bufBytes))
			}
			t.Logf("pong received\n")

			return nil
		},
	)

	wg.Wait()
}

func TestReadWriteLargeMsg(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	// 100MB message
	largeMsg := utils.MakeRandomMsg(100 * 1024 * 1024)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(conn *Connection, key *ecdsa.PrivateKey) error {
			rpubkey, err := conn.Handshake(key, false)
			if err != nil {
				return err
			}

			if !isPubKeyEqual(rpubkey, &clientKey.PublicKey) {
				return fmt.Errorf(
					"received public key %s not match with client public key %s",
					pubkeyToString(rpubkey),
					pubkeyToString(&clientKey.PublicKey))
			}

			buf := new(bytes.Buffer)
			if n, err := conn.Read(buf); err != nil {
				return fmt.Errorf(
					"failed to receive msg from client: %v",
					err)
			} else if n != uint64(len(largeMsg)) {
				return fmt.Errorf(
					"mismatch msg size from client, get %d, expect %d",
					n, len(largeMsg))
			} else if !bytes.Equal(largeMsg, buf.Bytes()) {
				return fmt.Errorf(
					"received msg is incorrect")
			}

			if err := conn.Write(
				bytes.NewReader(largeMsg), uint64(len(largeMsg)),
			); err != nil {
				return fmt.Errorf(
					"failed to send msg to client: %v",
					err)
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(conn *Connection, key *ecdsa.PrivateKey) error {
			rpubkey, err := conn.Handshake(key, true)
			if err != nil {
				return err
			}

			if !isPubKeyEqual(rpubkey, &serverKey.PublicKey) {
				return fmt.Errorf(
					"received public key %s not match with server public key %s",
					pubkeyToString(rpubkey),
					pubkeyToString(&serverKey.PublicKey))
			}

			if err := conn.Write(
				bytes.NewReader(largeMsg), uint64(len(largeMsg)),
			); err != nil {
				return fmt.Errorf(
					"failed to send msg to server: %v",
					err)
			}

			buf := new(bytes.Buffer)
			if n, err := conn.Read(buf); err != nil {
				return fmt.Errorf(
					"failed to receive msg from server: %v",
					err)
			} else if n != uint64(len(largeMsg)) {
				return fmt.Errorf(
					"mismatch msg size from server, get %d, expect %d",
					n, len(largeMsg))
			}
			if !bytes.Equal(largeMsg, buf.Bytes()) {
				return fmt.Errorf(
					"received message is incorrect")
			}

			return nil
		},
	)

	wg.Wait()
}

func TestReadWriteLargeFile(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	msgSize := uint64(4 * 1024 * 1024 * 1024)
	src, dst, err := prepareFile(msgSize)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(src.Name())
		os.Remove(dst.Name())
	}()
	src.Seek(0, 0)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(conn *Connection, key *ecdsa.PrivateKey) error {
			rpubkey, err := conn.Handshake(key, false)
			if err != nil {
				return err
			}

			if !isPubKeyEqual(rpubkey, &clientKey.PublicKey) {
				return fmt.Errorf(
					"received public key %s not match with client public key %s",
					pubkeyToString(rpubkey),
					pubkeyToString(&clientKey.PublicKey))
			}

			if n, err := conn.Read(dst); err != nil {
				return fmt.Errorf(
					"failed to receive msg from client: %v",
					err)
			} else if n != msgSize {
				return fmt.Errorf(
					"mismatch msg size from client, get %d, expect %d",
					n, msgSize)
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(conn *Connection, key *ecdsa.PrivateKey) error {
			rpubkey, err := conn.Handshake(key, true)
			if err != nil {
				return err
			}

			if !isPubKeyEqual(rpubkey, &serverKey.PublicKey) {
				return fmt.Errorf(
					"received public key %s not match with server public key %s",
					pubkeyToString(rpubkey),
					pubkeyToString(&serverKey.PublicKey))
			}

			if err := conn.Write(src, msgSize); err != nil {
				return fmt.Errorf(
					"failed to send msg to server: %v",
					err)
			}

			return nil
		},
	)

	wg.Wait()
}

func prepareFile(size uint64) (*os.File, *os.File, error) {
	needCleanup := false

	srcFile, err := ioutil.TempFile("", "pod_net_large_message_src")
	if err != nil {
		needCleanup = true
		return nil, nil, fmt.Errorf("failed to create source file: %v", err)
	}
	defer func(need *bool) {
		if !*need {
			return
		}
		os.Remove(srcFile.Name())
	}(&needCleanup)

	dstFile, err := ioutil.TempFile("", "pod_net_large_message_dst")
	if err != nil {
		needCleanup = true
		return nil, nil, fmt.Errorf("failed to create destination file: %v", err)
	}
	defer func(need *bool) {
		if !*need {
			return
		}
		os.Remove(dstFile.Name())
	}(&needCleanup)

	remaining := size
	for remaining > 0 {
		sz := uint64(4096)
		if sz > remaining {
			sz = remaining
		}
		random := utils.MakeRandomMsg(uint(sz))
		n, err := srcFile.Write(random)
		if err != nil {
			needCleanup = true
			return nil, nil, fmt.Errorf(
				"failed to write source file: %v", err)
		}

		remaining -= uint64(n)
	}

	return srcFile, dstFile, nil
}
