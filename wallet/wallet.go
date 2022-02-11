package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/nazeemnato/stonkcoin/utils"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    string
}

type Transaction struct {
	privateKey       *ecdsa.PrivateKey
	publicKey        *ecdsa.PublicKey
	senderAddress    string
	recipientAddress string
	amount           float32
}
type TransactionRequest struct {
	SenderPrivateKey *string `json:"sender_private_key"`
	SenderPublicKey  *string `json:"sender_public_key"`
	SenderAddress    *string `json:"sender_address"`
	ReceiverAddress  *string `json:"receiver_address"`
	Amount           *string `json:"amount"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderPrivateKey == nil || tr.SenderPublicKey == nil || tr.SenderAddress == nil || tr.ReceiverAddress == nil || tr.Amount == nil {
		return false
	}
	return true
}

// create new transaction
func NewTransaction(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, senderAddr string, recipientAddr string, amount float32) *Transaction {
	return &Transaction{privateKey, publicKey, senderAddr, recipientAddr, amount}
}

// create marshal json
func (t *Transaction) MarshalJSON() ([]byte, error) {
	// this json marshal must be in the same order as block/transaction.go MarshalJSON()
	// otherwise, the signature will be invalid
	return json.Marshal(struct {
		SenderAddress    string  `json:"senderAddress"`
		RecipientAddress string  `json:"recipientAddress"`
		Amount           float32 `json:"amount"`
	}{
		t.senderAddress,
		t.recipientAddress,
		t.amount,
	})
}

// generate signature
func (t *Transaction) GenerateSignature() *utils.Signature {
	m, _ := t.MarshalJSON()
	h := sha256.Sum256(m)
	r, s, _ := ecdsa.Sign(rand.Reader, t.privateKey, h[:])
	return &utils.Signature{R: r, S: s}
}

// create new wallet
func NewWallet() *Wallet {
	w := new(Wallet)
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := &privateKey.PublicKey
	w.privateKey = privateKey
	w.publicKey = publicKey
	// perform sha256 on public key
	h2 := sha256.New()
	h2.Write(w.publicKey.X.Bytes())
	h2.Write(w.publicKey.Y.Bytes())
	digest2 := h2.Sum(nil)
	// perform ripemd160 on sha256
	h3 := ripemd160.New()
	h3.Write(digest2)
	digest3 := h3.Sum(nil)
	// add version byte (0x00)
	vd4 := make([]byte, 21)
	vd4[0] = 0x00
	copy(vd4[1:], digest3)
	// perform sha256 on extended version byte
	h5 := sha256.New()
	h5.Write(vd4)
	digest5 := h5.Sum(nil)
	// perform sha256 on sha256
	h6 := sha256.New()
	h6.Write(digest5)
	digest6 := h6.Sum(nil)
	// take first 4 bytes of sha256 for checksum
	checksum := digest6[:4]
	// add checksum to extended version byte from above
	dc8 := make([]byte, 25)
	copy(dc8[:21], vd4[:])
	copy(dc8[21:], checksum[:])
	// base58 encode
	w.address = base58.Encode(dc8)
	return w
}

// get private key
func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

// get private key in hex
func (w *Wallet) PrivateKeyStr() string {
	// return w.privateKey.D.Text(16)
	return fmt.Sprintf("%x", w.privateKey.D.Bytes())
}

// get public key
func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return w.publicKey
}

// get private key in hex
func (w *Wallet) PublicKeyStr() string {
	// return w.privateKey.D.Text(16)
	return fmt.Sprintf("%x%x", w.publicKey.X.Bytes(), w.publicKey.Y.Bytes())
}

func (w *Wallet) Address() string {
	return w.address
}

// marshal json
func (w *Wallet) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PrivateKey string `json:"privateKey"`
		PublicKey  string `json:"publicKey"`
		Address    string `json:"address"`
	}{
		w.PrivateKeyStr(),
		w.PublicKeyStr(),
		w.Address(),
	})
}
