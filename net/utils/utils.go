package utils

import (
	"fmt"
	"math/rand"
)

func MakeRandomMsg(n uint) []byte {
	msg := make([]byte, n)
	for i := uint(0); i < n; i++ {
		msg[i] = byte(rand.Intn(255))
	}
	return msg
}

func EncodeUint8(v uint8) []byte {
	return []byte{v}
}

func DecodeUint8(bs []byte) (uint8, error) {
	if bs == nil || len(bs) != 1 {
		return 0, fmt.Errorf("cannot convert %v to uint8", bs)
	}
	return bs[0], nil
}

func EncodeUint16(v uint16) []byte {
	bs := make([]byte, 2)

	for i := 0; i < 2; i++ {
		bs[i] = uint8(v)
		v >>= 8
	}

	return bs
}

func DecodeUint16(bs []byte) (uint16, error) {
	if bs == nil || len(bs) != 2 {
		return 0, fmt.Errorf("cannot convert %v to uint16", bs)
	}

	v := uint16(0)
	for i := 1; i >= 0; i-- {
		v <<= 8
		v += uint16(bs[i])
	}

	return v, nil
}

func EncodeUint64(v uint64) []byte {
	bs := make([]byte, 8)

	for i := 0; i < 8; i++ {
		bs[i] = uint8(v)
		v >>= 8
	}

	return bs
}

func DecodeUint64(bs []byte) (uint64, error) {
	if bs == nil || len(bs) != 8 {
		return 0, fmt.Errorf("cannot convert %v to uint64", bs)
	}

	v := uint64(0)
	for i := 7; i >= 0; i-- {
		v <<= 8
		v += uint64(bs[i])
	}

	return v, nil
}
