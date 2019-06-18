package net

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sec-bit/zkPoD-node/net/rlpx"
	"github.com/sec-bit/zkPoD-node/net/utils"
)

func TestTxRequest(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := rlpx.NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	mklroot := utils.MakeRandomMsg(32)
	mode := uint8(ModePlainComplaintPoD)
	sessionID := uint64(0xdeadbeaf)
	txReq := utils.MakeRandomMsg(32)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if _, err := node.RecvSessionRequest(); err != nil {
				return fmt.Errorf(
					"failed to receive session request: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, mode, mklroot, true,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from server: %v",
					err)
			}

			if _, err := node.RecvSessionAck(false); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on server node: %v",
					err)
			}

			reqBuf := new(bytes.Buffer)
			if n, err := node.RecvTxRequest(reqBuf); err != nil {
				return err
			} else if n != uint64(len(txReq)) {
				return fmt.Errorf(
					"invalid request size, get %d bytes, expect %d bytes",
					n, len(txReq))
			}
			if node.state != stateTxRequestRecvd {
				return fmt.Errorf("server node not in TxRequestRecvd state")
			}
			req := reqBuf.Bytes()
			if !bytes.Equal(req, txReq) {
				return fmt.Errorf(
					"mismatch Tx request, get %s, expect %s",
					hex.EncodeToString(req),
					hex.EncodeToString(txReq))
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if err := node.SendNewSessionRequest(
				ModePlainComplaintPoD, mklroot,
			); err != nil {
				return fmt.Errorf(
					"failed to send session request: %v",
					err)
			}

			if _, err := node.RecvSessionAck(true); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on client node: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, ModePlainComplaintPoD, mklroot, false,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from client: %v",
					err)
			}

			if err := node.SendTxRequest(
				bytes.NewReader(txReq), uint64(len(txReq)),
			); err != nil {
				return fmt.Errorf(
					"failed to send Tx request: %v", err)
			}
			if node.state != stateTxResponseWait {
				return fmt.Errorf(
					"client node not in TxResponse state")
			}

			return nil
		},
	)

	wg.Wait()
}

func TestTxResponse(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := rlpx.NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	mklroot := utils.MakeRandomMsg(32)
	mode := uint8(ModePlainComplaintPoD)
	sessionID := uint64(0xdeadbeaf)
	txReq := utils.MakeRandomMsg(32)
	txResponse := utils.MakeRandomMsg(128)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if _, err := node.RecvSessionRequest(); err != nil {
				return fmt.Errorf(
					"failed to receive session request: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, mode, mklroot, true,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from server: %v",
					err)
			}

			if _, err := node.RecvSessionAck(false); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on server node: %v",
					err)
			}

			if _, err := node.RecvTxRequest(new(bytes.Buffer)); err != nil {
				return err
			}

			if err := node.SendTxResponse(
				bytes.NewReader(txResponse), uint64(len(txResponse)),
			); err != nil {
				return err
			}
			if node.state != stateTxReceiptWait {
				return fmt.Errorf(
					"server node not in TxReceiptWait state")
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if err := node.SendNewSessionRequest(
				ModePlainComplaintPoD, mklroot,
			); err != nil {
				return fmt.Errorf(
					"failed to send session request: %v",
					err)
			}

			if _, err := node.RecvSessionAck(true); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on client node: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, ModePlainComplaintPoD, mklroot, false,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from client: %v",
					err)
			}

			if err := node.SendTxRequest(
				bytes.NewReader(txReq), uint64(len(txReq)),
			); err != nil {
				return fmt.Errorf(
					"failed to send Tx request: %v", err)
			}

			respBuf := new(bytes.Buffer)
			if n, err := node.RecvTxResponse(respBuf); err != nil {
				return err
			} else if n != uint64(len(txResponse)) {
				return fmt.Errorf(
					"invalid tx response size, get %d bytes, expect %d bytes",
					n, len(txResponse))
			}
			if node.state != stateTxResponseRecvd {
				return fmt.Errorf(
					"client node not in TxResponseRecvd state")
			}
			resp := respBuf.Bytes()
			if !bytes.Equal(resp, txResponse) {
				return fmt.Errorf(
					"mismatch Tx response, get %s, expect %s",
					hex.EncodeToString(resp),
					hex.EncodeToString(txResponse))
			}

			return nil
		},
	)

	wg.Wait()
}

