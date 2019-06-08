package main

import (
	"errors"
	

	tableotvrf "github.com/sec-bit/zkPoD-lib/pod-go/table/otvrfq"
)

type PoDSellerTOQ struct {
	SellerSession *tableotvrf.SellerSession `json:"sellerSession"`
}

// sellerNewSessForTOQ prepares seller's session while mode is table_ot_vrf_query.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerTOQ struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForTOQ(publishPath string, Log ILogger) (PoDSellerTOQ, error) {
	var toq PoDSellerTOQ
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("failed to check. err=%v", err)
		return toq, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return toq, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	toq.SellerSession, err = tableotvrf.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return toq, errors.New("failed to create session for seller")
	}
	Log.Debugf("success to create session")
	return toq, nil
}

// sellerGeneNegoReq generates nego request file for seller while mode is table_ot_vrf_query.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generate nego request successfully, return true.
//  Otherwise, return false.
func (toq PoDSellerTOQ) sellerGeneNegoReq(requestFile string, Log ILogger) bool {

	err := toq.SellerSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to generate nego request. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego request. requestFile=%v", requestFile)
	return true
}

// sellerGeneNegoResp generates nego response file for seller while mode is table_ot_vrf_query.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If verify nego request and generate nego response successfully, return true.
//  Otherwise, return false.
func (toq PoDSellerTOQ) sellerGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := toq.SellerSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to verify nego request and generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify nego request and generate nego response")
	return true
}

// sellerDealNegoResp deals with buyer's nego response for seller while mode is table_ot_vrf_query.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deal with buyer's nego response successfully, return true.
//  Otherwise, return false.
func (toq PoDSellerTOQ) sellerDealNegoResp(responseFile string, Log ILogger) bool {
	err := toq.SellerSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to deal with nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to deal with nego response")
	return true
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is table_ot_vrf_query.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (toq PoDSellerTOQ) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := toq.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is table_ot_vrf_query.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (toq PoDSellerTOQ) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := toq.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerTOQ struct {
	BuyerSession    *tableotvrf.BuyerSession `json:"buyerSession"`
	KeyName         string                   `json:"keyName"`
	KeyValue        []string                 `json:"keyValue"`
	PhantomKeyValue []string                 `json:"phantomKeyValue"`
}

// buyerNewSessForTOQ prepares buyer's session while mode is table_ot_vrf_query.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a PoDBuyerTOQ struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForTOQ(keyName string, keyValue []string, phantomKeyValue []string, tableBulletin string, tablePublicPath string, Log ILogger) (PoDBuyerTOQ, error) {
	var toq PoDBuyerTOQ
	session, err := tableotvrf.NewBuyerSession(tableBulletin, tablePublicPath, sellerID, buyerID, keyName, keyValue, phantomKeyValue)
	if err != nil {
		Log.Warnf("failed to create session for buyer. err=%v", err)
		return toq, errors.New("failed to create session for buyer")
	}
	Log.Debugf("success to creatr session for buyer.")
	toq = PoDBuyerTOQ{session, keyName, keyValue, phantomKeyValue}
	return toq, nil
}

// buyerGeneNegoReq generates nego request file for buyer while mode is table_ot_vrf_query.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generates nego request for buyer successfully, return true.
//  Otherwise, return false.
func (toq PoDBuyerTOQ) buyerGeneNegoReq(requestFile string, Log ILogger) bool {

	err := toq.BuyerSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to generate nego request. err=%v", err)
		return false
	}
	Log.Debugf("generate nego request. requestFile=%v", requestFile)
	return true
}

// buyerGeneNegoResp verifies nego request and generates nego response for buyer while mode is table_ot_vrf_query.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If generates nego response for buyer successfully, return true.
//  Otherwise, return false.
func (toq PoDBuyerTOQ) buyerGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := toq.BuyerSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego response. requestFile=%v, responseFile=%v", requestFile, responseFile)
	return true
}

// buyerNewReq creates request file for buyer while mode is table_ot_vrf_query.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (toq PoDBuyerTOQ) buyerNewReq(requestFile string, Log ILogger) error {
	err := toq.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request. requestFile=%v", requestFile)
	return nil
}

// buyerDealNegoResp deals with seller's nego response for buyer while mode is table_ot_vrf_query.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deals with seller's nego response successfully, return true.
//  Otherwise, return false.
func (toq PoDBuyerTOQ) buyerDealNegoResp(responseFile string, Log ILogger) bool {
	err := toq.BuyerSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to deal with nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to deal with nego response")
	return true
}

// buyerVerifyResp verifies response data for buyer while mode is table_ot_vrf_query.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (toq PoDBuyerTOQ) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := toq.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("Verify response failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt")
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is table_ot_vrf_query.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (toq PoDBuyerTOQ) buyerVerifySecret(secretFile string, tablePosition string, Log ILogger) bool {
	err := toq.BuyerSession.OnSecret(secretFile, tablePosition)
	if err != nil {
		Log.Warnf("Verify secret failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret and generate position file. secretFile=%v, tablePosition=%v", secretFile, tablePosition)
	return true
}
