package main

import (
	"errors"

	plain_ot_complaint "github.com/sec-bit/zkPoD-lib/pod_go/plain/ot_complaint"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDAlicePOC struct {
	AliceSession *plain_ot_complaint.AliceSession `json:"AliceSession"`
}

// AliceNewSessForPOC prepares Alice's session while mode is plain_ot_complaint.
//
// It is provides an interface for NewAliceSession.
//
// Return:
//  If no error occurs, return a PoDAlicePOC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func AliceNewSessForPOC(publishPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDAlicePOC, error) {
	var poc PoDAlicePOC
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

	poc.AliceSession, err = plain_ot_complaint.NewAliceSession(publishPath, AliceID, BobID)
	if err != nil {
		Log.Warnf("failed to create session for Alice. err=%v", err)
		return poc, errors.New("failed to create session for Alice")
	}
	return poc, nil
}

// AliceGeneNegoReq generates nego request file for Alice while mode is plain_ot_complaint.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generate nego request successfully, return true.
//  Otherwise, return false.
func (poc PoDAlicePOC) AliceGeneNegoReq(requestFile string, Log ILogger) bool {

	err := poc.AliceSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to generate nego request for Alice. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego request for Alice. requestFile=%v", requestFile)
	return true
}

// AliceGeneNegoResp generates nego response file for Alice while mode is plain_ot_complaint.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If verify nego request and generate nego response successfully, return true.
//  Otherwise, return false.
func (poc PoDAlicePOC) AliceGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := poc.AliceSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("failed to verify nego request and generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify nego request and generate nego response. requestFile=%v, responseFile=%v", requestFile, responseFile)
	return true
}

// AliceDealNegoResp deals with Bob's nego response for Alice while mode is plain_ot_complaint.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deal with Bob's nego response successfully, return true.
//  Otherwise, return false.
func (poc PoDAlicePOC) AliceDealNegoResp(responseFile string, Log ILogger) bool {
	err := poc.AliceSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("failed to deal with Bob's nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to deal with Bob's nego response. responseFile=%v", responseFile)
	return true
}

// AliceVerifyReq verifies request file and generates response file for Alice while mode is plain_ot_complaint.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (poc PoDAlicePOC) AliceVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := poc.AliceSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response. requestFile=%v, responseFile=%v", requestFile, responseFile)
	return true
}

// AliceVerifyReceipt verifies receipt file and generate secret file for Alice while mode is plain_ot_complaint.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (poc PoDAlicePOC) AliceVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := poc.AliceSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret file. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBobPOC struct {
	BobSession *plain_ot_complaint.BobSession `json:"BobSession"`
	Demands      []Demand                         `json:"demands"`
	Phantoms     []Phantom                        `json:"phantoms"`
}

// BobNewSessForPOC prepares Bob's session while mode is plain_ot_complaint.
//
// It is provides an interface for NewBobSession.
//
// Return:
//  If no error occurs, return a PoDBobPOC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func BobNewSessForPOC(demandArr []Demand, phantomArr []Phantom, plainBulletin string, plainPublicPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDBobPOC, error) {

	var poc PoDBobPOC
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

	session, err := plain_ot_complaint.NewBobSession(plainBulletin, plainPublicPath, AliceID, BobID, demands, phantoms)
	if err != nil {
		Log.Warnf("Failed to create session for Bob. err=%v", err)
		return poc, errors.New("Failed to create session for Bob")
	}
	poc.BobSession = session
	Log.Debugf("success to create session for Alice.")
	return poc, nil
}

// BobGeneNegoReq generates nego request file for Bob while mode is plain_ot_complaint.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generates nego request for Bob successfully, return true.
//  Otherwise, return false.
func (poc PoDBobPOC) BobGeneNegoReq(requestFile string, Log ILogger) bool {

	err := poc.BobSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to generates nego request for Bob. err=%v", err)
		return false
	}
	Log.Debugf("success to generates nego request for Bob. requestFile=%v", requestFile)
	return true
}

// BobGeneNegoResp verifies nego request and generates nego response for Bob while mode is plain_ot_complaint.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If generates nego response for Bob successfully, return true.
//  Otherwise, return false.
func (poc PoDBobPOC) BobGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := poc.BobSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to verify nego request and generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify nego request and generate nego response. requestFile=%v, responseFile=%v", requestFile, responseFile)
	return true
}

// BobDealNegoResp deals with Alice's nego response for Bob while mode is plain_ot_complaint.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deals with Alice's nego response successfully, return true.
//  Otherwise, return false.
func (poc PoDBobPOC) BobDealNegoResp(responseFile string, Log ILogger) bool {
	err := poc.BobSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to deals with Alice's nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to deal with Alice's nego response. responseFile=%v", responseFile)
	return true
}

// BobNewReq creates request file for Bob while mode is plain_ot_complaint.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (poc PoDBobPOC) BobNewReq(requestFile string, Log ILogger) error {
	err := poc.BobSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request. requestFile=%v", requestFile)
	return nil
}

// BobVerifyResp verifies response data for Bob while mode is plain_ot_complaint.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (poc PoDBobPOC) BobVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := poc.BobSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. responseFile=%v, receiptFile=%v", responseFile, receiptFile)
	return true
}

// BobVerifySecret verifies secret for Bob while mode is plain_ot_complaint.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (poc PoDBobPOC) BobVerifySecret(secretFile string, Log ILogger) bool {
	err := poc.BobSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret. secretFile=%v", secretFile)
	return true
}

// BobGeneClaim generates claim with incorrect secret for Bob while mode is plain_ot_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If generate claim successfully, return true.
//  Otherwise, return false.
func (poc PoDBobPOC) BobGeneClaim(claimFile string, Log ILogger) bool {
	err := poc.BobSession.GenerateClaim(claimFile)
	if err != nil {
		Log.Warnf("failed to generate claim. err=%v", err)
		return false
	}
	Log.Debugf("success to generate claim. claimFile=%v", claimFile)
	return true
}

//BobDecrypt decrypts file for Bob while mode is plain_ot_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (poc PoDBobPOC) BobDecrypt(outFile string, Log ILogger) bool {
	err := poc.BobSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
