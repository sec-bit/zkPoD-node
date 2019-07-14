package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
)

func nodeRecovery(w http.ResponseWriter, Log ILogger) {

	if err := recover(); err != nil {
		Log.Errorf("exception unexpected: error=%v", err)
		fmt.Fprintf(w, string(RECOVERY_ERROR))
		return
	}
}

type Response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

//ReadCfgAPIHandler reads config info.
func ReadCfgAPIHandler(w http.ResponseWriter, r *http.Request) {

	Log := Logger.NewSessionLogger()

	defer nodeRecovery(w, Log)

	var resp Response
	resp.Data = BConf
	resp.Code = "0"
	resp.Message = "read config successfully"

	Log.Debugf("get config info success.")
	respByte, err := json.Marshal(&resp)
	if err != nil {
		Log.Warnf("failed to marshal response. err=%v", err)
		fmt.Fprintf(w, RESPONSE_FAILED_TO_RESPONSE)
		return
	}
	Log.Debugf("success to read config")
	fmt.Fprintf(w, string(respByte))
	return
}

//ReLoadConfigAPIHandler loads config info for web.
func ReLoadConfigAPIHandler(w http.ResponseWriter, r *http.Request) {
	Log := Logger.NewSessionLogger()

	if AliceNodeStart && BobNodeStart {
		//TODO: be capable to modify
		Log.Warnf("the node has started and can not modify config.")
		fmt.Fprintf(w, RESPONSE_HAS_STARTED)
		return
	}
	var plog PodLog
	plog.Result = LOG_RESULT_FAILED
	plog.Operation = LOG_OPERATION_TYPE_CONFIG_SETTING

	defer func() {
		err := insertLogToDB(plog)
		if err != nil {
			Log.Warnf("insert log error! %v", err)
			return
		}
		nodeRecovery(w, Log)
	}()

	if !AliceNodeStart {
		ip := r.FormValue("ip")
		if ip != "" {
			Log.Debugf("ip=%v", ip)
			BConf.NetIP = ip
		}
		plog.Detail = "modify ip =" + ip + ";"
	}

	if !AliceNodeStart && !BobNodeStart {
		password := r.FormValue("password")
		keystoreFile := r.FormValue("keystore")
		privkey := r.FormValue("privkey")
		if privkey != "" {
			Log.Debugf("import private key. privkey=%v", privkey)
			PrivateKeyECDSA, err := crypto.HexToECDSA(privkey)
			if err != nil {
				Log.Warnf("failed to parse private key. err=%v", err)
				fmt.Fprintf(w, RESPONSE_UPLOAD_KEY_FAILED)
				return
			}
			publicKey := PrivateKeyECDSA.Public()
			Log.Debugf("publicKey=%v", publicKey)
			plog.Detail = plog.Detail + "modify eth key, address=%v" + "TODO" + ";"
			publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
			if !ok {
				Log.Warnf("invalid private key.")
				fmt.Fprintf(w, RESPONSE_UPLOAD_KEY_FAILED)
				return
			} else {
				Log.Debugf("privateKeyECDSA=%v", PrivateKeyECDSA)
				key := &keystore.Key{
					Address:    crypto.PubkeyToAddress(*publicKeyECDSA),
					PrivateKey: PrivateKeyECDSA,
				}
				err = ConnectToProvider(key, BConf.ContractAddr, Log)
				if err != nil {
					Log.Warnf("failed to connect to provider for contract. err=%v", err)
					fmt.Fprintf(w, RESPONSE_INITIALIZE_FAILED)
					return
				}
				ETHKey = key
				Log.Infof("success to connect to provider for contract")
			}
			//TODO: save keystore
		} else if password != "" && keystoreFile != "" {
			plog.Detail = plog.Detail + "modify eth key, keystoreFile=" + keystoreFile
			key, err := initKeyStore(keystoreFile, password, Log)
			if err != nil {
				Log.Errorf("Failed to initialize key store file. err=%v", err)
				fmt.Fprintf(w, RESPONSE_INITIALIZE_FAILED)
				return
			}
			Log.Infof("initialize key store finish. ethaddr=%v.", key.Address.Hex())
			plog.Detail = plog.Detail + ", address=" + key.Address.Hex()
			err = ConnectToProvider(key, BConf.ContractAddr, Log)
			if err != nil {
				Log.Warnf("Failed to connect to provider for contract. err=%v", err)
				fmt.Fprintf(w, RESPONSE_INITIALIZE_FAILED)
				return
			}
			ETHKey = key
			Log.Infof("success to connect to provider for contract")
			BConf.KeyStoreFile = keystoreFile
		}
	}

	err := saveBasicConfig(BConf, DEFAULT_BASIC_CONFGI_FILE)
	if err != nil {
		Log.Warnf("failed to save basic config. file = %v, err=%v", DEFAULT_BASIC_CONFGI_FILE, err)
		fmt.Fprintf(w, RESPONSE_SAVE_CONFIG_FILE_FAILED)
		return
	}
	plog.Result = LOG_RESULT_SUCCESS
	Log.Infof("success to set basic config. config=%v", BConf)
	fmt.Fprintf(w, `{"code":"0","message":"set config successfully"}`)
	return
}

