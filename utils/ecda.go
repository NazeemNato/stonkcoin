package utils

import (
	"fmt"
	"math/big"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

// string signature
func (s *Signature) String() string {
	return fmt.Sprintf("%x%x", s.R.Bytes(), s.S.Bytes())
}