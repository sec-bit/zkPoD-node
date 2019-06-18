package main

import (
	"errors"

	plain_complaint "github.com/sec-bit/zkPoD-lib/pod_go/plain/complaint"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDSellerPC struct {
	SellerSession *plain_complaint.SellerSession `json:"sellerSession"`
}

// sellerNewSessForPC prepares seller's session while mode is plain_complaint.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerPC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForPC(publishPath string, Log ILogger) (PoDSellerPC, error) {
	var pc PoDSellerPC
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return pc, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return pc, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	pc.SellerSession, err = plain_complaint.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return pc, errors.New("failed to create session for seller")
	}
	Log.Debugf("success to create session for seller")
	return pc, nil
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is plain_complaint.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (pc PoDSellerPC) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := pc.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify transaction requset and generate transaction response")
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is plain_complaint.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (pc PoDSellerPC) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := pc.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt file and generate secret file. receipt file=%v, secret file=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerPC struct {
	BuyerSession *plain_complaint.BuyerSession `json:"buyerSession"`
	Demands      []Demand                      `json:"demands"`
}

// buyerNewSessForPC prepares buyer's session while mode is plain_complaint.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a PoDBuyerPC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForPC(demandArr []Demand, plainBulletin string, plainPublicPath string, Log ILogger) (PoDBuyerPC, error) {

	var pc PoDBuyerPC
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	pc.Demands = demandArr

	session, err := plain_complaint.NewBuyerSession(plainBulletin, plainPublicPath, sellerID, buyerID, demands)
	if err != nil {
		Log.Warnf("failed to create session for buyer. err=%v", err)
		return pc, errors.New("failed to create session for buyer")
	}
	pc.BuyerSession = session
	Log.Debugf("success to create session for buyer")
	return pc, nil
}

// buyerNewReq creates request file for buyer while mode is plain_complaint.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (pc PoDBuyerPC) buyerNewReq(requestFile string, Log ILogger) error {
	err := pc.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request file. filepath=%v", requestFile)
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is plain_complaint.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (pc PoDBuyerPC) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := pc.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. responseFile=%v, receiptFile=%v", responseFile, receiptFile)
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is plain_complaint.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (pc PoDBuyerPC) buyerVerifySecret(secretFile string, Log ILogger) bool {
	err := pc.BuyerSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret")
	return true
}

// buyerGeneClaim generates claim with incorrect secret for buyer while mode is plain_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If generate claim successfully, return true.
//  Otherwise, return false.
func (pc PoDBuyerPC) buyerGeneClaim(claimFile string, Log ILogger) bool {
	err := pc.BuyerSession.GenerateClaim(claimFile)
	if err != nil {
		Log.Warnf("generate claim failure. err=%v", err)
		return false
	}
	Log.Debugf("success to generate claim, claimFile=%v", claimFile)
	return true
}

// buyerDecrypt decrypts file for buyer while mode is plain_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (pc PoDBuyerPC) buyerDecrypt(outFile string, Log ILogger) bool {
	err := pc.BuyerSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt claim. outFile=%v", outFile)
	return true
}