type PublishExtraInfo struct {
	MerkleRoot   string   `json:"mklroot"`
	Mode         string   `json:"mode"`
	SubMode      []string `json:"sub_mode"`
	UnitPrice    int64    `json:"uprice"`
	Description  string   `json:"description"`
	ContractAddr string   `json:"contract_addr"`
}

type InitPublishConfig struct {
	Mode        string   `json:"mode"`
	SubMode     []string `json:"sub_mode"`
	Column      string   `json:"column"`
	Keys        []int    `json:"keys"`
	UnitPrice   int64    `json:"unit_price"`
	Description string   `json:"description"`
	FilePath    string   `json:"file_path"`
}

//InitPublishDataAPIHandler is a api handler for Alice to initializing data for publishing.
func InitPublishDataAPIHandler(w http.ResponseWriter, r *http.Request) {
	Log := Logger.NewSessionLogger()

	var plog PodLog
	plog.Result = LOG_RESULT_FAILED
	plog.Operation = LOG_OPERATION_TYPE_ALICE_PUBLISH_INIT

	defer func() {
		err := insertLogToDB(plog)
		if err != nil {
			Log.Warnf("insert log error! %v", err)
			return
		}
		nodeRecovery(w, Log)
	}()

	Log.Infof("start to init data for publishing...")
	request_data := r.FormValue("request_data")

	var config InitPublishConfig
	err := json.Unmarshal([]byte(request_data), &config)
	if err != nil {
		Log.Warnf("failed to parse parameters. err=%v", err)
		fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
		return
	}

	if len(config.SubMode) == 0 {
		Log.Warnf("parameters are incomplete, submode is nil. mode=%v, subMode=%v, column=%v, keys=%v", config.Mode, config.SubMode, config.Column, config.Keys)
		fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
		return
	}

	if config.Mode == TRANSACTION_MODE_PLAIN_POD {
		for _, s := range config.SubMode {
			if s != TRANSACTION_SUB_MODE_COMPLAINT && s != TRANSACTION_SUB_MODE_ATOMIC_SWAP && s != TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC {
				Log.Warnf("parameters are invalid. mode=%v, subMode=%v, column=%v, keys=%v", config.Mode, config.SubMode, config.Column, config.Keys)
				fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
				return
			}
		}
	} else if config.Mode == TRANSACTION_MODE_TABLE_POD {
		for _, s := range config.SubMode {
			if s != TRANSACTION_SUB_MODE_COMPLAINT && s != TRANSACTION_SUB_MODE_ATOMIC_SWAP && s != TRANSACTION_SUB_MODE_VRF && s != TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC {
				Log.Warnf("parameters are invalid. mode=%v, subMode=%v, column=%v, keys=%v", config.Mode, config.SubMode, config.Column, config.Keys)
				fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
				return
			}
		}
	}
	Log.Debugf("parameter verified. mode=%v, subMode=%v, column=%v, keys=%v", config.Mode, config.SubMode, config.Column, config.Keys)
	plog.Detail = fmt.Sprintf("mode=%v, subMode=%v, column=%v, keys=%v", config.Mode, config.SubMode, config.Column, config.Keys)

	var dir string = BConf.AliceDir + "/publish"
	var fileBytes []byte
	var fileName string
	if config.FilePath != "" {
		fileBytes, err = ioutil.ReadFile(config.FilePath)
		if err != nil {
			Log.Warnf("failed to read upload data file. err=%v", err)
			fmt.Fprintf(w, RESPONSE_SAVE_FILE_FAILED)
			return
		}
		_, fileName = filepath.Split(config.FilePath)
	} else {
		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		r.ParseMultipartForm(10 << 20)
		// FormFile returns the first file for the given key `myFile`
		// it also returns the FileHeader so we can get the Filename,
		// the Header and the size of the file
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			Log.Warnf("failed to read upload data file. err=%v", err)
			fmt.Fprintf(w, RESPONSE_SAVE_FILE_FAILED)
			return
		}
		defer file.Close()

		fileBytes, err = ioutil.ReadAll(file)
		if err != nil {
			Log.Warnf("failed to read keystore file. err=%v", err)
			fmt.Fprintf(w, RESPONSE_SAVE_FILE_FAILED)
			return
		}
		fileName = fileHeader.Filename
	}

	folder, err := savePublishRawFile(fileBytes, dir, fileName, Log)
	if err != nil {
		Log.Warnf("failed to save keystore file. err=%v", err)
		fmt.Fprintf(w, RESPONSE_SAVE_FILE_FAILED)
		return
	}
	Log.Debugf("save raw file file...file dictoinary=%v", folder)

	err = publishRawData(folder, fileName, config.Mode, config.Column, config.Keys)
	if err != nil {
		Log.Warnf("failed to generate publish data. err=%v", err)
		fmt.Fprintf(w, RESPONSE_GENERATE_PUBLISH_FAILED)
		return
	}

	bulletin, err := readBulletinFile(folder+"/bulletin", Log)
	if err != nil {
		Log.Warnf("failed to read bulletin file.")
		fmt.Fprintf(w, RESPONSE_GENERATE_PUBLISH_FAILED)
		return
	}
	Log.Debugf("data has been published...merkle root=%v", bulletin.SigmaMKLRoot)
	plog.Detail = fmt.Sprintf("%v, merkle root=%v", plog.Detail, bulletin.SigmaMKLRoot)

	var extra PublishExtraInfo
	extra.Mode = config.Mode
	extra.SubMode = config.SubMode
	extra.UnitPrice = config.UnitPrice
	extra.Description = config.Description
	extra.MerkleRoot = bulletin.SigmaMKLRoot
	err = savePublishExtraInfo(extra, folder+"/extra.json")
	if err != nil {
		Log.Warnf("failed to save publish extra info. err=%v", err)
		fmt.Fprintf(w, RESPONSE_GENERATE_PUBLISH_FAILED)
		return
	}
	Log.Debugf("save extra info for publish data...filepath=%v", folder+"/extra.json")

	err = reNameFolder(folder, dir+"/"+bulletin.SigmaMKLRoot)
	if err != nil {
		Log.Warnf("failed to rename folder. err=%v", err)
		fmt.Fprintf(w, RESPONSE_GENERATE_PUBLISH_FAILED)
		return
	}
	Log.Infof("success to initialize data for publishing. merkle root=%v, filepath=%v", bulletin.SigmaMKLRoot, dir+"/"+bulletin.SigmaMKLRoot)

	plog.Result = LOG_RESULT_SUCCESS
	fmt.Fprintf(w, fmt.Sprintf(RESPONSE_SUCCESS, "success to initialize data, merkle root="+bulletin.SigmaMKLRoot))
	return
}

