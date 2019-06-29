package main

import (
	"errors"

	plain_atomic_swap_vc "github.com/sec-bit/zkPoD-lib/pod_go/plain/atomic_swap_vc"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDAlicePASVC struct {
	AliceSession *plain_atomic_swap_vc.AliceSession `json:"AliceSession"`
}

// AliceNewSessForPASVC prepares Alice's session while mode is plain_atomic_swap_vc.
//
// It is provides an interface for NewAliceSession.
//
// Return:
//  If no error occurs, return a PoDAlicePASVC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func AliceNewSessForPASVC(publishPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDAlicePASVC, error) {
	var pasvc PoDAlicePASVC
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return pasvc, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return pasvc, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	pasvc.AliceSession, err = plain_atomic_swap_vc.NewAliceSession(publishPath, AliceID, BobID)
	if err != nil {
		Log.Warnf("failed to create session for Alice. err=%v", err)
		return pasvc, errors.New("failed to create session for Alice")
	}
	Log.Debugf("success to create session for Alice")
	return pasvc, nil
}

// AliceVerifyReq verifies request file and generates response file for Alice while mode is plain_atomic_swap_vc.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (pasvc PoDAlicePASVC) AliceVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := pasvc.AliceSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// AliceVerifyReceipt verifies receipt file and generate secret file for Alice while mode is plain_atomic_swap_vc.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (pasvc PoDAlicePASVC) AliceVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := pasvc.AliceSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt file and generate secret file. receipt file=%v, secret file=%v", receiptFile, secretFile)
	return true
}

type PoDBobPASVC struct {
	BobSession *plain_atomic_swap_vc.BobSession `json:"BobSession"`
	Demands    []Demand                         `json:"demands"`
}

// BobNewSessForPASVC prepares Bob's session while mode is plain_atomic_swap_vc.
//
// It is provides an interface for NewBobSession.
//
// Return:
//  If no error occurs, return a Bob's session and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func BobNewSessForPASVC(demandArr []Demand, plainBulletin string, plainPublicPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDBobPASVC, error) {

	var pasvc PoDBobPASVC
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	pasvc.Demands = demandArr

	session, err := plain_atomic_swap_vc.NewBobSession(plainBulletin, plainPublicPath, AliceID, BobID, demands)
	if err != nil {
		Log.Warnf("failed to create session for Bob. err=%v", err)
		return pasvc, errors.New("failed to create session for Bob")
	}
	pasvc.BobSession = session
	Log.Debugf("success to create session for Bob")
	return pasvc, nil
}

// BobNewReq creates request file for Bob while mode is plain_atomic_swap_vc.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (pasvc PoDBobPASVC) BobNewReq(requestFile string, Log ILogger) error {
	err := pasvc.BobSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request file. reqeuestFile=%v", requestFile)
	return nil
}

// BobVerifyResp verifies response data for Bob while mode is plain_atomic_swap_vc.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (pasvc PoDBobPASVC) BobVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := pasvc.BobSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. receiptFile=%v", receiptFile)
	return true
}

// BobVerifySecret verifies secret for Bob while mode is plain_atomic_swap_vc.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (pasvc PoDBobPASVC) BobVerifySecret(secretFile string, Log ILogger) bool {
	err := pasvc.BobSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret. secretFile=%v", secretFile)
	return true
}

// BobDecrypt decrypts file for Bob while mode is plain_atomic_swap_vc.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (pasvc PoDBobPASVC) BobDecrypt(outFile string, Log ILogger) bool {
	err := pasvc.BobSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
