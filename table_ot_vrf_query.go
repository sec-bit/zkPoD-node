package main

import (
	"errors"

	table_ot_vrf "github.com/sec-bit/zkPoD-lib/pod_go/table/ot_vrfq"
)

type PoDAliceTOQ struct {
	AliceSession *table_ot_vrf.AliceSession `json:"AliceSession"`
}

// AliceNewSessForTOQ prepares Alice's session while mode is table_ot_vrf_query.
//
// It is provides an interface for NewAliceSession.
//
// Return:
//  If no error occurs, return a PoDAliceTOQ struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func AliceNewSessForTOQ(publishPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDAliceTOQ, error) {
	var toq PoDAliceTOQ
	rs, err := pathExists(publishPath)
	if err != nil {
		Log.Warnf("failed to check. err=%v", err)
		return toq, err
	}
	if !rs {
		Log.Warnf("the path=%v does not exist.", publishPath)
		return toq, errors.New("the path does not exist")
	}
	Log.Debugf("publishPath=%v", publishPath)

	toq.AliceSession, err = table_ot_vrf.NewAliceSession(publishPath, AliceID, BobID)
	if err != nil {
		Log.Warnf("failed to create session for Alice. err=%v", err)
		return toq, errors.New("failed to create session for Alice")
	}
	Log.Debugf("success to create session")
	return toq, nil
}

// AliceGeneNegoReq generates nego request file for Alice while mode is table_ot_vrf_query.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generate nego request successfully, return true.
//  Otherwise, return false.
func (toq PoDAliceTOQ) AliceGeneNegoReq(requestFile string, Log ILogger) bool {

	err := toq.AliceSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to generate nego request. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego request. requestFile=%v", requestFile)
	return true
}

// AliceGeneNegoResp generates nego response file for Alice while mode is table_ot_vrf_query.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If verify nego request and generate nego response successfully, return true.
//  Otherwise, return false.
func (toq PoDAliceTOQ) AliceGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := toq.AliceSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to verify nego request and generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to verify nego request and generate nego response")
	return true
}

// AliceDealNegoResp deals with Bob's nego response for Alice while mode is table_ot_vrf_query.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deal with Bob's nego response successfully, return true.
//  Otherwise, return false.
func (toq PoDAliceTOQ) AliceDealNegoResp(responseFile string, Log ILogger) bool {
	err := toq.AliceSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to deal with nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to deal with nego response")
	return true
}

// AliceVerifyReq verifies request file and generates response file for Alice while mode is table_ot_vrf_query.
//
// It is provides an interface for OnRequest.
//
// Return:
//  If verify transaction requset and generate transaction response successfully, return true.
//  Otherwise, return false.
func (toq PoDAliceTOQ) AliceVerifyReq(requestFile string, responseFile string, Log ILogger) bool {

	err := toq.AliceSession.OnRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Verify request and generate response....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify request and generate response")
	return true
}

// AliceVerifyReceipt verifies receipt file and generate secret file for Alice while mode is table_ot_vrf_query.
//
// It is provides an interface for OnReceipt.
//
// Return:
//  If verify receipt file and generate secret file successfully, return true.
//  Otherwise, return false.
func (toq PoDAliceTOQ) AliceVerifyReceipt(receiptFile string, secretFile string, Log ILogger) bool {

	err := toq.AliceSession.OnReceipt(receiptFile, secretFile)
	if err != nil {
		Log.Warnf("Verify receipt file and generate secret file.....Failed. err=%v", err)
		return false
	}
	Log.Debugf("success to verify receipt and generate secret. receiptFile=%v, secretFile=%v", receiptFile, secretFile)
	return true
}

type PoDBobTOQ struct {
	BobSession    *table_ot_vrf.BobSession `json:"BobSession"`
	KeyName         string                     `json:"keyName"`
	KeyValue        []string                   `json:"keyValue"`
	PhantomKeyValue []string                   `json:"phantomKeyValue"`
}

