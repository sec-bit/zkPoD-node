package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"

	pod_net "github.com/sec-bit/zkPoD-node/net"
	"github.com/sec-bit/zkPoD-node/net/rlpx"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// SellerStartNode starts p2p node for seller node,
// listen to request from buyer
func SellerStartNode(sellerIPAddr string, key *keystore.Key, Log ILogger) error {

	var wg sync.WaitGroup

	serveAddr, err := rlpx.NewAddr(sellerIPAddr, key.PrivateKey.PublicKey)
	if err != nil {
		Log.Errorf("failed to initialize seller's server address. err=%v", err)
		return fmt.Errorf("failed to start seller's node")
	}
	// Log.Debugf("Initialize seller's server address finish.")

	l, err := rlpx.Listen(serveAddr)
	if err != nil {
		Log.Errorf("failed to listen on %s: %v", serveAddr, err)
		return fmt.Errorf("failed to start seller's node")
	}
	defer func() {
		if err := l.Close(); err != nil {
			Log.Errorf("failed to close listener on %s: %v",
				serveAddr, err)
		}
		if err := recover(); err != nil {
			Log.Errorf("exception unexpected: error=%v", err)
			return
		}
	}()
	SellerNodeStart = true
	fmt.Printf("===>>>Listen to %v\n\n", serveAddr)

	for {
		//TODO: reload
		conn, err := l.Accept()
		if err != nil {
			Log.Errorf("failed to accept connection on %s: %v",
				serveAddr, err)
			continue
		}
		wg.Add(1)
		go func() {
			sellerAcceptTx(&wg, conn, key, Log)
			if err := conn.Close(); err != nil {
				Log.Errorf("failed to close connection on server side: %v",
					err)
				return
			}
		}()
	}
	wg.Wait()
	return nil
}

