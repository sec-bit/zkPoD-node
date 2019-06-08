package rlpx

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io"
	"net"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sec-bit/zkPoD-node/net/rlpx/p2p"
	"github.com/sec-bit/zkPoD-node/net/utils"
)

type (
	// Addr defines the address of a PoD node.
	Addr struct {
		// TCPAddr provides the TCP address of the PoD node
		TCPAddr *net.TCPAddr
		// EthAddr provides the Ethereum address of the
		// counterpart of the PoD node on the Ethereum
		// network.
		EthAddr common.Address
	}

	// Connection represents a connection between PoD nodes.
	Connection struct {
		laddr    *Addr
		raddr    *Addr
		tcpConn  net.Conn
		rlpxConn *p2p.RLPx
	}

	// Listener represents a network agent listening for incoming
	// connections.
	Listener struct {
		addr        *Addr
		tcpListener net.Listener
	}

	preAuthMsg struct {
		rev     uint8
		typ     uint8
		size    uint64
		payload []byte
	}
)

const (
	preAuthRequst = iota
	preAuthAck
	preAuthMax // Leave at the end intentionally
)

const (
	preAuthRev     = 1
	preAuthReqSize = 1024
	preAuthSigSize = 65
	preAuthPubSize = 65
	preAuthAckSize = preAuthSigSize + preAuthPubSize

	// RLPx message code used in PoD transport protocol, which has
	// nothing to do with the message codes used in Ethereum. We
	// just (randomly) select the following value.
	rlpxMsgCodeSize = 0xf0
	rlpxMsgCodeData = 0xf1

	rlpxMsgDataPacketSize = 16777210
)

// NewAddr creates a new PoD address from the given TCP address and
// the ethereum public key.
//
// Parameters:
//  - tcpAddrStr: the TCP address of the node, in the form of
//                "IP address:port" or "hostname:port"
//  - pubKey: the Ethereum public key of the node
//
// Return:
//  If no error occurs, return an Addr object and a nil error.
//  Otherwise, return a nil Addr object and a non-nil error.
func NewAddr(tcpAddrStr string, pubKey ecdsa.PublicKey) (*Addr, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", tcpAddrStr)
	if err != nil {
		return nil, err
	}

	return &Addr{
		TCPAddr: tcpAddr,
		EthAddr: crypto.PubkeyToAddress(pubKey),
	}, nil
}

// Network returns the network name of PoD node.
func (addr *Addr) Network() string {
	return "tcp+rlpx"
}

func (addr *Addr) String() string {
	return fmt.Sprintf("(tcp %s, eth %s)",
		addr.TCPAddr,
		addr.EthAddr.Hex())
}

// Listen listens on the specified address for incoming connections.
//
// Parameters:
//  - addr: the address being listened on
//
// Return:
//  If no error occurs, return a Listener object and nil error. Otherwise,
//  return a nil Listener object and a non-nil error.
func Listen(addr *Addr) (*Listener, error) {
	tcpListener, err := net.Listen("tcp", addr.TCPAddr.String())
	if err != nil {
		return nil, fmt.Errorf(
			"failed to listen on tcp address %s: %v", addr.TCPAddr, err)
	}

	return &Listener{
		addr: &Addr{
			TCPAddr: tcpListener.Addr().(*net.TCPAddr),
			EthAddr: addr.EthAddr,
		},
		tcpListener: tcpListener,
	}, nil
}

// Addr returns the node address being listened on.
func (l *Listener) Addr() *Addr {
	return l.addr
}

// Close terminates the Listener.
//
// Return:
//  If no error occurs, return nil. Otherwise, return the error.
func (l *Listener) Close() error {
	return l.tcpListener.Close()
}

// Accept accepts an incoming connection.
//
// Return:
//  If no error occurs, return the incoming connection and a nil error.
//  Otherwise, return a nil connection and a non-nil error.
func (l *Listener) Accept() (*Connection, error) {
	conn, err := l.tcpListener.Accept()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to accept connection on local address: %v", err)
	}

	return &Connection{
		laddr:   l.addr,
		tcpConn: conn,
	}, nil
}