func TestTxReceipt(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := rlpx.NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	mklroot := utils.MakeRandomMsg(32)
	mode := uint8(ModePlainComplaintPoD)
	sessionID := uint64(0xdeadbeaf)
	txReq := utils.MakeRandomMsg(32)
	txResponse := utils.MakeRandomMsg(128)
	txReceipt := utils.MakeRandomMsg(128)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if _, err := node.RecvSessionRequest(); err != nil {
				return fmt.Errorf(
					"failed to receive session request: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, mode, mklroot, true,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from server: %v",
					err)
			}

			if _, err := node.RecvSessionAck(false); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on server node: %v",
					err)
			}

			if _, err := node.RecvTxRequest(new(bytes.Buffer)); err != nil {
				return err
			}

			if err := node.SendTxResponse(
				bytes.NewReader(txResponse), uint64(len(txResponse)),
			); err != nil {
				return err
			}

			receipt, sig, err := node.RecvTxReceipt()
			if err != nil {
				return err
			}
			if node.state != stateTxReceiptRecvd {
				return fmt.Errorf(
					"server node not in TxReceiptRecvd state")
			}
			if !bytes.Equal(receipt, txReceipt) {
				return fmt.Errorf(
					"mismatch receipt, get %s, expect %s",
					hex.EncodeToString(receipt),
					hex.EncodeToString(txReceipt))
			}

			hash := crypto.Keccak256Hash(receipt).Bytes()
			sigPubkey, err := crypto.Ecrecover(hash, sig)
			if err != nil {
				return fmt.Errorf(
					"Ecrecover(%v, %v) failed: %v",
					hex.EncodeToString(hash),
					hex.EncodeToString(sig),
					err)
			}
			pubkeyBytes := crypto.FromECDSAPub(&clientKey.PublicKey)
			if !bytes.Equal(sigPubkey, pubkeyBytes) {
				return fmt.Errorf(
					"mismatch public keys, get %s, expect %s",
					sigPubkey, pubkeyBytes)
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if err := node.SendNewSessionRequest(
				ModePlainComplaintPoD, mklroot,
			); err != nil {
				return fmt.Errorf(
					"failed to send session request: %v",
					err)
			}

			if _, err := node.RecvSessionAck(true); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on client node: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, ModePlainComplaintPoD, mklroot, false,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from client: %v",
					err)
			}

			if err := node.SendTxRequest(
				bytes.NewReader(txReq), uint64(len(txReq)),
			); err != nil {
				return fmt.Errorf(
					"failed to send Tx request: %v", err)
			}

			if _, err := node.RecvTxResponse(new(bytes.Buffer)); err != nil {
				return err
			}

			if err := node.SendTxReceipt(
				bytes.NewReader(txReceipt), uint64(len(txReceipt)),
			); err != nil {
				return err
			}
			if node.state != stateTxSecretWait {
				return fmt.Errorf(
					"client side not in TxSecretWait state")
			}

			return nil
		},
	)

	wg.Wait()
}

func TestOTNegoRequest(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := rlpx.NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	mklroot := utils.MakeRandomMsg(32)
	mode := uint8(ModePlainOTComplaintPoD)
	sessionID := uint64(0xdeadbeaf)

	negoReq := utils.MakeRandomMsg(32)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if _, err := node.RecvSessionRequest(); err != nil {
				return fmt.Errorf(
					"failed to receive session request: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, mode, mklroot, true,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from server: %v",
					err)
			}

			if _, err := node.RecvSessionAck(false); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on server node: %v",
					err)
			}

			reqBuf := new(bytes.Buffer)
			if n, err := node.RecvNegoRequest(reqBuf); err != nil {
				return fmt.Errorf(
					"failed to receive negotiation request: %v",
					err)
			} else if n != uint64(len(negoReq)) {
				return fmt.Errorf(
					"invalid ngo req size, get %d bytes, expect %d bytes",
					n, len(negoReq))
			}
			req := reqBuf.Bytes()
			if !bytes.Equal(req, negoReq) {
				return fmt.Errorf(
					"mismatch negotiation request, get %s, expect %s",
					hex.EncodeToString(req),
					hex.EncodeToString(negoReq))
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if err := node.SendNewSessionRequest(mode, mklroot); err != nil {
				return fmt.Errorf(
					"failed to send session request: %v",
					err)
			}

			if _, err := node.RecvSessionAck(true); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on client node: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, mode, mklroot, false,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from client: %v",
					err)
			}

			if err := node.SendNegoRequest(
				bytes.NewReader(negoReq), uint64(len(negoReq)),
			); err != nil {
				return fmt.Errorf(
					"failed to send negotiation request: %v",
					err)
			}

			return nil
		},
	)

	wg.Wait()
}

