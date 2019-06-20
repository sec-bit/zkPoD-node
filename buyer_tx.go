package main

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	pod_net "github.com/sec-bit/zkPoD-node/net"
)

func converAddr(address string) [40]uint8 {
	var uAddr [40]uint8
	if len(address) < 40 {
		return uAddr
	}
	var addrByte = []byte(address[len(address)-40:])

	for i, b := range addrByte {
		uAddr[i] = b
	}

	return uAddr
}

// BobTransaction shows the transaction data for Bob.
type BobTransaction struct {
	SessionID        string    `json:"sessionId"`
	Status           string    `json:"status"`
	AliceIP         string    `json:"AliceIp"`
	AliceAddr       string    `json:"AliceAddr"`
	BobAddr          string    `json:"BobAddr"`
	Mode             string    `json:"mode"`
	SubMode          string    `json:"sub_mode"`
	OT               bool      `json:"ot"`
	Bulletin         Bulletin  `json:"bulletin"`
	Price            int64     `json:"price" xorm:"INTEGER"`
	UnitPrice        int64     `json:"unit_price"`
	ExpireAt         int64     `json:"expireAt" xorm:"INTEGER"`
	PlainComplaint   PoDBobPC  `json:"PlainComplaint"`
	PlainOTComplaint PoDBobPOC `json:"PlainOTComplaint"`
	PlainAtomicSwap  PoDBobPAS `json:"PlainAtomicSwap"`
	TableComplaint   PoDBobTC  `json:"TableComplaint"`
	TableOTComplaint PoDBobTOC `json:"TableOTComplaint"`
	TableAtomicSwap  PoDBobTAS `json:"TableAtomicSwap"`
	TableVRF         PoDBobTQ  `json:"tablevrf"`
	TableOTVRF       PoDBobTOQ `json:"tableOTvrf"`
}

