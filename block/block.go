package block

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type Block struct {
	timestamp    int64
	nonce        int
	prevHash     [32]byte
	transactions []*Transaction
}

// create hash of block
func (b *Block) Hash() [32]byte {
	m, _ := b.MarshalJSON()
	return sha256.Sum256(m)
}

// print function for block
func (b *Block) Print() {
	fmt.Printf("nonce: %d\n", b.nonce)
	fmt.Printf("prevHash: %x\n", b.prevHash)
	fmt.Printf("timestamp: %d\n", b.timestamp)
	for _, t := range b.transactions {
		t.Print()
	}
}

// create new block
func NewBlock(nonce int, prevHash [32]byte, transactions []*Transaction) *Block {
	block := new(Block)
	block.timestamp = time.Now().UnixNano()
	block.nonce = nonce
	block.prevHash = prevHash
	block.transactions = transactions
	return block
}

// marshal block

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Nonce       int            `json:"nonce"`
		PrevHash    string         `json:"prevHash"`
		Timestamp   int64          `json:"timestamp"`
		Transaction []*Transaction `json:"transactions"`
	}{
		Nonce:       b.nonce,
		PrevHash:    fmt.Sprintf("%x", b.prevHash),
		Timestamp:   b.timestamp,
		Transaction: b.transactions,
	})
}
