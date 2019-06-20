package net

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sec-bit/zkPoD-node/net/utils"
)

// Transaction State Transitions
//
// 1. Non-OT mode
//
//  Alice (A)                                           Bob (B)
//  ==========                                           =========
//
//  SessionEstablished           request                 SessionEstablished
//      | (2) RecvTxRequest   <-----------  (1) SendTxRequest |
//      V                                                     V
//  TxRequestRecvd              response                 TxResponseWait
//      | (3) SendTxResponse  -----------> (4) RecvTxResponse |
//      V                                                     V
//  TxReceiptWait                receipt                 TxResponseRecvd
//      | (6) RecvTxReceipt   <-----------  (5) SendTxReceipt |
//      V                                                     V
//  TxReceiptRecvd                                       TxSecretWait
//
// 2. OT mode
//
//  Alice (A)                                             Bob (B)
//  ==========                                             =========
//
//  SessionEstablished          nego req                   SessionEstablished
//      | (2) RecvNegoRequest <-----------  (1) SendNegoRequest |
//      V                                                       V
//  NegoRequestRecvd          nego req+ack                 NegoAckReqWait
//      | (3) SendNegoAckReq  ----------->   (4) RecvNegoAckReq |
//      V                                                       V
//  NegoAckWait                 nego ack                   NegoAckReqRecvd
//      | (6) RecvNegoAck     <-----------      (5) SendNegoAck |
//      V                                                       V
//  Negotiated                   request                   Negotiated
//      | (8) RecvTxRequest   <-----------    (7) SendTxRequest |
//      V                                                       V
//  TxRequestRecvd              response                   TxResponseWait
//      | (9) SendTxResponse  -----------> (10) RecvTxResponse |
//      V                                                      V
//  TxReceiptWait                receipt                   TxResponseRecvd
//      | (12) RecvTxReceipt  <-----------  (11) SendTxReceipt |
//      V                                                      V
//  TxReceiptRecvd                                         TxSecretWait

const (
	txReceiptSigSize = 65
)

func (node *Node) sendTxMsg(
	typ uint16,
	payload io.Reader, size uint64,
	preState, postState nodeState,
	name string,
) error {
	if node.state != preState {
		return invalidStateError(node, preState)
	}

	if err := node.sendMsg(typ, payload, size); err != nil {
		return fmt.Errorf("failed to send tx msg: %v", err)
	}

	node.state = postState

	return nil
}

func (node *Node) recvTxMsg(
	dst io.Writer, typ uint16, preState, postState nodeState, name string,
) (uint64, error) {
	if node.state != preState {
		return 0, invalidStateError(node, preState)
	}

	msgTyp, length, err := node.recvMsgHeader()
	if err != nil {
		return 0, fmt.Errorf(
			"failed to receive %s header: %v", name, err)
	}
	if msgTyp != typ {
		return 0, fmt.Errorf(
			"mismatch msg type, get %d, expect %d",
			msgTyp, typ)
	}

	// TODO: check payload length in order to avoid extremely large message
	_ = length

	if err := node.recvMsgPayload(dst, length); err != nil {
		return 0, fmt.Errorf(
			"failed to receive %s payload: %v", name, err)
	}

	node.state = postState

	return length, nil
}

// SendTxRequest sends the transaction request.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the request is sent out
//  - any timeout occurs
//  - any error occurs
//
// In non-OT mode:
//   PreState : SessionEstablished
//   PostState: TxResponseWait
//
// In OT mode:
//   PreState : Negotiated
//   PostState: TxResponseWait
//
// Parameters:
//  - req: the request returned from the PoD library
//  - size: the number of bytes of the request
//
// Return:
//  If no error occurs, return nil. Otherwise, return a non-nil error.
func (node *Node) SendTxRequest(req io.Reader, size uint64) error {
	preState := nodeState(stateSessionEstablished)
	if node.isOTMode() {
		preState = stateNegotiated
	}

	return node.sendTxMsg(
		msgTxRequest,
		req,
		size,
		preState,
		stateTxResponseWait,
		"Tx request")
}

