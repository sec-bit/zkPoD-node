package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"

	_ "github.com/mattn/go-sqlite3"
)

var DBConn *xorm.Engine

func connectSQLite(dbpath string) (*xorm.Engine, error) {
	Log := Logger.NewSessionLogger()

	db, err := xorm.NewEngine("sqlite3", dbpath)
	if err != nil {
		Log.Errorf("failed to connect to db:%v", err.Error())
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		Log.Errorf("failed to connect to db: %v", err.Error())
		return nil, err
	}
	Log.Debugf("ping db success")
	db.SetMapper(core.SnakeMapper{})
	db.ShowSQL(false)
	db.Logger().SetLevel(core.LOG_DEBUG)

	err = db.CreateTables(&BobTx{})
	if err != nil {
		Log.Errorf("failed to create table Bob_tx, err=%v", err)
		return db, errors.New("failed to create table Bob_tx")
	}
	Log.Debugf("initialize table Bob_tx success")

	err = db.CreateTables(&AliceTx{})
	if err != nil {
		Log.Errorf("failed to create table Alice_tx, err=%v", err)
		return db, errors.New("failed to create table Alice_tx")
	}
	Log.Debugf("initialize table Alice_tx success")

	err = db.CreateTables(&PodLog{})
	if err != nil {
		Log.Errorf("failed to create table pod_log, err=%v", err)
		return db, errors.New("failed to create table pod_log")
	}
	Log.Debugf("initialize table pod_log success")

	err = db.Sync2(new(BobTx))
	if err != nil {
		Log.Warnf("table Bob_tx sync error, err=%v", err)
		return nil, err
	}
	err = db.Sync2(new(AliceTx))
	if err != nil {
		Log.Warnf("table Alice_tx sync error, err=%v", err)
		return nil, err
	}
	err = db.Sync2(new(PodLog))
	if err != nil {
		Log.Warnf("table pod_log sync error, err=%v", err)
		return nil, err
	}
	db.ShowSQL(false)
	db.ShowExecTime(true)
	return db, err
}

type BobTx struct {
	SessionId        string    `json:"sessionId" xorm:"text pk not null"`
	Status           string    `json:"status" xorm:"text not null"`
	AliceIP          string    `json:"AliceIP" xorm:"text not null"`
	AliceAddr        string    `json:"AliceAddr" xorm:"text not null"`
	AliceSyncthingId string    `json:"AliceSyncthingId" xorm:"text"`
	BobAddr          string    `json:"BobAddr" xorm:"text not null"`
	Mode             string    `json:"mode" xorm:"text not null"`
	SubMode          string    `json:"subMode" xorm:"text not null"`
	OT               bool      `json:"ot" xorm:"bool"`
	Size             string    `json:"size" xorm:"text not null"`
	S                string    `json:"s" xorm:"text not null"`
	N                string    `json:"n" xorm:"text not null"`
	SigmaMKLRoot     string    `json:"sigmaMklRoot" xorm:"text not null"`
	Price            int64     `json:"price" xorm:"INTEGER"`
	UnitPrice        int64     `json:"unit_price" xorm:"INTEGER"`
	ExpireAt         int64     `json:"expireAt" xorm:"INTEGER"`
	Count            int64     `json:"count" xorm:"INTEGER"`
	Demands          string    `json:"demands" xorm:"text"`
	Phantoms         string    `json:"phantoms" xorm:"text"`
	KeyName          string    `json:"keyName" xorm:"text"`
	KeyValue         string    `json:"keyValue" xorm:"text"`
	PhantomKeyValue  string    `json:"phantomKeyValue" xorm:"text"`
	CreateDate       time.Time `json:"createDate" xorm:"created"`
	UpdateDate       time.Time `json:"updateDate" xorm:"updated"`
}

