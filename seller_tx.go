package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"

	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	pod_net "github.com/sec-bit/zkPoD-node/net"
)

//Transaction shows the transaction data for Alice.
type Transaction struct {
	SessionID         string           `json:"sessionId"`
	Status            string           `json:"status"`
	Bulletin          Bulletin         `json:"bulletin"`
	BobPubKey         *ecdsa.PublicKey `json:"BobPubkey"`
	BobAddr           string           `json:"BobAddr"`
	AliceAddr         string           `json:"AliceAddr"`
	Mode              string           `json:"mode"`
	SubMode           string           `json:"sub_mode"`
	OT                bool             `json:"ot"`
	Price             int64            `json:"price"`
	UnitPrice         int64            `json:"unitPrice"`
	Count             int64            `json:"count"`
	ExpireAt          int64            `json:"expireAt"`
	PlainComplaint    PoDAlicePC       `json:"PlainComplaint"`
	PlainOTComplaint  PoDAlicePOC      `json:"PlainOTComplaint"`
	PlainAtomicSwap   PoDAlicePAS      `json:"TableAtomicSwap"`
	PlainAtomicSwapVc PoDAlicePASVC    `json:"TableAtomicSwapVc"`
	TableComplaint    PoDAliceTC       `json:"TableComplaint"`
	TableOTComplaint  PoDAliceTOC      `json:"TableOTComplaint"`
	TableAtomicSwap   PoDAliceTAS      `json:"TableAtomicSwap"`
	TableAtomicSwapVc PoDAliceTASVC    `json:"TableAtomicSwapVc"`
	TableVRF          PoDAliceTQ       `json:"Tablevrf"`
	TableOTVRF        PoDAliceTOQ      `json:"TableOTvrf"`
}

func newSessID() (string, error) {
	bigRand, err := rand.Int(rand.Reader, big.NewInt(1e10))
	if err != nil {
		return "", err
	}
	randStr := bigRand.String()
	return randStr, nil
}

//preAliceTx prepares for transaction.
func preAliceTx(mklroot string, re requestExtra, Log ILogger) (AliceConnParam, requestExtra, error) {
	var params AliceConnParam
	bulletinPath := BConf.AliceDir + "/publish/" + mklroot + "/bulletin"
	bulletin, err := readBulletinFile(bulletinPath, Log)
	if err != nil {
		Log.Warnf("failed to read bulletin. err=%v", err)
		return params, re, fmt.Errorf("failed to read bulletin")
	}
	extraPath := BConf.AliceDir + "/publish/" + mklroot + "/extra.json"
	extra, err := readExtraFile(extraPath)
	if err != nil {
		Log.Warnf("failed to extra file. err=%v", err)
		return params, re, fmt.Errorf("failed to extra file")
	}
	Log.Debugf("read extra info...%v", extra)

	if extra.UnitPrice > re.Price {
		Log.Warnf("unit price is lower. unit price=%v, lowest unit price=%v", re.Price, extra.UnitPrice)
		return params, re, fmt.Errorf("the price is too low")
	}
	params.UnitPrice = re.Price
	re.Mode = bulletin.Mode
	var subMode string
	if len(extra.SubMode) == 0 {
		return params, re, fmt.Errorf("no subMode")
	}
	for _, s := range extra.SubMode {
		if re.SubMode == s && s != "" {
			subMode = s
			break
		}
	}
	if subMode == "" {
		subMode = extra.SubMode[0]
	}
	re.SubMode = subMode
	params.Mode = bulletin.Mode
	params.SubMode = subMode
	if params.SubMode == TRANSACTION_SUB_MODE_ATOMIC_SWAP || params.SubMode == TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC {
		re.Ot = false
	}
	params.OT = re.Ot
	params.Bulletin = bulletin

	bltByte, err := calcuBltKey(bulletin)
	if err != nil {
		Log.Warnf("failed to calculate bltKey. err=%v", err)
		return params, re, errors.New("failed to calculate bltKey")
	}

	status, err := readDataStatusAtContract(bltByte)
	if err != nil {
		Log.Warnf("failed to check whether the data(merkle root = %v) is in sale. err=%v", bulletin.SigmaMKLRoot, err)
		return params, re, errors.New("Failed to check data")
	}
	if status != "OK" {
		Log.Warnf("The data merkel root = %v) is not in salse", bulletin.SigmaMKLRoot)
		return params, re, errors.New("The data is not in salse")
	}

	sessionID, err := newSessID()
	if err != nil {
		Log.Warnf("Failed to create session Id. err=%v", err)
		return params, re, errors.New("failed to create session Id")
	}
	Log.Debugf("success to create session ID. sessionID=%v", sessionID)

	params.SessionID = sessionID
	// err = savePublishFileForTransaction(sessionID, bulletin.SigmaMKLRoot, Log)
	// if err != nil {
	// 	Log.Warnf("Failed to save bulletin for Alice. err=%v", err)
	// 	return params, re, errors.New("failed to save bulletin")
	// }
	// Log.Debugf("success to save publish file.")
	dir := BConf.AliceDir + "/transaction/" + sessionID
	err = os.Mkdir(dir, os.ModePerm)
	if err != nil {
		Log.Errorf("create folder %v error. err=%v", dir, err)
		return params, re, errors.New("failed to create folder")
	}
	Log.Debugf("success to create folder. dir=%v", dir)

	return params, re, nil
}

