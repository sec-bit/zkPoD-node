package net

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"io"

	"github.com/sec-bit/zkPoD-node/net/rlpx"
	"github.com/sec-bit/zkPoD-node/net/utils"
)

type (
	// Node represents a communication node.
	Node struct {
		conn    *rlpx.Connection
		state   nodeState
		session session
		lkey    *ecdsa.PrivateKey
		rkey    *ecdsa.PublicKey
	}

	session struct {
		id   uint64
		mode uint8
	}
)

const (
	msgRev = 1
)

const (
	msgSessionRequest = iota
	msgSessionAck
	msgSessionClose
	msgTxRequest
	msgTxResponse
	msgTxReceipt
	msgNegoRequest
	msgNegoAck
	msgNegoAckReq // NegoAck + NegoRequest
	msgTypeMax    // Leave at the end intentionally
)

const (
	// XXX: Remember to update isOTMode() whenever a new OT mode is added.
	// XXX: Always add new modes at the end of existing modes in order to
	//      to keep the backwards compatibility.

	ModePlainComplaintPoD = iota
	ModePlainOTComplaintPoD
	ModePlainAtomicSwapPoD

	ModeTableVRFQuery
	ModeTableOTVRFQuery
	ModeTableComplaintPoD
	ModeTableOTComplaintPoD
	ModeTableAtomicSwapPoD

	modeMax // Leave at the end intentionally
)

// Create a node for communication. A non-nil connection must be provided.
//
// PreState : N/A
// PostState: Connected
//
// Parameters:
//  - conn: a network connection
//  - lkey: the ethereum private key of the node to be created
//  - rkey: the ethereum public key of the remote node
//
// Return:
//  If no error occurs, return a node for communication and a nil error.
//  Otherwise, return a nil node and the non-nil error.
//
// Examples:
// 1. Create a communication node for Alice.
//    ```
//    import (
//        "crypto/ecdsa"
//        pod_net "github.com/sec-bit/zkPoD-node/net"
//        "github.com/sec-bit/zkPoD-node/net/rlpx"
//    )
//
//    func Alice(AliceTCPAddr string, AliceEthPrivkey *ecdsa.PrivateKey) error {
//        addr, err := rlpx.NewAddr(AliceTCOAddr, AliceEthPrivkey.PublicKey)
//        if err != nil {
//            return err
//        }
//
//        l, err := rlpx.Listen(addr)
//        if err != nil {
//            return err
//        }
//        defer l.Close()
//
//        conn, err := l.Accept()
//        if err != nil {
//            return err
//        }
//        // Note: conn should be alive until the node is closed.
//        defer conn.Close()
//
//        BobEthPubkey, err := conn.Handshake(AliceEthPrivkey, false)
//        if err != nil {
//            return err
//        }
//
//        node, err := pod_net.NewNode(conn, AliceEthPrivkey, BobEthPubkey)
//        if err != nil {
//            return err
//        }
//        defer node.Close()
//
//        // do something with the node
//        // ...
//
//        return nil
//    }
//    ```
//
// 2. Create a communication node for Bob.
//    ```
//    import (
//        "crypto/ecdsa"
//        "net"
//        "github.com/ethereum/go-ethereum/common"
//        pod_net "github.com/sec-bit/zkPoD-node/net"
//        "github.com/sec-bit/zkPoD-node/net/rlpx"
//    )
//
//    func Bob(AliceTCPAddr, AliceEthAddr string, BobEthPrivkey *ecdsa.PrivateKey) error {
//        tcpAddr, err := net.ResolveTCPAddr("tcp", AliceTCPAddr)
//        if err != nil {
//            return err
//        }
//        ethAddr := common.HexToAddress(AliceEthAddr)
//        addr := rlpx.Addr{TCPAddr: tcpAddr, EthAddr: ethAddr}
//
//        conn, err := rlpx.Dial(addr)
//        if err != nil {
//            return err
//        }
//        defer conn.Close()
//
//        AliceEthPubkey, err := conn.Handshake(BobEthPrivkey, true)
//        if err != nil {
//            return err
//        }
//
//        node, err := pod_net.NewNode(conn, BobEthPrivkey, AliceEthPubkey)
//        if err != nil {
//            return nil
//        }
//        defer node.Close()
//
//        // do something with the node
//        // ...
//
//        return nil
//    }
//    ```
func NewNode(
	conn *rlpx.Connection, lkey *ecdsa.PrivateKey, rkey *ecdsa.PublicKey,
) (*Node, error) {
	if conn == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}

	return &Node{
		conn:  conn,
		state: stateConnected,
		lkey:  lkey,
		rkey:  rkey,
	}, nil
}