//PublishDataAPIHandler is the api handler for Alice to publish data to contract.
func PublishDataAPIHandler(w http.ResponseWriter, r *http.Request) {
	Log := Logger.NewSessionLogger()

	var plog PodLog
	plog.Result = LOG_RESULT_FAILED
	plog.Operation = LOG_OPERATION_TYPE_ALICE_PUBLISH

	defer func() {
		err := insertLogToDB(plog)
		if err != nil {
			Log.Warnf("insert log error! %v", err)
			return
		}
		nodeRecovery(w, Log)
	}()

	merkleRoot := r.FormValue("merkleRoot")
	value := r.FormValue("value")
	plog.Detail = fmt.Sprintf("merkleRoot=%v, deposit value=%v", merkleRoot, value)

	Log.Infof("start publish data to contract...merkleRoot=%v, value=%v", merkleRoot, value)
	dataFile := BConf.AliceDir + "/publish/" + merkleRoot

	valueInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		Log.Warnf("invalid value. value=%v, err=%v", value, err)
		fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
		return
	}
	if merkleRoot == "" {
		Log.Warnf("invalid merkle root. merkle root=%v, err=%v", merkleRoot, err)
		fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
		return
	}

	rs, err := pathExists(dataFile)
	if err != nil {
		Log.Errorf("check path exist error. err=%v", err)
		fmt.Fprintf(w, RESPONSE_READ_DATABASE_FAILED)
		return
	}
	if !rs {
		Log.Warnf("the data does not exist. filepath=%v", dataFile)
		fmt.Fprintf(w, RESPONSE_DATA_NOT_EXIST)
		return
	}

	b, err := readBulletinFile(dataFile+"/bulletin", Log)
	if err != nil {
		Log.Warnf("failed to read bulletin file. err=%v", err)
		fmt.Fprintf(w, RESPONSE_DATA_NOT_EXIST)
		return
	}
	Log.Debugf("read bulletin file...filepath=%v", dataFile+"/bulletin")

	extra, err := readExtraFile(dataFile + "/extra.json")
	if err != nil {
		Log.Errorf("failed to read publish extra info. err=%v", err)
		fmt.Fprintf(w, RESPONSE_DATA_NOT_EXIST)
		return
	}

	Log.Debugf("start send transaction to contract for publishing data...merkle root=%v, mode=%v, size=%v, n=%v, s=%v", b.SigmaMKLRoot, b.Mode, b.Size, b.N, b.S)
	t := time.Now()
	txid, err := publishDataToContract(b, valueInt)
	if err != nil {
		Log.Errorf("publish data to contract error! err=%v", err)
		fmt.Fprintf(w, RESPONSE_PUBLISH_TO_CONTRACT_FAILED)
		return
	}
	Log.Debugf("send publish data to contract successfully. txid=%v, merkle root=%v, time cost=%v", txid, b.SigmaMKLRoot, time.Since(t))

	extra.ContractAddr = BConf.ContractAddr
	err = savePublishExtraInfo(extra, dataFile+"/extra.json")
	if err != nil {
		Log.Errorf("failed to save publish extra info. err=%v")
	}
	Log.Infof("finish send transaction for publishing data...merkle root=%v", b.SigmaMKLRoot)

	plog.Result = LOG_RESULT_SUCCESS
	fmt.Fprintf(w, fmt.Sprintf(RESPONSE_SUCCESS, "send publish transaction to contract..."))
	return
}