func insertBobTxToDB(transaction BobTransaction) error {

	var tx BobTx
	tx.SessionId = transaction.SessionID
	tx.Status = transaction.Status
	tx.AliceIP = transaction.AliceIP
	tx.AliceAddr = transaction.AliceAddr
	tx.BobAddr = transaction.BobAddr
	tx.Mode = transaction.Mode
	tx.SubMode = transaction.SubMode
	tx.OT = transaction.OT
	tx.Size = transaction.Bulletin.Size
	tx.S = transaction.Bulletin.S
	tx.N = transaction.Bulletin.N
	tx.SigmaMKLRoot = transaction.Bulletin.SigmaMKLRoot
	tx.Price = transaction.Price
	tx.Count = transaction.Count
	tx.ExpireAt = transaction.ExpireAt

	if tx.Mode == TRANSACTION_MODE_TABLE_POD {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_COMPLAINT:
			if tx.OT {
				demands, err := json.Marshal(&transaction.PlainOTComplaint.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
				phantoms, err := json.Marshal(&transaction.PlainOTComplaint.Phantoms)
				if err != nil {
					return fmt.Errorf("failed to parse phantoms: %v", err)
				}
				tx.Phantoms = string(phantoms)
			} else {
				demands, err := json.Marshal(&transaction.PlainComplaint.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
			}
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP:
			demands, err := json.Marshal(&transaction.PlainAtomicSwap.Demands)
			if err != nil {
				return fmt.Errorf("failed to parse demands: %v", err)
			}
			tx.Demands = string(demands)
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC:
			demands, err := json.Marshal(&transaction.PlainAtomicSwapVc.Demands)
			if err != nil {
				return fmt.Errorf("failed to parse demands: %v", err)
			}
			tx.Demands = string(demands)
		}
	} else {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_COMPLAINT:
			if tx.OT {
				demands, err := json.Marshal(&transaction.TableOTComplaint.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
				phantoms, err := json.Marshal(&transaction.TableOTComplaint.Phantoms)
				if err != nil {
					return fmt.Errorf("failed to parse phantoms: %v", err)
				}
				tx.Phantoms = string(phantoms)
			} else {
				demands, err := json.Marshal(&transaction.TableComplaint.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
			}
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP:
			demands, err := json.Marshal(&transaction.TableAtomicSwap.Demands)
			if err != nil {
				return fmt.Errorf("failed to parse demands: %v", err)
			}
			tx.Demands = string(demands)
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC:
			demands, err := json.Marshal(&transaction.TableAtomicSwapVc.Demands)
			if err != nil {
				return fmt.Errorf("failed to parse demands: %v", err)
			}
			tx.Demands = string(demands)
		case TRANSACTION_SUB_MODE_VRF:
			if tx.OT {
				tx.KeyName = transaction.TableOTVRF.KeyName
				keyValue, err := json.Marshal(&transaction.TableOTVRF.KeyValue)
				if err != nil {
					return fmt.Errorf("failed to parse keyValue: %v", err)
				}
				tx.KeyValue = string(keyValue)
				phantomKeyValue, err := json.Marshal(&transaction.TableOTVRF.PhantomKeyValue)
				if err != nil {
					return fmt.Errorf("failed to parse PhantomKeyValue: %v", err)
				}
				tx.PhantomKeyValue = string(phantomKeyValue)
			} else {
				tx.KeyName = transaction.TableVRF.KeyName
				keyValue, err := json.Marshal(&transaction.TableVRF.KeyValue)
				if err != nil {
					return fmt.Errorf("failed to parse keyValue: %v", err)
				}
				tx.KeyValue = string(keyValue)
			}

		}
	}

	_, err := DBConn.Insert(&tx)
	if err != nil {
		return fmt.Errorf("failed to insert transaction to Bob_tx. err=%v", err)
	}
	return nil
}

func updateBobTxToDB(transaction BobTransaction) error {

	var tx BobTx
	tx.SessionId = transaction.SessionID
	tx.Status = transaction.Status
	tx.AliceIP = transaction.AliceIP
	tx.AliceAddr = transaction.AliceAddr
	tx.BobAddr = transaction.BobAddr
	tx.Mode = transaction.Mode
	tx.SubMode = transaction.SubMode
	tx.OT = transaction.OT
	tx.Size = transaction.Bulletin.Size
	tx.S = transaction.Bulletin.S
	tx.N = transaction.Bulletin.N
	tx.SigmaMKLRoot = transaction.Bulletin.SigmaMKLRoot
	tx.Price = transaction.Price
	tx.UnitPrice = transaction.UnitPrice
	tx.Count = transaction.Count
	tx.ExpireAt = transaction.ExpireAt

	if tx.Mode == TRANSACTION_MODE_TABLE_POD {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_COMPLAINT:
			if tx.OT {
				demands, err := json.Marshal(&transaction.PlainOTComplaint.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
				phantoms, err := json.Marshal(&transaction.PlainOTComplaint.Phantoms)
				if err != nil {
					return fmt.Errorf("failed to parse phantoms: %v", err)
				}
				tx.Phantoms = string(phantoms)
			} else {
				demands, err := json.Marshal(&transaction.PlainComplaint.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
			}
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP:
			demands, err := json.Marshal(&transaction.PlainAtomicSwap.Demands)
			if err != nil {
				return fmt.Errorf("failed to parse demands: %v", err)
			}
			tx.Demands = string(demands)
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC:
			demands, err := json.Marshal(&transaction.PlainAtomicSwapVc.Demands)
			if err != nil {
				return fmt.Errorf("failed to parse demands: %v", err)
			}
			tx.Demands = string(demands)
		}
	} else {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_COMPLAINT:
			if tx.OT {
				demands, err := json.Marshal(&transaction.TableOTComplaint.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
				phantoms, err := json.Marshal(&transaction.TableOTComplaint.Phantoms)
				if err != nil {
					return fmt.Errorf("failed to parse phantoms: %v", err)
				}
				tx.Phantoms = string(phantoms)
			} else {
				demands, err := json.Marshal(&transaction.TableComplaint.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
			}
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP:
			demands, err := json.Marshal(&transaction.TableAtomicSwap.Demands)
			if err != nil {
				return fmt.Errorf("failed to parse demands: %v", err)
			}
			tx.Demands = string(demands)
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC:
			demands, err := json.Marshal(&transaction.TableAtomicSwapVc.Demands)
			if err != nil {
				return fmt.Errorf("failed to parse demands: %v", err)
			}
			tx.Demands = string(demands)
		case TRANSACTION_SUB_MODE_VRF:
			if tx.OT {
				tx.KeyName = transaction.TableOTVRF.KeyName
				keyValue, err := json.Marshal(&transaction.TableOTVRF.KeyValue)
				if err != nil {
					return fmt.Errorf("failed to parse keyValue: %v", err)
				}
				tx.KeyValue = string(keyValue)
				phantomKeyValue, err := json.Marshal(&transaction.TableOTVRF.PhantomKeyValue)
				if err != nil {
					return fmt.Errorf("failed to parse PhantomKeyValue: %v", err)
				}
				tx.PhantomKeyValue = string(phantomKeyValue)
			} else {
				tx.KeyName = transaction.TableVRF.KeyName
				keyValue, err := json.Marshal(&transaction.TableVRF.KeyValue)
				if err != nil {
					return fmt.Errorf("failed to parse keyValue: %v", err)
				}
				tx.KeyValue = string(keyValue)
			}
		}
	}

	_, err := DBConn.Where("session_id = ?", tx.SessionId).Update(&tx)
	if err != nil {
		return fmt.Errorf("failed to insert transaction to Bob_tx. err=%v", err)
	}
	return nil
}

func loadBobTxFromDB(sessionID string) (BobTransaction, error) {
	var tx BobTx
	var transaction BobTransaction

	rs, err := DBConn.Where("session_id = ?", sessionID).Get(&tx)
	if err != nil {
		return transaction, fmt.Errorf("failed to read transaction for Bob. sessionID=%v, err=%v", sessionID, err)
	}

	if !rs {
		return transaction, nil
	}

	transaction.SessionID = tx.SessionId
	transaction.Status = tx.Status
	transaction.AliceIP = tx.AliceIP
	transaction.AliceAddr = tx.AliceAddr
	transaction.BobAddr = tx.BobAddr
	transaction.Mode = tx.Mode
	transaction.SubMode = tx.SubMode
	transaction.OT = tx.OT
	transaction.Bulletin.Size = tx.Size
	transaction.Bulletin.S = tx.S
	transaction.Bulletin.N = tx.N
	transaction.Bulletin.SigmaMKLRoot = tx.SigmaMKLRoot
	transaction.Bulletin.Mode = tx.Mode
	transaction.Price = tx.Price
	transaction.UnitPrice = tx.UnitPrice
	transaction.Count = tx.Count
	transaction.ExpireAt = tx.ExpireAt

	if transaction.Mode == TRANSACTION_MODE_TABLE_POD {
		switch transaction.SubMode {
		case TRANSACTION_SUB_MODE_COMPLAINT:
			if transaction.OT {
				err := json.Unmarshal([]byte(tx.Demands), &transaction.PlainOTComplaint.Demands)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse demands: %v", err)
				}
				err = json.Unmarshal([]byte(tx.Phantoms), &transaction.PlainOTComplaint.Phantoms)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse phantoms: %v", err)
				}
			} else {
				err := json.Unmarshal([]byte(tx.Demands), &transaction.PlainComplaint.Demands)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse demands: %v", err)
				}
			}
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP:
			err := json.Unmarshal([]byte(tx.Demands), &transaction.PlainAtomicSwap.Demands)
			if err != nil {
				return transaction, fmt.Errorf("failed to parse demands: %v", err)
			}
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC:
			err := json.Unmarshal([]byte(tx.Demands), &transaction.PlainAtomicSwapVc.Demands)
			if err != nil {
				return transaction, fmt.Errorf("failed to parse demands: %v", err)
			}
		}
	} else {
		switch transaction.SubMode {
		case TRANSACTION_SUB_MODE_COMPLAINT:
			if transaction.OT {
				err := json.Unmarshal([]byte(tx.Demands), &transaction.TableOTComplaint.Demands)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse demands: %v", err)
				}
				err = json.Unmarshal([]byte(tx.Phantoms), &transaction.TableOTComplaint.Phantoms)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse phantoms: %v", err)
				}
			} else {
				err := json.Unmarshal([]byte(tx.Demands), &transaction.TableComplaint.Demands)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse demands: %v", err)
				}
			}
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP:
			err := json.Unmarshal([]byte(tx.Demands), &transaction.TableAtomicSwap.Demands)
			if err != nil {
				return transaction, fmt.Errorf("failed to parse demands: %v", err)
			}
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC:
			err := json.Unmarshal([]byte(tx.Demands), &transaction.TableAtomicSwapVc.Demands)
			if err != nil {
				return transaction, fmt.Errorf("failed to parse demands: %v", err)
			}
		case TRANSACTION_SUB_MODE_VRF:
			if transaction.OT {
				transaction.TableOTVRF.KeyName = tx.KeyName
				err := json.Unmarshal([]byte(tx.KeyValue), &transaction.TableOTVRF.KeyValue)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse keyvalue: %v", err)
				}
				err = json.Unmarshal([]byte(tx.PhantomKeyValue), &transaction.TableOTVRF.PhantomKeyValue)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse phantomKeyValue: %v", err)
				}
			} else {
				transaction.TableVRF.KeyName = tx.KeyName
				err := json.Unmarshal([]byte(tx.KeyValue), &transaction.TableVRF.KeyValue)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse keyvalue: %v", err)
				}
			}
		}
	}

	return transaction, nil
}

