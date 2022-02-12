package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nazeemnato/stonkcoin/block"
	"github.com/nazeemnato/stonkcoin/utils"
	"github.com/nazeemnato/stonkcoin/wallet"
)

var cache map[string]*block.Blockchain = make(map[string]*block.Blockchain)

type Server struct {
	port uint16
}

func NewServer(port uint16) *Server {
	return &Server{port}
}

func (s *Server) Port() uint16 {
	return s.port
}

func (s *Server) GetBlockchain() *block.Blockchain {
	bc, ok := cache["blockchain"]
	if !ok {
		wallet := wallet.NewWallet()
		bc = block.NewBlockchain(wallet.Address(), s.port)
		cache["blockchain"] = bc

		log.Printf("Created new blockchain with address: %s\n", wallet.Address())
		log.Printf("Public key: %s\n", wallet.PublicKeyStr())
		log.Printf("Private key: %s\n", wallet.PrivateKeyStr())
	}
	return bc
}

func (s *Server) GetChain(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		bc := s.GetBlockchain()
		m, _ := bc.MarshalJSON()
		io.WriteString(w, string(m))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("Method not allowed: %s", req.Method)
	}
}

func (s *Server) Transaction(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		bc := s.GetBlockchain()
		transaction := bc.TransactionPool()
		m, _ := json.Marshal(struct {
			Transactions []*block.Transaction `json:"transactions"`
			Length       int                  `json:"length"`
		}{
			transaction,
			len(transaction),
		})
		io.WriteString(w, string(m))
	case http.MethodPost:
		decode := json.NewDecoder(req.Body)
		var t block.TransactionRequest
		err := decode.Decode(&t)
		if err != nil {
			log.Printf("Error decoding transaction: %s\n", err)
			io.WriteString(w, string(utils.Json("Error decoding transaction")))
			return
		}

		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		signature := utils.SignatureFromString(*t.Signature)

		bc := s.GetBlockchain()
		isCreated := bc.CreateTransaction(*t.SenderAddress, *t.ReceiverAddress, *t.Amount, publicKey, signature)

		w.Header().Set("Content-Type", "application/json")
		var m []byte
		if !isCreated {
			w.WriteHeader(http.StatusBadRequest)
			m = utils.Json("Transaction not created")
		} else {
			w.WriteHeader(http.StatusCreated)
			m = utils.Json("Transaction created")
		}
		io.WriteString(w, string(m))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("Method not allowed: %s", req.Method)
	}
}

func (s *Server) Mine(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		bc := s.GetBlockchain()
		isMined := bc.Mining()
		var m []byte
		if !isMined {
			w.WriteHeader(http.StatusBadRequest)
			m = utils.Json("Block not mined")
		} else {
			m = utils.Json("Block mined")
		}
		io.WriteString(w, string(m))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("Method not allowed: %s", req.Method)
	}
}

func (s *Server) StartMining(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		bc := s.GetBlockchain()
		bc.StartMining()
		m := utils.Json("Mining started")
		io.WriteString(w, string(m))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("Method not allowed: %s", req.Method)
	}
}

func (s *Server) AccountBalance(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		address := req.URL.Query().Get("address")
		amount := s.GetBlockchain().CalculateTransaction(address)
		res := &block.AmountRespone{Amount: amount}
		m, _ := res.MarshalJSON()
		io.WriteString(w, string(m))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("Method not allowed: %s", req.Method)
	}
}

func (s *Server) Run() {
	http.HandleFunc("/", s.GetChain)
	http.HandleFunc("/transaction", s.Transaction)
	http.HandleFunc("/mine", s.Mine)
	http.HandleFunc("/mine/start", s.StartMining)
	http.HandleFunc("/account/balance", s.AccountBalance)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil))
	log.Print("Server started")
}

func main() {
	port := flag.Uint("port", 5000, "TCP port to listen on")
	flag.Parse()
	app := NewServer(uint16(*port))
	app.Run()
}