//CloseDataAPIHandler is the api handler for Alice to close a published data in contract.
func CloseDataAPIHandler(w http.ResponseWriter, r *http.Request) {
	Log := Logger.NewSessionLogger()

	var plog PodLog
	plog.Result = LOG_RESULT_FAILED
	plog.Operation = LOG_OPERATION_TYPE_ALICE_CLOSE

	defer func() {
		err := insertLogToDB(plog)
		if err != nil {
			Log.Warnf("insert log error! %v", err)
			return
		}
		nodeRecovery(w, Log)
	}()

	merkleRoot := r.FormValue("merkle_root")
	Log.Infof("start close publish data from contract...merkleRoot=%v", merkleRoot)
	plog.Detail = fmt.Sprintf("merkleRoot=%v", merkleRoot)
	if merkleRoot == "" {
		Log.Warnf("invalid merkle root. merkle root=%v", merkleRoot)
		fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
		return
	}

	bulletinFile := BConf.AliceDir + "/publish/" + merkleRoot + "/bulletin"

	b, err := readBulletinFile(bulletinFile, Log)
	if err != nil {
		Log.Warnf("failed to read bulletin file. err=%v", err)
		fmt.Fprintf(w, RESPONSE_DATA_NOT_EXIST)
		return
	}
	Log.Debugf("read bulletin file...filepath=%v", bulletinFile)

	bltByte, err := calcuBltKey(b)
	if err != nil {
		Log.Warnf("failed to calculate bltKey. err=%v", err)
		fmt.Fprintf(w, RESPONSE_READ_CONTRACT_FAILED)
		return
	}

	status, err := readDataStatusAtContract(bltByte)
	if err != nil {
		Log.Errorf("read data status from contract error! err=%v", err)
		fmt.Fprintf(w, RESPONSE_READ_CONTRACT_FAILED)
		return
	}
	Log.Debugf("read data status from contract...status=%v", status)

	if status != "OK" {
		Log.Warnf("the data cannot been closed. status=%v", status)
		fmt.Fprintf(w, RESPONSE_UNPUBLISH_TO_CONTRACT_FAILED)
		return
	}

	Log.Debugf("start send transaction to contract for closing data...merkle root=%v, mode=%v, size=%v, n=%v, s=%v", b.SigmaMKLRoot, b.Mode, b.Size, b.N, b.S)
	t := time.Now()
	txid, rs, err := closeDataAtContract(bltByte)
	if err != nil {
		Log.Errorf("close data error. err=%v", err)
		fmt.Fprintf(w, RESPONSE_UNPUBLISH_TO_CONTRACT_FAILED)
		return
	}
	if !rs {
		Log.Warnf("failed to close the data.")
		fmt.Fprintf(w, RESPONSE_UNPUBLISH_TO_CONTRACT_FAILED)
		return
	}
	Log.Infof("success to send close transaction to contract. txid=%v, merkle root=%v, time cost=%v", txid, b.SigmaMKLRoot, time.Since(t))

	plog.Result = LOG_RESULT_SUCCESS
	fmt.Fprintf(w, fmt.Sprintf(RESPONSE_SUCCESS, "send transaction for closing data to contract..."))
	return
}

