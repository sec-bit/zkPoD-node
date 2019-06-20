package net

import (
	"bytes"
	"fmt"

	"github.com/sec-bit/zkPoD-node/net/utils"
)

// Session State Transitions:
//
//  Alice (A)                                                Bob (B)
//  ==========                                                =========
//
//  Connected                       req                       Connected
//      | (2) RecvSessionRequest  <-----   (1) SendSessionRequest |
//      V                                                         V
//  SessionReqRecvd                ack                        SessionAckWait
//      | (3) SendSessionAck      ----->       (4) RecvSessionAck |
//      V                                                         V
//  SessionAckWait                 ack                        SessionAckRecvd
//      | (6) RecvSessionAck      <-----       (5) SendSessionAck |
//      V                                                         V
//  SessionEstablished                                        SessionEstablished

type (
	// SessionRequest represents a session request.
	//
	// If `ID` is 0, request will be interpreted as a request to
	// create a new session.
	SessionRequest struct {
		ID           uint64
		Mode         uint8
		SigmaMklRoot []byte
		ExtraInfo    []byte
	}
)

const (
	sessionRequestSize = 41 // uint64 + uint8 + [32]byte
	sessionAckSize     = sessionRequestSize
)

// encodeSessionRequest encodes a session request into a byte array
// in the following format:
//   Byte 0 - 7 : session ID
//   Byte 8     : session mode (refer to Mode... constants in node.go)
//   Byte 9 - 40: sigma merkle root (256 bits)
func encodeSessionRequest(req *SessionRequest) []byte {
	buf := new(bytes.Buffer)

	buf.Write(utils.EncodeUint64(req.ID))
	buf.Write(utils.EncodeUint8(req.Mode))
	buf.Write(req.SigmaMklRoot)
	buf.Write(req.ExtraInfo)
	return buf.Bytes()
}

// decodeSessionRequest decodes a byte array to a session request.
// Refer to the comment of encodeSessionRequest for the format of the
// byte array.
func decodeSessionRequest(bs []byte) (*SessionRequest, error) {
	size := len(bs)

	if size < sessionRequestSize {
		return nil, fmt.Errorf(
			"invalid session request (%v) size %d, expect >= %d",
			bs, size, sessionRequestSize)
	}

	id, err := utils.DecodeUint64(bs[0:8])
	if err != nil {
		return nil, fmt.Errorf("failed to decode session ID: %v", err)
	}

	mode, err := utils.DecodeUint8(bs[8:9])
	if err != nil {
		return nil, fmt.Errorf("failed to decode session mode: %v", err)
	}

	return &SessionRequest{
		ID:           id,
		Mode:         mode,
		SigmaMklRoot: bs[9:41],
		ExtraInfo:    bs[41:],
	}, nil
}

// SendNewSessionRequest sends a request to create a new session that trades
// the data of the specified sigma merkle root in the specified mode.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the request is sent out
//  - any timeout occurs
//  - any error occurs
//
// PreState : Connected
// PostState: SessionAckWait
//
// Parameters:
//  - mode: the trading mode, refer to Mode... constants.
//  - sigmaMklRoot: the sigma merkle root of the data
//
// Return:
//  If no error occurs, return nil. Otherwise, return a non-nil error.
func (node *Node) SendNewSessionRequest(mode uint8, sigmaMklRoot []byte, extraInfo []byte) error {
	if node.state != stateConnected {
		return invalidStateError(node, stateConnected)
	}

	if mode < 0 || mode >= modeMax {
		return fmt.Errorf("invalid session request mode %d", mode)
	}

	if len(sigmaMklRoot) != 32 {
		return fmt.Errorf(
			"invalid session sigma merkle root (%v) size %d bytes, expect 32 bytes",
			sigmaMklRoot, len(sigmaMklRoot))
	}

	req := &SessionRequest{
		ID:           0,
		Mode:         uint8(mode),
		SigmaMklRoot: sigmaMklRoot,
		ExtraInfo:    extraInfo,
	}
	reqBytes := encodeSessionRequest(req)
	reqBuf := bytes.NewReader(reqBytes)
	if err := node.sendMsg(
		msgSessionRequest, reqBuf, uint64(len(reqBytes)),
	); err != nil {
		return fmt.Errorf("failed to send session request %v: %v",
			req, err)
	}

	node.state = stateSessionAckWait

	return nil
}

