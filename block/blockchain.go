package block

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/nazeemnato/stonkcoin/utils"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	MINING_DIFFICULTY = 4
	MINING_SENDER     = "0x0"
	MINING_REWARD     = 10
	MINING_EVERY_SEC  = 30
)

type Blockchain struct {
	transactionsPool []*Transaction
	chain            []*Block
	port             uint16
	address          string
	mux              sync.Mutex
}

type AmountRespone struct {
	Amount float32 `json:"amount"`
}

func (ar *AmountRespone) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount float32 `json:"amount"`
	}{
		ar.Amount,
	})
}

func (bc *Blockchain) CreateTransaction(sender string, recipient string, amount float32, senderPublicKey *ecdsa.PublicKey, signature *utils.Signature) bool {
	isOk := bc.AddTransaction(sender, recipient, amount, senderPublicKey, signature)
	return isOk
}

func (bc *Blockchain) TransactionPool() []*Transaction {
	return bc.transactionsPool
}

func (bc *Blockchain) AddTransaction(sender string, recipient string, amount float32, senderPublicKey *ecdsa.PublicKey, signature *utils.Signature) bool {
	t := NewTransaction(sender, recipient, amount)
	// check sender is mining address
	if sender == MINING_SENDER {
		log.Print("Transaction from the mining reward")
		bc.transactionsPool = append(bc.transactionsPool, t)
		return true
	}
	// calculate sender balance
	// for testing purpose only
	// if bc.CalculateTransaction(sender) < amount {
	// 	log.Println("Error: Not enough balance")
	// 	return false
	// }
	// verify transaction signature
	if bc.VerifyTransactionSignature(senderPublicKey, signature, t) {
		bc.transactionsPool = append(bc.transactionsPool, t)
		return true
	} else {
		log.Println("Invalid transaction signature")
		return false
	}

}

func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactionsPool {
		transactions = append(transactions, NewTransaction(t.senderAddress, t.recipientAddress, t.amount))
	}
	return transactions
}

func (bc *Blockchain) ValidProof(nonce int, prevHash [32]byte, transactions []*Transaction, difficulty int) bool {
	zeros := strings.Repeat("0", difficulty)
	guessBlock := Block{
		0,
		nonce,
		prevHash,
		transactions,
	}
	guessHash := fmt.Sprintf("%x", guessBlock.Hash())
	return guessHash[:difficulty] == zeros
}

func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactionPool()
	prevHash := bc.LastBlock().Hash()
	nonce := 0
	for !bc.ValidProof(nonce, prevHash, transactions, MINING_DIFFICULTY) {
		nonce += 1
	}
	return nonce
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		fmt.Printf("%s Chain %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
	}
	fmt.Printf("%s\n", strings.Repeat("*", 50))
}

func NewBlockchain(address string, port uint16) *Blockchain {
	block := &Block{}
	bc := new(Blockchain)
	bc.address = address
	bc.port = port
	bc.CreateBlock(0, block.Hash())
	return bc
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(bc.chain)
}

func (bc *Blockchain) CreateBlock(nonce int, prevHash [32]byte) *Block {
	block := NewBlock(nonce, prevHash, bc.transactionsPool)
	bc.chain = append(bc.chain, block)
	bc.transactionsPool = []*Transaction{}
	return block
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) Mining() bool {
	bc.mux.Lock()
	defer bc.mux.Unlock()
	if len(bc.transactionsPool) == 0 {
		return false
	}
	bc.AddTransaction(MINING_SENDER, bc.address, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	prevHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, prevHash)
	fmt.Println("Success!")
	return true
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	_ = time.AfterFunc(time.Second*MINING_EVERY_SEC, bc.StartMining)
}

func (bc *Blockchain) CalculateTransaction(address string) float32 {
	var amount float32 = 0
	for _, c := range bc.chain {
		for _, t := range c.transactions {
			if address == t.senderAddress {
				amount -= t.amount
			}

			if address == t.recipientAddress {
				amount += t.amount
			}
		}
	}
	return amount
}

func (bc *Blockchain) VerifyTransactionSignature(senderPublicKey *ecdsa.PublicKey, signature *utils.Signature, transaction *Transaction) bool {
	m, _ := json.Marshal(transaction)
	h := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublicKey, h[:], signature.R, signature.S)
}
