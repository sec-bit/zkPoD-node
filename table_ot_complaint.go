package main

import (
	"errors"

	table_ot_complaint "github.com/sec-bit/zkPoD-lib/pod_go/table/ot_complaint"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDSellerTOC struct {
	SellerSession *table_ot_complaint.SellerSession `json:"sellerSession"`
}

// sellerNewSessForTOC prepares seller's session while mode is table_ot_complaint.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerTOC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForTOC(publishPath string, Log ILogger) (PoDSellerTOC, error) {
	var toc PoDSellerTOC
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return toc, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return toc, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	toc.SellerSession, err = table_ot_complaint.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return toc, errors.New("failed to create session for seller")
	}
	Log.Debugf("success to create session for seller.")
	return toc, nil
}

// sellerGeneNegoReq generates nego request file for seller while mode is table_ot_complaint.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generate nego request successfully, return true.
//  Otherwise, return false.
func (toc PoDSellerTOC) sellerGeneNegoReq(requestFile string, Log ILogger) bool {

	err := toc.SellerSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to generate nego request. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego request. requestFile=%v", requestFile)
	return true
}

// sellerGeneNegoResp generates nego response file for seller while mode is table_ot_complaint.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If verify nego request and generate nego response successfully, return true.
//  Otherwise, return false.
func (toc PoDSellerTOC) sellerGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := toc.SellerSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// sellerDealNegoResp deals with buyer's nego response for seller while mode is table_ot_complaint.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deal with buyer's nego response successfully, return true.
//  Otherwise, return false.
func (toc PoDSellerTOC) sellerDealNegoResp(responseFile string, Log ILogger) bool {
	err := toc.SellerSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego response")
	return true
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is table_ot_complaint.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (toc PoDSellerTOC) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := toc.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is table_ot_complaint.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (toc PoDSellerTOC) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := toc.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerTOC struct {
	BuyerSession *table_ot_complaint.BuyerSession `json:"buyerSession"`
	Demands      []Demand                         `json:"demands"`
	Phantoms     []Phantom                        `json:"phantoms"`
}

// buyerNewSessForTOC prepares buyer's session while mode is table_ot_complaint.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a buyer's session and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForTOC(demandArr []Demand, phantomArr []Phantom, tableBulletin string, tablePublicPath string, Log ILogger) (PoDBuyerTOC, error) {

	var toc PoDBuyerTOC
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	toc.Demands = demandArr

	phantoms := make([]types.Range, 0)
	for _, p := range phantomArr {
		phantoms = append(phantoms, types.Range{p.PhantomStart, p.PhantomCount})
	}
	toc.Phantoms = phantomArr

	session, err := table_ot_complaint.NewBuyerSession(tableBulletin, tablePublicPath, sellerID, buyerID, demands, phantoms)
	if err != nil {
		Log.Warnf("failed to create session for buyer. err=%v", err)
		return toc, errors.New("failed to create session for buyer")
	}
	toc.BuyerSession = session
	Log.Debugf("success to create session for buyer.")
	return toc, nil
}

// buyerGeneNegoReq generates nego request file for buyer while mode is table_ot_complaint.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generates nego request for buyer successfully, return true.
//  Otherwise, return false.
func (toc PoDBuyerTOC) buyerGeneNegoReq(requestFile string, Log ILogger) bool {

	err := toc.BuyerSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to generate nego request. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego request")
	return true
}

// buyerGeneNegoResp verifies nego request and generates nego response for buyer while mode is table_ot_complaint.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If generates nego response for buyer successfully, return true.
//  Otherwise, return false.
func (toc PoDBuyerTOC) buyerGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := toc.BuyerSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to verify request ans generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request ans generate nego response.")
	return true
}

// buyerDealNegoResp deals with seller's nego response for buyer while mode is table_ot_complaint.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deals with seller's nego response successfully, return true.
//  Otherwise, return false.
func (toc PoDBuyerTOC) buyerDealNegoResp(responseFile string, Log ILogger) bool {
	err := toc.BuyerSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego resoponse")
	return true
}

// buyerNewReq creates request file for buyer while mode is table_ot_complaint.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (toc PoDBuyerTOC) buyerNewReq(requestFile string, Log ILogger) error {
	err := toc.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to create request file. err=%v", err)
		return errors.New("failed to create request file")
	}
	Log.Debugf("success to create request file. requestFile=%v", requestFile)
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is table_ot_complaint.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (toc PoDBuyerTOC) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := toc.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("verify response failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. responseFile=%v receiptFile=%v", responseFile, receiptFile)
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is table_ot_complaint.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (toc PoDBuyerTOC) buyerVerifySecret(secretFile string, Log ILogger) bool {
	err := toc.BuyerSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("verify secret failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret.")
	return true
}

// buyerGeneClaim generates claim with incorrect secret for buyer while mode is table_ot_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If generate claim successfully, return true.
//  Otherwise, return false.
func (toc PoDBuyerTOC) buyerGeneClaim(claimFile string, Log ILogger) bool {
	err := toc.BuyerSession.GenerateClaim(claimFile)
	if err != nil {
		Log.Warnf("generate claim failure. err=%v", err)
		return false
	}
	Log.Debugf("success to generate claim. claimFile=%v", claimFile)
	return true
}

// buyerDecrypt decrypts file for buyer while mode is table_ot_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (toc PoDBuyerTOC) buyerDecrypt(outFile string, Log ILogger) bool {
	err := toc.BuyerSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