//sellerAcceptTx connects with buyer and handle transaction for seller.
func sellerAcceptTx(wg *sync.WaitGroup, conn *rlpx.Connection, key *keystore.Key, Log ILogger) {
	Log.Debugf("start connect with buyer node....")
	defer func() {
		wg.Done()
	}()

	node, rkey, params, err := preSellerTxAndConn(conn, key, Log)
	if err != nil {
		Log.Warnf("failed to prepare for transaction connection. err=%v", err)
		return
	}
	defer func() {
		if err := node.Close(); err != nil {
			Log.Errorf("failed to close server node: %v", err)
			return
		}
	}()
	Log.Debugf("[%v]step0: prepare for transaction successfully....", params.SessionID)

	var tx Transaction
	tx.SessionID = params.SessionID
	tx.Status = TRANSACTION_STATUS_START
	tx.BuyerPubKey = rkey
	tx.BuyerAddr = crypto.PubkeyToAddress(*rkey).Hex()
	tx.Bulletin = params.Bulletin
	tx.Mode = params.Mode
	tx.SubMode = params.SubMode
	tx.OT = params.OT
	tx.UnitPrice = params.UnitPrice
	tx.SellerAddr = key.Address.Hex()

	SellerTxMap[tx.SessionID] = tx
	err = insertSellerTxToDB(tx)
	if err != nil {
		Log.Warnf("[%v]failed to save transaction to db for seller. err=%v", params.SessionID, err)
		return
	}

	publishPath := BConf.SellerDir + "/transaction/" + params.SessionID

	if tx.Mode == TRANSACTION_MODE_PLAIN_POD {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_BATCH1:
			if tx.OT {
				tx.PlainOTBatch1, err = sellerNewSessForPOB1(publishPath, Log)
				if err != nil {
					Log.Warnf("failed to prepare for seller's session. err=%v", err)
					return
				}
				defer func() {
					tx.PlainOTBatch1.SellerSession.Free()
				}()
				Log.Debugf("success to prepare seller session for plain_ot_batch1")

				err = sellerTxForPOB1(node, key, tx, Log)
				if err != nil {
					Log.Warnf("transaction error. err=%v", err)
					return
				}
				Log.Debugf("transaction finish...")
			} else {
				tx.PlainBatch1, err = sellerNewSessForPB1(publishPath, Log)
				if err != nil {
					Log.Warnf("failed to prepare for seller's session. err=%v", err)
					return
				}
				defer func() {
					tx.PlainBatch1.SellerSession.Free()
				}()
				Log.Debugf("success to prepare seller session for plain_batch1")

				err = sellerTxForPB1(node, key, tx, Log)
				if err != nil {
					Log.Warnf("transaction error. err=%v", err)
					return
				}
				Log.Debugf("transaction finish...")
			}
		case TRANSACTION_SUB_MODE_BATCH2:
			tx.PlainBatch2, err = sellerNewSessForPB2(publishPath, Log)
			if err != nil {
				Log.Warnf("Failed to prepare for seller's session. err=%v", err)
				return
			}
			defer func() {
				tx.PlainBatch2.SellerSession.Free()
			}()
			Log.Debugf("success to prepare seller session for plain_batch2")

			err = sellerTxForPB2(node, key, tx, Log)
			if err != nil {
				Log.Warnf("transaction error. err=%v", err)
				return
			}
			Log.Debugf("transaction finish...")
		}
	} else if tx.Mode == TRANSACTION_MODE_TABLE_POD {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_BATCH1:
			if tx.OT {
				tx.TableOTBatch1, err = sellerNewSessForTOB1(publishPath, Log)
				if err != nil {
					Log.Warnf("Failed to prepare for seller's session. err=%v", err)
					return
				}
				defer func() {
					tx.TableOTBatch1.SellerSession.Free()
				}()
				Log.Debugf("success to prepare seller session for table_ot_batch1")

				err = sellerTxForTOB1(node, key, tx, Log)
				if err != nil {
					Log.Warnf("transaction error. err=%v", err)
					return
				}
				Log.Debugf("transaction finish...")
			} else {
				tx.TableBatch1, err = sellerNewSessForTB1(publishPath, Log)
				if err != nil {
					Log.Warnf("Failed to prepare for seller's session. err=%v", err)
					return
				}
				defer func() {
					tx.TableBatch1.SellerSession.Free()
				}()
				Log.Debugf("success to prepare seller session for table_batch1")

				err = sellerTxForTB1(node, key, tx, Log)
				if err != nil {
					Log.Warnf("transaction error. err=%v", err)
					return
				}
				Log.Debugf("transaction finish...")
			}
		case TRANSACTION_SUB_MODE_BATCH2:
			tx.TableBatch2, err = sellerNewSessForTB2(publishPath, Log)
			if err != nil {
				Log.Warnf("Failed to prepare for seller's session. err=%v", err)
				return
			}
			defer func() {
				tx.TableBatch2.SellerSession.Free()
			}()
			Log.Debugf("success to prepare seller session for table_batch2")

			err = sellerTxForTB2(node, key, tx, Log)
			if err != nil {
				Log.Warnf("transaction error. err=%v", err)
				return
			}
			Log.Debugf("transaction finish...")
		case TRANSACTION_SUB_MODE_VRF:
			if tx.OT {
				tx.TableOTVRF, err = sellerNewSessForTOQ(publishPath, Log)
				if err != nil {
					Log.Warnf("Failed to prepare for seller's session. err=%v", err)
					return
				}
				defer func() {
					tx.TableOTVRF.SellerSession.Free()
				}()
				Log.Debugf("success to prepare seller session for table_ot_vrf")

				err = sellerTxForTOQ(node, key, tx, Log)
				if err != nil {
					Log.Warnf("transaction error. err=%v", err)
					return
				}
				Log.Debugf("transaction finish...")
			} else {
				tx.TableVRF, err = sellerNewSessForTQ(publishPath, Log)
				if err != nil {
					Log.Warnf("Failed to prepare for seller's session. err=%v", err)
					return
				}
				defer func() {
					tx.TableVRF.SellerSession.Free()
				}()
				Log.Debugf("success to prepare seller session for table_vrf")

				err = sellerTxForTQ(node, key, tx, Log)
				if err != nil {
					Log.Warnf("transaction error. err=%v", err)
					return
				}
				Log.Debugf("transaction finish...")
			}
		}
	}
}

type SellerConnParam struct {
	Mode      string
	SubMode   string
	OT        bool
	UnitPrice int64
	SessionID string
	Bulletin  Bulletin
}

