package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"MyBinancePro/utils"
	"strconv"
	"time"
)
type OrderBook struct {
	LastUpdateID int `json:"lastUpdateId"`
	Bids         []*Order
	Asks         []*Order
}

type Order struct {
	Price    float64
	Quantity float64
}

type SymbolPrice struct{
	Symbol string
	Price  float64
}


type ProcessedOrder struct {
	Symbol        string
	OrderID       int64
	ClientOrderID string
	TransactTime  time.Time
}

type Account struct {
	MakerCommision  int64
	TakerCommision  int64
	BuyerCommision  int64
	SellerCommision int64
	CanTrade        bool
	CanWithdraw     bool
	CanDeposit      bool
	Balances        []*Balance
}

// Balance groups balance-related information.
type Balance struct {
	Asset  string
	Free   float64
	Locked float64
}

type AddressInfo struct {
	Address string `json:"address"`
	Success bool   `json:"success"`
	AddressTag string `json:"addressTag"`
	Asset    string `json:"asset"`
}


func getDepth(limit string,symbol string) (*OrderBook,error){
	resp, err := http.Get("http://api.binance.com/api/v1/depth?limit="+limit+"&symbol="+symbol)
	
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	
	if err != nil {
		panic(err)
	}
	rawBook := &struct {
		LastUpdateID int             `json:"lastUpdateId"`
		Bids         [][]interface{} `json:"bids"`
		Asks         [][]interface{} `json:"asks"`
	}{}
	if err := json.Unmarshal(body, rawBook); err != nil {
		return nil, errors.New("timeResponse unmarshal failed")
	}
	ob := &OrderBook{
		LastUpdateID: rawBook.LastUpdateID,
	}
	extractOrder := func(rawPrice, rawQuantity interface{}) (*Order, error) {
		price, err := utils.FloatFromString(rawPrice)
		if err != nil {
			return nil, err
		}
		quantity, err := utils.FloatFromString(rawPrice)
		if err != nil {
			return nil, err
		}
		return &Order{
			Price:    price,
			Quantity: quantity,
		}, nil
	}
	for _, bid := range rawBook.Bids {
		order, err := extractOrder(bid[0], bid[1])
		if err != nil {
			return nil, err
		}
		ob.Bids = append(ob.Bids, order)
	}
	for _, ask := range rawBook.Asks {
		order, err := extractOrder(ask[0], ask[1])
		if err != nil {
			return nil, err
		}
		ob.Asks = append(ob.Asks, order)
	}
	
	return ob, nil
}

func getAllPrices() ([]*SymbolPrice,error)  {
	resp, err := http.Get("https://api.binance.com/api/v1/ticker/price")
	
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	if err != nil {
		panic(err)
	}
	rawSymbolPrices := []struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}{}
	if err := json.Unmarshal(body, &rawSymbolPrices); err != nil {
		return nil, errors.New("rawSymbolPrices unmarshal failed")
	}
	
	var sp []*SymbolPrice
	for _, rawSymbolPrice := range rawSymbolPrices {
		p, err := strconv.ParseFloat(rawSymbolPrice.Price, 64)
		if err != nil {
			return nil, errors.New("cannot parse TickerAllPrices.Price")
		}
		sp = append(sp, &SymbolPrice{
			Symbol: rawSymbolPrice.Symbol,
			Price:  p,
		})
	}
	return sp,nil
}

func postNewOrder(APIKey string,SecretKey string,params map[string]string) (*ProcessedOrder, error){
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest("POST", "https://api.binance.com/api/v3/order", nil)
	if err != nil {
		return nil, errors.New("unable to create request")
	}
	q := req.URL.Query()
	for key, val := range params {
		q.Add(key, val)
	}
	
	req.Header.Add("X-MBX-APIKEY",APIKey)
	
	q.Add("signature", (&utils.HmacSigner{}).Sign([]byte(q.Encode())))
	
	req.URL.RawQuery = q.Encode()
	
	fmt.Println("%+v",q.Encode())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	textRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("unable to read response")
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, errors.New("resp statuscode is not 200")
	}
	rawOrder := struct {
		Symbol        string  `json:"symbol"`
		OrderID       int64   `json:"orderId"`
		ClientOrderID string  `json:"clientOrderId"`
		TransactTime  float64 `json:"transactTime"`
	}{}
	if err := json.Unmarshal(textRes, &rawOrder); err != nil {
		return nil, errors.New("rawOrder unmarshal failed")
	}
	
	t, err := utils.TimeFromUnixTimestampFloat(rawOrder.TransactTime)
	if err != nil {
		return nil, err
	}
	
	return &ProcessedOrder{
		Symbol:        rawOrder.Symbol,
		OrderID:       rawOrder.OrderID,
		ClientOrderID: rawOrder.ClientOrderID,
		TransactTime:  t,
	}, nil
	
}

