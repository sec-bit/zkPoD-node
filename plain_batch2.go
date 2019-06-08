package main

import (
	"errors"

	plainbatch2 "github.com/sec-bit/zkPoD-lib/pod-go/plain/batch2"
	"github.com/sec-bit/zkPoD-lib/pod-go/types"
)

type PoDSellerPB2 struct {
	SellerSession *plainbatch2.SellerSession `json:"sellerSession"`
}

// sellerNewSessForPB2 prepares seller's session while mode is plain_batch2.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerPB2 struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForPB2(publishPath string, Log ILogger) (PoDSellerPB2, error) {
	var pb2 PoDSellerPB2
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return pb2, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return pb2, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	pb2.SellerSession, err = plainbatch2.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return pb2, errors.New("failed to create session for seller")
	}
	Log.Debugf("success to create session for seller")
	return pb2, nil
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is plain_batch2.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (pb2 PoDSellerPB2) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := pb2.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is plain_batch2.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (pb2 PoDSellerPB2) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := pb2.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt file and generate secret file. receipt file=%v, secret file=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerPB2 struct {
	BuyerSession *plainbatch2.BuyerSession `json:"buyerSession"`
	Demands      []Demand                  `json:"demands"`
}

// buyerNewSessForPB2 prepares buyer's session while mode is plain_batch2.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a buyer's session and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForPB2(demandArr []Demand, plainBulletin string, plainPublicPath string, Log ILogger) (PoDBuyerPB2, error) {

	var pb2 PoDBuyerPB2
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	pb2.Demands = demandArr

	session, err := plainbatch2.NewBuyerSession(plainBulletin, plainPublicPath, sellerID, buyerID, demands)
	if err != nil {
		Log.Warnf("Failed to create session for buyer. err=%v", err)
		return pb2, errors.New("Failed to create session for buyer")
	}
	pb2.BuyerSession = session
	Log.Debugf("success to create session for buyer")
	return pb2, nil
}

// buyerNewReq creates request file for buyer while mode is plain_batch2.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (pb2 PoDBuyerPB2) buyerNewReq(requestFile string, Log ILogger) error {
	err := pb2.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request file. reqeuestFile=%v", requestFile)
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is plain_batch2.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (pb2 PoDBuyerPB2) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := pb2.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. receiptFile=%v", receiptFile)
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is plain_batch2.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (pb2 PoDBuyerPB2) buyerVerifySecret(secretFile string, Log ILogger) bool {
	err := pb2.BuyerSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret. secretFile=%v", secretFile)
	return true
}

// buyerDecrypt decrypts file for buyer while mode is plain_batch2.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (pb2 PoDBuyerPB2) buyerDecrypt(outFile string, Log ILogger) bool {
	err := pb2.BuyerSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
