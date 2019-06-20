package main

import (
	"errors"

	plain_ot_complaint "github.com/sec-bit/zkPoD-lib/pod_go/plain/ot_complaint"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDSellerPOC struct {
	SellerSession *plain_ot_complaint.SellerSession `json:"sellerSession"`
}

// sellerNewSessForPOC prepares seller's session while mode is plain_ot_complaint.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerPOC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForPOC(publishPath string, sellerID [40]uint8, buyerID [40]uint8, Log ILogger) (PoDSellerPOC, error) {
	var poc PoDSellerPOC
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return poc, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return poc, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	poc.SellerSession, err = plain_ot_complaint.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return poc, errors.New("failed to create session for seller")
	}
	return poc, nil
}

// sellerGeneNegoReq generates nego request file for seller while mode is plain_ot_complaint.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generate nego request successfully, return true.
//  Otherwise, return false.
func (poc PoDSellerPOC) sellerGeneNegoReq(requestFile string, Log ILogger) bool {

	err := poc.SellerSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to generate nego request for seller. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego request for seller. requestFile=%v", requestFile)
	return true
}

// sellerGeneNegoResp generates nego response file for seller while mode is plain_ot_complaint.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If verify nego request and generate nego response successfully, return true.
//  Otherwise, return false.
func (poc PoDSellerPOC) sellerGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := poc.SellerSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("failed to verify nego request and generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify nego request and generate nego response. requestFile=%v, responseFile=%v", requestFile, responseFile)
	return true
}

// sellerDealNegoResp deals with buyer's nego response for seller while mode is plain_ot_complaint.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deal with buyer's nego response successfully, return true.
//  Otherwise, return false.
func (poc PoDSellerPOC) sellerDealNegoResp(responseFile string, Log ILogger) bool {
	err := poc.SellerSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("failed to deal with buyer's nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to deal with buyer's nego response. responseFile=%v", responseFile)
	return true
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is plain_ot_complaint.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (poc PoDSellerPOC) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := poc.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response. requestFile=%v, responseFile=%v", requestFile, responseFile)
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is plain_ot_complaint.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (poc PoDSellerPOC) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := poc.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret file. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerPOC struct {
	BuyerSession *plain_ot_complaint.BuyerSession `json:"buyerSession"`
	Demands      []Demand                         `json:"demands"`
	Phantoms     []Phantom                        `json:"phantoms"`
}

// buyerNewSessForPOC prepares buyer's session while mode is plain_ot_complaint.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a PoDBuyerPOC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForPOC(demandArr []Demand, phantomArr []Phantom, plainBulletin string, plainPublicPath string, sellerID [40]uint8, buyerID [40]uint8, Log ILogger) (PoDBuyerPOC, error) {

	var poc PoDBuyerPOC
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	poc.Demands = demandArr

	phantoms := make([]types.Range, 0)
	for _, p := range phantomArr {
		phantoms = append(phantoms, types.Range{p.PhantomStart, p.PhantomCount})
	}
	poc.Phantoms = phantomArr

	session, err := plain_ot_complaint.NewBuyerSession(plainBulletin, plainPublicPath, sellerID, buyerID, demands, phantoms)
	if err != nil {
		Log.Warnf("Failed to create session for buyer. err=%v", err)
		return poc, errors.New("Failed to create session for buyer")
	}
	poc.BuyerSession = session
	Log.Debugf("success to create session for seller.")
	return poc, nil
}

// buyerGeneNegoReq generates nego request file for buyer while mode is plain_ot_complaint.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generates nego request for buyer successfully, return true.
//  Otherwise, return false.
func (poc PoDBuyerPOC) buyerGeneNegoReq(requestFile string, Log ILogger) bool {

	err := poc.BuyerSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to generates nego request for buyer. err=%v", err)
		return false
	}
	Log.Debugf("success to generates nego request for buyer. requestFile=%v", requestFile)
	return true
}

// buyerGeneNegoResp verifies nego request and generates nego response for buyer while mode is plain_ot_complaint.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If generates nego response for buyer successfully, return true.
//  Otherwise, return false.
func (poc PoDBuyerPOC) buyerGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := poc.BuyerSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to verify nego request and generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify nego request and generate nego response. requestFile=%v, responseFile=%v", requestFile, responseFile)
	return true
}

// buyerDealNegoResp deals with seller's nego response for buyer while mode is plain_ot_complaint.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deals with seller's nego response successfully, return true.
//  Otherwise, return false.
func (poc PoDBuyerPOC) buyerDealNegoResp(responseFile string, Log ILogger) bool {
	err := poc.BuyerSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to deals with seller's nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to deal with seller's nego response. responseFile=%v", responseFile)
	return true
}

// buyerNewReq creates request file for buyer while mode is plain_ot_complaint.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (poc PoDBuyerPOC) buyerNewReq(requestFile string, Log ILogger) error {
	err := poc.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request. requestFile=%v", requestFile)
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is plain_ot_complaint.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (poc PoDBuyerPOC) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := poc.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. responseFile=%v, receiptFile=%v", responseFile, receiptFile)
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is plain_ot_complaint.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (poc PoDBuyerPOC) buyerVerifySecret(secretFile string, Log ILogger) bool {
	err := poc.BuyerSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret. secretFile=%v", secretFile)
	return true
}

// buyerGeneClaim generates claim with incorrect secret for buyer while mode is plain_ot_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If generate claim successfully, return true.
//  Otherwise, return false.
func (poc PoDBuyerPOC) buyerGeneClaim(claimFile string, Log ILogger) bool {
	err := poc.BuyerSession.GenerateClaim(claimFile)
	if err != nil {
		Log.Warnf("failed to generate claim. err=%v", err)
		return false
	}
	Log.Debugf("success to generate claim. claimFile=%v", claimFile)
	return true
}

//buyerDecrypt decrypts file for buyer while mode is plain_ot_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (poc PoDBuyerPOC) buyerDecrypt(outFile string, Log ILogger) bool {
	err := poc.BuyerSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