// BobTxForPC executes transaction for Bob while mode is plain_complaint.
//
// step1: prepare session,
// step2: create transaction request,
// step3: receive transaction response,
// step4: create transaction receipt,
// step5: read,save and verify secret from contract,
// step6: claim to contract or decrypt data.
//
// return: response string for api request.
func BobTxForPC(node *pod_net.Node, key *keystore.Key, tx BobTransaction, demands []Demand, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BobDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicPath := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"
	claimFile := dir + "/claim"

	defer func() {
		err := updateBobTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for Bob. err=%v", tx.SessionID, err)
			return
		}
		delete(BobTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for pod's session...", tx.SessionID)
	var err error
	tx.PlainComplaint, err = BobNewSessForPC(demands, bulletinFile, publicPath, converAddr(tx.AliceAddr), converAddr(tx.BobAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.PlainComplaint.BobSession.Free()
	}()
	Log.Debugf("[%v]step1: finish preparing for Bob's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start create and send request to Alice...", tx.SessionID)
	err = tx.PlainComplaint.BobNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to create request. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish create request to Alice...", tx.SessionID)

	err = BobSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send transaction request to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish send request to Alice...", tx.SessionID)

	Log.Debugf("[%v]step3: start receive and verify response from Alice...", tx.SessionID)
	err = BobRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to receive transaction response from Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish receive response from Alice...", tx.SessionID)

	rs := tx.PlainComplaint.BobVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: verify response failed...", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish verify response...", tx.SessionID)

	Log.Debugf("[%v]step4: start read, sign and send receipt to Alice...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForComplaint(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to read receipt for mode complaint. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish read receipt...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForComplaint(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish generate signature...", tx.SessionID)

	err = BobSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to send transaction receipt to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish send recipt to Alice...", tx.SessionID)

	Log.Debugf("[%v]step5: start read, save and verify secret from contract...", tx.SessionID)
	secret, err := readScrtForComplaint(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read secret from contract...", tx.SessionID)

	err = BobSaveSecretForComplaint(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to save secret for Bob. err=%v", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish save secret...", tx.SessionID)

	rs = tx.PlainComplaint.BobVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step5: finish verify secret.result=%v", tx.SessionID, rs)
	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: start claim from contract...", tx.SessionID)
		rs = tx.PlainComplaint.BobGeneClaim(claimFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to generate claim.")
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		Log.Debugf("[%v]finish generate claim...", tx.SessionID)

		txid, err := claimToContractForComplaint(tx.SessionID, tx.Bulletin, claimFile, tx.AliceAddr, Log)
		if err != nil {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to send claim to contract. err=%v", err)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_SEND_CLIAM
		BobTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step6: finish claim to contract...txid=%v", tx.SessionID, txid)
	} else {
		Log.Debugf("[%v]step6: start decrypt file...", tx.SessionID)
		rs = tx.PlainComplaint.BobDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to decrypt File.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_CLOSED
		BobTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step6: decrypt file successfully, path=%v", tx.SessionID, outputFile)
	}
	Log.Debugf("[%v]step6: purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// BobTxForPOC executes transaction for Bob while mode is plain_ot_complaint.
//
// step1: prepare session,
// step2: exchage keys with Alice,
// step3: create transaction request,
// step4: receive transaction response,
// step5: create transaction receipt,
// step6: read,save and verify secret from contract,
// step7: claim to contract or decrypt data.
//
// return: response string for api request.
func BobTxForPOC(node *pod_net.Node, key *keystore.Key, tx BobTransaction, demands []Demand, phantoms []Phantom, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BobDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"
	claimFile := dir + "/claim"
	BobNegoRequestFile := dir + "/Bob_nego_request"
	BobNegoResponseFile := dir + "/Bob_nego_response"
	AliceNegoRequestFile := dir + "/Alice_nego_request"
	AliceNegoResponseFile := dir + "/Alice_nego_response"

	defer func() {
		err := updateBobTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for Bob. err=%v", tx.SessionID, err)
			return
		}
		delete(BobTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for Bob's session...", tx.SessionID)
	var err error
	tx.PlainOTComplaint, err = BobNewSessForPOC(demands, phantoms, bulletinFile, publicPath, converAddr(tx.AliceAddr), converAddr(tx.BobAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: Failed to create session for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.PlainOTComplaint.BobSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for Bob's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start exchage key with Alice...", tx.SessionID)
	rs := tx.PlainOTComplaint.BobGeneNegoReq(BobNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego request data.", tx.SessionID)

	err = BobSendNegoReq(node, BobNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send transaction nego request to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego request data.", tx.SessionID)

	err = BobRcvNegoResp(node, AliceNegoResponseFile, AliceNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to receive transaction nego response for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to receive nego response and request data.", tx.SessionID)

	rs = tx.PlainOTComplaint.BobDealNegoResp(AliceNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to deal with nego response", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to deal with nego response.", tx.SessionID)

	rs = tx.PlainOTComplaint.BobGeneNegoResp(AliceNegoRequestFile, BobNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego response.", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego response.", tx.SessionID)

	err = BobSendNegoResp(node, BobNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send nego response to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego response data.", tx.SessionID)

	tx.Status = TRANSACTION_STATUS_NEGO
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish exchage key with Alice...", tx.SessionID)

	Log.Debugf("[%v]step3: start create and send transaction request to Alice...", tx.SessionID)
	err = tx.PlainOTComplaint.BobNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to create transaction request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish create transaction request...", tx.SessionID)

	err = BobSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to send transaction request data  to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish send transaction request to Alice...", tx.SessionID)

	Log.Debugf("[%v]step4: start receive and verify transaction response from Alice...", tx.SessionID)
	err = BobRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to receive transaction response from Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish receive transaction response from Alice...", tx.SessionID)

	rs = tx.PlainOTComplaint.BobVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: verify transaction response failed...", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish verify transaction response...", tx.SessionID)

	Log.Debugf("[%v]step5: start read, sign and send receipt to Alice...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForComplaint(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read receipt...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForComplaint(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish generate signature for receipt...", tx.SessionID)

	err = BobSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to send receipt to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish send recipt to Alice...", tx.SessionID)

	Log.Debugf("[%v]step6: start read, save and verify secret from contract...", tx.SessionID)
	secret, err := readScrtForComplaint(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	Log.Debugf("[%v]step6: finish read secret from contract...", tx.SessionID)

	err = BobSaveSecretForComplaint(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to save secret for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	Log.Debugf("[%v]step6: finish save secret...", tx.SessionID)

	rs = tx.PlainOTComplaint.BobVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step6: finish verify secret, result=%v", tx.SessionID, rs)
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BobTxMap[tx.SessionID] = tx

	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step7: start claim to contract...", tx.SessionID)
		rs = tx.PlainOTComplaint.BobGeneClaim(claimFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: failed to generate claim.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		Log.Debugf("[%v]step7: finish generate claim...", tx.SessionID)

		txid, err := claimToContractForComplaint(tx.SessionID, tx.Bulletin, claimFile, tx.AliceAddr, Log)
		if err != nil {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: Failed to read secret from contract. err=%v", tx.SessionID, err)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_SEND_CLIAM
		BobTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step7: finish send claim to contract...txid=%v", tx.SessionID, txid)
	} else {
		Log.Debugf("[%v]step7: start decrypt data...", tx.SessionID)
		rs = tx.PlainOTComplaint.BobDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: failed to decrypt data.")
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_CLOSED
		BobTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step7: decrypt file successfully, path=%v", tx.SessionID, outputFile)
	}
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// BobTxForPAS executes transaction for Bob while mode is plain_atomic_swap.
//
// step1: prepare session,
// step2: create transaction request,
// step3: receive transaction response,
// step4: create transaction receipt,
// step5: read,save and verify secret from contract,
// step6: decrypt data.
//
// return: response string for api request.
func BobTxForPAS(node *pod_net.Node, key *keystore.Key, tx BobTransaction, demands []Demand, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BobDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicPath := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"

	defer func() {
		err := updateBobTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for Bob. err=%v", tx.SessionID, err)
			return
		}
		delete(BobTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for Bob's session...", tx.SessionID)
	var err error
	tx.PlainAtomicSwap, err = BobNewSessForPAS(demands, bulletinFile, publicPath, converAddr(tx.AliceAddr), converAddr(tx.BobAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		Log.Warnf("[%v]step1: failed to create session for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.PlainAtomicSwap.BobSession.Free()
	}()
	Log.Debugf("[%v]step1: finish preparing for Bob's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start create and send transaction request to Alice...", tx.SessionID)
	err = tx.PlainAtomicSwap.BobNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to create transaction request. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish create transaction request...", tx.SessionID)

	err = BobSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send transaction request to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}

	Log.Debugf("[%v]step2: finish send transaction request to Alice...", tx.SessionID)
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BobTxMap[tx.SessionID] = tx

	Log.Debugf("[%v]step3: start receive and verify transaction response from Alice...", tx.SessionID)
	err = BobRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction response from Alice. err=%v", err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish receive response from Alice...", tx.SessionID)

	rs := tx.PlainAtomicSwap.BobVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("failed to invalid data. ")
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish verify response from Alice...", tx.SessionID)

	Log.Debugf("[%v]step4: start read, send and verify receipt to Alice...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForAtomicSwap(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish read receipt...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForAtomicSwap(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish generate signature for receipt...", tx.SessionID)

	err = BobSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to send transaction receipt to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish send receipt to Alice...", tx.SessionID)

	Log.Debugf("[%v]step5: start read, save and verify secret from contract...", tx.SessionID)
	secret, err := readScrtForAtomicSwap(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read secret from contract...", tx.SessionID)

	err = BobSaveSecretForAtomicSwap(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to save secret for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish save secret...", tx.SessionID)

	rs = tx.PlainAtomicSwap.BobVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step5: finish verify secret...result=%v", tx.SessionID, rs)
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BobTxMap[tx.SessionID] = tx

	if !rs {
		Log.Warnf("purchase failed.")
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	} else {
		Log.Debugf("[%v]step6: start decrypt data...", tx.SessionID)
		rs = tx.PlainAtomicSwap.BobDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to decrypt file.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_CLOSED
		BobTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step6: decrypt file successfully, path=%v", tx.SessionID, outputFile)
	}
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// BobTxForTC executes transaction for Bob while mode is table_complaint.
//
// step1: prepare session,
// step2: create transaction request,
// step3: receive transaction response,
// step4: create transaction receipt,
// step5: read,save and verify secret from contract,
// step6: claim to contract or decrypt data.
//
// return: response string for api request.
func BobTxForTC(node *pod_net.Node, key *keystore.Key, tx BobTransaction, demands []Demand, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BobDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"
	claimFile := dir + "/claim"

	defer func() {
		err := updateBobTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for Bob. err=%v", tx.SessionID, err)
			return
		}
		delete(BobTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for Bob's session...", tx.SessionID)
	var err error
	tx.TableComplaint, err = BobNewSessForTC(demands, bulletinFile, publicPath, converAddr(tx.AliceAddr), converAddr(tx.BobAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.TableComplaint.BobSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for Bob's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start create and send transaction request to Alice...", tx.SessionID)
	err = tx.TableComplaint.BobNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to create transaction request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish create transaction request to Alice...", tx.SessionID)

	err = BobSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send transaction request data  to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish send transaction request to Alice...", tx.SessionID)

	Log.Debugf("[%v]step3: start receive transaction response from Alice...", tx.SessionID)
	err = BobRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to receive transaction response from Alice. err=%v", err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish receive transaction response from Alice...", tx.SessionID)

	rs := tx.TableComplaint.BobVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to verify transaction response...", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish verify response...", tx.SessionID)
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BobTxMap[tx.SessionID] = tx

	Log.Debugf("[%v]step4: start read, sign and send receipt to Alice...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForComplaint(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish read receipt...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForComplaint(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("failed to generate signature. err=%v", err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish generate signature for receipt...", tx.SessionID)

	err = BobSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("Failed to send receipt to Alice. err=%v", err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish send recipt to Alice...", tx.SessionID)
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BobTxMap[tx.SessionID] = tx

	Log.Debugf("[%v]step5: start read save and verify secret from contract...", tx.SessionID)
	secret, err := readScrtForComplaint(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read secret from contract...", tx.SessionID)

	err = BobSaveSecretForComplaint(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("Failed to save secret for Bob. err=%v")
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish save secret...", tx.SessionID)
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BobTxMap[tx.SessionID] = tx

	rs = tx.TableComplaint.BobVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step5: finish verify secret...result=%v", tx.SessionID, rs)
	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step6: start claim to contract...", tx.SessionID)
		rs = tx.TableComplaint.BobGeneClaim(claimFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to generate claim.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		Log.Debugf("[%v]step6: finish generate claim...", tx.SessionID)

		txid, err := claimToContractForComplaint(tx.SessionID, tx.Bulletin, claimFile, tx.AliceAddr, Log)
		if err != nil {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to send claim to contract. err=%v", tx.SessionID, err)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		Log.Debugf("[%v]step6: finish claim to contract...txid=%v", tx.SessionID, txid)

		tx.Status = TRANSACTION_STATUS_SEND_CLIAM
		BobTxMap[tx.SessionID] = tx
	} else {
		Log.Debugf("[%v]step6: start decrypt data...", tx.SessionID)
		rs = tx.TableComplaint.BobDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to decrypt File.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		Log.Debugf("[%v]step6: finish decrypt file, path=%v", tx.SessionID, outputFile)
		tx.Status = TRANSACTION_STATUS_CLOSED
		BobTxMap[tx.SessionID] = tx
	}
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// BobTxForTOC executes transaction  for Bob while mode is table_ot_complaint.
//
// step1: prepare session,
// step2: exchage keys with Alice,
// step3: create transaction request,
// step4: receive transaction response,
// step5: create transaction receipt,
// step6: read, save and verify secret from contract,
// step7: claim to contract or decrypt data.
//
// return: response string for api request.
func BobTxForTOC(node *pod_net.Node, key *keystore.Key, tx BobTransaction, demands []Demand, phantoms []Phantom, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BobDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"
	claimFile := dir + "/claim"
	BobNegoRequestFile := dir + "/Bob_nego_request"
	BobNegoResponseFile := dir + "/Bob_nego_response"
	AliceNegoRequestFile := dir + "/Alice_nego_request"
	AliceNegoResponseFile := dir + "/Alice_nego_response"

	defer func() {
		err := updateBobTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for Bob. err=%v", tx.SessionID, err)
			return
		}
		delete(BobTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for Bob's session...", tx.SessionID)
	var err error
	tx.TableOTComplaint, err = BobNewSessForTOC(demands, phantoms, bulletinFile, publicPath, converAddr(tx.AliceAddr), converAddr(tx.BobAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.TableOTComplaint.BobSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for Bob's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start exchage key with Alice...", tx.SessionID)
	rs := tx.TableOTComplaint.BobGeneNegoReq(BobNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego request data", tx.SessionID)

	err = BobSendNegoReq(node, BobNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send nego request data  to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego request data", tx.SessionID)

	err = BobRcvNegoResp(node, AliceNegoResponseFile, AliceNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to receive nego response and nego ack request from Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to receive nego response and nego ack request data", tx.SessionID)

	rs = tx.TableOTComplaint.BobDealNegoResp(AliceNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to deal with nego response", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to deal with nego response", tx.SessionID)

	rs = tx.TableOTComplaint.BobGeneNegoResp(AliceNegoRequestFile, BobNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego response.", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego response", tx.SessionID)

	err = BobSendNegoResp(node, BobNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send nego response data to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego response data", tx.SessionID)

	tx.Status = TRANSACTION_STATUS_NEGO
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish exchage key with Alice...", tx.SessionID)

	Log.Debugf("[%v]step3: start create and send transaction request to Alice...", tx.SessionID)
	err = tx.TableOTComplaint.BobNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to create transaction request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish create transaction request...", tx.SessionID)

	err = BobSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to send transaction request data  to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish send transaction request to Alice...", tx.SessionID)

	Log.Debugf("[%v]step4: start receive and save response from Alice...", tx.SessionID)
	err = BobRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to receive data to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish receive transaction response from Alice...", tx.SessionID)

	rs = tx.TableOTComplaint.BobVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to verify transaction response", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish verify transaction response...", tx.SessionID)

	Log.Debugf("[%v]step5: start read, sign and send receipt to Alice...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForComplaint(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read receipt...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForComplaint(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish generate signature for receipt...", tx.SessionID)

	err = BobSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to send receipt to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish send recipt to Alice...", tx.SessionID)

	Log.Debugf("[%v]step6: start read, save and verify secret...", tx.SessionID)
	secret, err := readScrtForComplaint(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	Log.Debugf("[%v]step6: finish read secret from contract...", tx.SessionID)

	err = BobSaveSecretForComplaint(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to save secret for Bob. err=%v")
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step6: finish save secret...", tx.SessionID)

	rs = tx.TableOTComplaint.BobVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step6: finish verify secret... result=%v", tx.SessionID, rs)
	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step7: start generate and send claim to contract...", tx.SessionID)
		rs = tx.TableOTComplaint.BobGeneClaim(claimFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: failed to generate claim.")
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		Log.Debugf("[%v]step7: finish generate claim...", tx.SessionID)

		txid, err := claimToContractForComplaint(tx.SessionID, tx.Bulletin, claimFile, tx.AliceAddr, Log)
		if err != nil {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: failed to send claim to contract. err=%v", tx.SessionID, err)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		Log.Debugf("[%v]step7: finish send claim to contract...txid=%v", tx.SessionID, txid)

		tx.Status = TRANSACTION_STATUS_SEND_CLIAM
		BobTxMap[tx.SessionID] = tx
	} else {
		Log.Debugf("[%v]step7: start decrypt data...", tx.SessionID)
		rs = tx.TableOTComplaint.BobDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: failed to decrypt File.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		Log.Debugf("[%v]step7: finish decrypt file, path=%v", tx.SessionID, outputFile)

		tx.Status = TRANSACTION_STATUS_CLOSED
		BobTxMap[tx.SessionID] = tx
	}
	Log.Debugf("[%v]purchase successfully...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// BobTxForTAS executes transaction for Bob while mode is table_atomic_swap.
//
// step1: prepare session,
// step2: create transaction request,
// step3: receive transaction response,
// step4: create transaction receipt,
// step5: read,save and verify secret from contract,
// step6: claim to contract or decrypt data.
//
// return: response string for api request.
func BobTxForTAS(node *pod_net.Node, key *keystore.Key, tx BobTransaction, demands []Demand, bulletinFile string, publicPath string, Log ILogger) string {
	dir := BConf.BobDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"

	defer func() {
		err := updateBobTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for Bob. err=%v", tx.SessionID, err)
			return
		}
		delete(BobTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for Bob's session...", tx.SessionID)
	var err error
	tx.TableAtomicSwap, err = BobNewSessForTAS(demands, bulletinFile, publicPath, converAddr(tx.AliceAddr), converAddr(tx.BobAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.TableAtomicSwap.BobSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for Bob's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start create and send transaction request to Alice...", tx.SessionID)
	err = tx.TableAtomicSwap.BobNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to create request. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish create transaction request to Alice...", tx.SessionID)

	err = BobSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send request data to Alice. err=%v", err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish send request to Alice...", tx.SessionID)

	Log.Debugf("[%v]step3: start receive and verify transaction response from Alice...", tx.SessionID)
	err = BobRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		Log.Warnf("[%v]step3: Failed to receive data to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish receive response from Alice...", tx.SessionID)

	rs := tx.TableAtomicSwap.BobVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: verify transaction response failed.", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish receive transaction response from Alice...", tx.SessionID)

	Log.Debugf("[%v]step4: start read, sign and send receipt to Alice...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForAtomicSwap(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish read receipt from Alice...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForAtomicSwap(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish generate signature for receipt...", tx.SessionID)

	err = BobSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to send receipt to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish send receipt to Alice...", tx.SessionID)

	Log.Debugf("[%v]step5: start read and verify secret from contract...", tx.SessionID)
	secret, err := readScrtForAtomicSwap(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read secret from contract...", tx.SessionID)

	err = BobSaveSecretForAtomicSwap(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to save secret for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish save secret...", tx.SessionID)

	rs = tx.TableAtomicSwap.BobVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step5: finish verify secret...result=%v", tx.SessionID, rs)
	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: purchase failed.", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	} else {
		Log.Debugf("[%v]step6: start decrypt file...")
		rs = tx.TableAtomicSwap.BobDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BobTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to decrypt File.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_CLOSED
		BobTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step6: finish decrypt file, path=%v", tx.SessionID, outputFile)
	}
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// BobTxForTQ executes transaction  for Bob while mode is table_vrf_query.
//
// step1: prepare session,
// step2: create transaction request,
// step3: receive transaction response,
// step4: create transaction receipt,
// step5: read and save secret from contract,
// step6: verify secrypt and decrypt data.
//
// return: response string for api request.
func BobTxForTQ(node *pod_net.Node, key *keystore.Key, tx BobTransaction, keyName string, keyValue []string, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BobDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	positionFile := dir + "/positions"

	defer func() {
		err := updateBobTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for Bob. err=%v", tx.SessionID, err)
			return
		}
		delete(BobTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for Bob's session...", tx.SessionID)
	var err error
	tx.TableVRF, err = BobNewSessForTQ(keyName, keyValue, bulletinFile, publicPath, converAddr(tx.AliceAddr), converAddr(tx.BobAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.TableVRF.BobSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for Bob's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start create and send transaction request to Alice...", tx.SessionID)
	err = tx.TableVRF.BobNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to create request, err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish create transaction request to Alice...", tx.SessionID)

	err = BobSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send request data to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish send transaction request to Alice...", tx.SessionID)

	Log.Debugf("[%v]step3: start receive transaction response from Alice...", tx.SessionID)
	err = BobRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to receive data to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish receive transaction response from Alice...", tx.SessionID)

	rs := tx.TableVRF.BobVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: verify transaction response failed.", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish verify transaction response from Alice...", tx.SessionID)

	Log.Debugf("[%v]step4: start read, sign and send receipt to Alice...", tx.SessionID)
	receiptByte, receipt, err := readVRFReceipt(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to read vrf receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish read receipt to Alice...", tx.SessionID)

	tx.Price = tx.UnitPrice
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForVRFQ(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish generate signature for receipt...", tx.SessionID)

	err = BobSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to send receipt to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish send receipt to Alice...", tx.SessionID)

	Log.Debugf("[%v]step5: start read and save secret from contract...", tx.SessionID)
	secret, err := readScrtForVRFQ(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read secret from contract...", tx.SessionID)

	err = BobSaveSecretForVRFQ(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to save secret for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish save secret...", tx.SessionID)

	Log.Debugf("[%v]step6: start verify secret and generate position file...", tx.SessionID)
	rs = tx.TableVRF.BobVerifySecret(secretFile, positionFile, Log)
	if !rs {
		Log.Warnf("[%v]step6: failed to verify secret and generate position file.", tx.SessionID)
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BobTxMap[tx.SessionID] = tx
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	Log.Debugf("[%v]step6: finish verify secret and generate position file", tx.SessionID)

	tx.Status = TRANSACTION_STATUS_CLOSED
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// BobTxForTOQ executes transaction  for Bob while mode is table_ot_vrf_query.
//
// step1: prepare session,
// step2: exchage keys with Alice,
// step3: create transaction request,
// step4: receive transaction response,
// step5: create transaction receipt,
// step6: read, and save secret from contract,
// step7: verify secret and decrypt data.
//
// return: response string for api request.
func BobTxForTOQ(node *pod_net.Node, key *keystore.Key, tx BobTransaction, keyName string, keyValue []string, phantomKeyValue []string, bulletinFile string, publicPath string, Log ILogger) string {
	dir := BConf.BobDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	positionFile := dir + "/positions"
	BobNegoRequestFile := dir + "/Bob_nego_request"
	BobNegoResponseFile := dir + "/Bob_nego_response"
	AliceNegoRequestFile := dir + "/Alice_nego_request"
	AliceNegoResponseFile := dir + "/Alice_nego_response"

	defer func() {

		err := updateBobTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for Bob. err=%v", tx.SessionID, err)
			return
		}
		delete(BobTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for Bob's session...", tx.SessionID)
	var err error
	tx.TableOTVRF, err = BobNewSessForTOQ(keyName, keyValue, phantomKeyValue, bulletinFile, publicPath, converAddr(tx.AliceAddr), converAddr(tx.BobAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.TableOTVRF.BobSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for Bob's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start exchage key with Alice...", tx.SessionID)
	rs := tx.TableOTVRF.BobGeneNegoReq(BobNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego request data", tx.SessionID)

	err = BobSendNegoReq(node, BobNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send nego request data  to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego request data", tx.SessionID)

	err = BobRcvNegoResp(node, AliceNegoResponseFile, AliceNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to download nego response and nego ack request from Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to download nego response and nego ack request", tx.SessionID)

	rs = tx.TableOTVRF.BobDealNegoResp(AliceNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to deal with nego response", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to deal with nego response", tx.SessionID)

	rs = tx.TableOTVRF.BobGeneNegoResp(AliceNegoRequestFile, BobNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego response", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego response", tx.SessionID)

	err = BobSendNegoResp(node, BobNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send nego response data  to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego response data", tx.SessionID)

	tx.Status = TRANSACTION_STATUS_NEGO
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish exchage key with Alice...", tx.SessionID)

	Log.Debugf("[%v]step3: start create and send transaction request to Alice...", tx.SessionID)
	err = tx.TableOTVRF.BobNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to create transaction request. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish create transaction request to Alice...", tx.SessionID)

	err = BobSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to send transaction request data to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish send transaction request to Alice...", tx.SessionID)

	Log.Debugf("[%v]step4: start receive and verify transaction response from Alice...", tx.SessionID)
	err = BobRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to receive data to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish receive transaction response to Alice...", tx.SessionID)

	rs = tx.TableOTVRF.BobVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to verify response data")
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish verify response...", tx.SessionID)

	Log.Debugf("[%v]step5: start read, sign and send receipt to Alice...", tx.SessionID)
	receiptByte, receipt, err := readVRFReceipt(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Price = tx.UnitPrice
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForVRFQ(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read receipt...", tx.SessionID)

	err = BobSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to send receipt to Alice. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish send recipt to Alice...", tx.SessionID)

	Log.Debugf("[%v]step6: start read and save secret from contract...", tx.SessionID)
	secret, err := readScrtForVRFQ(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	Log.Debugf("[%v]step6: finish read secret from contract...", tx.SessionID)

	err = BobSaveSecretForVRFQ(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to save secret for Bob. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step6: start save secret...", tx.SessionID)

	Log.Debugf("[%v]step7: start verify secret and generate position file...", tx.SessionID)
	rs = tx.TableOTVRF.BobVerifySecret(secretFile, positionFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BobTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step7: failed to verify secret and generate position file", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
	}
	Log.Debugf("[%v]step7: finish verify secret...", tx.SessionID)

	tx.Status = TRANSACTION_STATUS_CLOSED
	BobTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}