// RecvTxRequest receives a transaction request.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the request is received
//  - any timeout occurs
//  - any error occurs
//
// In non-OT mode:
//   PreState : SessionEstablished
//   PostState: TxResponseWait
//
// In OT mode:
//   PreState : Negotiated
//   PostState: TxResponseWait
//
// Parameters:
//  - dst: the location where the request is saved
//
// Returns:
//  If no error occurs, return the number of bytes of the request
//  and a nil error. Otherwise, return 0 and the non-nil error.
func (node *Node) RecvTxRequest(dst io.Writer) (uint64, error) {
	preState := nodeState(stateSessionEstablished)
	if node.isOTMode() {
		preState = stateNegotiated
	}

	return node.recvTxMsg(
		dst,
		msgTxRequest,
		preState,
		stateTxRequestRecvd,
		"Tx request")
}

// SendTxResponse sends a transaction response.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the response is sent out
//  - any timeout occurs
//  - any error occurs
//
// PreState : TxRequestRecvd
// PostState: TxReceiptWait
//
// Parameters:
//  - response: the response returned from the PoD library
//  - size: the number of bytes of the response
//
// Returns:
//  If no error occurs, return nil. Otherwise, return a non-nil error.
func (node *Node) SendTxResponse(response io.Reader, size uint64) error {
	return node.sendTxMsg(
		msgTxResponse,
		response,
		size,
		stateTxRequestRecvd,
		stateTxReceiptWait,
		"Tx response")
}

// RecvTxResponse receives the transaction response.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the response is received
//  - any timeout occurs
//  - any error occurs
//
// PreState : TxResponseWait
// PostState: TxResponseRecvd
//
// Parameters:
//  - dst: the location where the response is saved
//
// Return:
//  If no error occurs, return the number of bytes of the response
//  and a nil error. Otherwise, return 0 and a non-nil error.
func (node *Node) RecvTxResponse(dst io.Writer) (uint64, error) {
	return node.recvTxMsg(
		dst,
		msgTxResponse,
		stateTxResponseWait,
		stateTxResponseRecvd,
		"Tx response")
}

// SendTxReceipt sends the transaction receipt and its signature.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the receipt is sent out
//  - any timeout occurs
//  - any error occurs
//
// PreState : TxResponseRecvd
// PostState: TxSecretWait
//
// Parameters:
//  - receipt: the receipt returned from the PoD library
//  - size: the number of bytes of the receipt
//
// Return:
//  If no error occurs, return nil. Otherwise, return a non-nil error.
func (node *Node) SendTxReceipt(receipt io.Reader, size uint64) error {
	receiptBuf := new(bytes.Buffer)
	if _, err := io.CopyN(receiptBuf, receipt, int64(size)); err != nil {
		return fmt.Errorf("failed to extract receipt: %v", err)
	}
	receiptBytes := receiptBuf.Bytes()

	hash := crypto.Keccak256Hash(receiptBytes).Bytes()
	signature, err := crypto.Sign(hash, node.lkey)
	if err != nil {
		return fmt.Errorf("failed to sign receipt: %v", err)
	}

	buf := new(bytes.Buffer)
	buf.Write(signature)
	buf.Write(receiptBytes)

	return node.sendTxMsg(
		msgTxReceipt,
		buf,
		uint64(txReceiptSigSize+len(receiptBytes)),
		stateTxResponseRecvd,
		stateTxSecretWait,
		"Tx receipt")
}