func (node Node) String() string {
	return fmt.Sprintf("node@%s", node.conn)
}

// Close closes the node.
//
// PreState : any
// PostState: Closed
func (node *Node) Close() error {
	node.state = stateClosed
	return nil
}

// sendMsg composes a message from the given message type and payload,
// and sends it out.
//
// The composed message is in the following form
//   Byte 0 - 1: message revision (1)
//   Byte 2 - 3: message type (typ)
//   Byte 4 - 11: payload length in byte
//   Byte 12 - _: payload
//
// The first 12 bytes composes the message header.
func (node *Node) sendMsg(typ uint16, payload io.Reader, size uint64) error {
	c := node.conn

	buf := new(bytes.Buffer)
	buf.Write(utils.EncodeUint16(msgRev))
	buf.Write(utils.EncodeUint16(typ))
	buf.Write(utils.EncodeUint64(size))

	if err := c.Write(buf, 12); err != nil {
		return fmt.Errorf("failed to send msg header: %v", err)
	}

	if err := c.Write(payload, size); err != nil {
		return fmt.Errorf("failed to send msg payload: %v", err)
	}

	return nil
}

// recvMsgHeader receives, checks and decomposes the message header.
//
// Return:
//  message type, payload length in byte, error (if any)
func (node *Node) recvMsgHeader() (uint16, uint64, error) {
	c := node.conn

	headerBuf := new(bytes.Buffer)
	if n, err := c.Read(headerBuf); err != nil {
		return 0, 0, fmt.Errorf(
			"failed to receive msg header: %v", err)
	} else if n != 12 {
		return 0, 0, fmt.Errorf(
			"invalid size of msg header, get %d bytes, expect 12 bytes",
			n)
	}

	header := headerBuf.Bytes()

	rev, err := utils.DecodeUint16(header[0:2])
	if err != nil {
		return 0, 0, fmt.Errorf(
			"failed to decode msg rev: %v", err)
	}
	if rev != msgRev {
		return 0, 0, fmt.Errorf(
			"unsupported msg rev %d", rev)
	}

	typ, err := utils.DecodeUint16(header[2:4])
	if err != nil {
		return 0, 0, fmt.Errorf(
			"failed to decode msg type: %v", err)
	}
	if typ >= msgTypeMax {
		return 0, 0, fmt.Errorf(
			"unsupported msg type %d", typ)
	}

	length, err := utils.DecodeUint64(header[4:12])
	if err != nil {
		return 0, 0, fmt.Errorf(
			"failed to decode msg payload size: %v", err)
	}

	return typ, length, nil
}

// recvMsgPayload receives at most `length` bytes of payload.
//
// Return:
//  error (if any)
func (node *Node) recvMsgPayload(dst io.Writer, length uint64) error {
	c := node.conn

	if n, err := c.Read(dst); err != nil {
		return fmt.Errorf(
			"failed to receive msg payload: %v", err)
	} else if n != length {
		return fmt.Errorf(
			"invalid payload size, get %d bytes, expect %d bytes",
			n, length)
	}

	return nil
}

// recvMsgWithCheck combines recvMsgHeader and recvMsgPayload.
//
// If the received message is not of the specified type `typ`, or the
// payload is not in the specified length `length`, this function will
// return error. Otherwise, it will return the payload.
func (node *Node) recvMsgWithCheck(dst io.Writer, typ uint16, length uint64) error {
	t, l, err := node.recvMsgHeader()
	if err != nil {
		return err
	}

	if t != typ {
		return fmt.Errorf(
			"mismatch msg type, get %d, expect %d", t, typ)
	}

	if l < length {
		return fmt.Errorf(
			"mismatch msg length, get %d, expect >= %d", l, length)
	}

	return node.recvMsgPayload(dst, l)
}

func (node *Node) isOTMode() bool {
	switch node.session.mode {
	case ModePlainOTComplaintPoD:
		return true
	case ModeTableOTComplaintPoD:
		return true
	case ModeTableOTVRFQuery:
		return true
	}
	return false
}