func (conn Connection) String() string {
	var from, to string
	if conn.laddr != nil {
		from = conn.laddr.String()
	}
	if conn.raddr != nil {
		to = conn.raddr.String()
	}
	return fmt.Sprintf("[%s <--> %s]", from, to)
}

// Dial tries to dial the specified PoD node.
//
// Parameters:
//  - addr: the address of the PoD node
//
// Return:
//  1. If no error occurs, return the connection and a nil error.
//  2. Otherwise, return a nil connection and a non-nil error.
func Dial(addr *Addr) (*Connection, error) {
	tcpConn, err := net.Dial("tcp", addr.TCPAddr.String())
	if err != nil {
		return nil, fmt.Errorf(
			"failed to dial %s: %v", addr, err)
	}

	return &Connection{
		laddr:   &Addr{TCPAddr: tcpConn.LocalAddr().(*net.TCPAddr)},
		raddr:   addr,
		tcpConn: tcpConn,
	}, nil
}

// Close terminates the connection.
func (conn *Connection) Close() error {
	return conn.tcpConn.Close()
}

// LocalAddr returns the local address of the connection.
func (conn *Connection) LocalAddr() *Addr {
	return conn.laddr
}

// LocalAddr returns the remote address of the connection.
func (conn *Connection) RemoteAddr() *Addr {
	return conn.raddr
}

// Handshake performs the handshake with the PoD node on the other
// side of the connection. It is usually buyer node initiates the
// handshake.
//
// Parameters:
//  - key: the Ethereum private key of the local PoD node
//  - initiator: If true, the local PoD node will initiating the handshake.
//               Otherwise, it will wait for the handshake from the other side.
//
// Return:
//  - If no error occurs, return the Ethereum public key of the other side,
//    and a nil error.
//  - Otherwise, return a nil public key and the non-nil error.
func (conn *Connection) Handshake(
	key *ecdsa.PrivateKey, initiator bool,
) (*ecdsa.PublicKey, error) {
	if initiator {
		return conn.initiatorHandshake(key)
	} else {
		return conn.receiverHandshake(key)
	}
}

func (conn *Connection) initiatorHandshake(key *ecdsa.PrivateKey) (*ecdsa.PublicKey, error) {
	rPubkey, err := conn.initiatorPreAuthHandshake()
	if err != nil {
		return nil, err
	}

	return rPubkey, conn.initiatorRLPXHandshake(key, rPubkey)
}

func (conn *Connection) receiverHandshake(key *ecdsa.PrivateKey) (*ecdsa.PublicKey, error) {
	if err := conn.receiverPreAuthHandshake(key); err != nil {
		return nil, err
	}

	return conn.receiverRLPXHandshake(key)
}

func (conn *Connection) initiatorPreAuthHandshake() (*ecdsa.PublicKey, error) {
	random := utils.MakeRandomMsg(preAuthReqSize)

	if err := conn.sendPreAuthRequest(random); err != nil {
		return nil, err
	}

	pubkey, signature, err := conn.recvPreAuthAck()
	if err != nil {
		return nil, err
	}

	if err := verifyPreAuthAck(random, conn.raddr, signature, pubkey); err != nil {
		return nil, err
	}

	return pubkey, nil
}

func (conn *Connection) receiverPreAuthHandshake(key *ecdsa.PrivateKey) error {
	random, err := conn.recvPreAuthRequest()
	if err != nil {
		return err
	}

	if err := conn.sendPreAuthAck(key, random); err != nil {
		return err
	}

	return nil
}

func (conn *Connection) sendPreAuthRequest(random []byte) error {
	msg := &preAuthMsg{
		rev:     preAuthRev,
		typ:     preAuthRequst,
		size:    preAuthReqSize,
		payload: random,
	}
	bs := encodePreAuthMsg(msg)

	if err := tcpSend(conn.tcpConn, bs); err != nil {
		return fmt.Errorf(
			"failed to send preauth request: %v", err)
	}

	return nil
}

func (conn *Connection) recvPreAuthRequest() ([]byte, error) {
	msg, err := decodePreAuthMsg(conn.tcpConn, preAuthReqSize)
	if err != nil {
		return nil, err
	}
	return msg.payload, nil
}

