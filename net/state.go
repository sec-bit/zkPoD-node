package net

import (
	"fmt"
)

const (
	stateConnected = iota
	stateClosed

	stateSessionAckWait
	stateSessionReqRecvd
	stateSessionAckRecvd
	stateSessionEstablished

	stateNegoAckReqWait
	stateNegoAckReqRecvd
	stateNegoRequestRecvd
	stateNegoAckWait
	stateNegotiated

	stateTxResponseWait
	stateTxRequestRecvd
	stateTxReceiptWait
	stateTxResponseRecvd
	stateTxSecretWait
	stateTxReceiptRecvd

	stateMax // Leave at the end intentionally
)

var (
	stateNames = [stateMax]string{
		stateConnected: "Connected",
		stateClosed:    "Closed",

		stateSessionAckWait:     "SessionAckWait",
		stateSessionReqRecvd:    "SessionReqRecvd",
		stateSessionAckRecvd:    "SessionAckRecvd",
		stateSessionEstablished: "SessionEstablished",

		stateNegoAckReqWait:   "NegoAckReqWait",
		stateNegoAckReqRecvd:  "NegoAckReqRecvd",
		stateNegoRequestRecvd: "NegoRequestRecvd",
		stateNegoAckWait:      "NegoAckWait",
		stateNegotiated:       "Negotiated",

		stateTxResponseWait:  "TxResponseWait",
		stateTxRequestRecvd:  "TxRequestRecvd",
		stateTxReceiptWait:   "TxReceiptWait",
		stateTxResponseRecvd: "TxResponseRecvd",
		stateTxSecretWait:    "TxSecretWait",
		stateTxReceiptRecvd:  "TxReceiptRecvd",
	}
)

type (
	nodeState int

	errInvalidState struct {
		nodeName      string
		currentState  nodeState
		expectedState nodeState
	}
)

func (state nodeState) String() string {
	if state >= stateMax {
		return "invalid state"
	}
	return stateNames[state]
}

func invalidStateError(node *Node, expected nodeState) *errInvalidState {
	return &errInvalidState{
		nodeName:      node.String(),
		currentState:  node.state,
		expectedState: expected,
	}
}

func (e *errInvalidState) Error() string {
	return fmt.Sprintf(
		"node %s in invalid state %s, expecting state %s",
		e.nodeName,
		e.currentState.String(),
		e.expectedState.String())
}