//AliceTxForPC is the transaction while mode is plain_complaint.
func AliceTxForPC(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {

	requestFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/request"
	responseFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/response"
	receiptFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/receipt"
	secretFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/secret"

	defer func() {
		err := updateAliceTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for Alice. err=%v", err)
			return
		}
		delete(AliceTxMap, tx.SessionID)
	}()

	err := AliceRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive transaction request for Alice.")

	rs := tx.PlainComplaint.AliceVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	tx.Count, err = calcuCntforComplaint(requestFile)
	if err != nil {
		Log.Warnf("failed to calculate count by request")
		return fmt.Errorf(
			"invalid request file or response file")
	}
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify transaction request and generate transaction response for Alice")

	tx.Price = tx.Count * tx.UnitPrice
	rs, err = calcuDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	defer func() {
		DepositLockMap[tx.AliceAddr+tx.BobAddr] -= tx.Price
	}()

	err = AliceSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send transaction response for Alice")

	var sign []byte
	var price int64
	sign, price, tx.ExpireAt, err = AliceRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive transaction receipt.")

	if price != tx.Price {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid price in signature. price=%v, real price=%v", price, tx.Price)
		return fmt.Errorf(
			"invalid price in signature")
	}

	rs = tx.PlainComplaint.AliceVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify receipt and generate secret.")

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForComplaint(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send secret to contract. err=%v", err)
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForComplaint(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}

	tx.Status = TRANSACTION_STATUS_CLOSED
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract")
	return nil
}

//AliceTxForPOC is the transaction while mode is plain_ot_complaint.
func AliceTxForPOC(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	dir := BConf.AliceDir + "/transaction/" + tx.SessionID
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	BobNegoRequestFile := dir + "/Bob_nego_request"
	BobNegoResponseFile := dir + "/Bob_nego_response"
	AliceNegoRequestFile := dir + "/Alice_nego_request"
	AliceNegoResponseFile := dir + "/Alice_nego_response"

	defer func() {
		err := updateAliceTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for Alice. err=%v", err)
			return
		}
		delete(AliceTxMap, tx.SessionID)
	}()

	err := AliceReceiveNegoReq(node, BobNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego request. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego request")
	}
	Log.Debugf("success to receive transaction nego request.")

	rs := tx.PlainOTComplaint.AliceGeneNegoResp(BobNegoRequestFile, AliceNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid nego request file or nego response file. err=%v", err)
		return fmt.Errorf(
			"invalid nego request file or nego response file")
	}
	Log.Debugf("success to verify nego request and generate nego response.")

	rs = tx.PlainOTComplaint.AliceGeneNegoReq(AliceNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to generate nego request file. err=%v", err)
		return fmt.Errorf(
			"failed to generate nego request file")
	}
	Log.Debugf("success to generate nego request for Alice.")

	err = AliceSendNegoResp(node, AliceNegoResponseFile, AliceNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("success to send transaction nego response.")

	err = AliceRcvNegoResp(node, BobNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("success to receive transaction nego response.")

	rs = tx.PlainOTComplaint.AliceDealNegoResp(BobNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to deal with nego response. err=%v", err)
		return fmt.Errorf(
			"failed to deal with nego response")
	}
	Log.Debugf("success to deal with nego response.")
	tx.Status = TRANSACTION_STATUS_NEGO
	AliceTxMap[tx.SessionID] = tx

	err = AliceRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	Log.Debugf("success to receive transaction request.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	AliceTxMap[tx.SessionID] = tx

	rs = tx.PlainOTComplaint.AliceVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid transaction request file or transaction response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file.")
	}
	Log.Debugf("success to verify transaction request and generate transaction response.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	tx.Count, err = calcuCntforOtComplaint(requestFile)
	if err != nil {
		Log.Warnf("failed to calculate count by request")
		return fmt.Errorf(
			"invalid request file or response file")
	}
	AliceTxMap[tx.SessionID] = tx

	tx.Price = tx.Count * tx.UnitPrice
	rs, err = calcuDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	defer func() {
		DepositLockMap[tx.AliceAddr+tx.BobAddr] -= tx.Price
	}()

	err = AliceSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	Log.Debugf("success to send transaction response.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	AliceTxMap[tx.SessionID] = tx

	var sign []byte
	var price int64
	sign, price, tx.ExpireAt, err = AliceRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	Log.Debugf("success to receive transaction receipt for Alice.")
	tx.Status = TRANSACTION_STATUS_RECEIPT
	AliceTxMap[tx.SessionID] = tx

	if price != tx.Price {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid price in signature. price=%v, real price=%v", price, tx.Price)
		return fmt.Errorf(
			"invalid price in signature")
	}

	rs = tx.PlainOTComplaint.AliceVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	Log.Debugf("success to verify receipt and generate secret.")
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	AliceTxMap[tx.SessionID] = tx

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	rs, err = checkDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("check deposit...unpass")
		return fmt.Errorf(
			"deposit is not enough or expired")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForComplaint(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract. err=%v", err)
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForComplaint(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}

	tx.Status = TRANSACTION_STATUS_CLOSED
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract.")
	return nil
}

//AliceTxForPAS is the transaction while mode is plain_atomic_swap.
func AliceTxForPAS(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	requestFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/request"
	responseFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/response"
	receiptFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/receipt"
	secretFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/secret"

	defer func() {
		err := updateAliceTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for Alice. err=%v", err)
			return
		}
		delete(AliceTxMap, tx.SessionID)
	}()

	err := AliceRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	Log.Debugf("success to receive transaction request.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	AliceTxMap[tx.SessionID] = tx

	rs := tx.PlainAtomicSwap.AliceVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	Log.Debugf("success to verify request and generate response.")
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	tx.Count, err = calcuCntforAS(requestFile)
	if err != nil {
		Log.Warnf("failed to calculate count by request")
		return fmt.Errorf(
			"invalid request file or response file")
	}
	AliceTxMap[tx.SessionID] = tx

	tx.Price = tx.Count * tx.UnitPrice
	rs, err = calcuDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	defer func() {
		DepositLockMap[tx.AliceAddr+tx.BobAddr] -= tx.Price
	}()

	err = AliceSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response to Bob. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	Log.Debugf("success to send transaction response to Bob.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	AliceTxMap[tx.SessionID] = tx

	var sign []byte
	var price int64
	sign, price, tx.ExpireAt, err = AliceRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt from Bob. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	Log.Debugf("success to receive transaction receipt from Bob.")
	tx.Status = TRANSACTION_STATUS_RECEIPT
	AliceTxMap[tx.SessionID] = tx

	if price != tx.Price {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid price in signature. price=%v, real price=%v", price, tx.Price)
		return fmt.Errorf(
			"invalid price in signature")
	}

	rs = tx.PlainAtomicSwap.AliceVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	Log.Debugf("success to verify receipt file and generate secret file.")
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	AliceTxMap[tx.SessionID] = tx

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	rs, err = checkDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("check deposit...unpass")
		return fmt.Errorf(
			"deposit is not enough or expired")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForAtomicSwap(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract")
		return fmt.Errorf(
			"failed to send secret")
	}

	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForAtomicSwap(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}

	tx.Status = TRANSACTION_STATUS_CLOSED
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract.")
	return nil
}

//AliceTxForPASVC is the transaction while mode is plain_atomic_swap_vc.
func AliceTxForPASVC(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	requestFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/request"
	responseFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/response"
	receiptFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/receipt"
	secretFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/secret"

	defer func() {
		err := updateAliceTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for Alice. err=%v", err)
			return
		}
		delete(AliceTxMap, tx.SessionID)
	}()

	err := AliceRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	Log.Debugf("success to receive transaction request.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	AliceTxMap[tx.SessionID] = tx

	rs := tx.PlainAtomicSwapVc.AliceVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	Log.Debugf("success to verify request and generate response.")
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	tx.Count, err = calcuCntforASVC(requestFile)
	if err != nil {
		Log.Warnf("failed to calculate count by request")
		return fmt.Errorf(
			"invalid request file or response file")
	}
	AliceTxMap[tx.SessionID] = tx

	tx.Price = tx.Count * tx.UnitPrice
	rs, err = calcuDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	defer func() {
		DepositLockMap[tx.AliceAddr+tx.BobAddr] -= tx.Price
	}()

	err = AliceSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response to Bob. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	Log.Debugf("success to send transaction response to Bob.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	AliceTxMap[tx.SessionID] = tx

	var sign []byte
	var price int64
	sign, price, tx.ExpireAt, err = AliceRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt from Bob. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	Log.Debugf("success to receive transaction receipt from Bob.")
	tx.Status = TRANSACTION_STATUS_RECEIPT
	AliceTxMap[tx.SessionID] = tx

	if price != tx.Price {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid price in signature. price=%v, real price=%v", price, tx.Price)
		return fmt.Errorf(
			"invalid price in signature")
	}

	rs = tx.PlainAtomicSwapVc.AliceVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	Log.Debugf("success to verify receipt file and generate secret file.")
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	AliceTxMap[tx.SessionID] = tx

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}
	rs, err = checkDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("check deposit...unpass")
		return fmt.Errorf(
			"deposit is not enough or expired")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForAtomicSwapVc(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract")
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForAtomicSwapVc(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}

	tx.Status = TRANSACTION_STATUS_CLOSED
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract.")
	return nil
}

//AliceTxForTC is the transaction while mode is table_complaint.
func AliceTxForTC(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	dir := BConf.AliceDir + "/transaction/" + tx.SessionID
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"

	defer func() {
		err := updateAliceTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for Alice. err=%v", err)
			return
		}
		delete(AliceTxMap, tx.SessionID)
	}()

	err := AliceRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	Log.Debugf("success to receive transaction request.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	AliceTxMap[tx.SessionID] = tx

	rs := tx.TableComplaint.AliceVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	tx.Count, err = calcuCntforComplaint(requestFile)
	if err != nil {
		Log.Warnf("failed to calculate count by request")
		return fmt.Errorf(
			"invalid request file or response file")
	}
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify transaction request and generate transaction response.")

	tx.Price = tx.Count * tx.UnitPrice
	rs, err = calcuDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	defer func() {
		DepositLockMap[tx.AliceAddr+tx.BobAddr] -= tx.Price
	}()

	err = AliceSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response to Bob. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send transaction response to Bob.")

	var sign []byte
	var price int64
	sign, price, tx.ExpireAt, err = AliceRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive receipt from Bob.")

	if price != tx.Price {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid price in signature. price=%v, real price=%v", price, tx.Price)
		return fmt.Errorf(
			"invalid price in signature")
	}

	rs = tx.TableComplaint.AliceVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify receipt and generate secret.")

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}
	rs, err = checkDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("check deposit...unpass")
		return fmt.Errorf(
			"deposit is not enough or expired")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForComplaint(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract. err=%v", err)
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForComplaint(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}

	tx.Status = TRANSACTION_STATUS_CLOSED
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract")
	return nil
}

//AliceTxForTOC is the transaction while mode is table_ot_complaint.
func AliceTxForTOC(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	dir := BConf.AliceDir + "/transaction/" + tx.SessionID
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	BobNegoRequestFile := dir + "/Bob_nego_request"
	BobNegoResponseFile := dir + "/Bob_nego_response"
	AliceNegoRequestFile := dir + "/Alice_nego_request"
	AliceNegoResponseFile := dir + "/Alice_nego_response"

	defer func() {
		err := updateAliceTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for Alice. err=%v", err)
			return
		}
		delete(AliceTxMap, tx.SessionID)
	}()

	err := AliceReceiveNegoReq(node, BobNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego request. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego request")
	}
	Log.Debugf("receive transaction nego request...")

	rs := tx.TableOTComplaint.AliceGeneNegoResp(BobNegoRequestFile, AliceNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid nego request file or nego response file. err=%v", err)
		return fmt.Errorf(
			"invalid nego request file or nego response file")
	}
	Log.Debugf("generate nego request file or nego response file...")

	rs = tx.TableOTComplaint.AliceGeneNegoReq(AliceNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to generate nego request file. err=%v", err)
		return fmt.Errorf(
			"failed to generate nego request file")
	}
	Log.Debugf("Alice generates nego request file...")

	err = AliceSendNegoResp(node, AliceNegoResponseFile, AliceNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("send nego response and nego request ...")

	err = AliceRcvNegoResp(node, BobNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("receive transaction nego response...")

	rs = tx.TableOTComplaint.AliceDealNegoResp(BobNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to deal with nego response. err=%v", err)
		return fmt.Errorf(
			"failed to deal with nego response")
	}
	tx.Status = TRANSACTION_STATUS_NEGO
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("deal with nego response...")

	err = AliceRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("receive transaction request for Alice...")

	rs = tx.TableOTComplaint.AliceVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	tx.Count, err = calcuCntforOtComplaint(requestFile)
	if err != nil {
		Log.Warnf("failed to calculate count by request")
		return fmt.Errorf(
			"invalid request file or response file")
	}
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("verify transaction request file and GENERATE response file...")

	tx.Price = tx.Count * tx.UnitPrice
	rs, err = calcuDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	defer func() {
		DepositLockMap[tx.AliceAddr+tx.BobAddr] -= tx.Price
	}()

	err = AliceSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("send transaction response...")

	var sign []byte
	var price int64
	sign, price, tx.ExpireAt, err = AliceRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("receive transaction receipt...")

	if price != tx.Price {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid price in signature. price=%v, real price=%v", price, tx.Price)
		return fmt.Errorf(
			"invalid price in signature")
	}

	rs = tx.TableOTComplaint.AliceVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("verify receipt file and GENERATE secret file...")

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}
	rs, err = checkDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("check deposit...unpass")
		return fmt.Errorf(
			"deposit is not enough or expired")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForComplaint(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract. err=%v", err)
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForComplaint(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}

	tx.Status = TRANSACTION_STATUS_CLOSED
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("send secret to contract...")
	return nil
}

//AliceTxForTAS is the transaction while mode is table_atomic_swap.
func AliceTxForTAS(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	dir := BConf.AliceDir + "/transaction/" + tx.SessionID
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"

	defer func() {
		err := updateAliceTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for Alice. err=%v", err)
			return
		}
		delete(AliceTxMap, tx.SessionID)
	}()

	err := AliceRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	Log.Debugf("success to receive transaction request.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	AliceTxMap[tx.SessionID] = tx

	rs := tx.TableAtomicSwap.AliceVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	tx.Count, err = calcuCntforAS(requestFile)
	if err != nil {
		Log.Warnf("failed to calculate count by request")
		return fmt.Errorf(
			"invalid request file or response file")
	}
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify transaction request and generate response.")

	tx.Price = tx.Count * tx.UnitPrice
	rs, err = calcuDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	defer func() {
		DepositLockMap[tx.AliceAddr+tx.BobAddr] -= tx.Price
	}()

	err = AliceSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response to Bob. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send transaction response to Bob.")

	var sign []byte
	var price int64
	sign, price, tx.ExpireAt, err = AliceRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt from Bob. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive receipt from Bob.")

	if price != tx.Price {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid price in signature. price=%v, real price=%v", price, tx.Price)
		return fmt.Errorf(
			"invalid price in signature")
	}

	rs = tx.TableAtomicSwap.AliceVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify receipt and generate secret.")

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}
	rs, err = checkDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("check deposit...unpass")
		return fmt.Errorf(
			"deposit is not enough or expired")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForAtomicSwap(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract")
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForAtomicSwap(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}

	tx.Status = TRANSACTION_STATUS_CLOSED
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract.")
	return nil
}

//AliceTxForTASVC is the transaction while mode is table_atomic_swap_vc.
func AliceTxForTASVC(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	dir := BConf.AliceDir + "/transaction/" + tx.SessionID
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"

	defer func() {
		err := updateAliceTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for Alice. err=%v", err)
			return
		}
		delete(AliceTxMap, tx.SessionID)
	}()

	err := AliceRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	Log.Debugf("success to receive transaction request.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	AliceTxMap[tx.SessionID] = tx

	rs := tx.TableAtomicSwapVc.AliceVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	tx.Count, err = calcuCntforASVC(requestFile)
	if err != nil {
		Log.Warnf("failed to calculate count by request")
		return fmt.Errorf(
			"invalid request file or response file")
	}
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify transaction request and generate response.")

	tx.Price = tx.Count * tx.UnitPrice
	rs, err = calcuDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	defer func() {
		DepositLockMap[tx.AliceAddr+tx.BobAddr] -= tx.Price
	}()

	err = AliceSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response to Bob. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send transaction response to Bob.")

	var sign []byte
	var price int64
	sign, price, tx.ExpireAt, err = AliceRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt from Bob. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive receipt from Bob.")

	if price != tx.Price {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid price in signature. price=%v, real price=%v", price, tx.Price)
		return fmt.Errorf(
			"invalid price in signature")
	}

	rs = tx.TableAtomicSwapVc.AliceVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify receipt and generate secret.")

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}
	rs, err = checkDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("check deposit...unpass")
		return fmt.Errorf(
			"deposit is not enough or expired")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForAtomicSwapVc(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract")
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForAtomicSwapVc(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}

	tx.Status = TRANSACTION_STATUS_CLOSED
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract.")
	return nil
}

//AliceTxForTQ is the transaction while mode is table_vrf_query.
func AliceTxForTQ(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	requestFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/request"
	responseFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/response"
	receiptFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/receipt"
	secretFile := BConf.AliceDir + "/transaction/" + tx.SessionID + "/secret"

	defer func() {
		err := updateAliceTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for Alice. err=%v", err)
			return
		}
		delete(AliceTxMap, tx.SessionID)
	}()

	tx.Price = tx.UnitPrice
	rs, err := calcuDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	defer func() {
		DepositLockMap[tx.AliceAddr+tx.BobAddr] -= tx.Price
	}()

	err = AliceRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive transaction request for Alice.")

	rs = tx.TableVRF.AliceVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify request file and generate respons file.")

	err = AliceSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send transaction response for Alice.")

	var sign []byte
	var price int64
	sign, price, tx.ExpireAt, err = AliceRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive receipt from Bob. err=%v", err)
		return fmt.Errorf(
			"failed to receive receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive receipt from Bob.")

	if price != tx.Price {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid price in signature. price=%v, real price=%v", price, tx.Price)
		return fmt.Errorf(
			"invalid price in signature")
	}

	rs = tx.TableVRF.AliceVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify receipt.")

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}
	rs, err = checkDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("check deposit...unpass")
		return fmt.Errorf(
			"deposit is not enough or expired")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForVRFQ(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract")
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForVRFQ(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}

	tx.Status = TRANSACTION_STATUS_CLOSED
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract.")
	return nil
}

//AliceTxForTOQ is the transaction while mode is table_ot_vrf_query.
func AliceTxForTOQ(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	dir := BConf.AliceDir + "/transaction/" + tx.SessionID
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	BobNegoRequestFile := dir + "/Bob_nego_request"
	BobNegoResponseFile := dir + "/Bob_nego_response"
	AliceNegoRequestFile := dir + "/Alice_nego_request"
	AliceNegoResponseFile := dir + "/Alice_nego_response"

	defer func() {
		err := updateAliceTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for Alice. err=%v", err)
			return
		}
		delete(AliceTxMap, tx.SessionID)
	}()

	tx.Price = tx.UnitPrice
	rs, err := calcuDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_RECEIVED_DEPOSIT_NOT_ENOUGH
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	defer func() {
		DepositLockMap[tx.AliceAddr+tx.BobAddr] -= tx.Price
	}()

	err = AliceReceiveNegoReq(node, BobNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego request. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego request")
	}
	Log.Debugf("success to receive transaction nego request.")

	rs = tx.TableOTVRF.AliceGeneNegoResp(BobNegoRequestFile, AliceNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid nego request file or nego response file. err=%v", err)
		return fmt.Errorf(
			"invalid nego request file or nego response file")
	}
	Log.Debugf("success to generate nego request and generate nego response")

	rs = tx.TableOTVRF.AliceGeneNegoReq(AliceNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to generate nego request file. err=%v", err)
		return fmt.Errorf(
			"failed to generate nego request file")
	}
	Log.Debugf("success to generate nego request for Alice")

	err = AliceSendNegoResp(node, AliceNegoResponseFile, AliceNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("success to send transaction nego response")

	err = AliceRcvNegoResp(node, BobNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("success to receive transaction nego response")

	rs = tx.TableOTVRF.AliceDealNegoResp(BobNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to deal with nego response. err=%v", err)
		return fmt.Errorf(
			"failed to deal with nego response")
	}
	tx.Status = TRANSACTION_STATUS_NEGO
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to deal with nego response.")

	err = AliceRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request from Bob. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive transaction request from Bob.")

	rs = tx.TableOTVRF.AliceVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify request and generate response.")

	err = AliceSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response to Bob. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send transaction response to Bob")

	var sign []byte
	var price int64
	sign, price, tx.ExpireAt, err = AliceRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive receipt for Alice. err=%v", err)
		return fmt.Errorf(
			"failed to receive receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive receipt for Alice")

	if price != tx.Price {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid price in signature. price=%v, real price=%v", price, tx.Price)
		return fmt.Errorf(
			"invalid price in signature")
	}

	rs = tx.TableOTVRF.AliceVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify receipt and generate secret")

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	rs, err = checkDeposit(tx.AliceAddr, tx.BobAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("check deposit...unpass")
		return fmt.Errorf(
			"deposit is not enough or expired")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForVRFQ(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract")
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForVRFQ(tx.SessionID, tx.AliceAddr, tx.BobAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		AliceTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}

	tx.Status = TRANSACTION_STATUS_CLOSED
	AliceTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract")
	return nil
}
