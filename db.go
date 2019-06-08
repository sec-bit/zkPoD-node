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
	db.ShowSQL(true)
	db.Logger().SetLevel(core.LOG_DEBUG)

	err = db.CreateTables(&BuyerTx{})
	if err != nil {
		Log.Errorf("failed to create table buyer_tx, err=%v", err)
		return db, errors.New("failed to create table buyer_tx")
	}
	Log.Debugf("initialize table buyer_tx success")

	err = db.CreateTables(&SellerTx{})
	if err != nil {
		Log.Errorf("failed to create table seller_tx, err=%v", err)
		return db, errors.New("failed to create table seller_tx")
	}
	Log.Debugf("initialize table seller_tx success")

	err = db.CreateTables(&PodLog{})
	if err != nil {
		Log.Errorf("failed to create table pod_log, err=%v", err)
		return db, errors.New("failed to create table pod_log")
	}
	Log.Debugf("initialize table pod_log success")

	err = db.Sync2(new(BuyerTx))
	if err != nil {
		Log.Warnf("table buyer_tx sync error, err=%v", err)
		return nil, err
	}
	err = db.Sync2(new(SellerTx))
	if err != nil {
		Log.Warnf("table seller_tx sync error, err=%v", err)
		return nil, err
	}
	err = db.Sync2(new(PodLog))
	if err != nil {
		Log.Warnf("table pod_log sync error, err=%v", err)
		return nil, err
	}
	db.ShowSQL(true)
	db.ShowExecTime(true)
	return db, err
}

type BuyerTx struct {
	SessionId         string    `json:"sessionId" xorm:"text pk not null"`
	Status            string    `json:"status" xorm:"text not null"`
	SellerIP          string    `json:"sellerIP" xorm:"text not null"`
	SellerAddr        string    `json:"sellerAddr" xorm:"text not null"`
	SellerSyncthingId string    `json:"sellerSyncthingId" xorm:"text"`
	BuyerAddr         string    `json:"buyerAddr" xorm:"text not null"`
	Mode              string    `json:"mode" xorm:"text not null"`
	SubMode           string    `json:"subMode" xorm:"text not null"`
	OT                bool      `json:"ot" xorm:"bool"`
	Size              string    `json:"size" xorm:"text not null"`
	S                 string    `json:"s" xorm:"text not null"`
	N                 string    `json:"n" xorm:"text not null"`
	SigmaMKLRoot      string    `json:"sigmaMklRoot" xorm:"text not null"`
	Price             int64     `json:"price" xorm:"INTEGER"`
	UnitPrice         int64     `json:"unit_price" xorm:"INTEGER"`
	ExpireAt          int64     `json:"expireAt" xorm:"INTEGER"`
	Demands           string    `json:"demands" xorm:"text"`
	Phantoms          string    `json:"phantoms" xorm:"text"`
	KeyName           string    `json:"keyName" xorm:"text"`
	KeyValue          string    `json:"keyValue" xorm:"text"`
	PhantomKeyValue   string    `json:"phantomKeyValue" xorm:"text"`
	CreateDate        time.Time `json:"createDate" xorm:"created"`
	UpdateDate        time.Time `json:"updateDate" xorm:"updated"`
}

