package main

import (
	"errors"

	table_ot_complaint "github.com/sec-bit/zkPoD-lib/pod_go/table/ot_complaint"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDAliceTOC struct {
	AliceSession *table_ot_complaint.AliceSession `json:"AliceSession"`
}

// AliceNewSessForTOC prepares Alice's session while mode is table_ot_complaint.
//
// It is provides an interface for NewAliceSession.
//
// Return:
//  If no error occurs, return a PoDAliceTOC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func AliceNewSessForTOC(publishPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDAliceTOC, error) {
	var toc PoDAliceTOC
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

	toc.AliceSession, err = table_ot_complaint.NewAliceSession(publishPath, AliceID, BobID)
	if err != nil {
		Log.Warnf("failed to create session for Alice. err=%v", err)
		return toc, errors.New("failed to create session for Alice")
	}
	Log.Debugf("success to create session for Alice.")
	return toc, nil
}

// AliceGeneNegoReq generates nego request file for Alice while mode is table_ot_complaint.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generate nego request successfully, return true.
//  Otherwise, return false.
func (toc PoDAliceTOC) AliceGeneNegoReq(requestFile string, Log ILogger) bool {

	err := toc.AliceSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to generate nego request. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego request. requestFile=%v", requestFile)
	return true
}

// AliceGeneNegoResp generates nego response file for Alice while mode is table_ot_complaint.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If verify nego request and generate nego response successfully, return true.
//  Otherwise, return false.
func (toc PoDAliceTOC) AliceGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := toc.AliceSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// AliceDealNegoResp deals with Bob's nego response for Alice while mode is table_ot_complaint.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deal with Bob's nego response successfully, return true.
//  Otherwise, return false.
func (toc PoDAliceTOC) AliceDealNegoResp(responseFile string, Log ILogger) bool {
	err := toc.AliceSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego response")
	return true
}

// AliceVerifyReq verifies request file and generates response file for Alice while mode is table_ot_complaint.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (toc PoDAliceTOC) AliceVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := toc.AliceSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// AliceVerifyReceipt verifies receipt file and generate secret file for Alice while mode is table_ot_complaint.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (toc PoDAliceTOC) AliceVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := toc.AliceSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBobTOC struct {
	BobSession *table_ot_complaint.BobSession `json:"BobSession"`
	Demands      []Demand                         `json:"demands"`
	Phantoms     []Phantom                        `json:"phantoms"`
}

// BobNewSessForTOC prepares Bob's session while mode is table_ot_complaint.
//
// It is provides an interface for NewBobSession.
//
// Return:
//  If no error occurs, return a Bob's session and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func BobNewSessForTOC(demandArr []Demand, phantomArr []Phantom, tableBulletin string, tablePublicPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDBobTOC, error) {

	var toc PoDBobTOC
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

	session, err := table_ot_complaint.NewBobSession(tableBulletin, tablePublicPath, AliceID, BobID, demands, phantoms)
	if err != nil {
		Log.Warnf("failed to create session for Bob. err=%v", err)
		return toc, errors.New("failed to create session for Bob")
	}
	toc.BobSession = session
	Log.Debugf("success to create session for Bob.")
	return toc, nil
}

// BobGeneNegoReq generates nego request file for Bob while mode is table_ot_complaint.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generates nego request for Bob successfully, return true.
//  Otherwise, return false.
func (toc PoDBobTOC) BobGeneNegoReq(requestFile string, Log ILogger) bool {

	err := toc.BobSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to generate nego request. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego request")
	return true
}

// BobGeneNegoResp verifies nego request and generates nego response for Bob while mode is table_ot_complaint.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If generates nego response for Bob successfully, return true.
//  Otherwise, return false.
func (toc PoDBobTOC) BobGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := toc.BobSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to verify request ans generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request ans generate nego response.")
	return true
}

// BobDealNegoResp deals with Alice's nego response for Bob while mode is table_ot_complaint.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deals with Alice's nego response successfully, return true.
//  Otherwise, return false.
func (toc PoDBobTOC) BobDealNegoResp(responseFile string, Log ILogger) bool {
	err := toc.BobSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego resoponse")
	return true
}

// BobNewReq creates request file for Bob while mode is table_ot_complaint.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (toc PoDBobTOC) BobNewReq(requestFile string, Log ILogger) error {
	err := toc.BobSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to create request file. err=%v", err)
		return errors.New("failed to create request file")
	}
	Log.Debugf("success to create request file. requestFile=%v", requestFile)
	return nil
}

// BobVerifyResp verifies response data for Bob while mode is table_ot_complaint.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (toc PoDBobTOC) BobVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := toc.BobSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("verify response failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. responseFile=%v receiptFile=%v", responseFile, receiptFile)
	return true
}

// BobVerifySecret verifies secret for Bob while mode is table_ot_complaint.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (toc PoDBobTOC) BobVerifySecret(secretFile string, Log ILogger) bool {
	err := toc.BobSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("verify secret failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret.")
	return true
}

// BobGeneClaim generates claim with incorrect secret for Bob while mode is table_ot_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If generate claim successfully, return true.
//  Otherwise, return false.
func (toc PoDBobTOC) BobGeneClaim(claimFile string, Log ILogger) bool {
	err := toc.BobSession.GenerateClaim(claimFile)
	if err != nil {
		Log.Warnf("generate claim failure. err=%v", err)
		return false
	}
	Log.Debugf("success to generate claim. claimFile=%v", claimFile)
	return true
}

// BobDecrypt decrypts file for Bob while mode is table_ot_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (toc PoDBobTOC) BobDecrypt(outFile string, Log ILogger) bool {
	err := toc.BobSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
