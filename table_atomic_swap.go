package main

import (
	"errors"

	table_atomic_swap "github.com/sec-bit/zkPoD-lib/pod_go/table/atomic_swap"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDSellerTAS struct {
	SellerSession *table_atomic_swap.SellerSession `json:"sellerSession"`
}

// sellerNewSessForTAS prepares seller's session while mode is table_atomic_swap.
//
// It is provides an interface for NewSellerSession.
//
// Return:
//  If no error occurs, return a PoDSellerTAS struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func sellerNewSessForTAS(publishPath string, sellerID [40]uint8, buyerID [40]uint8, Log ILogger) (PoDSellerTAS, error) {
	var tas PoDSellerTAS
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("failed to check. err=%v", err)
		return tas, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return tas, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	tas.SellerSession, err = table_atomic_swap.NewSellerSession(publishPath, sellerID, buyerID)
	if err != nil {
		Log.Warnf("failed to create session for seller. err=%v", err)
		return tas, errors.New("failed to create session for seller")
	}
	Log.Debugf("success to create session")
	return tas, nil
}

// sellerVerifyReq verifies request file and generates response file for seller while mode is table_atomic_swap.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (tas PoDSellerTAS) sellerVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := tas.SellerSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response.")
	return true
}

// sellerVerifyReceipt verifies receipt file and generate secret file for seller while mode is table_atomic_swap.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (tas PoDSellerTAS) sellerVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := tas.SellerSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt receipt and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret file. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBuyerTAS struct {
	BuyerSession *table_atomic_swap.BuyerSession `json:"buyerSession"`
	Demands      []Demand                        `json:"demands"`
}

// buyerNewSessForTAS prepares buyer's session while mode is table_atomic_swap.
//
// It is provides an interface for NewBuyerSession.
//
// Return:
//  If no error occurs, return a PoDBuyerTAS struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func buyerNewSessForTAS(demandArr []Demand, tableBulletin string, tablePublicPath string, sellerID [40]uint8, buyerID [40]uint8, Log ILogger) (PoDBuyerTAS, error) {

	var tas PoDBuyerTAS
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	tas.Demands = demandArr

	session, err := table_atomic_swap.NewBuyerSession(tableBulletin, tablePublicPath, sellerID, buyerID, demands)
	if err != nil {
		Log.Warnf("Failed to create session for buyer. err=%v", err)
		return tas, errors.New("Failed to create session for buyer")
	}
	tas.BuyerSession = session
	Log.Debugf("success to create session")
	return tas, nil
}

// buyerNewReq creates request file for buyer while mode is table_atomic_swap.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (tas PoDBuyerTAS) buyerNewReq(requestFile string, Log ILogger) error {
	err := tas.BuyerSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to generate request for buyer.")
	return nil
}

// buyerVerifyResp verifies response data for buyer while mode is table_atomic_swap.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (tas PoDBuyerTAS) buyerVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := tas.BuyerSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success tov verify response and generate receipt.")
	return true
}

// buyerVerifySecret verifies secret for buyer while mode is table_atomic_swap.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (tas PoDBuyerTAS) buyerVerifySecret(secretFile string, Log ILogger) bool {
	err := tas.BuyerSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret")
	return true
}

// buyerDecrypt decrypts file for buyer while mode is table_atomic_swap.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (tas PoDBuyerTAS) buyerDecrypt(outFile string, Log ILogger) bool {
	err := tas.BuyerSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
