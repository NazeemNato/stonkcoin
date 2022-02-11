package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nazeemnato/stonkcoin/block"
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

func (s *Server) Run() {
	http.HandleFunc("/", s.GetChain)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil))
	log.Print("Server started")
}


func main() {
	port := flag.Uint("port", 5000, "TCP port to listen on")
	flag.Parse()
	app := NewServer(uint16(*port))
	app.Run()
}