func insertBuyerTxToDB(transaction BuyerTransaction) error {

	var tx BuyerTx
	tx.SessionId = transaction.SessionID
	tx.Status = transaction.Status
	tx.SellerIP = transaction.SellerIP
	tx.SellerAddr = transaction.SellerAddr
	tx.BuyerAddr = transaction.BuyerAddr
	tx.Mode = transaction.Mode
	tx.SubMode = transaction.SubMode
	tx.OT = transaction.OT
	tx.Size = transaction.Bulletin.Size
	tx.S = transaction.Bulletin.S
	tx.N = transaction.Bulletin.N
	tx.SigmaMKLRoot = transaction.Bulletin.SigmaMKLRoot
	tx.Price = transaction.Price
	tx.ExpireAt = transaction.ExpireAt

	if tx.Mode == TRANSACTION_MODE_TABLE_POD {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_BATCH1:
			if tx.OT {
				demands, err := json.Marshal(&transaction.PlainOTBatch1.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
				phantoms, err := json.Marshal(&transaction.PlainOTBatch1.Phantoms)
				if err != nil {
					return fmt.Errorf("failed to parse phantoms: %v", err)
				}
				tx.Phantoms = string(phantoms)
			} else {
				demands, err := json.Marshal(&transaction.PlainBatch1.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
			}
		case TRANSACTION_SUB_MODE_BATCH2:
			demands, err := json.Marshal(&transaction.PlainBatch2.Demands)
			if err != nil {
				return fmt.Errorf("failed to parse demands: %v", err)
			}
			tx.Demands = string(demands)
		}
	} else {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_BATCH1:
			if tx.OT {
				demands, err := json.Marshal(&transaction.TableOTBatch1.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
				phantoms, err := json.Marshal(&transaction.TableOTBatch1.Phantoms)
				if err != nil {
					return fmt.Errorf("failed to parse phantoms: %v", err)
				}
				tx.Phantoms = string(phantoms)
			} else {
				demands, err := json.Marshal(&transaction.TableBatch1.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
			}
		case TRANSACTION_SUB_MODE_BATCH2:
			demands, err := json.Marshal(&transaction.TableBatch2.Demands)
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
		return fmt.Errorf("failed to insert transaction to buyer_tx. err=%v", err)
	}
	return nil
}

func updateBuyerTxToDB(transaction BuyerTransaction) error {

	var tx BuyerTx
	tx.SessionId = transaction.SessionID
	tx.Status = transaction.Status
	tx.SellerIP = transaction.SellerIP
	tx.SellerAddr = transaction.SellerAddr
	tx.BuyerAddr = transaction.BuyerAddr
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

	if tx.Mode == TRANSACTION_MODE_TABLE_POD {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_BATCH1:
			if tx.OT {
				demands, err := json.Marshal(&transaction.PlainOTBatch1.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
				phantoms, err := json.Marshal(&transaction.PlainOTBatch1.Phantoms)
				if err != nil {
					return fmt.Errorf("failed to parse phantoms: %v", err)
				}
				tx.Phantoms = string(phantoms)
			} else {
				demands, err := json.Marshal(&transaction.PlainBatch1.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
			}
		case TRANSACTION_SUB_MODE_BATCH2:
			demands, err := json.Marshal(&transaction.PlainBatch2.Demands)
			if err != nil {
				return fmt.Errorf("failed to parse demands: %v", err)
			}
			tx.Demands = string(demands)
		}
	} else {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_BATCH1:
			if tx.OT {
				demands, err := json.Marshal(&transaction.TableOTBatch1.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
				phantoms, err := json.Marshal(&transaction.TableOTBatch1.Phantoms)
				if err != nil {
					return fmt.Errorf("failed to parse phantoms: %v", err)
				}
				tx.Phantoms = string(phantoms)
			} else {
				demands, err := json.Marshal(&transaction.TableBatch1.Demands)
				if err != nil {
					return fmt.Errorf("failed to parse demands: %v", err)
				}
				tx.Demands = string(demands)
			}
		case TRANSACTION_SUB_MODE_BATCH2:
			demands, err := json.Marshal(&transaction.TableBatch2.Demands)
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
		return fmt.Errorf("failed to insert transaction to buyer_tx. err=%v", err)
	}
	return nil
}

func loadBuyerTxFromDB(sessionID string) (BuyerTransaction, error) {
	var tx BuyerTx
	var transaction BuyerTransaction

	rs, err := DBConn.Where("session_id = ?", sessionID).Get(&tx)
	if err != nil {
		return transaction, fmt.Errorf("failed to read transaction for buyer. sessionID=%v, err=%v", sessionID, err)
	}

	if !rs {
		return transaction, nil
	}

	transaction.SessionID = tx.SessionId
	transaction.Status = tx.Status
	transaction.SellerIP = tx.SellerIP
	transaction.SellerAddr = tx.SellerAddr
	transaction.BuyerAddr = tx.BuyerAddr
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

	if transaction.Mode == TRANSACTION_MODE_TABLE_POD {
		switch transaction.SubMode {
		case TRANSACTION_SUB_MODE_BATCH1:
			if transaction.OT {
				err := json.Unmarshal([]byte(tx.Demands), &transaction.PlainOTBatch1.Demands)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse demands: %v", err)
				}
				err = json.Unmarshal([]byte(tx.Phantoms), &transaction.PlainOTBatch1.Phantoms)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse phantoms: %v", err)
				}
			} else {
				err := json.Unmarshal([]byte(tx.Demands), &transaction.PlainBatch1.Demands)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse demands: %v", err)
				}
			}
		case TRANSACTION_SUB_MODE_BATCH2:
			err := json.Unmarshal([]byte(tx.Demands), &transaction.PlainBatch2.Demands)
			if err != nil {
				return transaction, fmt.Errorf("failed to parse demands: %v", err)
			}
		}
	} else {
		switch transaction.SubMode {
		case TRANSACTION_SUB_MODE_BATCH1:
			if transaction.OT {
				err := json.Unmarshal([]byte(tx.Demands), &transaction.TableOTBatch1.Demands)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse demands: %v", err)
				}
				err = json.Unmarshal([]byte(tx.Phantoms), &transaction.TableOTBatch1.Phantoms)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse phantoms: %v", err)
				}
			} else {
				err := json.Unmarshal([]byte(tx.Demands), &transaction.TableBatch1.Demands)
				if err != nil {
					return transaction, fmt.Errorf("failed to parse demands: %v", err)
				}
			}
		case TRANSACTION_SUB_MODE_BATCH2:
			err := json.Unmarshal([]byte(tx.Demands), &transaction.TableBatch2.Demands)
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

func loadBuyerTxListFromDB() ([]BuyerTransaction, error) {

	var txs []BuyerTx

	err := DBConn.Find(&txs)
	if err != nil {
		return nil, fmt.Errorf("failed to read transaction for buyer. err=%v", err)
	}

	var transactions = make([]BuyerTransaction, len(txs))
	for i, tx := range txs {
		transactions[i].SessionID = tx.SessionId
		transactions[i].Status = tx.Status
		transactions[i].SellerIP = tx.SellerIP
		transactions[i].SellerAddr = tx.SellerAddr
		transactions[i].BuyerAddr = tx.BuyerAddr
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

		if transactions[i].Mode == TRANSACTION_MODE_TABLE_POD {
			switch transactions[i].SubMode {
			case TRANSACTION_SUB_MODE_BATCH1:
				if transactions[i].OT {
					err := json.Unmarshal([]byte(tx.Demands), &transactions[i].PlainOTBatch1.Demands)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse demands: %v", err)
					}
					err = json.Unmarshal([]byte(tx.Phantoms), &transactions[i].PlainOTBatch1.Phantoms)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse phantoms: %v", err)
					}
				} else {
					err := json.Unmarshal([]byte(tx.Demands), &transactions[i].PlainBatch1.Demands)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse demands: %v", err)
					}
				}
			case TRANSACTION_SUB_MODE_BATCH2:
				err := json.Unmarshal([]byte(tx.Demands), &transactions[i].PlainBatch2.Demands)
				if err != nil {
					return transactions, fmt.Errorf("failed to parse demands: %v", err)
				}
			}
		} else {
			switch transactions[i].SubMode {
			case TRANSACTION_SUB_MODE_BATCH1:
				if transactions[i].OT {
					err := json.Unmarshal([]byte(tx.Demands), &transactions[i].TableOTBatch1.Demands)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse demands: %v", err)
					}
					err = json.Unmarshal([]byte(tx.Phantoms), &transactions[i].TableOTBatch1.Phantoms)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse phantoms: %v", err)
					}
				} else {
					err := json.Unmarshal([]byte(tx.Demands), &transactions[i].TableBatch1.Demands)
					if err != nil {
						return transactions, fmt.Errorf("failed to parse demands: %v", err)
					}
				}
			case TRANSACTION_SUB_MODE_BATCH2:
				err := json.Unmarshal([]byte(tx.Demands), &transactions[i].TableBatch2.Demands)
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

type SellerTx struct {
	SessionId    string `json:"sessionId" xorm:"text pk not null"`
	Status       string `json:"status" xorm:"text not null"`
	SellerAddr   string `json:"sellerAddr" xorm:"text not null"`
	BuyerPubKey  string `json:"buyerPubkey" xorm:"text not null"`
	BuyerAddr    string `json:"buyerAddr" xorm:"text not null"`
	Mode         string `json:"mode" xorm:"text not null"`
	SubMode      string `json:"subMode" xorm:"text not null"`
	OT           bool   `json:"ot" xorm:"bool"`
	Size         string `json:"size" xorm:"text not null"`
	S            string `json:"s" xorm:"text not null"`
	N            string `json:"n" xorm:"text not null"`
	SigmaMKLRoot string `json:"sigmaMklRoot" xorm:"text not null"`
	Price        int64  `json:"price" xorm:"INTEGER"`
	UnitPrice    int64  `json:"unit_price" xorm:"INTEGER"`
	ExpireAt     int64  `json:"expireAt" xorm:"text"`
	// Demands         string `json:"demands" xorm:"text"`
	// Phantoms        string `json:"phantoms" xorm:"text"`
	// KeyName         string `json:"keyName" xorm:"text"`
	// KeyValue        string `json:"keyValue" xorm:"text"`
	// PhantomKeyValue string `json:"phantomKeyValue" xorm:"text"`
	CreateDate time.Time `json:"createDate" xorm:"created"`
	UpdateDate time.Time `json:"updateDate" xorm:"updated"`
}

func insertSellerTxToDB(transaction Transaction) error {

	var tx SellerTx
	tx.SessionId = transaction.SessionID
	tx.Status = transaction.Status //StatusAtoi(transaction.Status)
	tx.SellerAddr = transaction.SellerAddr
	// tx.BuyerPubKey = fmt.Sprintf("0x%x", crypto.FromECDSAPub(transaction.BuyerPubKey)[1:])
	tx.BuyerAddr = transaction.BuyerAddr
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

	_, err := DBConn.Insert(&tx)
	if err != nil {
		return fmt.Errorf("failed to insert transaction to seller_tx. err=%v", err)
	}
	return nil
}

func updateSellerTxToDB(transaction Transaction) error {

	var tx SellerTx
	tx.SessionId = transaction.SessionID
	tx.Status = transaction.Status //StatusAtoi(transaction.Status)
	// tx.SellerIP = transaction.SellerIP
	tx.SellerAddr = transaction.SellerAddr
	// tx.BuyerPubKey = transaction.BuyerPubKey
	tx.BuyerAddr = transaction.BuyerAddr
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

	_, err := DBConn.Where("session_id = ?", tx.SessionId).Update(&tx)
	if err != nil {
		return fmt.Errorf("failed to insert transaction to seller_tx. err=%v", err)
	}
	return nil
}

func loadSellerFromDB(sessionID string) (Transaction, bool, error) {

	var tx SellerTx
	var transaction Transaction

	rs, err := DBConn.Where("session_id = ?", sessionID).Get(&tx)
	if err != nil {
		return transaction, false, fmt.Errorf("failed to read transaction for buyer. sessionID=%v, err=%v", sessionID, err)
	}
	if !rs {
		return transaction, false, nil
	}

	transaction.SessionID = tx.SessionId
	transaction.Status = tx.Status
	transaction.SellerAddr = tx.SellerAddr
	transaction.BuyerAddr = tx.BuyerAddr
	// transaction.BuyerPubKey = tx.BuyerPubKey
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

	return transaction, true, nil
}

func loadSellerTxListToDB() ([]Transaction, error) {

	var txs []BuyerTx

	err := DBConn.Find(&txs)
	if err != nil {
		return nil, fmt.Errorf("failed to read transaction for buyer. err=%v", err)
	}

	var transactions = make([]Transaction, len(txs))
	for i, tx := range txs {
		transactions[i].SessionID = tx.SessionId
		transactions[i].Status = tx.Status
		// transactions[i].SellerIP = tx.SellerIP
		transactions[i].SellerAddr = tx.SellerAddr
		transactions[i].BuyerAddr = tx.BuyerAddr
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
