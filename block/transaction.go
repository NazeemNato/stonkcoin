package block

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Transaction struct {
	senderAddress    string
	recipientAddress string
	amount           float32
}

func NewTransaction(senderAddress string, recipientAddress string, amount float32) *Transaction {
	transactions := new(Transaction)
	transactions.senderAddress = senderAddress
	transactions.recipientAddress = recipientAddress
	transactions.amount = amount
	return transactions
}

func (t *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 50))
	fmt.Printf("senderAddress: %s\n", t.senderAddress)
	fmt.Printf("recipientAddress: %s\n", t.recipientAddress)
	fmt.Printf("amount: %.1f\n", t.amount)
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SenderAddress    string  `json:"senderAddress"`
		RecipientAddress string  `json:"recipientAddress"`
		Amount           float32 `json:"amount"`
	}{
		SenderAddress:    t.senderAddress,
		RecipientAddress: t.recipientAddress,
		Amount:           t.amount,
	})
}