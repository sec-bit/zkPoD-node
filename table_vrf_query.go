package main

import (
	"errors"

	tablevrf "github.com/sec-bit/zkPoD-lib/pod_go/table/vrfq"
)

type PoDAliceTQ struct {
	AliceSession *tablevrf.AliceSession `json:"AliceSession"`
}

// AliceNewSessForTQ prepares Alice's session while mode is table_vrf_query.
//
// It is provides an interface for NewAliceSession.
//
// Return:
//  If no error occurs, return a PoDAliceTQ struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func AliceNewSessForTQ(publishPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDAliceTQ, error) {
	var tq PoDAliceTQ
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("Failed to check. err=%v", err)
		return tq, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return tq, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	tq.AliceSession, err = tablevrf.NewAliceSession(publishPath, AliceID, BobID)
	if err != nil {
		Log.Warnf("failed to create session for Alice. err=%v", err)
		return tq, errors.New("failed to create session for Alice")
	}
	Log.Debugf("success to create session")
	return tq, nil
}

// AliceVerifyReq verifies request file and generates response file for Alice while mode is table_vrf_query.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (tq PoDAliceTQ) AliceVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := tq.AliceSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// AliceVerifyReceipt verifies receipt file and generate secret file for Alice while mode is table_vrf_query.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (tq PoDAliceTQ) AliceVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := tq.AliceSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBobTQ struct {
	BobSession *tablevrf.BobSession `json:"BobSession"`
	KeyName    string               `json:"keyName"`
	KeyValue   []string             `json:"keyValue"`
}

// BobNewSessForTQ prepares Bob's session while mode is table_vrf_query.
//
// It is provides an interface for NewBobSession.
//
// Return:
//  If no error occurs, return a Bob's session and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func BobNewSessForTQ(keyName string, keyValue []string, tableBulletin string, tablePublicPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDBobTQ, error) {

	var tq PoDBobTQ
	Log.Debugf("tableBulletin=%v, tablePublicPath=%v, AliceID=%v, BobID=%v, keyName=%v,keyValue=%v",
		tableBulletin, tablePublicPath, AliceID, BobID, keyName, keyValue)
	session, err := tablevrf.NewBobSession(tableBulletin, tablePublicPath, AliceID, BobID, keyName, keyValue)
	if err != nil {
		Log.Warnf("Failed to create session for Bob. err=%v", err)
		return tq, errors.New("Failed to create session for Bob")
	}
	Log.Debugf("success to create session.")
	tq = PoDBobTQ{session, keyName, keyValue}
	return tq, nil
}

// BobNewReq creates request file for Bob while mode is table_vrf_query.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (tq PoDBobTQ) BobNewReq(requestFile string, Log ILogger) error {
	err := tq.BobSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request file. requestFile=%v", requestFile)
	return nil
}

// BobVerifyResp verifies response data for Bob while mode is table_vrf_query.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (tq PoDBobTQ) BobVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := tq.BobSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("Verify response failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and geerate receipt. responseFile=%v, receiptFile=%v", responseFile, receiptFile)
	return true
}

// BobVerifySecret verifies secret for Bob while mode is table_vrf_query.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (tq PoDBobTQ) BobVerifySecret(secretFile string, tablePosition string, Log ILogger) bool {
	err := tq.BobSession.OnSecret(secretFile, tablePosition)
	if err != nil {
		Log.Warnf("Verify secret failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret and generate table position. secretFile=%v, tablePosition=%v", secretFile, tablePosition)
	return true
}
