package main

import (
	"errors"

	table_complaint "github.com/sec-bit/zkPoD-lib/pod_go/table/complaint"
	"github.com/sec-bit/zkPoD-lib/pod_go/types"
)

type PoDAliceTC struct {
	AliceSession *table_complaint.AliceSession `json:"AliceSession"`
}

// AliceNewSessForTC prepares Alice's session while mode is table_complaint.
//
// It is provides an interface for NewAliceSession.
//
// Return:
//  If no error occurs, return a Alice's session and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func AliceNewSessForTC(publishPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDAliceTC, error) {
	var tc PoDAliceTC
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("failed to check. err=%v", err)
		return tc, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return tc, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	tc.AliceSession, err = table_complaint.NewAliceSession(publishPath, AliceID, BobID)
	if err != nil {
		Log.Warnf("failed to create session for Alice. err=%v", err)
		return tc, errors.New("failed to create session for Alice")
	}
	Log.Debugf("success to create session for Alice.")
	return tc, nil
}

// AliceVerifyReq verifies request file and generates response file for Alice while mode is table_complaint.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (tc PoDAliceTC) AliceVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := tc.AliceSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// AliceVerifyReceipt verifies receipt file and generate secret file for Alice while mode is table_complaint.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (tc PoDAliceTC) AliceVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := tc.AliceSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBobTC struct {
	BobSession *table_complaint.BobSession `json:"BobSession"`
	Demands      []Demand                      `json:"demands"`
}

// BobNewSessForTC prepares Bob's session while mode is table_complaint.
//
// It is provides an interface for NewBobSession.
//
// Return:
//  If no error occurs, return a PoDBob struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func BobNewSessForTC(demandArr []Demand, tableBulletin string, tablePublicPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDBobTC, error) {

	var tc PoDBobTC
	demands := make([]types.Range, 0)
	for _, d := range demandArr {
		demands = append(demands, types.Range{d.DemandStart, d.DemandCount})
	}
	tc.Demands = demandArr

	session, err := table_complaint.NewBobSession(tableBulletin, tablePublicPath, AliceID, BobID, demands)
	if err != nil {
		Log.Warnf("Failed to create session for Bob. err=%v", err)
		return tc, errors.New("Failed to create session for Bob")
	}
	tc.BobSession = session
	Log.Debugf("success to create session for Bob.")
	return tc, nil
}

// BobNewReq creates request file for Bob while mode is table_complaint.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (tc PoDBobTC) BobNewReq(requestFile string, Log ILogger) error {
	err := tc.BobSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request file for Bob")
	return nil
}

// BobVerifyResp verifies response data for Bob while mode is table_complaint.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (tc PoDBobTC) BobVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := tc.BobSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("Verify response failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt. responseFile=%v， receiptFile=%v", responseFile, receiptFile)
	return true
}

// BobVerifySecret verifies secret for Bob while mode is table_complaint.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (tc PoDBobTC) BobVerifySecret(secretFile string, Log ILogger) bool {
	err := tc.BobSession.OnSecret(secretFile)
	if err != nil {
		Log.Warnf("Verify secret failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret. secretFile=%v", secretFile)
	return true
}

// BobGeneClaim generates claim with incorrect secret for Bob while mode is table_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If generate claim successfully, return true.
//  Otherwise, return false.
func (tc PoDBobTC) BobGeneClaim(claimFile string, Log ILogger) bool {
	err := tc.BobSession.GenerateClaim(claimFile)
	if err != nil {
		Log.Warnf("generate claim failure. err=%v", err)
		return false
	}
	Log.Debugf("success to generate claim. claimFile=%v", claimFile)
	return true
}

// BobDecrypt decrypts file for Bob while mode is table_complaint.
//
// It is provides an interface for GenerateClaim.
//
// Return:
//  If decrypt file successfully, return true.
//  Otherwise, return false.
func (tc PoDBobTC) BobDecrypt(outFile string, Log ILogger) bool {
	err := tc.BobSession.Decrypt(outFile)
	if err != nil {
		Log.Warnf("Failed to decrypt file. err=%v", err)
		return false
	}
	Log.Debugf("success to decrypt file. outFile=%v", outFile)
	return true
}