func TestOTNegoAckReq(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := rlpx.NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	mklroot := utils.MakeRandomMsg(32)
	mode := uint8(ModePlainOTComplaintPoD)
	sessionID := uint64(0xdeadbeaf)

	buyerNegoReq := utils.MakeRandomMsg(32)
	sellerNegoReq := utils.MakeRandomMsg(32)
	sellerNegoAck := utils.MakeRandomMsg(128)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if _, err := node.RecvSessionRequest(); err != nil {
				return fmt.Errorf(
					"failed to receive session request: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, mode, mklroot, true,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from server: %v",
					err)
			}

			if _, err := node.RecvSessionAck(false); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on server node: %v",
					err)
			}

			if _, err := node.RecvNegoRequest(new(bytes.Buffer)); err != nil {
				return fmt.Errorf(
					"failed to receive negotiation request: %v",
					err)
			}

			if err := node.SendNegoAckReq(
				bytes.NewReader(sellerNegoAck),
				bytes.NewReader(sellerNegoReq),
				uint64(len(sellerNegoAck)),
				uint64(len(sellerNegoReq)),
			); err != nil {
				return fmt.Errorf(
					"failed to send nego ack+req: %v", err)
			}
			if node.state != stateNegoAckWait {
				return fmt.Errorf(
					"server node not in NegoAckWait state")
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if err := node.SendNewSessionRequest(mode, mklroot); err != nil {
				return fmt.Errorf(
					"failed to send session request: %v",
					err)
			}

			if _, err := node.RecvSessionAck(true); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on client node: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, mode, mklroot, false,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from client: %v",
					err)
			}

			if err := node.SendNegoRequest(
				bytes.NewReader(buyerNegoReq),
				uint64(len(buyerNegoReq)),
			); err != nil {
				return fmt.Errorf(
					"failed to send negotiation request: %v",
					err)
			}

			negoRespBuf := new(bytes.Buffer)
			negoReqBuf := new(bytes.Buffer)

			if m, n, err := node.RecvNegoAckReq(
				negoRespBuf, negoReqBuf,
			); err != nil {
				return fmt.Errorf(
					"failed to receive nego ack+req: %v",
					err)
			} else if m != uint64(len(sellerNegoAck)) {
				return fmt.Errorf(
					"invalid seller nego response size, get %d bytes, expect %d bytes",
					m, len(sellerNegoAck))
			} else if n != uint64(len(sellerNegoReq)) {
				return fmt.Errorf(
					"invalid seller nego request size, get %d bytes, expect %d bytes",
					n, len(sellerNegoReq))
			}
			if node.state != stateNegoAckReqRecvd {
				return fmt.Errorf(
					"client node not in NegoAckReqRecvd state")
			}
			negoResp := negoRespBuf.Bytes()
			if !bytes.Equal(negoResp, sellerNegoAck) {
				return fmt.Errorf(
					"mismatch server nego response, get %s, expect %s",
					hex.EncodeToString(negoResp),
					hex.EncodeToString(sellerNegoAck))
			}
			negoReq := negoReqBuf.Bytes()
			if !bytes.Equal(negoReq, sellerNegoReq) {
				return fmt.Errorf(
					"mismatch server nego reqeust, get %s, expect %s",
					hex.EncodeToString(negoReq),
					hex.EncodeToString(sellerNegoReq))
			}

			return nil
		},
	)

	wg.Wait()
}

