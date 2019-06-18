package main

import (
	"errors"

	plain_atomic_swap "github.com/sec-bit/zkPoD-lib/pod_go/plain/atomic_swap"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDSellerPAS struct {
	SellerSession *plain_atomic_swap.SellerSession `json:"sellerSession"`
}

// sellerNewSessForPAS prepares seller's session while mode is plain_atomic_swap.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerPAS struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForPAS(publishPath string, Log ILogger) (PoDSellerPAS, error) {
	var pas PoDSellerPAS
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return pas, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return pas, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	pas.SellerSession, err = plain_atomic_swap.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return pas, errors.New("failed to create session for seller")
	}
	Log.Debugf("success to create session for seller")
	return pas, nil
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is plain_atomic_swap.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (pas PoDSellerPAS) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := pas.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is plain_atomic_swap.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (pas PoDSellerPAS) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := pas.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt file and generate secret file. receipt file=%v, secret file=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerPAS struct {
	BuyerSession *plain_atomic_swap.BuyerSession `json:"buyerSession"`
	Demands      []Demand                        `json:"demands"`
}

// buyerNewSessForPAS prepares buyer's session while mode is plain_atomic_swap.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a buyer's session and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForPAS(demandArr []Demand, plainBulletin string, plainPublicPath string, Log ILogger) (PoDBuyerPAS, error) {

	var pas PoDBuyerPAS
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	pas.Demands = demandArr

	session, err := plain_atomic_swap.NewBuyerSession(plainBulletin, plainPublicPath, sellerID, buyerID, demands)
	if err != nil {
		Log.Warnf("Failed to create session for buyer. err=%v", err)
		return pas, errors.New("Failed to create session for buyer")
	}
	pas.BuyerSession = session
	Log.Debugf("success to create session for buyer")
	return pas, nil
}

// buyerNewReq creates request file for buyer while mode is plain_atomic_swap.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (pas PoDBuyerPAS) buyerNewReq(requestFile string, Log ILogger) error {
	err := pas.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request file. reqeuestFile=%v", requestFile)
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is plain_atomic_swap.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (pas PoDBuyerPAS) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := pas.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. receiptFile=%v", receiptFile)
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is plain_atomic_swap.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (pas PoDBuyerPAS) buyerVerifySecret(secretFile string, Log ILogger) bool {
	err := pas.BuyerSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret. secretFile=%v", secretFile)
	return true
}

// buyerDecrypt decrypts file for buyer while mode is plain_atomic_swap.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (pas PoDBuyerPAS) buyerDecrypt(outFile string, Log ILogger) bool {
	err := pas.BuyerSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
