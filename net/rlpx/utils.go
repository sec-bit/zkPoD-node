package rlpx

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sec-bit/zkPoD-node/net/utils"
)

const (
	tcpTimeout = 30 * time.Second
)

func tcpSend(conn net.Conn, bs []byte) error {
	remaining := len(bs)
	cur := 0

	conn.SetWriteDeadline(time.Now().Add(tcpTimeout))

	for remaining > 0 {
		n, err := conn.Write(bs[cur:])
		if err != nil {
			return err
		}
		remaining -= n
		cur += n
	}

	return nil
}

func tcpRecv(conn net.Conn, buf []byte) error {
	remaining := len(buf)
	cur := 0

	conn.SetReadDeadline(time.Now().Add(tcpTimeout))

	for remaining > 0 {
		n, err := conn.Read(buf[cur:])
		if err != nil {
			return err
		}
		remaining -= n
		cur += n
	}

	return nil
}

func encodePreAuthMsg(msg *preAuthMsg) []byte {
	buffer := new(bytes.Buffer)

	buffer.Write(utils.EncodeUint8(msg.rev))
	buffer.Write(utils.EncodeUint8(msg.typ))
	buffer.Write(utils.EncodeUint64(msg.size))
	if msg.payload != nil {
		buffer.Write(msg.payload)
	}

	return buffer.Bytes()
}

func decodePreAuthMsg(conn net.Conn, maxSize uint64) (*preAuthMsg, error) {
	u8Buf := make([]byte, 1)

	if err := tcpRecv(conn, u8Buf); err != nil {
		return nil, fmt.Errorf(
			"failed to receive preauth rev: %v", err)
	}

	rev, err := utils.DecodeUint8(u8Buf)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decode preauth rev: %v", err)
	}
	if rev != preAuthRev {
		return nil, fmt.Errorf(
			"unsupported preauth rev %d, expect %d", rev, preAuthRev)
	}

	if err := tcpRecv(conn, u8Buf); err != nil {
		return nil, fmt.Errorf(
			"failed to receive preauth type: %v", err)
	}

	typ, err := utils.DecodeUint8(u8Buf)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decode preauth type: %v", err)
	}
	if typ >= preAuthMax {
		return nil, fmt.Errorf(
			"unsupported preauth msg type %d", typ)
	}

	u64Buf := make([]byte, 8)

	if err := tcpRecv(conn, u64Buf); err != nil {
		return nil, fmt.Errorf(
			"failed to receive preauth size: %v", err)
	}

	size, err := utils.DecodeUint64(u64Buf)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to decode preauth size: %v", err)
	}
	if maxSize != 0 && size > maxSize {
		return nil, fmt.Errorf(
			"preauth size %d too large (>= %d)", size, maxSize)
	}

	payload := make([]byte, size)

	if err := tcpRecv(conn, payload); err != nil {
		return nil, fmt.Errorf(
			"failed to receive preauth payload: %v", err)
	}

	return &preAuthMsg{
		rev:     rev,
		typ:     typ,
		size:    size,
		payload: payload,
	}, nil
}

func isPubKeyEqual(k0, k1 *ecdsa.PublicKey) bool {
	bs0 := crypto.FromECDSAPub(k0)
	bs1 := crypto.FromECDSAPub(k1)
	return bytes.Equal(bs0, bs1)
}

func pubkeyToString(k *ecdsa.PublicKey) string {
	return hex.EncodeToString(crypto.FromECDSAPub(k))
}
