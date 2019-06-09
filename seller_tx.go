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

//Transaction shows the transaction data for seller.
type Transaction struct {
	SessionID     string           `json:"sessionId"`
	Status        string           `json:"status"`
	Bulletin      Bulletin         `json:"bulletin"`
	BuyerPubKey   *ecdsa.PublicKey `json:"buyerPubkey"`
	BuyerAddr     string           `json:"buyerAddr"`
	SellerAddr    string           `json:"sellerAddr"`
	Mode          string           `json:"mode"`
	SubMode       string           `json:"sub_mode"`
	OT            bool             `json:"ot"`
	Price         int64            `json:"price"`
	UnitPrice     int64            `json:"unitPrice"`
	ExpireAt      int64            `json:"expireAt"`
	PlainBatch1   PoDSellerPB1     `json:"plainBatch1"`
	PlainOTBatch1 PoDSellerPOB1    `json:"plainOTBatch1"`
	PlainBatch2   PoDSellerPB2     `json:"Tablebatch2"`
	TableBatch1   PoDSellerTB1     `json:"Tablebatch1"`
	TableOTBatch1 PoDSellerTOB1    `json:"TableOTBatch1"`
	TableBatch2   PoDSellerTB2     `json:"Tablebatch2"`
	TableVRF      PoDSellerTQ      `json:"Tablevrf"`
	TableOTVRF    PoDSellerTOQ     `json:"TableOTvrf"`
}

func newSessID() (string, error) {
	bigRand, err := rand.Int(rand.Reader, big.NewInt(1e10))
	if err != nil {
		return "", err
	}
	randStr := bigRand.String()
	return randStr, nil
}