func preSellerTxAndConn(conn *rlpx.Connection, key *keystore.Key, Log ILogger) (node *pod_net.Node, rkey *ecdsa.PublicKey, params SellerConnParam, err error) {

	node, rkey, err = sellerNewConn(conn, key, Log)
	if err != nil {
		Log.Warnf("%v", err)
		return
	}
	Log.Debugf("success to new connection...")

	req, re, err := sellerRcvSessReq(node, Log)
	if err != nil {
		node.Close()
		Log.Warnf("failed to receive session request. err=%v", err)
		return
	}

	mklroot := hex.EncodeToString(req.SigmaMklRoot)
	params, re, err = preSellerTx(mklroot, re, Log)
	if err != nil {
		node.Close()
		Log.Warnf("failed to prepare for transaction. err=%v", err)
		err = fmt.Errorf("failed to prepare for transaction")
		return
	}
	Log.Debugf("[%v]success to prepare for transaction...", params.SessionID)

	req.ExtraInfo, err = json.Marshal(&re)
	if err != nil {
		node.Close()
		Log.Warnf("failed to marshal extra info. err=%v")
		err = fmt.Errorf("failed to save extra info")
		return
	}

	err = preSellerConn(node, key, params, req, Log)
	if err != nil {
		node.Close()
		Log.Errorf("failed to establish session with buyer. err=%v", err)
		return
	}
	Log.Debugf("[%v]established connection session successfully....", params.SessionID)
	return
}

func sellerNewConn(conn *rlpx.Connection, key *keystore.Key, Log ILogger) (*pod_net.Node, *ecdsa.PublicKey, error) {

	rkey, err := conn.Handshake(key.PrivateKey, false)
	if err != nil {
		Log.Errorf("failed to server-side handshake: %v", err)
		return nil, rkey, errors.New("failed to server-side handshake")
	}
	Log.Debugf("establish connection's handshake successfully...")

	node, err := pod_net.NewNode(conn, key.PrivateKey, rkey)
	if err != nil {
		Log.Errorf("failed to create server node: %v", err)
		return nil, rkey, errors.New("failed to create server node")
	}
	Log.Debugf("create connection node successfully....")
	return node, rkey, nil
}

type requestExtra struct {
	Price   int64  `json:"price"`
	Mode    string `json:"mode"`
	SubMode string `json:"subMode"`
	Ot      bool   `json:"ot"`
}

func sellerRcvSessReq(node *pod_net.Node, Log ILogger) (req *pod_net.SessionRequest, re requestExtra, err error) {
	req, err = node.RecvSessionRequest()
	if err != nil {
		Log.Warnf("failed to receive session request. err=%v", err)
		err = fmt.Errorf("failed to receive session request")
		return
	}
	if req.ID != 0 {
		Log.Warnf("session ID (%d) not zero", req.ID)
		err = fmt.Errorf("session ID not zero")
		return
	}
	Log.Debugf("success to receive session request...")
	err = json.Unmarshal(req.ExtraInfo, &re)
	if err != nil {
		Log.Warnf("failed to parse extra info. err=%v")
		err = fmt.Errorf("failed to parse extra info")
		return
	}
	return
}

func preSellerConn(node *pod_net.Node, key *keystore.Key, params SellerConnParam, req *pod_net.SessionRequest, Log ILogger) (err error) {

	/////////////////////////RecvSessionRequest/////////////////////////
	sessionIDInt, err := strconv.ParseUint(params.SessionID, 16, 64)
	if err != nil {
		Log.Warnf("failed to convert sessionID. err=%v", err)
		err = fmt.Errorf("failed to convert sessionID")
		return
	}

	netMode, err := modeToInt(params.Mode, params.SubMode, params.OT)
	if err != nil {
		Log.Warnf("failed to convert mode to netMode. mode=%v, subMode=%v, ot=%v", params.Mode, params.SubMode, params.OT)
		err = fmt.Errorf("failed to convert mode to netMode")
		return
	}

	/////////////////////////SendSessionAck/////////////////////////
	if err = node.SendSessionAck(
		sessionIDInt, netMode, req.SigmaMklRoot, req.ExtraInfo, true,
	); err != nil {
		err = fmt.Errorf(
			"failed to send session ack from server: %v",
			err)
		return
	}
	Log.Debugf("success to send session ack...")
	// if node.state != pod_net.stateSessionAckWait {
	// 	return mode,mklroot,sessionID,fmt.Errorf(
	// 		"server node not in stateSessionAckWait state")
	// }

	/////////////////////////RecvSessionAck/////////////////////////
	ack, err := node.RecvSessionAck(false)
	if err != nil {
		err = fmt.Errorf(
			"failed to receive session ack on server node: %v",
			err)
		return
	}
	if ack.ID != sessionIDInt {
		err = fmt.Errorf(
			"mismatch session ID on server node, get %d, expect %d",
			ack.ID, sessionIDInt)
		return
	}
	Log.Debugf("success to receive session ack...")
	// if node.session.id != sessionIDInt {
	// 	return mode, mklroot, sessionID, fmt.Errorf(
	// 		"server session ID %d != %d",
	// 		node.session.id, sessionID)
	// }
	// if node.session.mode != req.Mode {
	// 	return mode, mklroot, sessionID, fmt.Errorf(
	// 		"server session mode %d != %d",
	// 		node.session.mode, req.Mode)
	// }
	// if node.state != pod_net.stateSessionEstablished {
	// 	return mode,mklroot,sessionID,fmt.Errorf("server node not in stateSessionEstablished state")
	// }
	return
}

