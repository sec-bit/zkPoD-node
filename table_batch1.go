package main

import (
	"errors"

	tablebatch1 "github.com/sec-bit/zkPoD-lib/pod-go/table/batch"
	"github.com/sec-bit/zkPoD-lib/pod-go/types"
)

type PoDSellerTB1 struct {
	SellerSession *tablebatch1.SellerSession `json:"sellerSession"`
}

// sellerNewSessForTB1 prepares seller's session while mode is tablebatch1.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a seller's session and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForTB1(publishPath string, Log ILogger) (PoDSellerTB1, error) {
	var tb1 PoDSellerTB1
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("failed to check. err=%v", err)
		return tb1, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return tb1, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	tb1.SellerSession, err = tablebatch1.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return tb1, errors.New("failed to create session for seller")
	}
	Log.Debugf("success to create session for seller.")
	return tb1, nil
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is tablebatch1.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (tb1 PoDSellerTB1) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := tb1.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is table_batch1.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (tb1 PoDSellerTB1) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := tb1.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerTB1 struct {
	BuyerSession *tablebatch1.BuyerSession `json:"buyerSession"`
	Demands      []Demand                  `json:"demands"`
}

// buyerNewSessForTB1 prepares buyer's session while mode is tablebatch1.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a PoDBuyer struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForTB1(demandArr []Demand, tableBulletin string, tablePublicPath string, Log ILogger) (PoDBuyerTB1, error) {

	var tb1 PoDBuyerTB1
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	tb1.Demands = demandArr

	session, err := tablebatch1.NewBuyerSession(tableBulletin, tablePublicPath, sellerID, buyerID, demands)
	if err != nil {
		Log.Warnf("Failed to create session for buyer. err=%v", err)
		return tb1, errors.New("Failed to create session for buyer")
	}
	tb1.BuyerSession = session
	Log.Debugf("success to create session for buyer.")
	return tb1, nil
}

// buyerNewReq creates request file for buyer while mode is table_batch1.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (tb1 PoDBuyerTB1) buyerNewReq(requestFile string, Log ILogger) error {
	err := tb1.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request file for buyer")
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is table_batch1.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (tb1 PoDBuyerTB1) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := tb1.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("Verify response failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. responseFile=%v， receiptFile=%v", responseFile, receiptFile)
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is table_batch1.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (tb1 PoDBuyerTB1) buyerVerifySecret(secretFile string, Log ILogger) bool {
	err := tb1.BuyerSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("Verify secret failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret. secretFile=%v", secretFile)
	return true
}

// buyerGeneClaim generates claim with incorrect secret for buyer while mode is table_batch1.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If generate claim successfully, return true.
//  Otherwise, return false.
func (tb1 PoDBuyerTB1) buyerGeneClaim(claimFile string, Log ILogger) bool {
	err := tb1.BuyerSession.GenerateClaim(claimFile)
	if err != nil {
		Log.Warnf("generate claim failure. err=%v", err)
		return false
	}
	Log.Debugf("success to generate claim. claimFile=%v", claimFile)
	return true
}

// buyerDecrypt decrypts file for buyer while mode is table_batch1.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (tb1 PoDBuyerTB1) buyerDecrypt(outFile string, Log ILogger) bool {
	err := tb1.BuyerSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
