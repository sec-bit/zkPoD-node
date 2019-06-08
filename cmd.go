package main

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
)

//BuyerPurchaseData purchases data.
func BuyerPurchaseData(requestData RequestData, Log ILogger) {

	data, err := json.Marshal(requestData)
	if err != nil {
		Log.Warnf("failed to marshal request data. err=%v", err)
		return
	}
	Log.Debugf("data sigma merkle root=%v, seller ip=%v", requestData.MerkleRoot, requestData.SellerIP)

	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/b/purchase"
	body := make(url.Values)
	body.Add("request_data", string(data))

	Log.Debugf("start send request for purchasing data. url=%v, request data=%v", urlStr, string(data))
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("purchasing data finish.")
	Log.Debugf("%s", string(responseBody))
}

//BuyerDepositETH deposit ETH to a seller in contract.
func BuyerDepositETH(value string, sellerAddress string, Log ILogger) {
	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/b/deposit"
	body := make(url.Values)
	body.Add("value", value)
	body.Add("address", sellerAddress)
	Log.Debugf("buyer deposit to seller. value=%v, seller address=%v", value, sellerAddress)

	Log.Debugf("start send request to deposit ETH. url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request to deposit ETH finish.")
	Log.Debugf("%s", string(responseBody))
}

//BuyerUnDepositETH undeposits ETH from a seller in contract.
func BuyerUnDepositETH(sellerAddress string, Log ILogger) {
	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/b/undeposit"
	body := make(url.Values)
	body.Add("address", sellerAddress)
	Log.Debugf("buyer undeposit from seller. seller address=%v", sellerAddress)

	Log.Debugf("start send request to undeposit ETH. url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request to undeposit ETH finish.")
	Log.Debugf("%s", string(responseBody))
}

//BuyerWithdrawETH withdraw ETH from a seller in contract.
func BuyerWithdrawETH(sellerAddress string, Log ILogger) {
	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/b/withdraw"
	body := make(url.Values)
	body.Add("address", sellerAddress)
	Log.Debugf("buyer withdraws for contract. seller address=%v", sellerAddress)

	Log.Debugf("start send request to withdraw for contract. url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request to withdraw for contract finish.")
	Log.Debugf("%s", string(responseBody))
}

//SellerInitDataNode initialize data for publishing.
func SellerInitDataNode(filepath string, Log ILogger) {

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		Log.Warnf("read config file error! filepath=%s, err=%v", filepath, err)
		panic(err.Error())
	}
	Log.Debugf("data=%v", string(data))

	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/s/publish/init"
	body := make(url.Values)
	body.Add("request_data", string(data))

	Log.Debugf("start send request for publishing data...url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request for publishing data finish.")
	Log.Debugf("%s", string(responseBody))
}

// SellerPublishData sends a request to server
// to initializes data and publishes data to contract by seller.
func SellerPublishData(merkleRoot string, value string, Log ILogger) {

	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/s/publish"
	body := make(url.Values)
	body.Add("merkleRoot", merkleRoot)
	body.Add("value", value)

	Log.Debugf("start send request for publishing data...url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request for publishing data finish.")
	Log.Debugf("%s", string(responseBody))
}

//SellerCloseData closes pushlished data in contract.
func SellerCloseData(merkleRoot string, Log ILogger) {
	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/s/close"
	body := make(url.Values)
	body.Add("merkle_root", merkleRoot)

	Log.Debugf("start send request for closing data...url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request for closing data finish.")
	Log.Debugf("%s", string(responseBody))
}

//SellerWithdrawETHForData withdraw ETH from a seller in contract.
func SellerWithdrawETHForData(merkle_root string, Log ILogger) {
	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/s/withdraw/data"
	body := make(url.Values)
	body.Add("merkle_root", merkle_root)

	Log.Debugf("start send request for withdrawing ETH...url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request for withdrawing ETH finish.")
	Log.Debugf("%s", string(responseBody))
}

//SellerWithdrawETHForTx withdraw ETH from a seller in contract.
func SellerWithdrawETHForTx(sessionID string, Log ILogger) {
	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/s/withdraw/tx"
	body := make(url.Values)
	body.Add("session_id", sessionID)
	Log.Debugf("buyer withdraws for contract. sessionID=%v", sessionID)

	Log.Debugf("start send request to withdraw for contract...url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request to withdraw for contract finish.")
	Log.Debugf("%s", string(responseBody))
}
