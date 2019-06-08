package main

import (
	"errors"
	

	tablevrf "github.com/sec-bit/zkPoD-lib/pod-go/table/vrfq"
)

type PoDSellerTQ struct {
	SellerSession *tablevrf.SellerSession `json:"sellerSession"`
}

// sellerNewSessForTQ prepares seller's session while mode is table_vrf_query.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerTQ struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForTQ(publishPath string, Log ILogger) (PoDSellerTQ, error) {
	var tq PoDSellerTQ
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return tq, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return tq, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	tq.SellerSession, err = tablevrf.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return tq, errors.New("failed to create session for seller")
	}
	Log.Debugf("success to create session")
	return tq, nil
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is table_vrf_query.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (tq PoDSellerTQ) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := tq.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is table_vrf_query.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (tq PoDSellerTQ) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := tq.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerTQ struct {
	BuyerSession *tablevrf.BuyerSession `json:"buyerSession"`
	KeyName      string                 `json:"keyName"`
	KeyValue     []string               `json:"keyValue"`
}

// buyerNewSessForTQ prepares buyer's session while mode is table_vrf_query.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a buyer's session and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForTQ(keyName string, keyValue []string, tableBulletin string, tablePublicPath string, Log ILogger) (PoDBuyerTQ, error) {

	var tq PoDBuyerTQ
	Log.Debugf("tableBulletin=%v, tablePublicPath=%v, sellerID=%v, buyerID=%v, keyName=%v,keyValue=%v",
		tableBulletin, tablePublicPath, sellerID, buyerID, keyName, keyValue)
	session, err := tablevrf.NewBuyerSession(tableBulletin, tablePublicPath, sellerID, buyerID, keyName, keyValue)
	if err != nil {
		Log.Warnf("Failed to create session for buyer. err=%v", err)
		return tq, errors.New("Failed to create session for buyer")
	}
	Log.Debugf("success to create session.")
	tq = PoDBuyerTQ{session, keyName, keyValue}
	return tq, nil
}

// buyerNewReq creates request file for buyer while mode is table_vrf_query.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (tq PoDBuyerTQ) buyerNewReq(requestFile string, Log ILogger) error {
	err := tq.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request file. requestFile=%v", requestFile)
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is table_vrf_query.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (tq PoDBuyerTQ) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := tq.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("Verify response failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and geerate receipt. responseFile=%v, receiptFile=%v", responseFile, receiptFile)
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is table_vrf_query.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (tq PoDBuyerTQ) buyerVerifySecret(secretFile string, tablePosition string, Log ILogger) bool {
	err := tq.BuyerSession.OnSecret(secretFile, tablePosition)
	if err != nil {
		Log.Warnf("Verify secret failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret and generate table position. secretFile=%v, tablePosition=%v", secretFile, tablePosition)
	return true
}