//AliceWithdrawFromDataAPIHandler is the api handler for Alice to withdraw ETH from closed data.
func AliceWithdrawFromDataAPIHandler(w http.ResponseWriter, r *http.Request) {
	Log := Logger.NewSessionLogger()

	var plog PodLog
	plog.Result = LOG_RESULT_FAILED
	plog.Operation = LOG_OPERATION_TYPE_ALICE_CLOSE

	defer func() {
		err := insertLogToDB(plog)
		if err != nil {
			Log.Warnf("insert log error! %v", err)
			return
		}
		nodeRecovery(w, Log)
	}()

	merkleRoot := r.FormValue("merkle_root")
	Log.Infof("start withdraw eth from closed publish data in contract...merkleRoot=%v", merkleRoot)
	if merkleRoot == "" {
		Log.Warnf("invalid merkle root. merkle root=%v", merkleRoot)
		fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
		return
	}
	plog.Detail = fmt.Sprintf("merkleRoot=%v", merkleRoot)

	bulletinFile := BConf.AliceDir + "/publish/" + merkleRoot + "/bulletin"
	b, err := readBulletinFile(bulletinFile, Log)
	if err != nil {
		Log.Warnf("failed to read bulletin file. err=%v", err)
		fmt.Fprintf(w, RESPONSE_DATA_NOT_EXIST)
		return
	}
	Log.Debugf("read bulletin file...filepath=%v", bulletinFile)

	bltByte, err := calcuBltKey(b)
	if err != nil {
		Log.Warnf("failed to calculate bltKey. err=%v", err)
		fmt.Fprintf(w, RESPONSE_READ_CONTRACT_FAILED)
		return
	}

	Log.Debugf("start send transaction to withdraw eth from closed publish data in contract...merkle root=%v, bltByte=%v", b.SigmaMKLRoot, bltByte)
	t := time.Now()
	txid, err := withdrawAETHFromContract(bltByte, Log)
	if err != nil {
		Log.Warnf("failed to withdraw ETH for Alice from contract. err=%v", err)
		fmt.Fprintf(w, RESPONSE_READ_CONTRACT_FAILED)
		return
	}
	Log.Debugf("success to send transaction to withdraw eth...txid=%v, merkle root=%v, bltByte=%v, time cost=%v", txid, b.SigmaMKLRoot, bltByte, time.Since(t))

	plog.Result = LOG_RESULT_SUCCESS
	fmt.Fprintf(w, fmt.Sprintf(RESPONSE_SUCCESS, "send transaction for withdrawing from contract..."))
	return
}

//AliceWithdrawFromTxAPIHandler is the api handler for Alice to withdraw ETH from transacion.
func AliceWithdrawFromTxAPIHandler(w http.ResponseWriter, r *http.Request) {
	Log := Logger.NewSessionLogger()

	var plog PodLog
	plog.Result = LOG_RESULT_FAILED
	plog.Operation = LOG_OPERATION_TYPE_ALICE_WITHDRAW_FROM_TX

	defer func() {
		err := insertLogToDB(plog)
		if err != nil {
			Log.Warnf("insert log error! %v", err)
			return
		}
		nodeRecovery(w, Log)
	}()

	sessionID := r.FormValue("session_id")
	Log.Infof("start withdraw eth from transaction in contract...sessionID=%v", sessionID)
	plog.Detail = fmt.Sprintf("sessionID=%v", sessionID)

	if sessionID == "" {
		Log.Warnf("invalid sessionID. sessionID=%v", sessionID)
		fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
		return
	}
	tx, rs, err := loadAliceFromDB(sessionID)
	if err != nil {
		Log.Warnf("failed to load transaction info from db. sessionID=%v, err=%v", sessionID, err)
		fmt.Fprintf(w, RESPONSE_READ_DATABASE_FAILED)
		return
	}
	if !rs {
		Log.Warnf("no transaction info loaded. sessionID=%v,", sessionID)
		fmt.Fprintf(w, RESPONSE_NO_NEED_TO_WITHDRAW)
		return
	}
	if tx.SubMode != TRANSACTION_SUB_MODE_COMPLAINT {
		Log.Warnf("the mode does not need withdraw eth.")
		fmt.Fprintf(w, RESPONSE_NO_NEED_TO_WITHDRAW)
		return
	}
	Log.Debugf("start send transaction to withdraw eth...sessionID=%v", sessionID)
	t := time.Now()
	txid, err := settleDealForComplaint(sessionID, tx.AliceAddr, tx.BobAddr)
	if err != nil {
		Log.Warnf("failed to withdraw eth for Alice from contract. err=%v", err)
		fmt.Fprintf(w, RESPONSE_READ_CONTRACT_FAILED)
		return
	}
	Log.Debugf("success to send transaction to withdraw eth...txid=%v, sessionID=%v, time cost=%v", txid, sessionID, time.Since(t))

	plog.Result = LOG_RESULT_SUCCESS
	fmt.Fprintf(w, fmt.Sprintf(RESPONSE_SUCCESS, "send transaction for withdrawing from contract..."))
	return
}