func modeToInt(mode string, subMode string, ot bool) (netMode uint8, err error) {

	if mode == TRANSACTION_MODE_PLAIN_POD {
		switch subMode {
		case TRANSACTION_SUB_MODE_BATCH1:
			if !ot {
				netMode = pod_net.ModePlainBatchPoD
			} else {
				netMode = pod_net.ModePlainOTBatchPoD
			}
		case TRANSACTION_SUB_MODE_BATCH2:
			netMode = pod_net.ModePlainBatch2PoD
		default:
			err = errors.New("invalid mode")
		}
	} else if mode == TRANSACTION_MODE_TABLE_POD {
		switch subMode {
		case TRANSACTION_SUB_MODE_BATCH1:
			if !ot {
				netMode = pod_net.ModeTableBatchPoD
			} else {
				netMode = pod_net.ModeTableOTBatchPoD
			}
		case TRANSACTION_SUB_MODE_BATCH2:
			netMode = pod_net.ModeTableBatch2PoD
		case TRANSACTION_SUB_MODE_VRF:
			if !ot {
				netMode = pod_net.ModeTableVRFQuery
			} else {
				netMode = pod_net.ModeTableOTVRFQuery
			}
		default:
			err = errors.New("invalid mode")
		}
	} else {
		err = errors.New("invalid mode")
	}
	return
}

func modeFromInt(netMode uint8) (mode string, subMode string, ot bool, err error) {
	switch netMode {
	case pod_net.ModePlainBatchPoD:
		mode = TRANSACTION_MODE_PLAIN_POD
		subMode = TRANSACTION_SUB_MODE_BATCH1
		ot = false
	case pod_net.ModePlainOTBatchPoD:
		mode = TRANSACTION_MODE_PLAIN_POD
		subMode = TRANSACTION_SUB_MODE_BATCH1
		ot = true
	case pod_net.ModePlainBatch2PoD:
		mode = TRANSACTION_MODE_PLAIN_POD
		subMode = TRANSACTION_SUB_MODE_BATCH2
		ot = false
	case pod_net.ModeTableBatchPoD:
		mode = TRANSACTION_MODE_TABLE_POD
		subMode = TRANSACTION_SUB_MODE_BATCH1
		ot = false
	case pod_net.ModeTableOTBatchPoD:
		mode = TRANSACTION_MODE_TABLE_POD
		subMode = TRANSACTION_SUB_MODE_BATCH1
		ot = true
	case pod_net.ModeTableBatch2PoD:
		mode = TRANSACTION_MODE_TABLE_POD
		subMode = TRANSACTION_SUB_MODE_BATCH2
		ot = false
	case pod_net.ModeTableVRFQuery:
		mode = TRANSACTION_MODE_TABLE_POD
		subMode = TRANSACTION_SUB_MODE_VRF
		ot = false
	case pod_net.ModeTableOTVRFQuery:
		mode = TRANSACTION_MODE_TABLE_POD
		subMode = TRANSACTION_SUB_MODE_VRF
		ot = true
	default:
		err = fmt.Errorf("invalid mode=%v", netMode)
	}
	return
}