func loadBobTxListFromDB() ([]BobTransaction, error) {

	var txs []BobTx

	err := DBConn.Find(&txs)
	if err != nil {
		return nil, fmt.Errorf("failed to read transaction for Bob. err=%v", err)
	}

	var transactions = make([]BobTransaction, len(txs))
	for i, tx := range txs {
		transactions[i].SessionID = tx.SessionId
		transactions[i].Status = tx.Status
		transactions[i].AliceIP = tx.AliceIP
		transactions[i].AliceAddr = tx.AliceAddr
		transactions[i].BobAddr = tx.BobAddr
		transactions[i].Mode = tx.Mode
		transactions[i].SubMode = tx.SubMode
		transactions[i].OT = tx.OT
		transactions[i].Bulletin.Size = tx.Size
		transactions[i].Bulletin.S = tx.S
		transactions[i].Bulletin.N = tx.N
		transactions[i].Bulletin.SigmaMKLRoot = tx.SigmaMKLRoot
		transactions[i].Bulletin.Mode = tx.Mode
		transactions[i].Price = tx.Price
		transactions[i].UnitPrice = tx.UnitPrice
		transactions[i].Count = tx.Count
		transactions[i].ExpireAt = tx.ExpireAt

		if transactions[i].Mode == TRANSACTION_MODE_TABLE_POD {
			switch transactions[i].SubMode {
			case TRANSACTION_SUB_MODE_COMPLAINT:
				if transactions[i].OT {
					err := json.Unmarshal([]byte(tx.Demands), &transactions[i].PlainOTComplaint.Demands)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse demands: %v", err)
					}
					err = json.Unmarshal([]byte(tx.Phantoms), &transactions[i].PlainOTComplaint.Phantoms)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse phantoms: %v", err)
					}
				} else {
					err := json.Unmarshal([]byte(tx.Demands), &transactions[i].PlainComplaint.Demands)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse demands: %v", err)
					}
				}
			case TRANSACTION_SUB_MODE_ATOMIC_SWAP:
				err := json.Unmarshal([]byte(tx.Demands), &transactions[i].PlainAtomicSwap.Demands)
				if err != nil {
					return transactions, fmt.Errorf("failed to parse demands: %v", err)
				}
			case TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC:
				err := json.Unmarshal([]byte(tx.Demands), &transactions[i].PlainAtomicSwapVc.Demands)
				if err != nil {
					return transactions, fmt.Errorf("failed to parse demands: %v", err)
				}
			}
		} else {
			switch transactions[i].SubMode {
			case TRANSACTION_SUB_MODE_COMPLAINT:
				if transactions[i].OT {
					err := json.Unmarshal([]byte(tx.Demands), &transactions[i].TableOTComplaint.Demands)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse demands: %v", err)
					}
					err = json.Unmarshal([]byte(tx.Phantoms), &transactions[i].TableOTComplaint.Phantoms)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse phantoms: %v", err)
					}
				} else {
					err := json.Unmarshal([]byte(tx.Demands), &transactions[i].TableComplaint.Demands)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse demands: %v", err)
					}
				}
			case TRANSACTION_SUB_MODE_ATOMIC_SWAP:
				err := json.Unmarshal([]byte(tx.Demands), &transactions[i].TableAtomicSwap.Demands)
				if err != nil {
					return transactions, fmt.Errorf("failed to parse demands: %v", err)
				}
			case TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC:
				err := json.Unmarshal([]byte(tx.Demands), &transactions[i].TableAtomicSwapVc.Demands)
				if err != nil {
					return transactions, fmt.Errorf("failed to parse demands: %v", err)
				}
			case TRANSACTION_SUB_MODE_VRF:
				if transactions[i].OT {
					transactions[i].TableOTVRF.KeyName = tx.KeyName
					err := json.Unmarshal([]byte(tx.KeyValue), &transactions[i].TableOTVRF.KeyValue)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse keyvalue: %v", err)
					}
					err = json.Unmarshal([]byte(tx.PhantomKeyValue), &transactions[i].TableOTVRF.PhantomKeyValue)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse phantomKeyValue: %v", err)
					}
				} else {
					transactions[i].TableVRF.KeyName = tx.KeyName
					err := json.Unmarshal([]byte(tx.KeyValue), &transactions[i].TableVRF.KeyValue)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse keyvalue: %v", err)
					}
				}
			}
		}
	}

	return transactions, nil
}