// RecvTxReceipt receives the transaction receipt and it signature.
// This function also validates the signature against the public key
// of the remote node.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the receipt is received
//  - any timeout occurs
//  - any error occurs
//
// PreState : TxReceiptWait
// PostState: TxReceiptRecvd
//
// Returns:
//  If no error occurs, return the receipt, the receipt signature,
//  and a nil error. Otherwise, return a nil receipt, a nil signature,
//  and a non-nil error.
func (node *Node) RecvTxReceipt() ([]byte, []byte, error) {
	buf := new(bytes.Buffer)
	if _, err := node.recvTxMsg(
		buf,
		msgTxReceipt,
		stateTxReceiptWait,
		stateTxReceiptRecvd,
		"Tx receipt",
	); err != nil {
		return nil, nil, err
	}
	payload := buf.Bytes()

	size := len(payload)
	if size <= txReceiptSigSize {
		return nil, nil, fmt.Errorf(
			"signed receipt too short, get %d bytes, expect > %d bytes",
			size, txReceiptSigSize)
	}

	signature := payload[:txReceiptSigSize]
	receipt := payload[txReceiptSigSize:]

	hash := crypto.Keccak256Hash(receipt).Bytes()
	sigPubkey, err := crypto.Ecrecover(hash, signature)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"Ecrecover(%v, %v) failed: %v",
			hex.EncodeToString(hash),
			hex.EncodeToString(signature),
			err)
	}
	pubkeyBytes := crypto.FromECDSAPub(node.rkey)
	if !bytes.Equal(sigPubkey, pubkeyBytes) {
		return nil, nil, fmt.Errorf(
			"mismatch public keys, get %s, expect %s",
			sigPubkey, pubkeyBytes)
	}

	return receipt, signature, nil
}

// SendNegoRequest sends the OT negotiation request. It can be used
// only in OT mode.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the request is sent out
//  - any timeout occurs
//  - any error occurs
//
// PreState : stateSessionEstablished
// PostState: stateNegoAckReqWait
//
// Parameters:
//  - request: the negotiation request returned from PoD library
//  - size: the number of bytes of the request
//
// Return:
//  If no error occurs, return nil. Otherwise, return a non-nil error.
func (node *Node) SendNegoRequest(request io.Reader, size uint64) error {
	if !node.isOTMode() {
		return fmt.Errorf(
			"node %s not in any OT mode, current mode %d",
			node, node.session.mode)
	}

	return node.sendTxMsg(
		msgNegoRequest,
		request,
		size,
		stateSessionEstablished,
		stateNegoAckReqWait,
		"nego request")
}

// RecvNegoRequest receives the OT negotiation request. It can be used
// only in OT mode.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the request is received
//  - any timeout occurs
//  - any error occurs
//
// PreState : stateSessionEstablished
// PostState: stateNegoRequestRecvd
//
// Parameters:
//  - dst: the location where the request is saved
//
// Return:
//  If no error occurs, return the number of bytes of the request and
//  a nil error. Otherwise, return 0 and a non-nil error.
func (node *Node) RecvNegoRequest(dst io.Writer) (uint64, error) {
	if !node.isOTMode() {
		return 0, fmt.Errorf(
			"node %s not in any OT mode, current mode %d",
			node, node.session.mode)
	}

	return node.recvTxMsg(
		dst,
		msgNegoRequest,
		stateSessionEstablished,
		stateNegoRequestRecvd,
		"nego request")
}

// SendNegoAckReq sends the OT negotiation ack+request. It can be used
// only in OT mode.
//
// The ack is the response to the OT negotiation request received from
// the remote PoD node.
//
// The request is the OT negotiation request from the current PoD node.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the ack+request is sent out
//  - any timeout occurs
//  - any error occurs
//
// PreState : stateNegoRequestRecvd
// PostState: stateNegoAckWait
//
// Parameters:
//  - resp: the OT negotiation response returned from PoD library
//  - req: the OT negotiation request returned from Pod library
//  - respSize: the number of bytes of the response
//  - reqSize: the number of bytes of the request
//
// Return:
//  If no error occurs, return nil. Otherwise, return a non-nil error.
func (node *Node) SendNegoAckReq(
	resp, req io.Reader, respSize, reqSize uint64,
) error {
	if !node.isOTMode() {
		return fmt.Errorf(
			"node %s not in any OT mode, current mode %d",
			node, node.session.mode)
	}

	payload := new(bytes.Buffer)
	payload.Write(utils.EncodeUint64(respSize))
	if _, err := io.CopyN(payload, resp, int64(respSize)); err != nil {
		return err
	}
	if _, err := io.CopyN(payload, req, int64(reqSize)); err != nil {
		return err
	}

	return node.sendTxMsg(
		msgNegoAckReq,
		payload,
		8+respSize+reqSize,
		stateNegoRequestRecvd,
		stateNegoAckWait,
		"nego ack+req")
}