func sellerRcvPODReq(node *pod_net.Node, requestFile string) error {
	reqBuf := new(bytes.Buffer)
	if _, err := node.RecvTxRequest(reqBuf); err != nil {
		return err
	}
	// if node.state != pod_net.stateTxRequestRecvd {
	// 	return fmt.Errorf("server node not in TxRequestRecvd state")
	// }

	reqf, err := os.OpenFile(requestFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}
	defer reqf.Close()

	_, err = reqf.Write(reqBuf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write to request to file: %v", err)
	}
	return nil
}

func sellerSendPODResp(node *pod_net.Node, responseFile string) error {
	txResponse, err := ioutil.ReadFile(responseFile)
	if err != nil {
		return fmt.Errorf("failed to read response file: %v", err)
	}
	if err := node.SendTxResponse(
		bytes.NewReader(txResponse), uint64(len(txResponse)),
	); err != nil {
		return err
	}
	// if node.state != pod_net.stateTxReceiptWait {
	// 	return fmt.Errorf(
	// 		"server node not in TxReceiptWait state")
	// }
	return nil
}

func sellerRcvPODRecpt(node *pod_net.Node, receiptFile string) (receiptSign []byte, price int64, expireAt int64, err error) {
	receipt, _, err := node.RecvTxReceipt()
	if err != nil {
		return
	}
	var receiptConn ReceiptForConnection
	err = json.Unmarshal(receipt, &receiptConn)
	if err != nil {
		err = fmt.Errorf("failed to parse receipt. err=%v", err)
		return
	}
	price = receiptConn.Price
	expireAt = receiptConn.ExpireAt
	receiptSign = receiptConn.ReceiptSign

	recf, err := os.OpenFile(receiptFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		err = fmt.Errorf("failed to save file: %v", err)
		return
	}
	defer recf.Close()

	_, err = recf.Write(receiptConn.ReceiptByte)
	if err != nil {
		err = fmt.Errorf("failed to write to receipt to file: %v", err)
		return
	}
	return
}

func sellerReceiveNegoReq(node *pod_net.Node, buyerNegoRequestFile string) error {
	reqBuf := new(bytes.Buffer)
	if _, err := node.RecvNegoRequest(reqBuf); err != nil {
		return fmt.Errorf(
			"failed to receive negotiation request: %v",
			err)
	}
	reqf, err := os.OpenFile(buyerNegoRequestFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}
	defer reqf.Close()

	_, err = reqf.Write(reqBuf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write to nego request to file: %v", err)
	}
	return nil
}

func sellerSendNegoResp(node *pod_net.Node, negoResponseFile string, negoRequestFile string) error {
	txResponse, err := ioutil.ReadFile(negoResponseFile)
	if err != nil {
		return fmt.Errorf("failed to read response file: %v", err)
	}
	txRequest, err := ioutil.ReadFile(negoRequestFile)
	if err != nil {
		return fmt.Errorf("failed to read ack file: %v", err)
	}

	if err := node.SendNegoAckReq(
		bytes.NewReader(txResponse),
		bytes.NewReader(txRequest),
		uint64(len(txResponse)),
		uint64(len(txRequest)),
	); err != nil {
		return fmt.Errorf(
			"failed to send nego ack+req: %v", err)
	}
	return nil
}

func sellerRcvNegoResp(node *pod_net.Node, buyerNegoResponseFile string) error {

	negoRespBuf := new(bytes.Buffer)
	if _, err := node.RecvNegoAck(negoRespBuf); err != nil {
		return fmt.Errorf(
			"failed to receive nego ack: %v", err)
	}
	// if node.state != pod_net.stateNegotiated {
	// 	return fmt.Errorf(
	// 		"server node not in Negotiated state")
	// }

	reqf, err := os.OpenFile(buyerNegoResponseFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}
	defer reqf.Close()

	_, err = reqf.Write(negoRespBuf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write to nego request to file: %v", err)
	}
	return nil
}

/////////////////////////////////Buyer////////////////////////////////////////
type BuyerConnParam struct {
	SellerIPAddr string
	SellerAddr   string
	Mode         string
	SubMode      string
	OT           bool
	UnitPrice    int64
	SessionID    string
	MerkleRoot   string
}

