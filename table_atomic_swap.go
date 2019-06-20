package main

import (
	"errors"

	table_atomic_swap "github.com/sec-bit/zkPoD-lib/pod_go/table/atomic_swap"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDAliceTAS struct {
	AliceSession *table_atomic_swap.AliceSession `json:"AliceSession"`
}

// AliceNewSessForTAS prepares Alice's session while mode is table_atomic_swap.
//
// It is provides an interface for NewAliceSession.
//
// Return:
//  If no error occurs, return a PoDAliceTAS struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func AliceNewSessForTAS(publishPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDAliceTAS, error) {
	var tas PoDAliceTAS
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

	tas.AliceSession, err = table_atomic_swap.NewAliceSession(publishPath, AliceID, BobID)
	if err != nil {
		Log.Warnf("failed to create session for Alice. err=%v", err)
		return tas, errors.New("failed to create session for Alice")
	}
	Log.Debugf("success to create session")
	return tas, nil
}

// AliceVerifyReq verifies request file and generates response file for Alice while mode is table_atomic_swap.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (tas PoDAliceTAS) AliceVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := tas.AliceSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response.")
	return true
}

// AliceVerifyReceipt verifies receipt file and generate secret file for Alice while mode is table_atomic_swap.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (tas PoDAliceTAS) AliceVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := tas.AliceSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt receipt and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret file. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBobTAS struct {
	BobSession *table_atomic_swap.BobSession `json:"BobSession"`
	Demands      []Demand                        `json:"demands"`
}

// BobNewSessForTAS prepares Bob's session while mode is table_atomic_swap.
//
// It is provides an interface for NewBobSession.
//
// Return:
//  If no error occurs, return a PoDBobTAS struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func BobNewSessForTAS(demandArr []Demand, tableBulletin string, tablePublicPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDBobTAS, error) {

	var tas PoDBobTAS
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	tas.Demands = demandArr

	session, err := table_atomic_swap.NewBobSession(tableBulletin, tablePublicPath, AliceID, BobID, demands)
	if err != nil {
		Log.Warnf("Failed to create session for Bob. err=%v", err)
		return tas, errors.New("Failed to create session for Bob")
	}
	tas.BobSession = session
	Log.Debugf("success to create session")
	return tas, nil
}

// BobNewReq creates request file for Bob while mode is table_atomic_swap.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (tas PoDBobTAS) BobNewReq(requestFile string, Log ILogger) error {
	err := tas.BobSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to generate request for Bob.")
	return nil
}

// BobVerifyResp verifies response data for Bob while mode is table_atomic_swap.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (tas PoDBobTAS) BobVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := tas.BobSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("failed to verify response and generate receipt. err=%v", err)
		return false
	}
	Log.Debugf("success tov verify response and generate receipt.")
	return true
}

// BobVerifySecret verifies secret for Bob while mode is table_atomic_swap.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (tas PoDBobTAS) BobVerifySecret(secretFile string, Log ILogger) bool {
	err := tas.BobSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("failed to verify secret. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret")
	return true
}

// BobDecrypt decrypts file for Bob while mode is table_atomic_swap.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (tas PoDBobTAS) BobDecrypt(outFile string, Log ILogger) bool {
	err := tas.BobSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