type AliceTx struct {
	SessionId    string    `json:"sessionId" xorm:"text pk not null"`
	Status       string    `json:"status" xorm:"text not null"`
	AliceAddr    string    `json:"AliceAddr" xorm:"text not null"`
	BobPubKey    string    `json:"BobPubkey" xorm:"text not null"`
	BobAddr      string    `json:"BobAddr" xorm:"text not null"`
	Mode         string    `json:"mode" xorm:"text not null"`
	SubMode      string    `json:"subMode" xorm:"text not null"`
	OT           bool      `json:"ot" xorm:"bool"`
	Size         string    `json:"size" xorm:"text not null"`
	S            string    `json:"s" xorm:"text not null"`
	N            string    `json:"n" xorm:"text not null"`
	SigmaMKLRoot string    `json:"sigmaMklRoot" xorm:"text not null"`
	Price        int64     `json:"price" xorm:"INTEGER"`
	UnitPrice    int64     `json:"unit_price" xorm:"INTEGER"`
	ExpireAt     int64     `json:"expireAt" xorm:"text"`
	Count        int64     `json:"count" xorm:"INTEGER"`
	CreateDate   time.Time `json:"createDate" xorm:"created"`
	UpdateDate   time.Time `json:"updateDate" xorm:"updated"`
}

