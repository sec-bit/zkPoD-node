package main

import (
	"errors"

	table_atomic_swap_vc "github.com/sec-bit/zkPoD-lib/pod_go/table/atomic_swap_vc"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDAliceTASVC struct {
	AliceSession *table_atomic_swap_vc.AliceSession `json:"AliceSession"`
}

// AliceNewSessForTASVC prepares Alice's session while mode is table_atomic_swap_vc.
//
// It is provides an interface for NewAliceSession.
//
// Return:
//  If no error occurs, return a PoDAliceTASVC struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func AliceNewSessForTASVC(publishPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDAliceTASVC, error) {
	var tasvc PoDAliceTASVC
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("failed to check. err=%v", err)
		return tasvc, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return tasvc, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	tasvc.AliceSession, err = table_atomic_swap_vc.NewAliceSession(publishPath, AliceID, BobID)
	if err != nil {
		Log.Warnf("failed to create session for Alice. err=%v", err)
		return tasvc, errors.New("failed to create session for Alice")
	}
	Log.Debugf("success to create session")
	return tasvc, nil
}

// AliceVerifyReq verifies request file and generates response file for Alice while mode is table_atomic_swap_vc.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (tasvc PoDAliceTASVC) AliceVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := tasvc.AliceSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response.")
	return true
}

// AliceVerifyReceipt verifies receipt file and generate secret file for Alice while mode is table_atomic_swap_vc.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (tasvc PoDAliceTASVC) AliceVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := tasvc.AliceSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt receipt and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret file. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBobTASVC struct {
	BobSession *table_atomic_swap_vc.BobSession `json:"BobSession"`
	Demands    []Demand                         `json:"demands"`
}

// BobNewSessForTASVC prepares Bob's session while mode is table_atomic_swap_vc.
//
// It is provides an interface for NewBobSession.
//
// Return:
//  If no error occurs, return a PoDBobTAS struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func BobNewSessForTASVC(demandArr []Demand, tableBulletin string, tablePublicPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDBobTASVC, error) {

	var tasvc PoDBobTASVC
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	tasvc.Demands = demandArr

	session, err := table_atomic_swap_vc.NewBobSession(tableBulletin, tablePublicPath, AliceID, BobID, demands)
	if err != nil {
		Log.Warnf("Failed to create session for Bob. err=%v", err)
		return tasvc, errors.New("Failed to create session for Bob")
	}
	tasvc.BobSession = session
	Log.Debugf("success to create session")
	return tasvc, nil
}

// BobNewReq creates request file for Bob while mode is table_atomic_swap_vc.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (tasvc PoDBobTASVC) BobNewReq(requestFile string, Log ILogger) error {
	err := tasvc.BobSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to generate request for Bob.")
	return nil
}

// BobVerifyResp verifies response data for Bob while mode is table_atomic_swap_vc.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (tasvc PoDBobTASVC) BobVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := tasvc.BobSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success tov verify response and generate receipt.")
	return true
}

// BobVerifySecret verifies secret for Bob while mode is table_atomic_swap_vc.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (tasvc PoDBobTASVC) BobVerifySecret(secretFile string, Log ILogger) bool {
	err := tasvc.BobSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret")
	return true
}

// BobDecrypt decrypts file for Bob while mode is table_atomic_swap_vc.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (tasvc PoDBobTASVC) BobDecrypt(outFile string, Log ILogger) bool {
	err := tasvc.BobSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
