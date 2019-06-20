package main

import (
	"errors"

	plain_complaint "github.com/sec-bit/zkPoD-lib/pod_go/plain/complaint"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDAlicePC struct {
	AliceSession *plain_complaint.AliceSession `json:"AliceSession"`
}

// AliceNewSessForPC prepares Alice's session while mode is plain_complaint.
//
// It is provides an interface for NewAliceSession.
//
// Return:
//  If no error occurs, return a PoDAlicePC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func AliceNewSessForPC(publishPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDAlicePC, error) {
	var pc PoDAlicePC
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

	pc.AliceSession, err = plain_complaint.NewAliceSession(publishPath, AliceID, BobID)
	if err != nil {
		Log.Warnf("failed to create session for Alice. err=%v", err)
		return pc, errors.New("failed to create session for Alice")
	}
	Log.Debugf("success to create session for Alice")
	return pc, nil
}

// AliceVerifyReq verifies request file and generates response file for Alice while mode is plain_complaint.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (pc PoDAlicePC) AliceVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := pc.AliceSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify transaction requset and generate transaction response")
	return true
}

// AliceVerifyReceipt verifies receipt file and generate secret file for Alice while mode is plain_complaint.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (pc PoDAlicePC) AliceVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := pc.AliceSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt file and generate secret file. receipt file=%v, secret file=%v", receiptFile, secretFile)
	return true
}

type PoDBobPC struct {
	BobSession *plain_complaint.BobSession `json:"BobSession"`
	Demands      []Demand                      `json:"demands"`
}

// BobNewSessForPC prepares Bob's session while mode is plain_complaint.
//
// It is provides an interface for NewBobSession.
//
// Return:
//  If no error occurs, return a PoDBobPC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func BobNewSessForPC(demandArr []Demand, plainBulletin string, plainPublicPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDBobPC, error) {

	var pc PoDBobPC
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	pc.Demands = demandArr

	session, err := plain_complaint.NewBobSession(plainBulletin, plainPublicPath, AliceID, BobID, demands)
	if err != nil {
		Log.Warnf("failed to create session for Bob. err=%v", err)
		return pc, errors.New("failed to create session for Bob")
	}
	pc.BobSession = session
	Log.Debugf("success to create session for Bob")
	return pc, nil
}

// BobNewReq creates request file for Bob while mode is plain_complaint.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (pc PoDBobPC) BobNewReq(requestFile string, Log ILogger) error {
	err := pc.BobSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request file. filepath=%v", requestFile)
	return nil
}

// BobVerifyResp verifies response data for Bob while mode is plain_complaint.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (pc PoDBobPC) BobVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := pc.BobSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. responseFile=%v, receiptFile=%v", responseFile, receiptFile)
	return true
}

// BobVerifySecret verifies secret for Bob while mode is plain_complaint.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (pc PoDBobPC) BobVerifySecret(secretFile string, Log ILogger) bool {
	err := pc.BobSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret")
	return true
}

// BobGeneClaim generates claim with incorrect secret for Bob while mode is plain_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If generate claim successfully, return true.
//  Otherwise, return false.
func (pc PoDBobPC) BobGeneClaim(claimFile string, Log ILogger) bool {
	err := pc.BobSession.GenerateClaim(claimFile)
	if err != nil {
		Log.Warnf("generate claim failure. err=%v", err)
		return false
	}
	Log.Debugf("success to generate claim, claimFile=%v", claimFile)
	return true
}

// BobDecrypt decrypts file for Bob while mode is plain_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (pc PoDBobPC) BobDecrypt(outFile string, Log ILogger) bool {
	err := pc.BobSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt claim. outFile=%v", outFile)
	return true
}
