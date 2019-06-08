package main

import (
	"errors"
	

	plainbatch1 "github.com/sec-bit/zkPoD-lib/pod-go/plain/batch"
	"github.com/sec-bit/zkPoD-lib/pod-go/types"
)

type PoDSellerPB1 struct {
	SellerSession *plainbatch1.SellerSession `json:"sellerSession"`
}

// sellerNewSessForPB1 prepares seller's session while mode is plain_batch1.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerPB1 struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForPB1(publishPath string, Log ILogger) (PoDSellerPB1, error) {
	var pb1 PoDSellerPB1
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return pb1, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return pb1, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	pb1.SellerSession, err = plainbatch1.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return pb1, errors.New("failed to create session for seller")
	}
	Log.Debugf("success to create session for seller")
	return pb1, nil
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is plain_batch1.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (pb1 PoDSellerPB1) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := pb1.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify transaction requset and generate transaction response")
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is plain_batch1.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (pb1 PoDSellerPB1) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := pb1.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt file and generate secret file. receipt file=%v, secret file=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerPB1 struct {
	BuyerSession *plainbatch1.BuyerSession `json:"buyerSession"`
	Demands      []Demand                  `json:"demands"`
}

// buyerNewSessForPB1 prepares buyer's session while mode is plain_batch1.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a PoDBuyerPB1 struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForPB1(demandArr []Demand, plainBulletin string, plainPublicPath string, Log ILogger) (PoDBuyerPB1, error) {

	var pb1 PoDBuyerPB1
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	pb1.Demands = demandArr

	session, err := plainbatch1.NewBuyerSession(plainBulletin, plainPublicPath, sellerID, buyerID, demands)
	if err != nil {
		Log.Warnf("failed to create session for buyer. err=%v", err)
		return pb1, errors.New("failed to create session for buyer")
	}
	pb1.BuyerSession = session
	Log.Debugf("success to create session for buyer")
	return pb1, nil
}

// buyerNewReq creates request file for buyer while mode is plain_batch1.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (pb1 PoDBuyerPB1) buyerNewReq(requestFile string, Log ILogger) error {
	err := pb1.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request file. filepath=%v", requestFile)
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is plain_batch1.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (pb1 PoDBuyerPB1) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := pb1.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. responseFile=%v, receiptFile=%v", responseFile, receiptFile)
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is plain_batch1.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (pb1 PoDBuyerPB1) buyerVerifySecret(secretFile string, Log ILogger) bool {
	err := pb1.BuyerSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret")
	return true
}

// buyerGeneClaim generates claim with incorrect secret for buyer while mode is plain_batch1.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If generate claim successfully, return true.
//  Otherwise, return false.
func (pb1 PoDBuyerPB1) buyerGeneClaim(claimFile string, Log ILogger) bool {
	err := pb1.BuyerSession.GenerateClaim(claimFile)
	if err != nil {
		Log.Warnf("generate claim failure. err=%v", err)
		return false
	}
	Log.Debugf("success to generate claim, claimFile=%v", claimFile)
	return true
}

// buyerDecrypt decrypts file for buyer while mode is plain_batch1.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (pb1 PoDBuyerPB1) buyerDecrypt(outFile string, Log ILogger) bool {
	err := pb1.BuyerSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt claim. outFile=%v", outFile)
	return true
}