//preSellerTx prepares for transaction.
func preSellerTx(mklroot string, re requestExtra, Log ILogger) (SellerConnParam, requestExtra, error) {
	var params SellerConnParam
	bulletinPath := BConf.SellerDir + "/publish/" + mklroot + "/bulletin"
	bulletin, err := readBulletinFile(bulletinPath, Log)
	if err != nil {
		Log.Warnf("failed to read bulletin. err=%v", err)
		return params, re, fmt.Errorf("failed to read bulletin")
	}
	extraPath := BConf.SellerDir + "/publish/" + mklroot + "/extra.json"
	extra, err := readExtraFile(extraPath)
	if err != nil {
		Log.Warnf("failed to read bulletin. err=%v", err)
		return params, re, fmt.Errorf("failed to read bulletin")
	}
	Log.Debugf("read extra info...%v", extra)

	if extra.UnitPrice > re.Price {
		Log.Warnf("unit price is lower. unit price=%v, lowest unit price=%v", re.Price, extra.UnitPrice)
		return params, re, fmt.Errorf("the price is too low")
	}
	params.UnitPrice = re.Price
	re.Mode = bulletin.Mode
	re.SubMode = extra.SubMode
	params.Mode = bulletin.Mode
	params.SubMode = extra.SubMode
	if params.SubMode == TRANSACTION_SUB_MODE_BATCH2 {
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
		Log.Warnf("Failed to check whether the data(merkle root = %v) is in sale. err=%v", bulletin.SigmaMKLRoot, err)
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
	// 	Log.Warnf("Failed to save bulletin for seller. err=%v", err)
	// 	return params, re, errors.New("failed to save bulletin")
	// }
	// Log.Debugf("success to save publish file.")
	dir := BConf.SellerDir + "/transaction/" + sessionID
	err = os.Mkdir(dir, os.ModePerm)
	if err != nil {
		Log.Errorf("create folder %v error. err=%v", dir, err)
		return params, re, errors.New("failed to create folder")
	}
	Log.Debugf("success to create folder. dir=%v", dir)

	return params, re, nil
}

//sellerTxForPB1 is the transaction while mode is plain_range.
func sellerTxForPB1(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {

	requestFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/request"
	responseFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/response"
	receiptFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/receipt"
	secretFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/secret"

	defer func() {
		err := updateSellerTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for seller. err=%v", err)
			return
		}
		delete(SellerTxMap, tx.SessionID)
	}()

	err := sellerRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive transaction request for seller.")

	rs := tx.PlainBatch1.sellerVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify transaction request and generate transaction response for seller")

	err = sellerSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response for seller. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send transaction response for seller")

	var sign []byte
	sign, tx.Price, tx.ExpireAt, err = sellerRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive transaction receipt.")

	rs = tx.PlainBatch1.sellerVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify receipt and generate secret.")

	rs, err = verifyDeposit(tx.SellerAddr, tx.BuyerAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForBatch1(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send secret to contract. err=%v", err)
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForBatch1(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}
	DepositLockMap[tx.SellerAddr+tx.BuyerAddr] -= tx.Price

	tx.Status = TRANSACTION_STATUS_CLOSED
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract")
	return nil
}

//sellerTxForPOB1 is the transaction while mode is plain_ot_range.
func sellerTxForPOB1(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	dir := BConf.SellerDir + "/transaction/" + tx.SessionID
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	buyerNegoRequestFile := dir + "/buyer_nego_request"
	buyerNegoResponseFile := dir + "/buyer_nego_response"
	sellerNegoRequestFile := dir + "/seller_nego_request"
	sellerNegoResponseFile := dir + "/seller_nego_response"

	defer func() {
		err := updateSellerTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for seller. err=%v", err)
			return
		}
		delete(SellerTxMap, tx.SessionID)
	}()

	err := sellerReceiveNegoReq(node, buyerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego request. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego request")
	}
	Log.Debugf("success to receive transaction nego request.")

	rs := tx.PlainOTBatch1.sellerGeneNegoResp(buyerNegoRequestFile, sellerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid nego request file or nego response file. err=%v", err)
		return fmt.Errorf(
			"invalid nego request file or nego response file")
	}
	Log.Debugf("success to verify nego request and generate nego response.")

	rs = tx.PlainOTBatch1.sellerGeneNegoReq(sellerNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to generate nego request file. err=%v", err)
		return fmt.Errorf(
			"failed to generate nego request file")
	}
	Log.Debugf("success to generate nego request for seller.")

	err = sellerSendNegoResp(node, sellerNegoResponseFile, sellerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("success to send transaction nego response.")

	err = sellerRcvNegoResp(node, buyerNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("success to receive transaction nego response.")

	rs = tx.PlainOTBatch1.sellerDealNegoResp(buyerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to deal with nego response. err=%v", err)
		return fmt.Errorf(
			"failed to deal with nego response")
	}
	Log.Debugf("success to deal with nego response.")
	tx.Status = TRANSACTION_STATUS_NEGO
	SellerTxMap[tx.SessionID] = tx

	err = sellerRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	Log.Debugf("success to receive transaction request.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	SellerTxMap[tx.SessionID] = tx

	rs = tx.PlainOTBatch1.sellerVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid transaction request file or transaction response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file.")
	}
	Log.Debugf("success to verify transaction request and generate transaction response.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	SellerTxMap[tx.SessionID] = tx

	err = sellerSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response for seller. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	Log.Debugf("success to send transaction response.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	SellerTxMap[tx.SessionID] = tx

	var sign []byte
	sign, tx.Price, tx.ExpireAt, err = sellerRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	Log.Debugf("success to receive transaction receipt for seller.")
	tx.Status = TRANSACTION_STATUS_RECEIPT
	SellerTxMap[tx.SessionID] = tx

	rs = tx.PlainOTBatch1.sellerVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	Log.Debugf("success to verify receipt and generate secret.")
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	SellerTxMap[tx.SessionID] = tx

	rs, err = verifyDeposit(tx.SellerAddr, tx.BuyerAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForBatch1(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract. err=%v", err)
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForBatch1(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}
	DepositLockMap[tx.SellerAddr+tx.BuyerAddr] -= tx.Price

	tx.Status = TRANSACTION_STATUS_CLOSED
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract.")
	return nil
}

//sellerTxForPB2 is the transaction while mode is plain_range.
func sellerTxForPB2(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	requestFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/request"
	responseFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/response"
	receiptFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/receipt"
	secretFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/secret"

	defer func() {
		err := updateSellerTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for seller. err=%v", err)
			return
		}
		delete(SellerTxMap, tx.SessionID)
	}()

	err := sellerRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	Log.Debugf("success to receive transaction request.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	SellerTxMap[tx.SessionID] = tx

	rs := tx.PlainBatch2.sellerVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	Log.Debugf("success to verify request and generate response.")
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	SellerTxMap[tx.SessionID] = tx

	err = sellerSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response to buyer. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	Log.Debugf("success to send transaction response to buyer.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	SellerTxMap[tx.SessionID] = tx

	var sign []byte
	sign, tx.Price, tx.ExpireAt, err = sellerRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt from buyer. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	Log.Debugf("success to receive transaction receipt from buyer.")
	tx.Status = TRANSACTION_STATUS_RECEIPT
	SellerTxMap[tx.SessionID] = tx

	rs = tx.PlainBatch2.sellerVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	Log.Debugf("success to verify receipt file and generate secret file.")
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	SellerTxMap[tx.SessionID] = tx

	rs, err = verifyDeposit(tx.SellerAddr, tx.BuyerAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForBatch2(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract")
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForBatch2(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}
	DepositLockMap[tx.SellerAddr+tx.BuyerAddr] -= tx.Price

	tx.Status = TRANSACTION_STATUS_CLOSED
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract.")
	return nil
}

//sellerTxForTB1 is the transaction while mode is plain_range.
func sellerTxForTB1(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	dir := BConf.SellerDir + "/transaction/" + tx.SessionID
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"

	defer func() {
		err := updateSellerTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for seller. err=%v", err)
			return
		}
		delete(SellerTxMap, tx.SessionID)
	}()

	err := sellerRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	Log.Debugf("success to receive transaction request.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	SellerTxMap[tx.SessionID] = tx

	rs := tx.TableBatch1.sellerVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify transaction request and generate transaction response.")

	err = sellerSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response to buyer. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send transaction response to buyer.")

	var sign []byte
	sign, tx.Price, tx.ExpireAt, err = sellerRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive receipt from buyer.")

	rs = tx.TableBatch1.sellerVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify receipt and generate secret.")

	rs, err = verifyDeposit(tx.SellerAddr, tx.BuyerAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForBatch1(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract. err=%v", err)
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForBatch1(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}
	DepositLockMap[tx.SellerAddr+tx.BuyerAddr] -= tx.Price

	tx.Status = TRANSACTION_STATUS_CLOSED
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract")
	return nil
}

//sellerTxForTOB1 is the transaction while mode is plain_range.
func sellerTxForTOB1(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	dir := BConf.SellerDir + "/transaction/" + tx.SessionID
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	buyerNegoRequestFile := dir + "/buyer_nego_request"
	buyerNegoResponseFile := dir + "/buyer_nego_response"
	sellerNegoRequestFile := dir + "/seller_nego_request"
	sellerNegoResponseFile := dir + "/seller_nego_response"

	defer func() {
		err := updateSellerTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for seller. err=%v", err)
			return
		}
		delete(SellerTxMap, tx.SessionID)
	}()

	err := sellerReceiveNegoReq(node, buyerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego request. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego request")
	}
	Log.Debugf("receive transaction nego request...")

	rs := tx.TableOTBatch1.sellerGeneNegoResp(buyerNegoRequestFile, sellerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid nego request file or nego response file. err=%v", err)
		return fmt.Errorf(
			"invalid nego request file or nego response file")
	}
	Log.Debugf("generate nego request file or nego response file...")

	rs = tx.TableOTBatch1.sellerGeneNegoReq(sellerNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to generate nego request file. err=%v", err)
		return fmt.Errorf(
			"failed to generate nego request file")
	}
	Log.Debugf("seller generates nego request file...")

	err = sellerSendNegoResp(node, sellerNegoResponseFile, sellerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("send nego response and nego request ...")

	err = sellerRcvNegoResp(node, buyerNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("receive transaction nego response...")

	rs = tx.TableOTBatch1.sellerDealNegoResp(buyerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to deal with nego response. err=%v", err)
		return fmt.Errorf(
			"failed to deal with nego response")
	}
	tx.Status = TRANSACTION_STATUS_NEGO
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("deal with nego response...")

	err = sellerRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("receive transaction request for seller...")

	rs = tx.TableOTBatch1.sellerVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("verify transaction request file and GENERATE response file...")

	err = sellerSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response for seller. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("send transaction response...")

	var sign []byte
	sign, tx.Price, tx.ExpireAt, err = sellerRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("receive transaction receipt...")

	rs = tx.TableOTBatch1.sellerVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("verify receipt file and GENERATE secret file...")

	rs, err = verifyDeposit(tx.SellerAddr, tx.BuyerAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForBatch1(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract. err=%v", err)
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForBatch1(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}
	DepositLockMap[tx.SellerAddr+tx.BuyerAddr] -= tx.Price

	tx.Status = TRANSACTION_STATUS_CLOSED
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("send secret to contract...")
	return nil
}

//sellerTxForTB2 is the transaction while mode is plain_range.
func sellerTxForTB2(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	dir := BConf.SellerDir + "/transaction/" + tx.SessionID
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"

	defer func() {
		err := updateSellerTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for seller. err=%v", err)
			return
		}
		delete(SellerTxMap, tx.SessionID)
	}()

	err := sellerRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	Log.Debugf("success to receive transaction request.")
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	SellerTxMap[tx.SessionID] = tx

	rs := tx.TableBatch2.sellerVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify transaction request and generate response.")

	err = sellerSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response to buyer. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send transaction response to buyer.")

	var sign []byte
	sign, tx.Price, tx.ExpireAt, err = sellerRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction receipt from buyer. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive receipt from buyer.")

	rs = tx.TableBatch2.sellerVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify receipt and generate secret.")

	rs, err = verifyDeposit(tx.SellerAddr, tx.BuyerAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForBatch2(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract")
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForBatch2(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}
	DepositLockMap[tx.SellerAddr+tx.BuyerAddr] -= tx.Price

	tx.Status = TRANSACTION_STATUS_CLOSED
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract.")
	return nil
}

//sellerTxForTQ is the transaction while mode is plain_range.
func sellerTxForTQ(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	requestFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/request"
	responseFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/response"
	receiptFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/receipt"
	secretFile := BConf.SellerDir + "/transaction/" + tx.SessionID + "/secret"

	defer func() {
		err := updateSellerTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for seller. err=%v", err)
			return
		}
		delete(SellerTxMap, tx.SessionID)
	}()

	err := sellerRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive transaction request for seller.")

	rs := tx.TableVRF.sellerVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify request file and generate respons file.")

	err = sellerSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response for seller. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send transaction response for seller.")

	var sign []byte
	sign, tx.Price, tx.ExpireAt, err = sellerRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive receipt from buyer. err=%v", err)
		return fmt.Errorf(
			"failed to receive receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIPT
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive receipt from buyer.")

	rs = tx.TableVRF.sellerVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify receipt.")

	rs, err = verifyDeposit(tx.SellerAddr, tx.BuyerAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForVRFQ(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract")
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForVRFQ(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}
	DepositLockMap[tx.SellerAddr+tx.BuyerAddr] -= tx.Price

	tx.Status = TRANSACTION_STATUS_CLOSED
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract.")
	return nil
}

//sellerTxForTOQ is the transaction while mode is plain_range.
func sellerTxForTOQ(node *pod_net.Node, key *keystore.Key, tx Transaction, Log ILogger) error {
	dir := BConf.SellerDir + "/transaction/" + tx.SessionID
	requestFile := dir + "/request"
	responseFile := dir + "/response"
	receiptFile := dir + "/receipt"
	secretFile := dir + "/secret"
	buyerNegoRequestFile := dir + "/buyer_nego_request"
	buyerNegoResponseFile := dir + "/buyer_nego_response"
	sellerNegoRequestFile := dir + "/seller_nego_request"
	sellerNegoResponseFile := dir + "/seller_nego_response"

	defer func() {
		err := updateSellerTxToDB(tx)
		if err != nil {
			Log.Warnf("failed to update transaction for seller. err=%v", err)
			return
		}
		delete(SellerTxMap, tx.SessionID)
	}()

	err := sellerReceiveNegoReq(node, buyerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego request. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego request")
	}
	Log.Debugf("success to receive transaction nego request.")

	rs := tx.TableOTVRF.sellerGeneNegoResp(buyerNegoRequestFile, sellerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid nego request file or nego response file. err=%v", err)
		return fmt.Errorf(
			"invalid nego request file or nego response file")
	}
	Log.Debugf("success to generate nego request and generate nego response")

	rs = tx.TableOTVRF.sellerGeneNegoReq(sellerNegoRequestFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to generate nego request file. err=%v", err)
		return fmt.Errorf(
			"failed to generate nego request file")
	}
	Log.Debugf("success to generate nego request for seller")

	err = sellerSendNegoResp(node, sellerNegoResponseFile, sellerNegoRequestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("success to send transaction nego response")

	err = sellerRcvNegoResp(node, buyerNegoResponseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction nego response. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction nego response")
	}
	Log.Debugf("success to receive transaction nego response")

	rs = tx.TableOTVRF.sellerDealNegoResp(buyerNegoResponseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_NEGO_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to deal with nego response. err=%v", err)
		return fmt.Errorf(
			"failed to deal with nego response")
	}
	tx.Status = TRANSACTION_STATUS_NEGO
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to deal with nego response.")

	err = sellerRcvPODReq(node, requestFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive transaction request from buyer. err=%v", err)
		return fmt.Errorf(
			"failed to receive transaction request")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_REQUEST
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive transaction request from buyer.")

	rs = tx.TableOTVRF.sellerVerifyReq(requestFile, responseFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_INVALID_REQUEST
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid request file or response file. err=%v", err)
		return fmt.Errorf(
			"invalid request file or response file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify request and generate response.")

	err = sellerSendPODResp(node, responseFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_RESPONSE_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send transaction response to buyer. err=%v", err)
		return fmt.Errorf(
			"failed to send transaction response")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send transaction response to buyer")

	var sign []byte
	sign, tx.Price, tx.ExpireAt, err = sellerRcvPODRecpt(node, receiptFile)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to receive receipt for seller. err=%v", err)
		return fmt.Errorf(
			"failed to receive receipt")
	}
	tx.Status = TRANSACTION_STATUS_RECEIVED_RESPONSE
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to receive receipt for seller")

	rs = tx.TableOTVRF.sellerVerifyReceipt(receiptFile, secretFile, Log)
	if !rs {
		tx.Status = TRANSACTION_STATUS_GENERATE_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("invalid receipt file or secret file. err=%v", err)
		return fmt.Errorf(
			"invalid receipt file or secret file")
	}
	tx.Status = TRANSACTION_STATUS_GENERATE_SECRET
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to verify receipt and generate secret")

	rs, err = verifyDeposit(tx.SellerAddr, tx.BuyerAddr, tx.Price)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to verify deposit eth. err=%v", err)
		return fmt.Errorf(
			"failed to verify deposit eth")
	}
	if !rs {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("no enough deposit eth. err=%v", err)
		return fmt.Errorf(
			"no enough deposit eth")
	}

	if time.Now().Unix()+600 > tx.ExpireAt {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_TERMINATED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("the receipt signature will timeout soon.")
		return fmt.Errorf(
			"the receipt signature timeout")
	}

	Log.Debugf("start send transaction to submit contract from contract...")
	t := time.Now()
	txid, err := submitScrtForVRFQ(tx, sign, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to send Secret to contract")
		return fmt.Errorf(
			"failed to send secret")
	}
	Log.Debugf("success to submit secret to contract...txid=%v, time cost=%v", txid, time.Since(t))

	_, err = readScrtForVRFQ(tx.SessionID, tx.SellerAddr, tx.BuyerAddr, Log)
	if err != nil {
		tx.Status = TRANSACTION_STATUS_SEND_SECRET_FAILED
		SellerTxMap[tx.SessionID] = tx
		Log.Warnf("failed to read secret from contract.")
		return fmt.Errorf(
			"failed to send secret")
	}
	DepositLockMap[tx.SellerAddr+tx.BuyerAddr] -= tx.Price

	tx.Status = TRANSACTION_STATUS_CLOSED
	SellerTxMap[tx.SessionID] = tx
	Log.Debugf("success to send secret to contract")
	return nil
}