func insertAliceTxToDB(transaction Transaction) error {

	var tx AliceTx
	tx.SessionId = transaction.SessionID
	tx.Status = transaction.Status //StatusAtoi(transaction.Status)
	tx.AliceAddr = transaction.AliceAddr
	// tx.BobPubKey = fmt.Sprintf("0x%x", crypto.FromECDSAPub(transaction.BobPubKey)[1:])
	tx.BobAddr = transaction.BobAddr
	tx.Mode = transaction.Mode
	tx.SubMode = transaction.SubMode
	tx.OT = transaction.OT
	tx.Size = transaction.Bulletin.Size
	tx.S = transaction.Bulletin.S
	tx.N = transaction.Bulletin.N
	tx.SigmaMKLRoot = transaction.Bulletin.SigmaMKLRoot
	tx.Price = transaction.Price
	tx.UnitPrice = transaction.UnitPrice
	tx.ExpireAt = transaction.ExpireAt
	tx.Count = transaction.Count

	_, err := DBConn.Insert(&tx)
	if err != nil {
		return fmt.Errorf("failed to insert transaction to Alice_tx. err=%v", err)
	}
	return nil
}

func updateAliceTxToDB(transaction Transaction) error {

	var tx AliceTx
	tx.SessionId = transaction.SessionID
	tx.Status = transaction.Status //StatusAtoi(transaction.Status)
	// tx.AliceIP = transaction.AliceIP
	tx.AliceAddr = transaction.AliceAddr
	// tx.BobPubKey = transaction.BobPubKey
	tx.BobAddr = transaction.BobAddr
	tx.Mode = transaction.Mode
	tx.SubMode = transaction.SubMode
	tx.OT = transaction.OT
	tx.Size = transaction.Bulletin.Size
	tx.S = transaction.Bulletin.S
	tx.N = transaction.Bulletin.N
	tx.SigmaMKLRoot = transaction.Bulletin.SigmaMKLRoot
	tx.Price = transaction.Price
	tx.UnitPrice = transaction.UnitPrice
	tx.ExpireAt = transaction.ExpireAt
	tx.Count = transaction.Count

	_, err := DBConn.Where("session_id = ?", tx.SessionId).Update(&tx)
	if err != nil {
		return fmt.Errorf("failed to insert transaction to Alice_tx. err=%v", err)
	}
	return nil
}