// BobNewSessForTOQ prepares Bob's session while mode is table_ot_vrf_query.
//
// It is provides an interface for NewBobSession.
//
// Return:
//  If no error occurs, return a PoDBobTOQ struct and a nil error.
//  Otherwise, return a nil session and the non-nil error.
func BobNewSessForTOQ(keyName string, keyValue []string, phantomKeyValue []string, tableBulletin string, tablePublicPath string, AliceID [40]uint8, BobID [40]uint8, Log ILogger) (PoDBobTOQ, error) {
	var toq PoDBobTOQ
	session, err := table_ot_vrf.NewBobSession(tableBulletin, tablePublicPath, AliceID, BobID, keyName, keyValue, phantomKeyValue)
	if err != nil {
		Log.Warnf("failed to create session for Bob. err=%v", err)
		return toq, errors.New("failed to create session for Bob")
	}
	Log.Debugf("success to creatr session for Bob.")
	toq = PoDBobTOQ{session, keyName, keyValue, phantomKeyValue}
	return toq, nil
}

// BobGeneNegoReq generates nego request file for Bob while mode is table_ot_vrf_query.
//
// It is provides an interface for GetNegoRequest.
//
// Return:
//  If generates nego request for Bob successfully, return true.
//  Otherwise, return false.
func (toq PoDBobTOQ) BobGeneNegoReq(requestFile string, Log ILogger) bool {

	err := toq.BobSession.GetNegoRequest(requestFile)
	if err != nil {
		Log.Warnf("failed to generate nego request. err=%v", err)
		return false
	}
	Log.Debugf("generate nego request. requestFile=%v", requestFile)
	return true
}

// BobGeneNegoResp verifies nego request and generates nego response for Bob while mode is table_ot_vrf_query.
//
// It is provides an interface for OnNegoRequest.
//
// Return:
//  If generates nego response for Bob successfully, return true.
//  Otherwise, return false.
func (toq PoDBobTOQ) BobGeneNegoResp(requestFile string, responseFile string, Log ILogger) bool {
	err := toq.BobSession.OnNegoRequest(requestFile, responseFile)
	if err != nil {
		Log.Warnf("Failed to generate nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to generate nego response. requestFile=%v, responseFile=%v", requestFile, responseFile)
	return true
}

// BobNewReq creates request file for Bob while mode is table_ot_vrf_query.
//
// It is provides an interface for GetRequest.
//
// Return:
//  If no error occurs, generate a request file and return a nil error.
//  Otherwise, return the non-nil error.
func (toq PoDBobTOQ) BobNewReq(requestFile string, Log ILogger) error {
	err := toq.BobSession.GetRequest(requestFile)
	if err != nil {
		Log.Warnf("Failed to create request file. err=%v", err)
		return errors.New("Failed to create request file")
	}
	Log.Debugf("success to create request. requestFile=%v", requestFile)
	return nil
}

// BobDealNegoResp deals with Alice's nego response for Bob while mode is table_ot_vrf_query.
//
// It is provides an interface for OnNegoResponse.
//
// Return:
//  If deals with Alice's nego response successfully, return true.
//  Otherwise, return false.
func (toq PoDBobTOQ) BobDealNegoResp(responseFile string, Log ILogger) bool {
	err := toq.BobSession.OnNegoResponse(responseFile)
	if err != nil {
		Log.Warnf("Failed to deal with nego response. err=%v", err)
		return false
	}
	Log.Debugf("success to deal with nego response")
	return true
}

// BobVerifyResp verifies response data for Bob while mode is table_ot_vrf_query.
//
// It is provides an interface for OnResponse.
//
// Return:
//  If verify response and generate receipt successfully, return true.
//  Otherwise, return false.
func (toq PoDBobTOQ) BobVerifyResp(responseFile string, receiptFile string, Log ILogger) bool {
	err := toq.BobSession.OnResponse(responseFile, receiptFile)
	if err != nil {
		Log.Warnf("Verify response failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify response and generate receipt")
	return true
}

// BobVerifySecret verifies secret for Bob while mode is table_ot_vrf_query.
//
// It is provides an interface for OnSecret.
//
// Return:
//  If verify secret successfully, return true.
//  Otherwise, return false.
func (toq PoDBobTOQ) BobVerifySecret(secretFile string, tablePosition string, Log ILogger) bool {
	err := toq.BobSession.OnSecret(secretFile, tablePosition)
	if err != nil {
		Log.Warnf("Verify secret failure. err=%v", err)
		return false
	}
	Log.Debugf("success to verify secret and generate position file. secretFile=%v, tablePosition=%v", secretFile, tablePosition)
	return true
}