//BobPurchaseDataAPIHandler provides api for purchasing data for Bob.
func BobPurchaseDataAPIHandler(w http.ResponseWriter, r *http.Request) {
	Log := Logger.NewSessionLogger()

	Log.Infof("start purchase data...")
	var plog PodLog
	plog.Result = LOG_RESULT_FAILED
	plog.Operation = LOG_OPERATION_TYPE_BOB_TX
	defer func() {
		err := insertLogToDB(plog)
		if err != nil {
			Log.Warnf("insert log error! %v", err)
			return
		}
		nodeRecovery(w, Log)
	}()

	requestData := r.FormValue("request_data")
	var data RequestData
	err := json.Unmarshal([]byte(requestData), &data)
	if err != nil {
		Log.Warnf("invalid parameter. data=%v, err=%v", requestData, err)
		fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
		return
	}
	Log.Debugf("success to parse request data. data=%v", requestData)

	if data.MerkleRoot == "" || data.AliceIP == "" || data.AliceAddr == "" || data.BulletinFile == "" || data.PubPath == "" {
		Log.Warnf("invalid parameter. merkleRoot=%v, AliceIP=%v, AliceAddr=%v, bulletinFile=%v, PubPath=%v",
			data.MerkleRoot, data.AliceIP, data.AliceAddr, data.BulletinFile, data.PubPath)
		fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
		return
	}
	Log.Debugf("read parameters. merkleRoot=%v, AliceIP=%v, AliceAddr=%v, bulletinFile=%v, PubPath=%v",
		data.MerkleRoot, data.AliceIP, data.AliceAddr, data.BulletinFile, data.PubPath)

	plog.Detail = fmt.Sprintf("merkleRoot=%v, AliceIP=%v, AliceAddr=%v, bulletinFile=%v, PubPath=%v",
		data.MerkleRoot, data.AliceIP, data.AliceAddr, data.BulletinFile, data.PubPath)

	bulletin, err := readBulletinFile(data.BulletinFile, Log)
	if err != nil {
		Log.Warnf("failed to read bulletin File. err=%v", err)
		fmt.Fprintf(w, RESPONSE_PURCHASE_FAILED)
		return
	}
	plog.Detail = fmt.Sprintf("%v, merkle root=%v,", plog.Detail, bulletin.SigmaMKLRoot)

	Log.Debugf("step0: prepare for transaction...")
	var params = BobConnParam{data.AliceIP, data.AliceAddr, bulletin.Mode, data.SubMode, data.OT, data.UnitPrice, "", bulletin.SigmaMKLRoot}
	node, conn, params, err := preBobConn(params, ETHKey, Log)
	if err != nil {
		Log.Warnf("failed to prepare net for transaction. err=%v", err)
		fmt.Fprintf(w, RESPONSE_PURCHASE_FAILED)
		return
	}
	defer func() {
		if err := node.Close(); err != nil {
			fmt.Errorf("failed to close client node: %v", err)
		}
		if err := conn.Close(); err != nil {
			Log.Errorf("failed to close connection on client side: %v", err)
		}
	}()
	Log.Debugf("[%v]step0: success to establish connecting session with Alice. Alice IP=%v, Alice address=%v", params.SessionID, params.AliceIPAddr, params.AliceAddr)
	plog.Detail = fmt.Sprintf("%v, sessionID=%v,", plog.Detail, params.SessionID)
	plog.SessionId = params.SessionID

	var tx BobTransaction
	tx.SessionID = params.SessionID
	tx.Status = TRANSACTION_STATUS_START
	tx.Bulletin = bulletin
	tx.AliceIP = params.AliceIPAddr
	tx.AliceAddr = params.AliceAddr
	tx.Mode = params.Mode
	tx.SubMode = params.SubMode
	tx.OT = params.OT
	tx.UnitPrice = params.UnitPrice
	tx.BobAddr = fmt.Sprintf("%v", ETHKey.Address.Hex())

	Log.Debugf("[%v]step0: success to prepare for transaction...", params.SessionID)
	tx.Status = TRANSACTION_STATUS_START
	err = insertBobTxToDB(tx)
	if err != nil {
		Log.Warnf("failed to save transaction  to db for Bob. err=%v", err)
		fmt.Fprintf(w, fmt.Sprintf(RESPONSE_TRANSACTION_FAILED, "failed to save transaction to db for Bob."))
		return
	}

	var response string
	if tx.Mode == TRANSACTION_MODE_PLAIN_POD {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_COMPLAINT:
			if tx.OT {
				tx.Count = calcuPCnt(data.Phantoms)
				response = BobTxForPOC(node, ETHKey, tx, data.Demands, data.Phantoms, data.BulletinFile, data.PubPath, Log)
			} else {
				tx.Count = calcuDCnt(data.Demands)
				response = BobTxForPC(node, ETHKey, tx, data.Demands, data.BulletinFile, data.PubPath, Log)
			}
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP:
			tx.Count = calcuDCnt(data.Demands)
			response = BobTxForPAS(node, ETHKey, tx, data.Demands, data.BulletinFile, data.PubPath, Log)
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC:
			tx.Count = calcuDCnt(data.Demands)
			response = BobTxForPASVC(node, ETHKey, tx, data.Demands, data.BulletinFile, data.PubPath, Log)
		}
	} else if tx.Mode == TRANSACTION_MODE_TABLE_POD {
		switch tx.SubMode {
		case TRANSACTION_SUB_MODE_COMPLAINT:
			if tx.OT {
				tx.Count = calcuPCnt(data.Phantoms)
				response = BobTxForTOC(node, ETHKey, tx, data.Demands, data.Phantoms, data.BulletinFile, data.PubPath, Log)
			} else {
				tx.Count = calcuDCnt(data.Demands)
				response = BobTxForTC(node, ETHKey, tx, data.Demands, data.BulletinFile, data.PubPath, Log)
			}
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP:
			tx.Count = calcuDCnt(data.Demands)
			response = BobTxForTAS(node, ETHKey, tx, data.Demands, data.BulletinFile, data.PubPath, Log)
		case TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC:
			tx.Count = calcuDCnt(data.Demands)
			response = BobTxForTASVC(node, ETHKey, tx, data.Demands, data.BulletinFile, data.PubPath, Log)
		case TRANSACTION_SUB_MODE_VRF:
			tx.Count = 1
			if tx.OT {
				response = BobTxForTOQ(node, ETHKey, tx, data.KeyName, data.KeyValue, data.PhantomKeyValue, data.BulletinFile, data.PubPath, Log)
			} else {
				response = BobTxForTQ(node, ETHKey, tx, data.KeyName, data.KeyValue, data.BulletinFile, data.PubPath, Log)
			}
		}
	}
	var resp Response
	err = json.Unmarshal([]byte(response), &resp)
	if err != nil {
		Log.Warnf("failed to parse response. response=%v, err=%v", response, err)
		fmt.Fprintf(w, RESPONSE_FAILED_TO_RESPONSE)
		return
	}
	if resp.Code == "0" {
		plog.Result = LOG_RESULT_SUCCESS
	}
	Log.Debugf("[%v]the transaction finish. merkel root=%v, response=%v", params.SessionID, bulletin.SigmaMKLRoot, response)
	fmt.Fprintf(w, response)
	return
}

