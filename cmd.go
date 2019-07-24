package main

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
)

//BobPurchaseData purchases data.
func BobPurchaseData(requestData RequestData, Log ILogger) {

	data, err := json.Marshal(requestData)
	if err != nil {
		Log.Warnf("failed to marshal request data. err=%v", err)
		return
	}
	Log.Debugf("data sigma merkle root=%v, Alice ip=%v", requestData.MerkleRoot, requestData.AliceIP)

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

//BobDepositETH deposit ETH to a Alice in contract.
func BobDepositETH(value string, AliceAddress string, Log ILogger) {
	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/b/deposit"
	body := make(url.Values)
	body.Add("value", value)
	body.Add("address", AliceAddress)
	Log.Debugf("Bob deposit to Alice. value=%v, Alice address=%v", value, AliceAddress)

	Log.Debugf("start send request to deposit ETH. url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request to deposit ETH finish.")
	Log.Debugf("%s", string(responseBody))
}

//BobUnDepositETH undeposits ETH from a Alice in contract.
func BobUnDepositETH(AliceAddress string, Log ILogger) {
	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/b/undeposit"
	body := make(url.Values)
	body.Add("address", AliceAddress)
	Log.Debugf("Bob undeposit from Alice. Alice address=%v", AliceAddress)

	Log.Debugf("start send request to undeposit ETH. url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request to undeposit ETH finish.")
	Log.Debugf("%s", string(responseBody))
}

//BobWithdrawETH withdraw ETH from a Alice in contract.
func BobWithdrawETH(AliceAddress string, Log ILogger) {
	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/b/withdraw"
	body := make(url.Values)
	body.Add("address", AliceAddress)
	Log.Debugf("Bob withdraws for contract. Alice address=%v", AliceAddress)

	Log.Debugf("start send request to withdraw for contract. url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request to withdraw for contract finish.")
	Log.Debugf("%s", string(responseBody))
}

//AliceInitDataNode initialize data for publishing.
func AliceInitDataNode(filepath string, Log ILogger) {

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		Log.Warnf("read config file error! filepath=%s, err=%v", filepath, err)
		panic(err.Error())
	}
	// Log.Debugf("data=%v", string(data))

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

// AlicePublishData sends a request to server
// to initializes data and publishes data to contract by Alice.
func AlicePublishData(merkleRoot string, value string, Log ILogger) {

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

//AliceCloseData closes pushlished data in contract.
func AliceCloseData(merkleRoot string, Log ILogger) {
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

//AliceWithdrawETHForData withdraw ETH from a Alice in contract.
func AliceWithdrawETHForData(merkle_root string, Log ILogger) {
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

//AliceWithdrawETHForTx withdraw ETH from a Alice in contract.
func AliceWithdrawETHForTx(sessionID string, Log ILogger) {
	urlStr := REQUEST_URL_LOCAL_HOST + ":" + BConf.Port + "/s/withdraw/tx"
	body := make(url.Values)
	body.Add("session_id", sessionID)
	Log.Debugf("Bob withdraws for contract. sessionID=%v", sessionID)

	Log.Debugf("start send request to withdraw for contract...url=%v", urlStr)
	responseBody, err := sendRequest(body, REQUEST_METHOD_POST, urlStr, Log)
	if err != nil {
		Log.Warnf("read response error! url=%s, err=%v", urlStr, err)
		panic(err.Error())
	}
	Log.Debugf("send request to withdraw for contract finish.")
	Log.Debugf("%s", string(responseBody))
}
