package main

import (
	"errors"
	

	tablebatch2 "github.com/sec-bit/zkPoD-lib/pod-go/table/batch2"
	"github.com/sec-bit/zkPoD-lib/pod-go/types"
)

type PoDSellerTB2 struct {
	SellerSession *tablebatch2.SellerSession `json:"sellerSession"`
}

// sellerNewSessForTB2 prepares seller's session while mode is table_batch2.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerTB2 struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForTB2(publishPath string, Log ILogger) (PoDSellerTB2, error) {
	var tb2 PoDSellerTB2
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("failed to check. err=%v", err)
		return tb2, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return tb2, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	tb2.SellerSession, err = tablebatch2.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return tb2, errors.New("failed to create session for seller")
	}
	Log.Debugf("success to create session")
	return tb2, nil
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is table_batch2.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (tb2 PoDSellerTB2) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := tb2.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response.")
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is table_batch2.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (tb2 PoDSellerTB2) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := tb2.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt receipt and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret file. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerTB2 struct {
	BuyerSession *tablebatch2.BuyerSession `json:"buyerSession"`
	Demands      []Demand                  `json:"demands"`
}

// buyerNewSessForTB2 prepares buyer's session while mode is table_batch2.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a PoDBuyerTB2 struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForTB2(demandArr []Demand, tableBulletin string, tablePublicPath string, Log ILogger) (PoDBuyerTB2, error) {

	var tb2 PoDBuyerTB2
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	tb2.Demands = demandArr

	session, err := tablebatch2.NewBuyerSession(tableBulletin, tablePublicPath, sellerID, buyerID, demands)
	if err != nil {
		Log.Warnf("Failed to create session for buyer. err=%v", err)
		return tb2, errors.New("Failed to create session for buyer")
	}
	tb2.BuyerSession = session
	Log.Debugf("success to create session")
	return tb2, nil
}

// buyerNewReq creates request file for buyer while mode is table_batch2.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (tb2 PoDBuyerTB2) buyerNewReq(requestFile string, Log ILogger) error {
	err := tb2.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to generate request for buyer.")
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is table_batch2.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (tb2 PoDBuyerTB2) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := tb2.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success tov verify response and generate receipt.")
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is table_batch2.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (tb2 PoDBuyerTB2) buyerVerifySecret(secretFile string, Log ILogger) bool {
	err := tb2.BuyerSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret")
	return true
}

// buyerDecrypt decrypts file for buyer while mode is table_batch2.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (tb2 PoDBuyerTB2) buyerDecrypt(outFile string, Log ILogger) bool {
	err := tb2.BuyerSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