//BobDepositETHAPIHandler provides api for depositing eth for Bob from contract.
func BobDepositETHAPIHandler(w http.ResponseWriter, r *http.Request) {
	Log := Logger.NewSessionLogger()

	var plog PodLog
	plog.Result = LOG_RESULT_FAILED
	plog.Operation = LOG_OPERATION_TYPE_BOB_DEPOSIT

	defer func() {
		err := insertLogToDB(plog)
		if err != nil {
			Log.Warnf("insert log error! %v", err)
			return
		}
		nodeRecovery(w, Log)
	}()

	value := r.FormValue("value")
	AliceAddr := r.FormValue("address")
	Log.Infof("start deposit eth to Alice. value=%v, Alice address=%v", value, AliceAddr)
	plog.Detail = fmt.Sprintf("undeposit value=%v, Alice address=%v", value, AliceAddr)

	valueInt, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		Log.Warnf("parse value parameter failed. value=%v, err=%v", value, err)
		fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
		return
	}

	if AliceAddr == "" {
		Log.Warnf("incomplete parameter. Alice address=%v", AliceAddr)
		fmt.Fprintf(w, RESPONSE_INCOMPLETE_PARAM)
		return
	}
	Log.Debugf("read parameters. value=%v, Alice address=%v", value, AliceAddr)

	Log.Debugf("start send transaction to deposit eth to Alice in contract...value=%v, Alice address=%v", value, AliceAddr)
	t := time.Now()
	txid, err := BobDeposit(valueInt, AliceAddr)
	if err != nil {
		Log.Warnf("failed to deposit eth to contract. err=%v", err)
		fmt.Fprintf(w, RESPONSE_DEPOSIT_CONTRACT_FAILED)
		return
	}
	Log.Infof("success to send transaction to deposit eth...txid=%v, value=%v, Alice address=%v, time cost=%v", txid, value, AliceAddr, time.Since(t))

	plog.Result = LOG_RESULT_SUCCESS
	fmt.Fprintf(w, fmt.Sprintf(RESPONSE_SUCCESS, "depositing eth to contract..."))
	return

}

