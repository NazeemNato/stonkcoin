package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"text/template"

	"github.com/nazeemnato/stonkcoin/utils"
	"github.com/nazeemnato/stonkcoin/wallet"
)

type WalletServer struct {
	port    uint16
	gateway string
}

func NewWalletServer(port uint16, gateway string) *WalletServer {
	return &WalletServer{port, gateway}
}

func (ws *WalletServer) Gateway() string {
	return ws.gateway
}

func (ws *WalletServer) Port() uint16 {
	return ws.port
}

func (ws *WalletServer) Index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		t, _ := template.ParseFiles("templates/index.html")
		t.Execute(w, "")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (ws *WalletServer) CreateWallet(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		w.Header().Set("Content-Type", "application/json")
		wlt := wallet.NewWallet()
		m, _ := wlt.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (ws *WalletServer) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var t wallet.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Println("Error")
			io.WriteString(w, string(utils.Json("Error")))
			return
		}
		if !t.Validate() {
			log.Println("Missing fields")
			io.WriteString(w, string(utils.Json("Missing fields")))
			return
		}

		fmt.Printf("%+v\n", t)
		io.WriteString(w, string(utils.Json("Success")))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (ws *WalletServer) Start() {
	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/create", ws.CreateWallet)
	http.HandleFunc("/transaction", ws.CreateTransaction)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", ws.port), nil))
}

func main() {
	port := flag.Uint("port", 8000, "TCP port to listen on")
	gateway := flag.String("gateway", "http://120.0.0.1:5000", "Gateway URL")
	flag.Parse()
	s := NewWalletServer(uint16(*port), *gateway)
	s.Start()
}
