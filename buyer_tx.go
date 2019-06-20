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

// BuyerTransaction shows the transaction data for buyer.
type BuyerTransaction struct {
	SessionID        string      `json:"sessionId"`
	Status           string      `json:"status"`
	SellerIP         string      `json:"sellerIp"`
	SellerAddr       string      `json:"sellerAddr"`
	BuyerAddr        string      `json:"buyerAddr"`
	Mode             string      `json:"mode"`
	SubMode          string      `json:"sub_mode"`
	OT               bool        `json:"ot"`
	Bulletin         Bulletin    `json:"bulletin"`
	Price            int64       `json:"price" xorm:"INTEGER"`
	UnitPrice        int64       `json:"unit_price"`
	ExpireAt         int64       `json:"expireAt" xorm:"INTEGER"`
	PlainComplaint   PoDBuyerPC  `json:"PlainComplaint"`
	PlainOTComplaint PoDBuyerPOC `json:"PlainOTComplaint"`
	PlainAtomicSwap  PoDBuyerPAS `json:"PlainAtomicSwap"`
	TableComplaint   PoDBuyerTC  `json:"TableComplaint"`
	TableOTComplaint PoDBuyerTOC `json:"TableOTComplaint"`
	TableAtomicSwap  PoDBuyerTAS `json:"TableAtomicSwap"`
	TableVRF         PoDBuyerTQ  `json:"tablevrf"`
	TableOTVRF       PoDBuyerTOQ `json:"tableOTvrf"`
}

