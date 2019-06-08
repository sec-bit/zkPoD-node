// A wrapper around rlpx.go.

package p2p

import (
	"crypto/ecdsa"
	"net"
)

type RLPx struct {
	raw *rlpx
}

func NewRLPx(fd net.Conn) *RLPx {
	return &RLPx{
		raw: newRLPX(fd).(*rlpx),
	}
}

func (t *RLPx) ReadMsg() (Msg, error) {
	return t.raw.ReadMsg()
}

func (t *RLPx) WriteMsg(msg Msg) error {
	return t.raw.WriteMsg(msg)
}

func (t *RLPx) Close(err error) {
	t.raw.close(err)
}

func (t *RLPx) DoEncHandshake(prv *ecdsa.PrivateKey, dial *ecdsa.PublicKey) (*ecdsa.PublicKey, error) {
	return t.raw.doEncHandshake(prv, dial)
}