func loadAliceFromDB(sessionID string) (Transaction, bool, error) {

	var tx AliceTx
	var transaction Transaction

	rs, err := DBConn.Where("session_id = ?", sessionID).Get(&tx)
	if err != nil {
		return transaction, false, fmt.Errorf("failed to read transaction for Bob. sessionID=%v, err=%v", sessionID, err)
	}
	if !rs {
		return transaction, false, nil
	}

	transaction.SessionID = tx.SessionId
	transaction.Status = tx.Status
	transaction.AliceAddr = tx.AliceAddr
	transaction.BobAddr = tx.BobAddr
	// transaction.BobPubKey = tx.BobPubKey
	transaction.Mode = tx.Mode
	transaction.SubMode = tx.SubMode
	transaction.OT = tx.OT
	transaction.Bulletin.Size = tx.Size
	transaction.Bulletin.S = tx.S
	transaction.Bulletin.N = tx.N
	transaction.Bulletin.SigmaMKLRoot = tx.SigmaMKLRoot
	transaction.Bulletin.Mode = tx.Mode
	transaction.Price = tx.Price
	transaction.UnitPrice = tx.UnitPrice
	transaction.ExpireAt = tx.ExpireAt
	transaction.Count = tx.Count

	return transaction, true, nil
}

func loadAliceTxListToDB() ([]Transaction, error) {

	var txs []BobTx

	err := DBConn.Find(&txs)
	if err != nil {
		return nil, fmt.Errorf("failed to read transaction for Bob. err=%v", err)
	}

	var transactions = make([]Transaction, len(txs))
	for i, tx := range txs {
		transactions[i].SessionID = tx.SessionId
		transactions[i].Status = tx.Status
		// transactions[i].AliceIP = tx.AliceIP
		transactions[i].AliceAddr = tx.AliceAddr
		transactions[i].BobAddr = tx.BobAddr
		transactions[i].Mode = tx.Mode
		transactions[i].SubMode = tx.SubMode
		transactions[i].OT = tx.OT
		transactions[i].Bulletin.Size = tx.Size
		transactions[i].Bulletin.S = tx.S
		transactions[i].Bulletin.N = tx.N
		transactions[i].Bulletin.SigmaMKLRoot = tx.SigmaMKLRoot
		transactions[i].Bulletin.Mode = tx.Mode
		transactions[i].Price = tx.Price
		transactions[i].UnitPrice = tx.UnitPrice
		transactions[i].ExpireAt = tx.ExpireAt
		transactions[i].Count = tx.Count
	}

	return transactions, nil
}

type PodLog struct {
	Id        int64     `json:"id" xorm:"autoincr pk"`
	SessionId string    `json:"session_id" xorm:"text"`
	Operation int       `json:"operation" xorm:"INTEGER not null"`
	Detail    string    `json:"detail" xorm:"text"`
	Result    int       `json:"result" xorm:"INTEGER not null"`
	CreatedAt time.Time `json:"created_at" xorm:"created"`
}

func insertLogToDB(log PodLog) error {
	_, err := DBConn.Insert(&log)
	if err != nil {
		return fmt.Errorf("failed to insert log to pod_log. err=%v", err)
	}
	return nil
}

func loadLogListByOperatorFromDB(operator string) ([]PodLog, error) {

	var list []PodLog

	err := DBConn.Where("operator = ?", operator).Find(&list)
	if err != nil {
		return list, fmt.Errorf("failed to load logs. err=%v", err)
	}
	return list, nil
}
