package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/nazeemnato/stonkcoin/block"
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
		value, err := strconv.ParseFloat(*t.Amount, 32)
		if err != nil {
			log.Println("Error")
			io.WriteString(w, string(utils.Json("Error")))
			return
		}
		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		privateKey := utils.PrivateKeyFromString(*t.SenderPrivateKey, publicKey)
		amount := float32(value)

		transaction := wallet.NewTransaction(privateKey, publicKey, *t.SenderAddress, *t.ReceiverAddress, amount)
		signature := transaction.GenerateSignature()
		signatureStr := signature.String()

		bt := block.TransactionRequest{
			ReceiverAddress: t.ReceiverAddress,
			SenderAddress:   t.SenderAddress,
			SenderPublicKey: t.SenderPublicKey,
			Amount:          &amount,
			Signature:       &signatureStr,
		}

		m, _ := json.Marshal(bt)
		buf := bytes.NewBuffer(m)
		req, _ := http.Post(ws.Gateway()+"/transaction", "application/json", buf)
		if req.StatusCode == http.StatusCreated {
			io.WriteString(w, string(utils.Json("Success")))
			return
		}
		io.WriteString(w, string(utils.Json("Error")))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (ws *WalletServer) WalletBalance(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		address := r.URL.Query().Get("address")
		endpoint := fmt.Sprintf("%s/account/balance", ws.Gateway())
		client := http.Client{}
		bcReq, _ := http.NewRequest("GET", endpoint, nil)
		q := bcReq.URL.Query()
		q.Add("address", address)
		bcReq.URL.RawQuery = q.Encode()
		res, err := client.Do(bcReq)
		if err != nil {
			log.Println("Error")
			io.WriteString(w, string(utils.Json("Error")))
			return
		} else {
			decoder := json.NewDecoder(res.Body)
			var bar block.AmountRespone
			err := decoder.Decode(&bar)
			if err != nil {
				log.Println("Error")
				io.WriteString(w, string(utils.Json("Error")))
				return
			} else {
				m, _ := json.Marshal(struct {
					Balance float32 `json:"balance"`
				}{
					Balance: bar.Amount,
				})
				io.WriteString(w, string(m[:]))
				return
			}
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)

	}
}

func (ws *WalletServer) Start() {
	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/create", ws.CreateWallet)
	http.HandleFunc("/transaction", ws.CreateTransaction)
	http.HandleFunc("/balance", ws.WalletBalance)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", ws.port), nil))
}

func main() {
	port := flag.Uint("port", 8000, "TCP port to listen on")
	gateway := flag.String("gateway", "http://localhost:5000", "Gateway URL")
	flag.Parse()
	s := NewWalletServer(uint16(*port), *gateway)
	s.Start()
}