//BobUnDepositETHAPIHandler provides api for undepositing eth for Bob from contract.
func BobUnDepositETHAPIHandler(w http.ResponseWriter, r *http.Request) {
	Log := Logger.NewSessionLogger()

	var plog PodLog
	plog.Result = LOG_RESULT_FAILED
	plog.Operation = LOG_OPERATION_TYPE_BOB_UNDEPOSIT

	defer func() {
		err := insertLogToDB(plog)
		if err != nil {
			Log.Warnf("insert log error! %v", err)
			return
		}
		nodeRecovery(w, Log)
	}()

	AliceAddr := r.FormValue("address")
	Log.Infof("start undeposit eth from Alice in contract. Alice address=%v", AliceAddr)

	if AliceAddr == "" {
		Log.Warnf("incomplete parameter. Alice address=%v", AliceAddr)
		fmt.Fprintf(w, RESPONSE_UNDEPOSIT_CONTRACT_FAILED)
		return
	}
	plog.Detail = fmt.Sprintf("Alice address=%v", AliceAddr)

	Log.Debugf("start send transaction to undeposit eth to Alice in contract...Alice address=%v", AliceAddr)
	t := time.Now()
	txid, err := BobUnDeposit(AliceAddr)
	if err != nil {
		Log.Warnf("failed to undeposit eth to contract. err=%v", err)
		fmt.Fprintf(w, RESPONSE_UNDEPOSIT_CONTRACT_FAILED)
		return
	}
	Log.Infof("success to send transaction to undeposit eth...txid=%v, Alice address=%v, time cost=%v", txid, AliceAddr, time.Since(t))

	plog.Result = LOG_RESULT_SUCCESS
	fmt.Fprintf(w, fmt.Sprintf(RESPONSE_SUCCESS, "undepositing eth from contract..."))
	return
}

// BobWithdrawETHAPIHandler provides api for withdrawing ETH for Bob from contract.
func BobWithdrawETHAPIHandler(w http.ResponseWriter, r *http.Request) {
	Log := Logger.NewSessionLogger()

	var plog PodLog
	plog.Result = LOG_RESULT_FAILED
	plog.Operation = LOG_OPERATION_TYPE_BOB_WITHDRAW

	defer func() {
		err := insertLogToDB(plog)
		if err != nil {
			Log.Warnf("insert log error! %v", err)
			return
		}
		nodeRecovery(w, Log)
	}()

	AliceAddr := r.FormValue("address")
	Log.Infof("start withdraw eth from contract. Alice address=%v", AliceAddr)
	if AliceAddr == "" {
		Log.Warnf("incomplete parameter. Alice address=%v", AliceAddr)
		fmt.Fprintf(w, RESPONSE_WITHDRAW_CONTRACT_FAILED)
		return
	}
	plog.Detail = fmt.Sprintf("Alice address=%v", AliceAddr)

	Log.Debugf("start send transaction to withdraw eth from contract...Alice address=%v", AliceAddr)
	t := time.Now()
	txid, rs, err := BobWithdrawETHFromContract(AliceAddr)
	if err != nil {
		Log.Warnf("failed to deposit ETH to contract. err=%v", err)
		fmt.Fprintf(w, RESPONSE_WITHDRAW_CONTRACT_FAILED)
		return
	}
	if !rs {
		Log.Warnf("failed to deposit ETH to contract.")
		fmt.Fprintf(w, RESPONSE_WITHDRAW_CONTRACT_FAILED)
		return
	}
	Log.Infof("success to send transaction to withdraw eth...txid=%v, Alice address=%v, time cost=%v", txid, AliceAddr, time.Since(t))

	plog.Result = LOG_RESULT_SUCCESS
	fmt.Fprintf(w, fmt.Sprintf(RESPONSE_SUCCESS, "withdrawing eth from contract..."))
	return
}
