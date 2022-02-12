package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

// string signature
func (s *Signature) String() string {
	return fmt.Sprintf("%064x%064x", s.R.Bytes(), s.S.Bytes())
}

// convert string to bytes tuple
func String2BytesTuple(s string) (big.Int, big.Int) {
	// 128 bytes so x will be 64 bytes
	bx, _ := hex.DecodeString(s[:64])
	by, _ := hex.DecodeString(s[64:])

	var bix big.Int
	var biy big.Int
	_, _ = bix.SetBytes(bx), biy.SetBytes(by)
	return bix, biy
}

func PublicKeyFromString(s string) *ecdsa.PublicKey {
	bx, by := String2BytesTuple(s)
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     &bx,
		Y:     &by,
	}
}

func PrivateKeyFromString(s string, publicKey *ecdsa.PublicKey) *ecdsa.PrivateKey {
	b, _ := hex.DecodeString(s[:])
	var bi big.Int
	_ = bi.SetBytes(b)
	return &ecdsa.PrivateKey{
		PublicKey: *publicKey,
		D:         &bi,
	}
}

func SignatureFromString(s string) *Signature {
	x, y := String2BytesTuple(s)
	return &Signature{
		R: &x,
		S: &y,
	}
}