// RecvNegoAckReq receives the OT negotiation ack+request. It can be used
// only in OT mode.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the ack+request is received
//  - any timeout occurs
//  - any error occurs
//
// PreState : stateNegoAckReqWait
// PostState: stateNegoAckReqRecvd
//
// Parameters:
//  - resp: the OT negotiation response returned from PoD library
//  - req: the OT negotiation request returned from Pod library
//
// Return:
//  If no error occurs, return the number of bytes of the OT
//  negotiation response, the number of bytes of the OT negotiation
//  request and a nil error. Otherwise, return 0, 0 and the non-nil
//  error.
func (node *Node) RecvNegoAckReq(resp, req io.Writer) (uint64, uint64, error) {
	if !node.isOTMode() {
		return 0, 0, fmt.Errorf(
			"node %s not in any OT mode, current mode %d",
			node, node.session.mode)
	}

	buf := new(bytes.Buffer)
	bufSize, err := node.recvTxMsg(
		buf,
		msgNegoAckReq,
		stateNegoAckReqWait,
		stateNegoAckReqRecvd,
		"nego ack+req",
	)
	if err != nil {
		return 0, 0, err
	}
	if bufSize <= 8 {
		return 0, 0, fmt.Errorf(
			"received nego ack+req too short, get %d bytes, expect > 8 bytes",
			bufSize)
	}
	// TODO: check the upper bound of payload size
	payload := buf.Bytes()
	payloadLength := uint64(len(payload))

	respLength, err := utils.DecodeUint64(payload[:8])
	if err != nil {
		return 0, 0, fmt.Errorf(
			"failed to decode response length: %v", err)
	}
	if respLength == 0 {
		return 0, 0, fmt.Errorf("empty response in nego ack+req")
	}
	// TODO: check the upper bound of respLength

	if payloadLength < respLength+8 {
		return 0, 0, fmt.Errorf(
			"received nego ack+req too short, get %d bytes, expect > %d bytes",
			payloadLength, respLength+8)
	}

	endRespIdx := 8 + respLength
	if _, err := resp.Write(payload[8:endRespIdx]); err != nil {
		return 0, 0, fmt.Errorf(
			"failed to save response in nego ack+req: %v", err)
	}
	if _, err := req.Write(payload[endRespIdx:]); err != nil {
		return 0, 0, fmt.Errorf(
			"failed to save request in nego ack+req: %v", err)
	}

	return respLength, payloadLength - respLength - 8, nil
}

// SendNegoAck sends the OT negotiation ack. It can be used
// only in OT mode.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the ack is sent out
//  - any timeout occurs
//  - any error occurs
//
// PreState : NegoAckReqRecvd
// PostState: Negotiated
//
// Parameters:
//  - resp: the OT negotiation response returned from PoD library
//  - size: the number of bytes of the response
//
// Return:
//  If no error occurs, return nil. Otherwise, return a non-nil error.
func (node *Node) SendNegoAck(resp io.Reader, size uint64) error {
	if !node.isOTMode() {
		return fmt.Errorf(
			"node %s not in any OT mode, current mode %d",
			node, node.session.mode)
	}

	return node.sendTxMsg(
		msgNegoAck,
		resp,
		size,
		stateNegoAckReqRecvd,
		stateNegotiated,
		"nego ack")
}

// RecvNegoAck receives the OT negotiation ack. It can be used
// only in OT mode.
//
// This function is a block operation, which returns when any of
// following events occurs:
//  - the ack+request is received
//  - any timeout occurs
//  - any error occurs
//
// PreState : NegoAckWait
// PostState: Negotiated
//
// Parameters:
//  - dst: the location where the response is saved
//
// Return:
//  If no error occurs, return the number of bytes of the response
//  and a nil error. Otherwise, return 0 and the non-nil error.
func (node *Node) RecvNegoAck(dst io.Writer) (uint64, error) {
	if !node.isOTMode() {
		return 0, fmt.Errorf(
			"node %s not in any OT mode, current mode %d",
			node, node.session.mode)
	}

	return node.recvTxMsg(
		dst,
		msgNegoAck,
		stateNegoAckWait,
		stateNegotiated,
		"nego ack")
}
