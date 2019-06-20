package main

import (
	"errors"

	plain_atomic_swap "github.com/sec-bit/zkPoD-lib/pod_go/plain/atomic_swap"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDAlicePAS struct {
	AliceSession *plain_atomic_swap.AliceSession `json:"AliceSession"`
}

// AliceNewSessForPAS prepares Alice's session while mode is plain_atomic_swap.
//
// It is provides an interface for NewAliceSession.
//
// Return:
//  If no error occurs, return a PoDAlicePAS struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func AliceNewSessForPAS(publishPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDAlicePAS, error) {
	var pas PoDAlicePAS
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return pas, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return pas, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	pas.AliceSession, err = plain_atomic_swap.NewAliceSession(publishPath, AliceID, BobID)
	if err != nil {
		Log.Warnf("failed to create session for Alice. err=%v", err)
		return pas, errors.New("failed to create session for Alice")
	}
	Log.Debugf("success to create session for Alice")
	return pas, nil
}

// AliceVerifyReq verifies request file and generates response file for Alice while mode is plain_atomic_swap.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (pas PoDAlicePAS) AliceVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := pas.AliceSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// AliceVerifyReceipt verifies receipt file and generate secret file for Alice while mode is plain_atomic_swap.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (pas PoDAlicePAS) AliceVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := pas.AliceSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt file and generate secret file. receipt file=%v, secret file=%v", receiptFile, secretFile)
	return true
}

type PoDBobPAS struct {
	BobSession *plain_atomic_swap.BobSession `json:"BobSession"`
	Demands      []Demand                        `json:"demands"`
}

// BobNewSessForPAS prepares Bob's session while mode is plain_atomic_swap.
//
// It is provides an interface for NewBobSession.
//
// Return:
//  If no error occurs, return a Bob's session and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func BobNewSessForPAS(demandArr []Demand, plainBulletin string, plainPublicPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDBobPAS, error) {

	var pas PoDBobPAS
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	pas.Demands = demandArr

	session, err := plain_atomic_swap.NewBobSession(plainBulletin, plainPublicPath, AliceID, BobID, demands)
	if err != nil {
		Log.Warnf("Failed to create session for Bob. err=%v", err)
		return pas, errors.New("Failed to create session for Bob")
	}
	pas.BobSession = session
	Log.Debugf("success to create session for Bob")
	return pas, nil
}

// BobNewReq creates request file for Bob while mode is plain_atomic_swap.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (pas PoDBobPAS) BobNewReq(requestFile string, Log ILogger) error {
	err := pas.BobSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request file. reqeuestFile=%v", requestFile)
	return nil
}

// BobVerifyResp verifies response data for Bob while mode is plain_atomic_swap.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (pas PoDBobPAS) BobVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := pas.BobSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. receiptFile=%v", receiptFile)
	return true
}

// BobVerifySecret verifies secret for Bob while mode is plain_atomic_swap.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (pas PoDBobPAS) BobVerifySecret(secretFile string, Log ILogger) bool {
	err := pas.BobSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret. secretFile=%v", secretFile)
	return true
}

// BobDecrypt decrypts file for Bob while mode is plain_atomic_swap.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (pas PoDBobPAS) BobDecrypt(outFile string, Log ILogger) bool {
	err := pas.BobSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
