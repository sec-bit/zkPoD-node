package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/sec-bit/zkPoD-node/net/rlpx"
)

var Logger = NewSimpleLogger("zkPoD-node")

var ETHKey *keystore.Key
var BConf BasicConfig
var ServerAddr *rlpx.Addr
var AliceNodeStart bool = false
var BobNodeStart bool = false
var AliceTxMap map[string]Transaction
var BobTxMap map[string]BobTransaction
var DepositLockMap map[string]int64

func main() {
	var err error
	// read command
	config := initCli()
	// read basic config in file.
	BConf, err = readBasicFile(config.BasicConfig)
	if err != nil {
		panic(err)
	}

	if config.Operation == OPERATION_START {
		initMap()
		err = saveBasicFile(BConf)
		if err != nil {
			panic(err)
		}
		ETHKey, err = readKeyStore(BConf, config.Password)
		if err != nil {
			panic(err)
		}
		b := preparePOD(BConf.InitKeyPath)
		if !b {
			panic("failed to prepare setup bin")
		}
		err := initDir(BConf.BobDir, BConf.AliceDir)
		if err != nil {
			panic(err)
		}
		DBConn, err = connectSQLite(DEFAULT_DB_PATH)
		if err != nil {
			panic(err)
		}
		SetupPOD()
	} else {
		HandleCmdReq(config)
	}
}

// SetupPOD setups api server and node-
func SetupPOD() {

	Log := Logger.NewSessionLogger()
	fmt.Printf("         ██╗      ████████╗             ████████╗ \n")
	fmt.Printf("         ██║      ██╔════██╗            ██╔════███╗\n")
	fmt.Printf("███████╗ ██║  ██╗ ██║    ██║    ███╗    ██║      ██║\n")
	fmt.Printf("╚══███╔╝ ██║ ██╔╝ ████████╔╝  ██╔══██╗  ██║      ██║\n")
	fmt.Printf("  ███╔╝  █████╔╝  ██╔═════╝  ██╔╝   ██║ ██║      ██║\n")
	fmt.Printf(" ███╔╝   ██╔═██╗  ██║         ██╗  ██╔╝ ██║    ███╔╝\n")
	fmt.Printf("███████╗ ██║  ██╗ ██║          ╚███╔═╝  ████████╔═╝\n")
	fmt.Printf("╚══════╝ ╚═╝  ╚═╝ ╚═╝           ╚══╝    ╚═══════╝\n")

	go SetupAPIServe(Log)
	SetupNode(Log)
}

//SetupAPIServe setups the api server for command line requests and web requests.
func SetupAPIServe(Log ILogger) {
	http.HandleFunc("/config/read", ReadCfgAPIHandler)
	http.HandleFunc("/config/reload", ReLoadConfigAPIHandler)

	http.HandleFunc("/s/publish/init", InitPublishDataAPIHandler)
	http.HandleFunc("/s/publish", PublishDataAPIHandler)
	http.HandleFunc("/s/close", CloseDataAPIHandler)
	http.HandleFunc("/s/withdraw/data", AliceWithdrawFromDataAPIHandler)
	http.HandleFunc("/s/withdraw/tx", AliceWithdrawFromTxAPIHandler)

	http.HandleFunc("/b/purchase", BobPurchaseDataAPIHandler)
	http.HandleFunc("/b/deposit", BobDepositETHAPIHandler)
	http.HandleFunc("/b/undeposit", BobUnDepositETHAPIHandler)
	http.HandleFunc("/b/withdraw", BobWithdrawETHAPIHandler)

	fmt.Printf("====================≖‿≖✧====================\n")
	fmt.Printf("API server start...\n")
	fmt.Printf("port=%v\n\n", BConf.Port)

	http.ListenAndServe(":"+BConf.Port, nil)
}