func TestOTNegoAck(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := rlpx.NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	mklroot := utils.MakeRandomMsg(32)
	mode := uint8(ModePlainOTComplaintPoD)
	sessionID := uint64(0xdeadbeaf)

	buyerNegoReq := utils.MakeRandomMsg(32)
	buyerNegoAck := utils.MakeRandomMsg(128)
	sellerNegoReq := utils.MakeRandomMsg(32)
	sellerNegoAck := utils.MakeRandomMsg(128)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if _, err := node.RecvSessionRequest(); err != nil {
				return fmt.Errorf(
					"failed to receive session request: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, mode, mklroot, true,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from server: %v",
					err)
			}

			if _, err := node.RecvSessionAck(false); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on server node: %v",
					err)
			}

			if _, err := node.RecvNegoRequest(new(bytes.Buffer)); err != nil {
				return fmt.Errorf(
					"failed to receive negotiation request: %v",
					err)
			}

			if err := node.SendNegoAckReq(
				bytes.NewReader(sellerNegoAck),
				bytes.NewReader(sellerNegoReq),
				uint64(len(sellerNegoAck)),
				uint64(len(sellerNegoReq)),
			); err != nil {
				return fmt.Errorf(
					"failed to send nego ack+req: %v", err)
			}

			negoRespBuf := new(bytes.Buffer)
			if n, err := node.RecvNegoAck(negoRespBuf); err != nil {
				return fmt.Errorf(
					"failed to receive nego ack: %v", err)
			} else if n != uint64(len(buyerNegoAck)) {
				return fmt.Errorf(
					"invalid buyer nego ack size, get %d bytes, expect %d bytes",
					n, len(buyerNegoAck))
			}
			if node.state != stateNegotiated {
				return fmt.Errorf(
					"server node not in Negotiated state")
			}
			negoResp := negoRespBuf.Bytes()
			if !bytes.Equal(negoResp, buyerNegoAck) {
				return fmt.Errorf(
					"mismatch nego ack, get %s, expect %v",
					hex.EncodeToString(negoResp),
					hex.EncodeToString(buyerNegoAck))
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if err := node.SendNewSessionRequest(mode, mklroot); err != nil {
				return fmt.Errorf(
					"failed to send session request: %v",
					err)
			}

			if _, err := node.RecvSessionAck(true); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on client node: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, mode, mklroot, false,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from client: %v",
					err)
			}

			if err := node.SendNegoRequest(
				bytes.NewReader(buyerNegoReq),
				uint64(len(buyerNegoReq)),
			); err != nil {
				return fmt.Errorf(
					"failed to send negotiation request: %v",
					err)
			}

			if _, _, err := node.RecvNegoAckReq(
				new(bytes.Buffer), new(bytes.Buffer),
			); err != nil {
				return fmt.Errorf(
					"failed to receive nego ack+req: %v",
					err)
			}

			if err := node.SendNegoAck(
				bytes.NewReader(buyerNegoAck),
				uint64(len(buyerNegoAck)),
			); err != nil {
				return fmt.Errorf(
					"failed to send nego ack: %v", err)
			}
			if node.state != stateNegotiated {
				return fmt.Errorf(
					"client node not in Negotiated state")
			}

			return nil
		},
	)

	wg.Wait()
}