func preBuyerConn(params BuyerConnParam, key *keystore.Key, Log ILogger) (*pod_net.Node, *rlpx.Connection, BuyerConnParam, error) {
	node, conn, err := buyNewConn(params.SellerIPAddr, params.SellerAddr, key, Log)
	if err != nil {
		Log.Warnf("failed to new connection. err=%v", err)
		return node, conn, params, errors.New("failed to new connection")
	}
	params.SessionID, params.Mode, params.SubMode, params.OT, err = buyerCreateSess(node, params.MerkleRoot, params.Mode, params.SubMode, params.OT, params.UnitPrice, Log)
	if err != nil {
		Log.Warnf("failed to create net session. err=%v", err)
		return node, conn, params, errors.New("failed to create net session")
	}
	return node, conn, params, nil
}

func buyNewConn(sellerIPAddr string, sellerAddr string, key *keystore.Key, Log ILogger) (*pod_net.Node, *rlpx.Connection, error) {

	Log.Debugf("sellerIPAddr=%v", sellerIPAddr)
	Log.Debugf("PublicKey=%v", key.PrivateKey.PublicKey)
	commonAddr := common.HexToAddress(sellerAddr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", sellerIPAddr)
	if err != nil {
		return nil, nil, err
	}
	serveAddr := &rlpx.Addr{
		TCPAddr: tcpAddr,
		EthAddr: commonAddr,
	}
	Log.Debugf("serveAddr=%v", serveAddr)

	conn, err := rlpx.Dial(serveAddr)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to dial %s: %v", sellerIPAddr, err)
	}
	Log.Debugf("create dial with seller successfully. sellerAddr=%v, sellerIP=%v. ", serveAddr, sellerIPAddr)

	rkey, err := conn.Handshake(key.PrivateKey, true)
	if err != nil {
		return nil, nil, fmt.Errorf("client-side handshake failed: %v", err)
	}
	Log.Debugf("establish connection handshake successfully...")

	node, err := pod_net.NewNode(conn, key.PrivateKey, rkey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create client node: %v", err)
	}
	Log.Debugf("connect to seller...")
	return node, conn, nil
}

func buyerCreateSess(node *pod_net.Node, mklroot string, mode string, subMode string, ot bool, unitPrice int64, Log ILogger) (string, string, string, bool, error) {
	mklrootByte, err := hex.DecodeString(mklroot)
	if err != nil {
		return "", mode, subMode, ot, fmt.Errorf(
			"failed to decode merkle root: %v",
			err)
	}

	var extra requestExtra
	extra.Price = unitPrice
	extra.Mode = mode
	extra.Ot = ot
	extraByte, err := json.Marshal(&extra)
	if err != nil {
		return "", mode, subMode, ot, fmt.Errorf(
			"failed to decode extra info: %v",
			err)
	}
	Log.Debugf("extra = %v", string(extraByte))

	/////////////////////////SendNewSessionRequest/////////////////////////
	if err = node.SendNewSessionRequest(
		uint8(0), mklrootByte, extraByte,
	); err != nil {
		return "", mode, subMode, ot, fmt.Errorf(
			"failed to send session request: %v",
			err)
	}
	Log.Debugf("success to send session request...")
	// if node.state != pod_net.stateSessionAckWait {
	// 	return fmt.Errorf(
	// 		"client node not in SessionAckWait state")
	// }

	/////////////////////////RecvSessionAck/////////////////////////
	ack, err := node.RecvSessionAck(true)
	if err != nil {
		return "", mode, subMode, ot, fmt.Errorf(
			"failed to receive session ack on client node: %v",
			err)
	}
	Log.Debugf("success to receive session ack...%v", ack.ID)

	mode, subMode, ot, err = modeFromInt(ack.Mode)
	if err != nil {
		return "", mode, subMode, ot, fmt.Errorf(
			"invalid net mode: %v",
			ack.Mode)
	}

	// if node.state != pod_net.stateSessionAckRecvd {
	// 	return fmt.Errorf("client node not in stateSessionAckRecvd state")
	// }

	/////////////////////////SendSessionAck/////////////////////////
	if err := node.SendSessionAck(
		ack.ID, ack.Mode, mklrootByte, ack.ExtraInfo, false,
	); err != nil {
		return "", mode, subMode, ot, fmt.Errorf(
			"failed to send session ack from client: %v",
			err)
	}
	Log.Debugf("success to send session ack...")
	// if node.session.id != ack.ID {
	// 	return fmt.Errorf(
	// 		"client session ID %d != %d",
	// 		node.session.id, ack.ID)
	// }
	// if node.session.mode != ack.Mode {
	// 	return fmt.Errorf(
	// 		"client session mode %d != %d",
	// 		node.session.mode, ack.Mode)
	// }
	// if node.state != pod_net.stateSessionEstablished {
	// 	return fmt.Errorf(
	// 		"client node not in stateSessionEstablished state")
	// }

	sessionID := fmt.Sprintf("%x", ack.ID)
	return sessionID, mode, subMode, ot, nil
}