//SetupNode setups the node for p2p net.
func SetupNode(Log ILogger) {

	Log.Infof("setup node...")
	for {
		if GAdminAuth == nil || ZkPoDExchangeClient == nil {
			time.Sleep(2 * time.Second)
			continue
		}
		BobNodeStart = true
		fmt.Printf("\n====================(●'◡'●)ﾉ♥====================\n")
		fmt.Printf("Bob node start....\n\n")
		break
	}
	for {
		if BConf.NetIP == "" || GAdminAuth == nil || ZkPoDExchangeClient == nil {
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Printf("\n====================(づ｡◕‿‿◕｡)づ====================\n")
		fmt.Printf("Alice node start....\n\n")
		err := AliceStartNode(BConf.NetIP, ETHKey, Log)
		if err != nil {
			Log.Errorf("failed to start node.")
			AliceNodeStart = false
		}
	}
}

//HandleCmdReq handles requests from command line.
func HandleCmdReq(config Config) {
	Log := Logger.NewSessionLogger()
	switch config.Operation {
	case OPERATION_ALICE_INITDATA:
		Log.Debugf("start initdata for publishing...config path=%v", config.InitPublishConfigPath)
		if config.InitPublishConfigPath == "" {
			Log.Warnf("parameter is incomplete. please input 'init'")
			panic("parameter is incomplete")
		}
		AliceInitDataNode(config.InitPublishConfigPath, Log)
	case OPERATION_ALICE_PUBLISH:
		Log.Debugf("start publish data...merkle root=%v, eth value=%v", config.MerkleRoot, config.ETHValue)
		if config.MerkleRoot == "" {
			Log.Warnf("parameter is incomplete. please input 'mkl' and 'eth'")
			panic("parameter is incomplete")
		}
		AlicePublishData(config.MerkleRoot, config.ETHValue, Log)
	case OPERATION_ALICE_CLOSE:
		Log.Debugf("start close published data...merkle root=%v", config.MerkleRoot)
		if config.MerkleRoot == "" {
			Log.Warnf("parameter is incomplete. please input 'mkl'")
			panic("parameter is incomplete")
		}
		AliceCloseData(config.MerkleRoot, Log)
	case OPERATION_WITHDRAW:
		if config.MerkleRoot != "" {
			Log.Debugf("start withdraw for Alice...merkle root=%v", config.MerkleRoot)
			AliceWithdrawETHForData(config.MerkleRoot, Log)
		} else if config.SessionID != "" {
			Log.Debugf("start withdraw for Alice...session id=%v", config.SessionID)
			AliceWithdrawETHForTx(config.SessionID, Log)
		} else if config.AliceAddress != "" {
			Log.Debugf("start withdraw for Bob...Alice address=%v", config.AliceAddress)
			BobWithdrawETH(config.AliceAddress, Log)
		} else {
			Log.Warnf("parameters are incomplete. please input 'mkl' or 'addr'")
			panic("parameters are incomplete")
		}
	case OPERATION_BOB_PURCHASE:
		Log.Debugf("start purchase data...config path=%v", config.PurchaseConfigPath)
		if config.PurchaseConfigPath == "" {
			Log.Warnf("parameter is incomplete. please input 'data'")
			panic("parameter is incomplete")
		}
		requestData, err := readRequestData(config.PurchaseConfigPath)
		if err != nil {
			panic(err)
		}
		BobPurchaseData(requestData, Log)
	case OPERATION_BOB_DEPOSIT:
		Log.Debugf("start deposit for Bob...Alice address=%v, eth value=%v", config.AliceAddress, config.ETHValue)
		if config.AliceAddress == "" || config.ETHValue == "" {
			Log.Warnf("parameters are incomplete. please input 'data'")
			panic("parameters are incomplete")
		}
		BobDepositETH(config.ETHValue, config.AliceAddress, Log)
	case OPERATION_BOB_UNDEPOSIT:
		Log.Debugf("start undeposit for Bob...Alice address=%v", config.AliceAddress)
		if config.AliceAddress == "" {
			Log.Warnf("parameter is incomplete. please input 'addr'")
			panic("parameter is incomplete")
		}
		BobUnDepositETH(config.AliceAddress, Log)
	default:
		Log.Warnf("invalid operation. operation=%v", config.Operation)
		panic("invalid operation")
	}
}
