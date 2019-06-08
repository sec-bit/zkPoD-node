package utils

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestMakeRandomMsg(t *testing.T) {
	msg := MakeRandomMsg(32)
	if len(msg) != 32 {
		t.Fatalf("invalid message length, expect 32 bytes, but got %d bytes",
			len(msg))
	}
}

func TestEncodeUint8(t *testing.T) {
	v := uint8(rand.Intn(255))
	b := EncodeUint8(v)
	e := []byte{v}

	if !bytes.Equal(b, e) {
		t.Fatalf("EncodeUint8(%d) returns %v, should be %v",
			v, e, b)
	}
}

func TestDecodeUint8(t *testing.T) {
	e := uint8(rand.Intn(255))
	b := []byte{e}
	v, err := DecodeUint8(b)

	if err != nil {
		t.Fatalf("DecodeUint8(%v) failed: %v", b, err)
	}

	if v != e {
		t.Fatalf("DecodeUint8(%v) returns %d, should be %d",
			b, v, e)
	}
}

func TestEncodeUint64(t *testing.T) {
	v := uint64(0x12345678deadbeaf)
	e := []byte{0xaf, 0xbe, 0xad, 0xde, 0x78, 0x56, 0x34, 0x12}
	bs := EncodeUint64(v)

	if !bytes.Equal(bs, e) {
		t.Fatalf("EncodeUint64(%d) returns %v, should be %v",
			v, bs, e)
	}
}

func TestDecodeUint64(t *testing.T) {
	bs := []byte{0xaf, 0xbe, 0xad, 0xde, 0x78, 0x56, 0x34, 0x12}
	e := uint64(0x12345678deadbeaf)
	v, err := DecodeUint64(bs)

	if err != nil {
		t.Fatalf("DecodeUint64(%d) failed: %v",
			bs, err)
	}

	if v != e {
		t.Fatalf("DecodeUint64(%v) returns %d, should be %d",
			bs, v, e)
	}
}