func buyerSendPODReq(node *pod_net.Node, requestFile string) error {
	txReq, err := ioutil.ReadFile(requestFile)
	if err != nil {
		return fmt.Errorf("failed to read transaction request file: %v", err)
	}
	if err = node.SendTxRequest(bytes.NewReader(txReq), uint64(len(txReq))); err != nil {
		return fmt.Errorf(
			"failed to send Tx request: %v", err)
	}
	// if node.state != pod_net.stateTxResponseWait {
	// 	return fmt.Errorf(
	// 		"client node not in TxResponse state")
	// }
	return nil
}

func buyerRcvPODResp(node *pod_net.Node, responseFile string) error {

	RespBuf := new(bytes.Buffer)
	_, err := node.RecvTxResponse(RespBuf)
	if err != nil {
		return err
	}
	// if node.state != pod_net.stateTxResponseRecvd {
	// 	return fmt.Errorf(
	// 		"client node not in TxResponseRecvd state")
	// }
	respf, err := os.OpenFile(responseFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to save file")
	}
	defer respf.Close()

	_, err = respf.Write(RespBuf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write to nego request to file: %v", err)
	}
	return nil
}

type ReceiptForConnection struct {
	ReceiptByte []byte `json:"receiptByte"`
	ReceiptSign []byte `json:"receiptSign"`
	Price       int64  `json:"price"`
	ExpireAt    int64  `json:"expireAt"`
}

func buyerSendPODRecpt(node *pod_net.Node, price int64, expireAt int64, receiptByte []byte, sign []byte) error {

	var receipt ReceiptForConnection
	receipt.ReceiptByte = receiptByte
	receipt.ReceiptSign = sign
	receipt.Price = price
	receipt.ExpireAt = expireAt
	receiptConnBytes, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	err = node.SendTxReceipt(bytes.NewReader(receiptConnBytes), uint64(len(receiptConnBytes)))
	if err != nil {
		return err
	}
	return nil
}

func buyerSendNegoReq(node *pod_net.Node, negoRequestFile string) error {

	buyerNegoReq, err := ioutil.ReadFile(negoRequestFile)
	if err != nil {
		return fmt.Errorf("failed to read transaction receipt file: %v", err)
	}

	if err := node.SendNegoRequest(
		bytes.NewReader(buyerNegoReq),
		uint64(len(buyerNegoReq)),
	); err != nil {
		return fmt.Errorf(
			"failed to send negotiation request: %v",
			err)
	}

	return nil
}

func buyerRcvNegoResp(node *pod_net.Node, negoResponseFile string, negoAckFile string) error {
	respBuf := new(bytes.Buffer)
	ackBuf := new(bytes.Buffer)
	if _, _, err := node.RecvNegoAckReq(
		respBuf, ackBuf,
	); err != nil {
		return fmt.Errorf(
			"failed to receive nego ack+req: %v",
			err)
	}

	respf, err := os.OpenFile(negoResponseFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to save file")
	}
	defer respf.Close()

	_, err = respf.Write(respBuf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write to nego request to file")
	}

	ackf, err := os.OpenFile(negoAckFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to save file")
	}
	defer ackf.Close()

	_, err = ackf.Write(ackBuf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write to nego request to file")
	}
	return nil
}

func buyerSendNegoResp(node *pod_net.Node, negoResponseFile string) error {

	buyerNegoResp, err := ioutil.ReadFile(negoResponseFile)
	if err != nil {
		return fmt.Errorf("failed to read transaction receipt file: %v", err)
	}

	if err := node.SendNegoAck(
		bytes.NewReader(buyerNegoResp),
		uint64(len(buyerNegoResp)),
	); err != nil {
		return fmt.Errorf(
			"failed to send nego ack: %v", err)
	}
	return nil
}
