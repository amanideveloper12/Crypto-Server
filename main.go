// The server can be run by using go run main.go. The API endpoints are hosted on port 8080

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"context"

	"github.com/gin-gonic/gin"    //Create HTTP server
	"github.com/gobwas/ws"        //Connect to websocket
	"github.com/gobwas/ws/wsutil" //Connect to websocket
)

type crypto struct {
	Id              string `json:"id"`
	Crypto_FullName string `json:"full_name"`
	Ask             string `json:"ask"`
	Bid             string `json:"bid"`
	Last            string `json:"last"`
	Open            string `json:"open"`
	Low             string `json:"low"`
	High            string `json:"high"`
	FeeCurrency     string `json:"currency"`
}

type ticker struct {
	Ask          string `json:"ask"`
	Bid          string `json:"bid"`
	Last         string `json:"last"`
	Open         string `json:"open"`
	Low          string `json:"low"`
	High         string `json:"high"`
	FeeCurrency  string `json:"currency"`
	Volume       string `json:"volume"`
	Volume_quote string `json:"volume_quote"`
	Timestamp    string `json:"timestamp"`
}

type currency struct {
	Crypto_FullName string `json:"full_name"`
}

var reqBody = map[string]interface{}{
	"method": "subscribe",
	"ch":     "orderbook/top/1000ms",
	"params": map[string]interface{}{
		"symbols": []string{"ETHBTC", "BTCUSDT"},
	},
	"id": 123,
}

func getCurrencies(context *gin.Context) {
	socketListener()
	cryptos, err := getAllCurrencies()
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "HTBC server cannot be connected"})
	}
	context.IndentedJSON(http.StatusOK, cryptos)
}

func getCurrency(context *gin.Context) {
	socketListener()
	symbol := context.Param("symbol")
	currency, err := getCurrencyBySymbol(symbol)
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Can't find crypto"})
	}
	context.IndentedJSON(http.StatusOK, currency)
}

func getCurrencyBySymbol(symbol string) (*crypto, error) {
	url := "https://api.hitbtc.com/api/3/public/ticker/" + symbol
	response, err := http.Get(url)
	if err != nil {
		log.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var tickerval ticker
	json.Unmarshal(responseData, &tickerval)

	url = "https://api.hitbtc.com/api/3/public/currency/" + symbol[:3]
	response, err = http.Get(url)
	if err != nil {
		log.Print(err.Error())
		os.Exit(1)
	}

	responseData, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var currencyval currency
	json.Unmarshal(responseData, &currencyval)
	var cryptosval crypto
	if symbol == "BTCUSDT" {
		cryptosval = crypto{Id: symbol[:3], Crypto_FullName: currencyval.Crypto_FullName, Ask: tickerval.Ask, Bid: tickerval.Bid, Last: tickerval.Last, Open: tickerval.Open, Low: tickerval.Low, High: tickerval.High, FeeCurrency: symbol[3:6]}
	} else {
		cryptosval = crypto{Id: symbol[:3], Crypto_FullName: currencyval.Crypto_FullName, Ask: tickerval.Ask, Bid: tickerval.Bid, Last: tickerval.Last, Open: tickerval.Open, Low: tickerval.Low, High: tickerval.High, FeeCurrency: symbol[len(symbol)-3:]}
	}
	if (symbol != "ETHBTC") && (symbol != "BTCUSDT") {
		log.Println("symbol not valid ", symbol)
		return nil, errors.New("Can't find crypto")
	}
	return &cryptosval, nil
}

func getAllCurrencies() ([]*crypto, error) {
	BTCUSD, err := getCurrencyBySymbol("BTCUSDT")
	if err != nil {
		return nil, err
	}
	ETHBTC, err := getCurrencyBySymbol("ETHBTC")
	if err != nil {
		return nil, err
	}
	var cryptos []*crypto
	cryptos = append(cryptos, BTCUSD)
	cryptos = append(cryptos, ETHBTC)
	return cryptos, nil
}

func socketListener() {
	log.Println("Start client web socket")
	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), "wss://api.hitbtc.com/api/3/ws/public")
	if err != nil {
		log.Println("Not connected: " + err.Error())
		return
	}
	defer conn.Close()
	log.Println("Web socket server connected")
	if bodyBytes, err := json.Marshal(reqBody); err != nil {
		log.Println("Marshal failed", err)
		return
	} else {
		err = wsutil.WriteClientMessage(conn, ws.OpText, bodyBytes)
		if err != nil {
			log.Println("Write failed", err)
			return
		}
	}

	msg, opCode, err := wsutil.ReadServerData(conn)
	if err != nil {
		log.Println("Data not received: " + err.Error())
		return
	}
	log.Println("Receive Server op code: ", opCode)
	log.Println("Receive Server message: ", string(msg))
	log.Println("Server Disconnected")
}

func main() {
	router := gin.Default()
	log.Println("Server running ")
	router.GET("/currency/all", getCurrencies)
	router.GET("/currency/:symbol", getCurrency)
	router.Run("localhost:8080")
}