// RecvSessionRequest receives a session request.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the request is received
//  - any timeout occurs
//  - any error occurs
//
// PreState : Connected
// PostState: SessionReqRecvd
//
// Return:
//  If no error occurs, return the session request and a nil error.
//  Otherwise, return nil request and a non-nil error.
func (node *Node) RecvSessionRequest() (*SessionRequest, error) {
	if node.state != stateConnected {
		return nil, invalidStateError(node, stateConnected)
	}

	buf := new(bytes.Buffer)
	if err := node.recvMsgWithCheck(
		buf, msgSessionRequest, sessionRequestSize,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to receive session request: %v", err)
	}
	payload := buf.Bytes()

	req, err := decodeSessionRequest(payload)
	if err != nil {
		return nil, err
	}
	if req.ID == 0 {
		if req.Mode >= modeMax {
			return nil, fmt.Errorf(
				"invalid mode %d in session request", req.Mode)
		}
	}

	node.state = stateSessionReqRecvd

	return req, nil
}

// SendSessionAck sends an acknowledgment for the session request.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the ack is sent out
//  - any timeout occurs
//  - any error occurs
//
// When called by the node that receives the session request:
//   PreState : SessionReqRecvd
//   PostState: SessionAckWait
//
// When called by the node that sends the session request:
//   PreState : SessionAckRecvd
//   PostState: SessionEstablished
//
// Parameters:
//  - id: the session ID
//  - mode: the trading mode
//  - mklroot: the sigma merkle root of the data
//  - needFurtherAck: does the caller need to receive a further session ack?
//
// Return:
//  If no error occurs, return nil. Otherwise, return the error.
func (node *Node) SendSessionAck(
	id uint64, mode uint8, mklroot []byte, extra []byte, needFurtherAck bool,
) error {
	if needFurtherAck && node.state != stateSessionReqRecvd {
		// server/Alice side
		return invalidStateError(node, stateSessionReqRecvd)
	} else if !needFurtherAck && node.state != stateSessionAckRecvd {
		return invalidStateError(node, stateSessionAckRecvd)
	}

	ack := &SessionRequest{
		ID:           id,
		Mode:         mode,
		SigmaMklRoot: mklroot,
		ExtraInfo:    extra,
	}
	ackBytes := encodeSessionRequest(ack)
	ackBuf := bytes.NewReader(ackBytes)
	if err := node.sendMsg(
		msgSessionAck, ackBuf, uint64(len(ackBytes)),
	); err != nil {
		return fmt.Errorf("failed to send session ack %v: %v",
			ack, err)
	}

	if needFurtherAck {
		node.state = stateSessionAckWait
	} else {
		node.session.id = ack.ID
		node.session.mode = ack.Mode
		node.state = stateSessionEstablished
	}

	return nil
}

// RecvSessionAck receives a session acknowledgment.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the ack is received
//  - any timeout occurs
//  - any error occurs
//
// When called by the node that receives the session request:
//   PreState : SessionAckWait
//   PostState: SessionEstablished
//
// When called by the node that sends the session request:
//   PreState : SessionAckWait
//   PostState: SessionAckRecvd
//
// Parameters:
//  - needFurtherAck: does the caller need to echo the acknowledgment?
//
// Return:
//  If no error occurs, return the acknowledgment and a nil error.
//  Otherwise, return a nil acknowledgment and the error.
func (node *Node) RecvSessionAck(needFurtherAck bool) (*SessionRequest, error) {
	if node.state != stateSessionAckWait {
		return nil, invalidStateError(node, stateSessionAckWait)
	}

	buf := new(bytes.Buffer)
	if err := node.recvMsgWithCheck(
		buf, msgSessionAck, sessionAckSize,
	); err != nil {
		return nil, fmt.Errorf("failed to receive session ack: %v", err)
	}
	payload := buf.Bytes()

	ack, err := decodeSessionRequest(payload)
	if err != nil {
		return nil, err
	}
	if ack.ID == 0 {
		return nil, fmt.Errorf("zero session ID in session ack")
	}

	if needFurtherAck {
		node.state = stateSessionAckRecvd
	} else {
		node.session.id = ack.ID
		node.session.mode = ack.Mode
		node.state = stateSessionEstablished
	}

	return ack, nil
}
