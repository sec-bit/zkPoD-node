package main

import (
	"errors"
	

	tableotbatch1 "github.com/sec-bit/zkPoD-lib/pod-go/table/otbatch"
	"github.com/sec-bit/zkPoD-lib/pod-go/types"
)

type PoDSellerTOB1 struct {
	SellerSession *tableotbatch1.SellerSession `json:"sellerSession"`
}

// sellerNewSessForTOB1 prepares seller's session while mode is table_ot_batch1.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerTOB1 struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForTOB1(publishPath string, Log ILogger) (PoDSellerTOB1, error) {
	var tob1 PoDSellerTOB1
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return tob1, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return tob1, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	tob1.SellerSession, err = tableotbatch1.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return tob1, errors.New("failed to create session for seller")
	}
	Log.Debugf("success to create session for seller.")
	return tob1, nil
}

// sellerGeneNegoReq generates nego request file for seller while mode is table_ot_batch1.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generate nego request successfully, return true.
//  Otherwise, return false.
func (tob1 PoDSellerTOB1) sellerGeneNegoReq(requestFile string, Log ILogger) bool {

	err := tob1.SellerSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to generate nego request. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego request. requestFile=%v", requestFile)
	return true
}

// sellerGeneNegoResp generates nego response file for seller while mode is table_ot_batch1.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If verify nego request and generate nego response successfully, return true.
//  Otherwise, return false.
func (tob1 PoDSellerTOB1) sellerGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := tob1.SellerSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// sellerDealNegoResp deals with buyer's nego response for seller while mode is table_ot_batch1.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deal with buyer's nego response successfully, return true.
//  Otherwise, return false.
func (tob1 PoDSellerTOB1) sellerDealNegoResp(responseFile string, Log ILogger) bool {
	err := tob1.SellerSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego response")
	return true
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is table_ot_batch1.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (tob1 PoDSellerTOB1) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := tob1.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is table_ot_batch1.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (tob1 PoDSellerTOB1) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := tob1.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerTOB1 struct {
	BuyerSession *tableotbatch1.BuyerSession `json:"buyerSession"`
	Demands      []Demand                    `json:"demands"`
	Phantoms     []Phantom                   `json:"phantoms"`
}

// buyerNewSessForTOB1 prepares buyer's session while mode is table_ot_batch1.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a buyer's session and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForTOB1(demandArr []Demand, phantomArr []Phantom, tableBulletin string, tablePublicPath string, Log ILogger) (PoDBuyerTOB1, error) {

	var tob1 PoDBuyerTOB1
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	tob1.Demands = demandArr

	phantoms := make([]types.Range, 0)
	for _, p := range phantomArr {
		phantoms = append(phantoms, types.Range{p.PhantomStart, p.PhantomCount})
	}
	tob1.Phantoms = phantomArr

	session, err := tableotbatch1.NewBuyerSession(tableBulletin, tablePublicPath, sellerID, buyerID, demands, phantoms)
	if err != nil {
		Log.Warnf("failed to create session for buyer. err=%v", err)
		return tob1, errors.New("failed to create session for buyer")
	}
	tob1.BuyerSession = session
	Log.Debugf("success to create session for buyer.")
	return tob1, nil
}

// buyerGeneNegoReq generates nego request file for buyer while mode is table_ot_batch1.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generates nego request for buyer successfully, return true.
//  Otherwise, return false.
func (tob1 PoDBuyerTOB1) buyerGeneNegoReq(requestFile string, Log ILogger) bool {

	err := tob1.BuyerSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to generate nego request. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego request")
	return true
}

// buyerGeneNegoResp verifies nego request and generates nego response for buyer while mode is table_ot_batch1.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If generates nego response for buyer successfully, return true.
//  Otherwise, return false.
func (tob1 PoDBuyerTOB1) buyerGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := tob1.BuyerSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to verify request ans generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request ans generate nego response.")
	return true
}

// buyerDealNegoResp deals with seller's nego response for buyer while mode is table_ot_batch1.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deals with seller's nego response successfully, return true.
//  Otherwise, return false.
func (tob1 PoDBuyerTOB1) buyerDealNegoResp(responseFile string, Log ILogger) bool {
	err := tob1.BuyerSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego resoponse")
	return true
}

// buyerNewReq creates request file for buyer while mode is table_ot_batch1.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (tob1 PoDBuyerTOB1) buyerNewReq(requestFile string, Log ILogger) error {
	err := tob1.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to create request file. err=%v", err)
		return errors.New("failed to create request file")
	}
	Log.Debugf("success to create request file. requestFile=%v", requestFile)
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is table_ot_batch1.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (tob1 PoDBuyerTOB1) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := tob1.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("verify response failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. responseFile=%v receiptFile=%v", responseFile, receiptFile)
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is table_ot_batch1.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (tob1 PoDBuyerTOB1) buyerVerifySecret(secretFile string, Log ILogger) bool {
	err := tob1.BuyerSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("verify secret failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret.")
	return true
}

// buyerGeneClaim generates claim with incorrect secret for buyer while mode is table_ot_batch1.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If generate claim successfully, return true.
//  Otherwise, return false.
func (tob1 PoDBuyerTOB1) buyerGeneClaim(claimFile string, Log ILogger) bool {
	err := tob1.BuyerSession.GenerateClaim(claimFile)
	if err != nil {
		Log.Warnf("generate claim failure. err=%v", err)
		return false
	}
	Log.Debugf("success to generate claim. claimFile=%v", claimFile)
	return true
}

// buyerDecrypt decrypts file for buyer while mode is table_ot_batch1.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (tob1 PoDBuyerTOB1) buyerDecrypt(outFile string, Log ILogger) bool {
	err := tob1.BuyerSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