// buyerTxForPC executes transaction for buyer while mode is plain_complaint.
//
// step1: prepare session,
// step2: create transaction request,
// step3: receive transaction response,
// step4: create transaction receipt,
// step5: read,save and verify secret from contract,
// step6: claim to contract or decrypt data.
//
// return: response string for api request.
func buyerTxForPC(node *pod_net.Node, key *keystore.Key, tx BuyerTransaction, demands []Demand, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BuyerDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicPath := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"
	claimFile := dir + "/claim"

	defer func() {
		err := updateBuyerTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for buyer. err=%v", tx.SessionID, err)
			return
		}
		delete(BuyerTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for pod's session...", tx.SessionID)
	var err error
	tx.PlainComplaint, err = buyerNewSessForPC(demands, bulletinFile, publicPath, converAddr(tx.SellerAddr), converAddr(tx.BuyerAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.PlainComplaint.BuyerSession.Free()
	}()
	Log.Debugf("[%v]step1: finish preparing for buyer's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start create and send request to seller...", tx.SessionID)
	err = tx.PlainComplaint.buyerNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to create request. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish create request to seller...", tx.SessionID)

	err = buyerSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send transaction request to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish send request to seller...", tx.SessionID)

	Log.Debugf("[%v]step3: start receive and verify response from seller...", tx.SessionID)
	err = buyerRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to receive transaction response from seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish receive response from seller...", tx.SessionID)

	rs := tx.PlainComplaint.buyerVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: verify response failed...", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish verify response...", tx.SessionID)

	Log.Debugf("[%v]step4: start read, sign and send receipt to seller...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForComplaint(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to read receipt for mode complaint. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish read receipt...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForComplaint(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish generate signature...", tx.SessionID)

	err = buyerSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to send transaction receipt to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish send recipt to seller...", tx.SessionID)

	Log.Debugf("[%v]step5: start read, save and verify secret from contract...", tx.SessionID)
	secret, err := readScrtForComplaint(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read secret from contract...", tx.SessionID)

	err = buyerSaveSecretForComplaint(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to save secret for buyer. err=%v", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish save secret...", tx.SessionID)

	rs = tx.PlainComplaint.buyerVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step5: finish verify secret.result=%v", tx.SessionID, rs)
	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: start claim from contract...", tx.SessionID)
		rs = tx.PlainComplaint.buyerGeneClaim(claimFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to generate claim.")
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		Log.Debugf("[%v]finish generate claim...", tx.SessionID)

		txid, err := claimToContractForComplaint(tx.SessionID, tx.Bulletin, claimFile, tx.SellerAddr, Log)
		if err != nil {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to send claim to contract. err=%v", err)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_SEND_CLIAM
		BuyerTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step6: finish claim to contract...txid=%v", tx.SessionID, txid)
	} else {
		Log.Debugf("[%v]step6: start decrypt file...", tx.SessionID)
		rs = tx.PlainComplaint.buyerDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to decrypt File.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_CLOSED
		BuyerTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step6: decrypt file successfully, path=%v", tx.SessionID, outputFile)
	}
	Log.Debugf("[%v]step6: purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// buyerTxForPOC executes transaction for buyer while mode is plain_ot_complaint.
//
// step1: prepare session,
// step2: exchage keys with seller,
// step3: create transaction request,
// step4: receive transaction response,
// step5: create transaction receipt,
// step6: read,save and verify secret from contract,
// step7: claim to contract or decrypt data.
//
// return: response string for api request.
func buyerTxForPOC(node *pod_net.Node, key *keystore.Key, tx BuyerTransaction, demands []Demand, phantoms []Phantom, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BuyerDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"
	claimFile := dir + "/claim"
	buyerNegoRequestFile := dir + "/buyer_nego_request"
	buyerNegoResponseFile := dir + "/buyer_nego_response"
	sellerNegoRequestFile := dir + "/seller_nego_request"
	sellerNegoResponseFile := dir + "/seller_nego_response"

	defer func() {
		err := updateBuyerTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for buyer. err=%v", tx.SessionID, err)
			return
		}
		delete(BuyerTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for buyer's session...", tx.SessionID)
	var err error
	tx.PlainOTComplaint, err = buyerNewSessForPOC(demands, phantoms, bulletinFile, publicPath, converAddr(tx.SellerAddr), converAddr(tx.BuyerAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: Failed to create session for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.PlainOTComplaint.BuyerSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for buyer's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start exchage key with seller...", tx.SessionID)
	rs := tx.PlainOTComplaint.buyerGeneNegoReq(buyerNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego request data.", tx.SessionID)

	err = buyerSendNegoReq(node, buyerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send transaction nego request to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego request data.", tx.SessionID)

	err = buyerRcvNegoResp(node, sellerNegoResponseFile, sellerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to receive transaction nego response for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to receive nego response and request data.", tx.SessionID)

	rs = tx.PlainOTComplaint.buyerDealNegoResp(sellerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to deal with nego response", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to deal with nego response.", tx.SessionID)

	rs = tx.PlainOTComplaint.buyerGeneNegoResp(sellerNegoRequestFile, buyerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego response.", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego response.", tx.SessionID)

	err = buyerSendNegoResp(node, buyerNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send nego response to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego response data.", tx.SessionID)

	tx.Status = TRANSACTION_STATUS_NEGO
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish exchage key with seller...", tx.SessionID)

	Log.Debugf("[%v]step3: start create and send transaction request to seller...", tx.SessionID)
	err = tx.PlainOTComplaint.buyerNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to create transaction request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish create transaction request...", tx.SessionID)

	err = buyerSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to send transaction request data  to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish send transaction request to seller...", tx.SessionID)

	Log.Debugf("[%v]step4: start receive and verify transaction response from seller...", tx.SessionID)
	err = buyerRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to receive transaction response from seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish receive transaction response from seller...", tx.SessionID)

	rs = tx.PlainOTComplaint.buyerVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: verify transaction response failed...", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish verify transaction response...", tx.SessionID)

	Log.Debugf("[%v]step5: start read, sign and send receipt to seller...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForComplaint(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read receipt...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForComplaint(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish generate signature for receipt...", tx.SessionID)

	err = buyerSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to send receipt to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish send recipt to seller...", tx.SessionID)

	Log.Debugf("[%v]step6: start read, save and verify secret from contract...", tx.SessionID)
	secret, err := readScrtForComplaint(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	Log.Debugf("[%v]step6: finish read secret from contract...", tx.SessionID)

	err = buyerSaveSecretForComplaint(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to save secret for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	Log.Debugf("[%v]step6: finish save secret...", tx.SessionID)

	rs = tx.PlainOTComplaint.buyerVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step6: finish verify secret, result=%v", tx.SessionID, rs)
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BuyerTxMap[tx.SessionID] = tx

	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step7: start claim to contract...", tx.SessionID)
		rs = tx.PlainOTComplaint.buyerGeneClaim(claimFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: failed to generate claim.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		Log.Debugf("[%v]step7: finish generate claim...", tx.SessionID)

		txid, err := claimToContractForComplaint(tx.SessionID, tx.Bulletin, claimFile, tx.SellerAddr, Log)
		if err != nil {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: Failed to read secret from contract. err=%v", tx.SessionID, err)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_SEND_CLIAM
		BuyerTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step7: finish send claim to contract...txid=%v", tx.SessionID, txid)
	} else {
		Log.Debugf("[%v]step7: start decrypt data...", tx.SessionID)
		rs = tx.PlainOTComplaint.buyerDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: failed to decrypt data.")
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_CLOSED
		BuyerTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step7: decrypt file successfully, path=%v", tx.SessionID, outputFile)
	}
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// buyerTxForPAS executes transaction for buyer while mode is plain_atomic_swap.
//
// step1: prepare session,
// step2: create transaction request,
// step3: receive transaction response,
// step4: create transaction receipt,
// step5: read,save and verify secret from contract,
// step6: decrypt data.
//
// return: response string for api request.
func buyerTxForPAS(node *pod_net.Node, key *keystore.Key, tx BuyerTransaction, demands []Demand, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BuyerDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicPath := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"

	defer func() {
		err := updateBuyerTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for buyer. err=%v", tx.SessionID, err)
			return
		}
		delete(BuyerTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for buyer's session...", tx.SessionID)
	var err error
	tx.PlainAtomicSwap, err = buyerNewSessForPAS(demands, bulletinFile, publicPath, converAddr(tx.SellerAddr), converAddr(tx.BuyerAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		Log.Warnf("[%v]step1: failed to create session for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.PlainAtomicSwap.BuyerSession.Free()
	}()
	Log.Debugf("[%v]step1: finish preparing for buyer's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start create and send transaction request to seller...", tx.SessionID)
	err = tx.PlainAtomicSwap.buyerNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to create transaction request. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish create transaction request...", tx.SessionID)

	err = buyerSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send transaction request to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}

	Log.Debugf("[%v]step2: finish send transaction request to seller...", tx.SessionID)
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BuyerTxMap[tx.SessionID] = tx

	Log.Debugf("[%v]step3: start receive and verify transaction response from seller...", tx.SessionID)
	err = buyerRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction response from seller. err=%v", err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish receive response from seller...", tx.SessionID)

	rs := tx.PlainAtomicSwap.buyerVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to invalid data. ")
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish verify response from seller...", tx.SessionID)

	Log.Debugf("[%v]step4: start read, send and verify receipt to seller...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForAtomicSwap(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish read receipt...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForAtomicSwap(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish generate signature for receipt...", tx.SessionID)

	err = buyerSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to send transaction receipt to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish send receipt to seller...", tx.SessionID)

	Log.Debugf("[%v]step5: start read, save and verify secret from contract...", tx.SessionID)
	secret, err := readScrtForAtomicSwap(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read secret from contract...", tx.SessionID)

	err = buyerSaveSecretForAtomicSwap(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to save secret for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish save secret...", tx.SessionID)

	rs = tx.PlainAtomicSwap.buyerVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step5: finish verify secret...result=%v", tx.SessionID, rs)
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BuyerTxMap[tx.SessionID] = tx

	if !rs {
		Log.Warnf("purchase failed.")
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	} else {
		Log.Debugf("[%v]step6: start decrypt data...", tx.SessionID)
		rs = tx.PlainAtomicSwap.buyerDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to decrypt file.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_CLOSED
		BuyerTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step6: decrypt file successfully, path=%v", tx.SessionID, outputFile)
	}
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// buyerTxForTC executes transaction for buyer while mode is table_complaint.
//
// step1: prepare session,
// step2: create transaction request,
// step3: receive transaction response,
// step4: create transaction receipt,
// step5: read,save and verify secret from contract,
// step6: claim to contract or decrypt data.
//
// return: response string for api request.
func buyerTxForTC(node *pod_net.Node, key *keystore.Key, tx BuyerTransaction, demands []Demand, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BuyerDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"
	claimFile := dir + "/claim"

	defer func() {
		err := updateBuyerTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for buyer. err=%v", tx.SessionID, err)
			return
		}
		delete(BuyerTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for buyer's session...", tx.SessionID)
	var err error
	tx.TableComplaint, err = buyerNewSessForTC(demands, bulletinFile, publicPath, converAddr(tx.SellerAddr), converAddr(tx.BuyerAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.TableComplaint.BuyerSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for buyer's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start create and send transaction request to seller...", tx.SessionID)
	err = tx.TableComplaint.buyerNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to create transaction request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish create transaction request to seller...", tx.SessionID)

	err = buyerSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send transaction request data  to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish send transaction request to seller...", tx.SessionID)

	Log.Debugf("[%v]step3: start receive transaction response from seller...", tx.SessionID)
	err = buyerRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to receive transaction response from seller. err=%v", err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish receive transaction response from seller...", tx.SessionID)

	rs := tx.TableComplaint.buyerVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to verify transaction response...", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish verify response...", tx.SessionID)
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BuyerTxMap[tx.SessionID] = tx

	Log.Debugf("[%v]step4: start read, sign and send receipt to seller...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForComplaint(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish read receipt...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForComplaint(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to generate signature. err=%v", err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish generate signature for receipt...", tx.SessionID)

	err = buyerSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("Failed to send receipt to seller. err=%v", err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish send recipt to seller...", tx.SessionID)
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BuyerTxMap[tx.SessionID] = tx

	Log.Debugf("[%v]step5: start read save and verify secret from contract...", tx.SessionID)
	secret, err := readScrtForComplaint(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read secret from contract...", tx.SessionID)

	err = buyerSaveSecretForComplaint(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("Failed to save secret for buyer. err=%v")
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish save secret...", tx.SessionID)
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BuyerTxMap[tx.SessionID] = tx

	rs = tx.TableComplaint.buyerVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step5: finish verify secret...result=%v", tx.SessionID, rs)
	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step6: start claim to contract...", tx.SessionID)
		rs = tx.TableComplaint.buyerGeneClaim(claimFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to generate claim.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		Log.Debugf("[%v]step6: finish generate claim...", tx.SessionID)

		txid, err := claimToContractForComplaint(tx.SessionID, tx.Bulletin, claimFile, tx.SellerAddr, Log)
		if err != nil {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to send claim to contract. err=%v", tx.SessionID, err)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		Log.Debugf("[%v]step6: finish claim to contract...txid=%v", tx.SessionID, txid)

		tx.Status = TRANSACTION_STATUS_SEND_CLIAM
		BuyerTxMap[tx.SessionID] = tx
	} else {
		Log.Debugf("[%v]step6: start decrypt data...", tx.SessionID)
		rs = tx.TableComplaint.buyerDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to decrypt File.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		Log.Debugf("[%v]step6: finish decrypt file, path=%v", tx.SessionID, outputFile)
		tx.Status = TRANSACTION_STATUS_CLOSED
		BuyerTxMap[tx.SessionID] = tx
	}
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// buyerTxForTOC executes transaction  for buyer while mode is table_ot_complaint.
//
// step1: prepare session,
// step2: exchage keys with seller,
// step3: create transaction request,
// step4: receive transaction response,
// step5: create transaction receipt,
// step6: read, save and verify secret from contract,
// step7: claim to contract or decrypt data.
//
// return: response string for api request.
func buyerTxForTOC(node *pod_net.Node, key *keystore.Key, tx BuyerTransaction, demands []Demand, phantoms []Phantom, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BuyerDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"
	claimFile := dir + "/claim"
	buyerNegoRequestFile := dir + "/buyer_nego_request"
	buyerNegoResponseFile := dir + "/buyer_nego_response"
	sellerNegoRequestFile := dir + "/seller_nego_request"
	sellerNegoResponseFile := dir + "/seller_nego_response"

	defer func() {
		err := updateBuyerTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for buyer. err=%v", tx.SessionID, err)
			return
		}
		delete(BuyerTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for buyer's session...", tx.SessionID)
	var err error
	tx.TableOTComplaint, err = buyerNewSessForTOC(demands, phantoms, bulletinFile, publicPath, converAddr(tx.SellerAddr), converAddr(tx.BuyerAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.TableOTComplaint.BuyerSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for buyer's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start exchage key with seller...", tx.SessionID)
	rs := tx.TableOTComplaint.buyerGeneNegoReq(buyerNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego request data", tx.SessionID)

	err = buyerSendNegoReq(node, buyerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send nego request data  to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego request data", tx.SessionID)

	err = buyerRcvNegoResp(node, sellerNegoResponseFile, sellerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to receive nego response and nego ack request from seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to receive nego response and nego ack request data", tx.SessionID)

	rs = tx.TableOTComplaint.buyerDealNegoResp(sellerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to deal with nego response", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to deal with nego response", tx.SessionID)

	rs = tx.TableOTComplaint.buyerGeneNegoResp(sellerNegoRequestFile, buyerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego response.", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego response", tx.SessionID)

	err = buyerSendNegoResp(node, buyerNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send nego response data to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego response data", tx.SessionID)

	tx.Status = TRANSACTION_STATUS_NEGO
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish exchage key with seller...", tx.SessionID)

	Log.Debugf("[%v]step3: start create and send transaction request to seller...", tx.SessionID)
	err = tx.TableOTComplaint.buyerNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to create transaction request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish create transaction request...", tx.SessionID)

	err = buyerSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to send transaction request data  to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish send transaction request to seller...", tx.SessionID)

	Log.Debugf("[%v]step4: start receive and save response from seller...", tx.SessionID)
	err = buyerRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to receive data to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish receive transaction response from seller...", tx.SessionID)

	rs = tx.TableOTComplaint.buyerVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to verify transaction response", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish verify transaction response...", tx.SessionID)

	Log.Debugf("[%v]step5: start read, sign and send receipt to seller...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForComplaint(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read receipt...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForComplaint(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish generate signature for receipt...", tx.SessionID)

	err = buyerSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to send receipt to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish send recipt to seller...", tx.SessionID)

	Log.Debugf("[%v]step6: start read, save and verify secret...", tx.SessionID)
	secret, err := readScrtForComplaint(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	Log.Debugf("[%v]step6: finish read secret from contract...", tx.SessionID)

	err = buyerSaveSecretForComplaint(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to save secret for buyer. err=%v")
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step6: finish save secret...", tx.SessionID)

	rs = tx.TableOTComplaint.buyerVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step6: finish verify secret... result=%v", tx.SessionID, rs)
	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step7: start generate and send claim to contract...", tx.SessionID)
		rs = tx.TableOTComplaint.buyerGeneClaim(claimFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: failed to generate claim.")
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		Log.Debugf("[%v]step7: finish generate claim...", tx.SessionID)

		txid, err := claimToContractForComplaint(tx.SessionID, tx.Bulletin, claimFile, tx.SellerAddr, Log)
		if err != nil {
			tx.Status = TRANSACTION_STATUS_SEND_CLIAM_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: failed to send claim to contract. err=%v", tx.SessionID, err)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		Log.Debugf("[%v]step7: finish send claim to contract...txid=%v", tx.SessionID, txid)

		tx.Status = TRANSACTION_STATUS_SEND_CLIAM
		BuyerTxMap[tx.SessionID] = tx
	} else {
		Log.Debugf("[%v]step7: start decrypt data...", tx.SessionID)
		rs = tx.TableOTComplaint.buyerDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step7: failed to decrypt File.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
		}
		Log.Debugf("[%v]step7: finish decrypt file, path=%v", tx.SessionID, outputFile)

		tx.Status = TRANSACTION_STATUS_CLOSED
		BuyerTxMap[tx.SessionID] = tx
	}
	Log.Debugf("[%v]purchase successfully...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// buyerTxForTAS executes transaction for buyer while mode is table_atomic_swap.
//
// step1: prepare session,
// step2: create transaction request,
// step3: receive transaction response,
// step4: create transaction receipt,
// step5: read,save and verify secret from contract,
// step6: claim to contract or decrypt data.
//
// return: response string for api request.
func buyerTxForTAS(node *pod_net.Node, key *keystore.Key, tx BuyerTransaction, demands []Demand, bulletinFile string, publicPath string, Log ILogger) string {
	dir := BConf.BuyerDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	outputFile := dir + "/output"

	defer func() {
		err := updateBuyerTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for buyer. err=%v", tx.SessionID, err)
			return
		}
		delete(BuyerTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for buyer's session...", tx.SessionID)
	var err error
	tx.TableAtomicSwap, err = buyerNewSessForTAS(demands, bulletinFile, publicPath, converAddr(tx.SellerAddr), converAddr(tx.BuyerAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.TableAtomicSwap.BuyerSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for buyer's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start create and send transaction request to seller...", tx.SessionID)
	err = tx.TableAtomicSwap.buyerNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to create request. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish create transaction request to seller...", tx.SessionID)

	err = buyerSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send request data to seller. err=%v", err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish send request to seller...", tx.SessionID)

	Log.Debugf("[%v]step3: start receive and verify transaction response from seller...", tx.SessionID)
	err = buyerRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		Log.Warnf("[%v]step3: Failed to receive data to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish receive response from seller...", tx.SessionID)

	rs := tx.TableAtomicSwap.buyerVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: verify transaction response failed.", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish receive transaction response from seller...", tx.SessionID)

	Log.Debugf("[%v]step4: start read, sign and send receipt to seller...", tx.SessionID)
	receiptByte, receipt, err := readReceiptForAtomicSwap(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish read receipt from seller...", tx.SessionID)

	tx.Price = tx.UnitPrice * int64(receipt.C)
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForAtomicSwap(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish generate signature for receipt...", tx.SessionID)

	err = buyerSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to send receipt to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish send receipt to seller...", tx.SessionID)

	Log.Debugf("[%v]step5: start read and verify secret from contract...", tx.SessionID)
	secret, err := readScrtForAtomicSwap(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read secret from contract...", tx.SessionID)

	err = buyerSaveSecretForAtomicSwap(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to save secret for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish save secret...", tx.SessionID)

	rs = tx.TableAtomicSwap.buyerVerifySecret(secretFile, Log)
	Log.Debugf("[%v]step5: finish verify secret...result=%v", tx.SessionID, rs)
	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: purchase failed.", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	} else {
		Log.Debugf("[%v]step6: start decrypt file...")
		rs = tx.TableAtomicSwap.buyerDecrypt(outputFile, Log)
		if !rs {
			tx.Status = TRANSACTION_STATUS_DECRYPT_FAILED
			BuyerTxMap[tx.SessionID] = tx
			Log.Warnf("[%v]step6: failed to decrypt File.", tx.SessionID)
			return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
		}
		tx.Status = TRANSACTION_STATUS_CLOSED
		BuyerTxMap[tx.SessionID] = tx
		Log.Debugf("[%v]step6: finish decrypt file, path=%v", tx.SessionID, outputFile)
	}
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// buyerTxForTQ executes transaction  for buyer while mode is table_vrf_query.
//
// step1: prepare session,
// step2: create transaction request,
// step3: receive transaction response,
// step4: create transaction receipt,
// step5: read and save secret from contract,
// step6: verify secrypt and decrypt data.
//
// return: response string for api request.
func buyerTxForTQ(node *pod_net.Node, key *keystore.Key, tx BuyerTransaction, keyName string, keyValue []string, bulletinFile string, publicPath string, Log ILogger) string {

	dir := BConf.BuyerDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	positionFile := dir + "/positions"

	defer func() {
		err := updateBuyerTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for buyer. err=%v", tx.SessionID, err)
			return
		}
		delete(BuyerTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for buyer's session...", tx.SessionID)
	var err error
	tx.TableVRF, err = buyerNewSessForTQ(keyName, keyValue, bulletinFile, publicPath, converAddr(tx.SellerAddr), converAddr(tx.BuyerAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.TableVRF.BuyerSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for buyer's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start create and send transaction request to seller...", tx.SessionID)
	err = tx.TableVRF.buyerNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to create request, err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: finish create transaction request to seller...", tx.SessionID)

	err = buyerSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send request data to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish send transaction request to seller...", tx.SessionID)

	Log.Debugf("[%v]step3: start receive transaction response from seller...", tx.SessionID)
	err = buyerRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to receive data to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish receive transaction response from seller...", tx.SessionID)

	rs := tx.TableVRF.buyerVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: verify transaction response failed.", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish verify transaction response from seller...", tx.SessionID)

	Log.Debugf("[%v]step4: start read, sign and send receipt to seller...", tx.SessionID)
	receiptByte, receipt, err := readVRFReceipt(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to read vrf receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish read receipt to seller...", tx.SessionID)

	tx.Price = tx.UnitPrice
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForVRFQ(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish generate signature for receipt...", tx.SessionID)

	err = buyerSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to send receipt to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish send receipt to seller...", tx.SessionID)

	Log.Debugf("[%v]step5: start read and save secret from contract...", tx.SessionID)
	secret, err := readScrtForVRFQ(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read secret from contract...", tx.SessionID)

	err = buyerSaveSecretForVRFQ(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to save secret for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish save secret...", tx.SessionID)

	Log.Debugf("[%v]step6: start verify secret and generate position file...", tx.SessionID)
	rs = tx.TableVRF.buyerVerifySecret(secretFile, positionFile, Log)
	if !rs {
		Log.Warnf("[%v]step6: failed to verify secret and generate position file.", tx.SessionID)
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BuyerTxMap[tx.SessionID] = tx
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	Log.Debugf("[%v]step6: finish verify secret and generate position file", tx.SessionID)

	tx.Status = TRANSACTION_STATUS_CLOSED
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}

// buyerTxForTOQ executes transaction  for buyer while mode is table_ot_vrf_query.
//
// step1: prepare session,
// step2: exchage keys with seller,
// step3: create transaction request,
// step4: receive transaction response,
// step5: create transaction receipt,
// step6: read, and save secret from contract,
// step7: verify secret and decrypt data.
//
// return: response string for api request.
func buyerTxForTOQ(node *pod_net.Node, key *keystore.Key, tx BuyerTransaction, keyName string, keyValue []string, phantomKeyValue []string, bulletinFile string, publicPath string, Log ILogger) string {
	dir := BConf.BuyerDir + "/transaction/" + tx.SessionID
	// bulletinFile := dir + "/bulletin"
	// publicFile := dir + "/public"
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	positionFile := dir + "/positions"
	buyerNegoRequestFile := dir + "/buyer_nego_request"
	buyerNegoResponseFile := dir + "/buyer_nego_response"
	sellerNegoRequestFile := dir + "/seller_nego_request"
	sellerNegoResponseFile := dir + "/seller_nego_response"

	defer func() {

		err := updateBuyerTxToDB(tx)
		if err != nil {
			Log.Warnf("[%v]failed to update transaction to db for buyer. err=%v", tx.SessionID, err)
			return
		}
		delete(BuyerTxMap, tx.SessionID)
	}()

	Log.Debugf("[%v]step1: prepare for buyer's session...", tx.SessionID)
	var err error
	tx.TableOTVRF, err = buyerNewSessForTOQ(keyName, keyValue, phantomKeyValue, bulletinFile, publicPath, converAddr(tx.SellerAddr), converAddr(tx.BuyerAddr), Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_START_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step1: failed to create session for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step1: failed to purchase data.")
	}
	defer func() {
		tx.TableOTVRF.BuyerSession.Free()
	}()
	Log.Debugf("[%v]step1: finish prepare for buyer's session...", tx.SessionID)

	Log.Debugf("[%v]step2: start exchage key with seller...", tx.SessionID)
	rs := tx.TableOTVRF.buyerGeneNegoReq(buyerNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego request", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego request data", tx.SessionID)

	err = buyerSendNegoReq(node, buyerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send nego request data  to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego request data", tx.SessionID)

	err = buyerRcvNegoResp(node, sellerNegoResponseFile, sellerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to download nego response and nego ack request from seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to download nego response and nego ack request", tx.SessionID)

	rs = tx.TableOTVRF.buyerDealNegoResp(sellerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to deal with nego response", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to deal with nego response", tx.SessionID)

	rs = tx.TableOTVRF.buyerGeneNegoResp(sellerNegoRequestFile, buyerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to generate nego response", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to generate nego response", tx.SessionID)

	err = buyerSendNegoResp(node, buyerNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step2: failed to send nego response data  to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step2: failed to purchase data.")
	}
	Log.Debugf("[%v]step2: success to send nego response data", tx.SessionID)

	tx.Status = TRANSACTION_STATUS_NEGO
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step2: finish exchage key with seller...", tx.SessionID)

	Log.Debugf("[%v]step3: start create and send transaction request to seller...", tx.SessionID)
	err = tx.TableOTVRF.buyerNewReq(requestFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to create transaction request. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	Log.Debugf("[%v]step3: finish create transaction request to seller...", tx.SessionID)

	err = buyerSendPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step3: failed to send transaction request data to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step3: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step3: finish send transaction request to seller...", tx.SessionID)

	Log.Debugf("[%v]step4: start receive and verify transaction response from seller...", tx.SessionID)
	err = buyerRcvPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to receive data to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	Log.Debugf("[%v]step4: finish receive transaction response to seller...", tx.SessionID)

	rs = tx.TableOTVRF.buyerVerifyResp(responseFile, receiptFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step4: failed to verify response data")
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step4: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step4: finish verify response...", tx.SessionID)

	Log.Debugf("[%v]step5: start read, sign and send receipt to seller...", tx.SessionID)
	receiptByte, receipt, err := readVRFReceipt(receiptFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to read receipt. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Price = tx.UnitPrice
	tx.ExpireAt = time.Now().Unix() + 36000
	sign, err := signRecptForVRFQ(key, tx.SessionID, receipt, tx.Price, tx.ExpireAt, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to generate signature. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	Log.Debugf("[%v]step5: finish read receipt...", tx.SessionID)

	err = buyerSendPODRecpt(node, tx.Price, tx.ExpireAt, receiptByte, sign)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step5: failed to send receipt to seller. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step5: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step5: finish send recipt to seller...", tx.SessionID)

	Log.Debugf("[%v]step6: start read and save secret from contract...", tx.SessionID)
	secret, err := readScrtForVRFQ(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to read secret from contract. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	Log.Debugf("[%v]step6: finish read secret from contract...", tx.SessionID)

	err = buyerSaveSecretForVRFQ(secret, secretFile, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_GOT_SECRET_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step6: failed to save secret for buyer. err=%v", tx.SessionID, err)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step6: failed to purchase data.")
	}
	tx.Status = TRANSACTION_STATUS_GOT_SECRET
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]step6: start save secret...", tx.SessionID)

	Log.Debugf("[%v]step7: start verify secret and generate position file...", tx.SessionID)
	rs = tx.TableOTVRF.buyerVerifySecret(secretFile, positionFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_VERIFY_FAILED
		BuyerTxMap[tx.SessionID] = tx
		Log.Warnf("[%v]step7: failed to verify secret and generate position file", tx.SessionID)
		return fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "step7: failed to purchase data.")
	}
	Log.Debugf("[%v]step7: finish verify secret...", tx.SessionID)

	tx.Status = TRANSACTION_STATUS_CLOSED
	BuyerTxMap[tx.SessionID] = tx
	Log.Debugf("[%v]purchase finish...", tx.SessionID)
	return fmt.Sprintf(RESPONSE_SUCCESS, "purchase data successfully. sessionID="+tx.SessionID)
}
