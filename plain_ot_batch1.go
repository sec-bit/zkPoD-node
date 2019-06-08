package main

import (
	"errors"
	

	plainotbatch1 "github.com/sec-bit/zkPoD-lib/pod-go/plain/otbatch"
	"github.com/sec-bit/zkPoD-lib/pod-go/types"
)

type PoDSellerPOB1 struct {
	SellerSession *plainotbatch1.SellerSession `json:"sellerSession"`
}

// sellerNewSessForPOB1 prepares seller's session while mode is plain_ot_batch1.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerPOB1 struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForPOB1(publishPath string, Log ILogger) (PoDSellerPOB1, error) {
	var pob1 PoDSellerPOB1
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return pob1, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return pob1, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	pob1.SellerSession, err = plainotbatch1.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return pob1, errors.New("failed to create session for seller")
	}
	return pob1, nil
}

// sellerGeneNegoReq generates nego request file for seller while mode is plain_ot_batch1.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generate nego request successfully, return true.
//  Otherwise, return false.
func (pob1 PoDSellerPOB1) sellerGeneNegoReq(requestFile string, Log ILogger) bool {

	err := pob1.SellerSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to generate nego request for seller. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego request for seller. requestFile=%v", requestFile)
	return true
}

// sellerGeneNegoResp generates nego response file for seller while mode is plain_ot_batch1.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If verify nego request and generate nego response successfully, return true.
//  Otherwise, return false.
func (pob1 PoDSellerPOB1) sellerGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := pob1.SellerSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("failed to verify nego request and generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify nego request and generate nego response. requestFile=%v, responseFile=%v", requestFile, responseFile)
	return true
}

// sellerDealNegoResp deals with buyer's nego response for seller while mode is plain_ot_batch1.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deal with buyer's nego response successfully, return true.
//  Otherwise, return false.
func (pob1 PoDSellerPOB1) sellerDealNegoResp(responseFile string, Log ILogger) bool {
	err := pob1.SellerSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("failed to deal with buyer's nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to deal with buyer's nego response. responseFile=%v", responseFile)
	return true
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is plain_ot_batch1.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (pob1 PoDSellerPOB1) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := pob1.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response. requestFile=%v, responseFile=%v", requestFile, responseFile)
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is plain_ot_batch1.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (pob1 PoDSellerPOB1) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := pob1.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret file. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerPOB1 struct {
	BuyerSession *plainotbatch1.BuyerSession `json:"buyerSession"`
	Demands      []Demand                    `json:"demands"`
	Phantoms     []Phantom                   `json:"phantoms"`
}

// buyerNewSessForPOB1 prepares buyer's session while mode is plain_ot_batch1.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a PoDBuyerPOB1 struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForPOB1(demandArr []Demand, phantomArr []Phantom, plainBulletin string, plainPublicPath string, Log ILogger) (PoDBuyerPOB1, error) {

	var pob1 PoDBuyerPOB1
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	pob1.Demands = demandArr

	phantoms := make([]types.Range, 0)
	for _, p := range phantomArr {
		phantoms = append(phantoms, types.Range{p.PhantomStart, p.PhantomCount})
	}
	pob1.Phantoms = phantomArr

	session, err := plainotbatch1.NewBuyerSession(plainBulletin, plainPublicPath, sellerID, buyerID, demands, phantoms)
	if err != nil {
		Log.Warnf("Failed to create session for buyer. err=%v", err)
		return pob1, errors.New("Failed to create session for buyer")
	}
	pob1.BuyerSession = session
	Log.Debugf("success to create session for seller.")
	return pob1, nil
}

// buyerGeneNegoReq generates nego request file for buyer while mode is plain_ot_batch1.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generates nego request for buyer successfully, return true.
//  Otherwise, return false.
func (pob1 PoDBuyerPOB1) buyerGeneNegoReq(requestFile string, Log ILogger) bool {

	err := pob1.BuyerSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to generates nego request for buyer. err=%v", err)
		return false
	}
	Log.Debugf("success to generates nego request for buyer. requestFile=%v", requestFile)
	return true
}

// buyerGeneNegoResp verifies nego request and generates nego response for buyer while mode is plain_ot_batch1.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If generates nego response for buyer successfully, return true.
//  Otherwise, return false.
func (pob1 PoDBuyerPOB1) buyerGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := pob1.BuyerSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to verify nego request and generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify nego request and generate nego response. requestFile=%v, responseFile=%v", requestFile, responseFile)
	return true
}

// buyerDealNegoResp deals with seller's nego response for buyer while mode is plain_ot_batch1.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deals with seller's nego response successfully, return true.
//  Otherwise, return false.
func (pob1 PoDBuyerPOB1) buyerDealNegoResp(responseFile string, Log ILogger) bool {
	err := pob1.BuyerSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to deals with seller's nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to deal with seller's nego response. responseFile=%v", responseFile)
	return true
}

// buyerNewReq creates request file for buyer while mode is plain_ot_batch1.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (pob1 PoDBuyerPOB1) buyerNewReq(requestFile string, Log ILogger) error {
	err := pob1.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request. requestFile=%v", requestFile)
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is plain_ot_batch1.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (pob1 PoDBuyerPOB1) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := pob1.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. responseFile=%v, receiptFile=%v", responseFile, receiptFile)
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is plain_ot_batch1.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (pob1 PoDBuyerPOB1) buyerVerifySecret(secretFile string, Log ILogger) bool {
	err := pob1.BuyerSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret. secretFile=%v", secretFile)
	return true
}

// buyerGeneClaim generates claim with incorrect secret for buyer while mode is plain_ot_batch1.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If generate claim successfully, return true.
//  Otherwise, return false.
func (pob1 PoDBuyerPOB1) buyerGeneClaim(claimFile string, Log ILogger) bool {
	err := pob1.BuyerSession.GenerateClaim(claimFile)
	if err != nil {
		Log.Warnf("failed to generate claim. err=%v", err)
		return false
	}
	Log.Debugf("success to generate claim. claimFile=%v", claimFile)
	return true
}

//buyerDecrypt decrypts file for buyer while mode is plain_ot_batch1.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (pob1 PoDBuyerPOB1) buyerDecrypt(outFile string, Log ILogger) bool {
	err := pob1.BuyerSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
