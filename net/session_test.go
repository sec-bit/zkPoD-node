package net

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/sec-bit/zkPoD-node/net/rlpx"
	"github.com/sec-bit/zkPoD-node/net/utils"
)

func TestNewSessionRequest(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := rlpx.NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	mklroot := utils.MakeRandomMsg(32)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(node *Node, key *ecdsa.PrivateKey) error {
			req, err := node.RecvSessionRequest()
			if err != nil {
				return fmt.Errorf(
					"failed to receive session request: %v",
					err)
			}

			if req.ID != 0 {
				return fmt.Errorf(
					"session ID (%d) not zero", req.ID)
			}

			if req.Mode != ModePlainBatchPoD {
				return fmt.Errorf(
					"mismatch session mode, get %d, expect %d",
					req.Mode, ModePlainBatchPoD)
			}

			if !bytes.Equal(mklroot, req.SigmaMklRoot) {
				return fmt.Errorf(
					"mismatch sigma merkle root, get %v, expect %v",
					hex.EncodeToString(req.SigmaMklRoot),
					hex.EncodeToString(mklroot))
			}

			if node.state != stateSessionReqRecvd {
				return fmt.Errorf(
					"server node not in SessionReqRecvd state")
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if err := node.SendNewSessionRequest(
				ModePlainBatchPoD, mklroot,
			); err != nil {
				return fmt.Errorf(
					"failed to send session request: %v",
					err)
			}

			if node.state != stateSessionAckWait {
				return fmt.Errorf(
					"client node not in SessionAckWait state")
			}

			return nil
		},
	)

	wg.Wait()
}

func TestSessionAck(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := rlpx.NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	mklroot := utils.MakeRandomMsg(32)
	sessionID := uint64(0xdeadbeaf)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(node *Node, key *ecdsa.PrivateKey) error {
			req, err := node.RecvSessionRequest()
			if err != nil {
				return fmt.Errorf(
					"failed to receive session request: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, req.Mode, req.SigmaMklRoot, true,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from server: %v",
					err)
			}
			if node.state != stateSessionAckWait {
				return fmt.Errorf(
					"server node not in stateSessionAckWait state")
			}

			ack, err := node.RecvSessionAck(false)
			if err != nil {
				return fmt.Errorf(
					"failed to receive session ack on server node: %v",
					err)
			}
			if ack.ID != sessionID {
				return fmt.Errorf(
					"mismatch session ID on server node, get %d, expect %d",
					ack.ID, sessionID)
			}
			if node.session.id != sessionID {
				return fmt.Errorf(
					"server session ID %d != %d",
					node.session.id, sessionID)
			}
			if node.session.mode != req.Mode {
				return fmt.Errorf(
					"server session mode %d != %d",
					node.session.mode, req.Mode)
			}
			if node.state != stateSessionEstablished {
				return fmt.Errorf("server node not in stateSessionEstablished state")
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if err := node.SendNewSessionRequest(
				ModePlainBatchPoD, mklroot,
			); err != nil {
				return fmt.Errorf(
					"failed to send session request: %v",
					err)
			}

			ack, err := node.RecvSessionAck(true)
			if err != nil {
				return fmt.Errorf(
					"failed to receive session ack on client node: %v",
					err)
			}
			if ack.ID != sessionID {
				return fmt.Errorf(
					"mismatch session ID on client node, get %d, expect %d",
					ack.ID, sessionID)
			}
			if node.state != stateSessionAckRecvd {
				return fmt.Errorf("client node not in stateSessionAckRecvd state")
			}

			if err := node.SendSessionAck(
				ack.ID, ModePlainBatchPoD, mklroot, false,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from client: %v",
					err)
			}
			if node.session.id != ack.ID {
				return fmt.Errorf(
					"client session ID %d != %d",
					node.session.id, ack.ID)
			}
			if node.session.mode != ack.Mode {
				return fmt.Errorf(
					"client session mode %d != %d",
					node.session.mode, ack.Mode)
			}
			if node.state != stateSessionEstablished {
				return fmt.Errorf(
					"client node not in stateSessionEstablished state")
			}

			return nil
		},
	)

	wg.Wait()
}