func (conn *Connection) sendPreAuthAck(key *ecdsa.PrivateKey, random []byte) error {
	signature, err := crypto.Sign(crypto.Keccak256Hash(random).Bytes(), key)
	if err != nil {
		return fmt.Errorf(
			"failed to sign %v by private key %s in preauth ack: %v",
			random,
			hex.EncodeToString(crypto.FromECDSA(key)),
			err)
	}
	pubkey := crypto.FromECDSAPub(&key.PublicKey)

	payloadBuf := new(bytes.Buffer)
	payloadBuf.Write(signature)
	payloadBuf.Write(pubkey)

	msg := &preAuthMsg{
		rev:     preAuthRev,
		typ:     preAuthAck,
		size:    preAuthAckSize,
		payload: payloadBuf.Bytes(),
	}
	bs := encodePreAuthMsg(msg)

	if err := tcpSend(conn.tcpConn, bs); err != nil {
		return fmt.Errorf("failed to send preauth ack: %v", err)
	}

	return nil
}

func (conn *Connection) recvPreAuthAck() (*ecdsa.PublicKey, []byte, error) {
	msg, err := decodePreAuthMsg(conn.tcpConn, preAuthAckSize)
	if err != nil {
		return nil, nil, err
	}

	payload := msg.payload

	signature := payload[0:preAuthSigSize]

	pubkeyBytes := payload[preAuthSigSize:preAuthAckSize]
	pubkey, err := crypto.UnmarshalPubkey(pubkeyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to extract remote public key from %v: %v",
			hex.EncodeToString(pubkeyBytes), err)
	}

	return pubkey, signature, nil
}

func verifyPreAuthAck(
	random []byte, raddr *Addr, signature []byte, pubkey *ecdsa.PublicKey,
) error {
	pubkeyBytes := crypto.FromECDSAPub(pubkey)

	ethAddr := crypto.PubkeyToAddress(*pubkey)
	if !bytes.Equal(ethAddr.Bytes(), raddr.EthAddr.Bytes()) {
		return fmt.Errorf(
			"remote public key %s not match with remote eth address %s",
			hex.EncodeToString(pubkeyBytes),
			raddr.EthAddr.Hex())
	}

	hash := crypto.Keccak256Hash(random).Bytes()
	sigPubkey, err := crypto.Ecrecover(hash, signature)
	if err != nil {
		return fmt.Errorf(
			"Ecrecover(%v, %v) failed: %v", random, signature, err)
	}
	if !bytes.Equal(sigPubkey, pubkeyBytes) {
		return fmt.Errorf(
			"remote pubkey %s not match with recovered pubkey %s",
			hex.EncodeToString(pubkeyBytes),
			hex.EncodeToString(sigPubkey))
	}

	return nil
}

func (conn *Connection) initiatorRLPXHandshake(
	key *ecdsa.PrivateKey, rPubkey *ecdsa.PublicKey,
) error {
	rlpxConn := p2p.NewRLPx(conn.tcpConn)

	pubkey, err := rlpxConn.DoEncHandshake(key, rPubkey)
	if err != nil {
		return fmt.Errorf(
			"rlpx enc handshake failed on initiator: %v", err)
	}

	rPubkeyBytes := crypto.FromECDSAPub(rPubkey)
	pubkeyBytes := crypto.FromECDSAPub(pubkey)
	if !bytes.Equal(rPubkeyBytes, pubkeyBytes) {
		return fmt.Errorf(
			"received receiver pubkey %s not match with %s",
			hex.EncodeToString(pubkeyBytes),
			hex.EncodeToString(rPubkeyBytes))
	}

	conn.rlpxConn = rlpxConn

	return nil
}

func (conn *Connection) receiverRLPXHandshake(
	key *ecdsa.PrivateKey,
) (*ecdsa.PublicKey, error) {
	rlpxConn := p2p.NewRLPx(conn.tcpConn)

	pubkey, err := rlpxConn.DoEncHandshake(key, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"rlpx enc handshake failed on receiver: %v", err)
	}

	conn.rlpxConn = rlpxConn

	return pubkey, nil
}