func TestOTPod(t *testing.T) {
	var wg sync.WaitGroup

	serverKey := genKey(t)
	clientKey := genKey(t)
	serverAddr, _ := rlpx.NewAddr("127.0.0.1:0", serverKey.PublicKey)

	addrChan := make(chan *net.TCPAddr)

	mklroot := utils.MakeRandomMsg(32)
	mode := uint8(ModePlainOTComplaintPoD)
	sessionID := uint64(0xdeadbeaf)

	buyerNegoReq := utils.MakeRandomMsg(32)
	buyerNegoAck := utils.MakeRandomMsg(128)
	sellerNegoReq := utils.MakeRandomMsg(32)
	sellerNegoAck := utils.MakeRandomMsg(128)

	txReq := utils.MakeRandomMsg(32)
	txResponse := utils.MakeRandomMsg(128)
	txReceipt := utils.MakeRandomMsg(128)

	wg.Add(1)
	go startServer(t, &wg, serverAddr, serverKey, addrChan,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if _, err := node.RecvSessionRequest(); err != nil {
				return fmt.Errorf(
					"failed to receive session request: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, mode, mklroot, true,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from server: %v",
					err)
			}

			if _, err := node.RecvSessionAck(false); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on server node: %v",
					err)
			}

			if _, err := node.RecvNegoRequest(new(bytes.Buffer)); err != nil {
				return fmt.Errorf(
					"failed to receive negotiation request: %v",
					err)
			}

			if err := node.SendNegoAckReq(
				bytes.NewReader(sellerNegoAck),
				bytes.NewReader(sellerNegoReq),
				uint64(len(sellerNegoAck)),
				uint64(len(sellerNegoReq)),
			); err != nil {
				return fmt.Errorf(
					"failed to send nego ack+req: %v", err)
			}

			if _, err := node.RecvNegoAck(new(bytes.Buffer)); err != nil {
				return fmt.Errorf(
					"failed to receive nego ack: %v", err)
			}

			if _, err := node.RecvTxRequest(new(bytes.Buffer)); err != nil {
				return err
			}

			if err := node.SendTxResponse(
				bytes.NewReader(txResponse),
				uint64(len(txResponse)),
			); err != nil {
				return err
			}

			receipt, sig, err := node.RecvTxReceipt()
			if err != nil {
				return err
			}
			if node.state != stateTxReceiptRecvd {
				return fmt.Errorf(
					"server node not in TxReceiptRecvd state")
			}
			if !bytes.Equal(receipt, txReceipt) {
				return fmt.Errorf(
					"mismatch receipt, get %s, expect %s",
					hex.EncodeToString(receipt),
					hex.EncodeToString(txReceipt))
			}

			hash := crypto.Keccak256Hash(receipt).Bytes()
			sigPubkey, err := crypto.Ecrecover(hash, sig)
			if err != nil {
				return fmt.Errorf(
					"Ecrecover(%v, %v) failed: %v",
					hex.EncodeToString(hash),
					hex.EncodeToString(sig),
					err)
			}
			pubkeyBytes := crypto.FromECDSAPub(&clientKey.PublicKey)
			if !bytes.Equal(sigPubkey, pubkeyBytes) {
				return fmt.Errorf(
					"mismatch public keys, get %s, expect %s",
					hex.EncodeToString(sigPubkey),
					hex.EncodeToString(pubkeyBytes))
			}

			return nil
		},
	)

	serverAddr.TCPAddr = <-addrChan

	wg.Add(1)
	go startClient(t, &wg, serverAddr, clientKey,
		func(node *Node, key *ecdsa.PrivateKey) error {
			if err := node.SendNewSessionRequest(mode, mklroot); err != nil {
				return fmt.Errorf(
					"failed to send session request: %v",
					err)
			}

			if _, err := node.RecvSessionAck(true); err != nil {
				return fmt.Errorf(
					"failed to receive session ack on client node: %v",
					err)
			}

			if err := node.SendSessionAck(
				sessionID, mode, mklroot, false,
			); err != nil {
				return fmt.Errorf(
					"failed to send session ack from client: %v",
					err)
			}

			if err := node.SendNegoRequest(
				bytes.NewReader(buyerNegoReq),
				uint64(len(buyerNegoReq)),
			); err != nil {
				return fmt.Errorf(
					"failed to send negotiation request: %v",
					err)
			}

			if _, _, err := node.RecvNegoAckReq(
				new(bytes.Buffer), new(bytes.Buffer),
			); err != nil {
				return fmt.Errorf(
					"failed to receive nego ack+req: %v",
					err)
			}

			if err := node.SendNegoAck(
				bytes.NewReader(buyerNegoAck),
				uint64(len(buyerNegoAck)),
			); err != nil {
				return fmt.Errorf(
					"failed to send nego ack: %v", err)
			}

			if err := node.SendTxRequest(
				bytes.NewReader(txReq),
				uint64(len(txReq)),
			); err != nil {
				return fmt.Errorf(
					"failed to send Tx request: %v", err)
			}

			if _, err := node.RecvTxResponse(new(bytes.Buffer)); err != nil {
				return err
			}

			if err := node.SendTxReceipt(
				bytes.NewReader(txReceipt),
				uint64(len(txReceipt)),
			); err != nil {
				return err
			}
			if node.state != stateTxSecretWait {
				return fmt.Errorf(
					"client side not in TxSecretWait state")
			}

			return nil
		},
	)

	wg.Wait()
}