func getAccount(APIKey string,SecretKey string,params map[string]string)(*Account, error){
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/account", nil)
	if err != nil {
		return nil, errors.New("unable to create request")
	}
	q := req.URL.Query()
	for key, val := range params {
		q.Add(key, val)
	}
	
	req.Header.Add("X-MBX-APIKEY",APIKey)
	
	q.Add("signature", (&utils.HmacSigner{}).Sign([]byte(q.Encode())))
	
	req.URL.RawQuery = q.Encode()
	
	fmt.Println("%+v",q.Encode())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	textRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("unable to read response")
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, errors.New("resp statuscode is not 200")
	}
	rawAccount := struct {
		MakerCommision   int64 `json:"makerCommision"`
		TakerCommission  int64 `json:"takerCommission"`
		BuyerCommission  int64 `json:"buyerCommission"`
		SellerCommission int64 `json:"sellerCommission"`
		CanTrade         bool  `json:"canTrade"`
		CanWithdraw      bool  `json:"canWithdraw"`
		CanDeposit       bool  `json:"canDeposit"`
		Balances         []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		}
	}{}
	if err := json.Unmarshal(textRes, &rawAccount); err != nil {
		return nil, errors.New( "rawAccount unmarshal failed")
	}
	
	acc := &Account{
		MakerCommision:  rawAccount.MakerCommision,
		TakerCommision:  rawAccount.TakerCommission,
		BuyerCommision:  rawAccount.BuyerCommission,
		SellerCommision: rawAccount.SellerCommission,
		CanTrade:        rawAccount.CanTrade,
		CanWithdraw:     rawAccount.CanWithdraw,
		CanDeposit:      rawAccount.CanDeposit,
	}
	for _, b := range rawAccount.Balances {
		f, err := utils.FloatFromString(b.Free)
		if err != nil {
			return nil, err
		}
		l, err := utils.FloatFromString(b.Locked)
		if err != nil {
			return nil, err
		}
		acc.Balances = append(acc.Balances, &Balance{
			Asset:  b.Asset,
			Free:   f,
			Locked: l,
		})
	}
	return acc, nil
}

func getDepositAddress(APIKey string,SecretKey string,params map[string]string)([]*AddressInfo , error){
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/account", nil)
	if err != nil {
		return nil, errors.New("unable to create request")
	}
	q := req.URL.Query()
	for key, val := range params {
		q.Add(key, val)
	}
	
	req.Header.Add("X-MBX-APIKEY",APIKey)
	
	q.Add("signature", (&utils.HmacSigner{}).Sign([]byte(q.Encode())))
	
	req.URL.RawQuery = q.Encode()
	
	fmt.Println("%+v",q.Encode())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	textRes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("unable to read response")
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, errors.New("resp statuscode is not 200")
	}
	var addressInfos []*AddressInfo
	if err := json.Unmarshal(textRes, &addressInfos); err != nil {
		return nil, errors.New("rawOrder unmarshal failed")
	}
	
	
	return addressInfos, nil
}


func main() {
	//1.获得深度
	//getDepth("100","BTCUSDT")
	// 2.获得价格
	//getAllPrices()
	
	// 3下单
	APIKey := "6XNN4nTBGpMFGVoeg8NNkHG8gBmnHh7A0kzxK2XwJQQlgmT6WBtCuVtmAGjYBZiy"
	SecretKey := "pBO57Cnuqp2vLOQ4q3pyY8cm174bmhsG0o0SUQfDrMKqE3uuw5G8YfxtS57GTCUN"
	params := make(map[string]string)
	params["symbol"] = "BNBETH"
	params["side"] = "BUY"
	params["type"] = "LIMIT"
	params["timeInForce"] = "IOC"
	params["quantity"] = strconv.FormatFloat(1,'f', 10, 64)
	params["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	//processedOrder , err := postNewOrder(APIKey,SecretKey,params)
	//if err != nil{
	//
	//}
	//fmt.Println("%+v",processedOrder)
	//
	// 4查询当前账号余额
	account,err := getAccount(APIKey,SecretKey,params)
	if err!=nil {
		fmt.Println("get account error")
	}
	fmt.Println("%+v",account)
	
	// 5查询充值地址
	params =make(map[string]string)
	params["asset"] = "BNB"
	params["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	depositAddressArr ,err:= getDepositAddress(APIKey,SecretKey,params)
	if err != nil{
		fmt.Println("get deposite address err")
	}
	depositAddressArrStr ,_ := json.Marshal(depositAddressArr)
	fmt.Println(depositAddressArrStr)
	
}