// Write writes data to the connection.
//
// Parameters:
//  - data: an io.Reader for the data to be sent
//  - size: the number of bytes to be sent
//
// Return:
//  If no error occurs, return nil. Otherwise, return the non-nil error.
func (conn *Connection) Write(data io.Reader, size uint64) error {
	if err := conn.writeSize(size); err != nil {
		return err
	}

	if err := conn.writeData(data, size); err != nil {
		return err
	}

	return nil
}

func (conn *Connection) writeSize(size uint64) error {
	sizeBytes := utils.EncodeUint64(size)

	if err := conn.rlpxConn.WriteMsg(
		p2p.Msg{
			Code:    rlpxMsgCodeSize,
			Size:    uint32(len(sizeBytes)),
			Payload: bytes.NewReader(sizeBytes),
		},
	); err != nil {
		return fmt.Errorf("failed to send payload size: %v", err)
	}

	return nil
}

func (conn *Connection) writeData(data io.Reader, size uint64) error {
	remaining := size
	cur := uint64(0)

	buf := new(bytes.Buffer)

	for remaining > 0 {
		buf.Reset()

		nrBytes := remaining
		if nrBytes > rlpxMsgDataPacketSize {
			nrBytes = rlpxMsgDataPacketSize
		}

		if _, err := io.CopyN(buf, data, int64(nrBytes)); err != nil {
			return fmt.Errorf(
				"failed to get data bytes %d - %d: %v",
				cur, cur+nrBytes, err)
		}

		if err := conn.rlpxConn.WriteMsg(
			p2p.Msg{
				Code:    rlpxMsgCodeData,
				Size:    uint32(nrBytes),
				Payload: buf,
			},
		); err != nil {
			return fmt.Errorf(
				"failed to send payload bytes %d - %d: %v",
				cur, cur+nrBytes, err)
		}

		cur += nrBytes
		remaining -= nrBytes
	}

	return nil
}

// Read reads data from the connection.
//
// Parameters:
//  - buf: the buffer to store the data
//
// Return:
//  - If no error occurs, return the number of received bytes and a nil error.
//  - Otherwise, return the number of received bytes and the non-nil error.
func (conn *Connection) Read(buf io.Writer) (uint64, error) {
	size, err := conn.readSize()
	if err != nil {
		return 0, err
	}

	if err := conn.readData(buf, size); err != nil {
		return 0, err
	}

	return size, nil
}

func (conn *Connection) readSize() (uint64, error) {
	msg, err := conn.rlpxConn.ReadMsg()
	if err != nil {
		return 0, err
	}
	if msg.Code != rlpxMsgCodeSize {
		return 0, fmt.Errorf(
			"not size message, get msg code %d, expect %d",
			msg.Code, rlpxMsgCodeSize)
	}

	payload := make([]byte, 8)
	payloadSize, err := msg.Payload.Read(payload)
	if err != nil {
		return 0, fmt.Errorf("failed to decode payload size: %v", err)
	}
	if payloadSize != 8 {
		return 0, fmt.Errorf(
			"partially received payload size: %d out of 8 bytes",
			payloadSize)
	}

	return utils.DecodeUint64(payload)
}

func (conn *Connection) readData(buf io.Writer, length uint64) error {
	remaining := length
	cur := uint32(0)

	for remaining > 0 {
		msg, err := conn.rlpxConn.ReadMsg()
		if err != nil {
			return fmt.Errorf(
				"failed to receive payload at byte %d: %v",
				cur, err)
		}
		if msg.Code != rlpxMsgCodeData {
			return fmt.Errorf(
				"not data message, get msg code %d, expect %d",
				msg.Code, rlpxMsgCodeData)
		}

		size := msg.Size
		if size == 0 {
			continue
		}

		if _, err := io.CopyN(buf, msg.Payload, int64(size)); err != nil {
			return fmt.Errorf(
				"failed to receive payload bytes %d - %d: %v",
				cur, cur+size, err)
		}

		remaining -= uint64(size)
		cur += size
	}

	return nil
}
