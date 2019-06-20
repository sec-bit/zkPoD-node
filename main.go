package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sec-bit/zkPoD-node/net/rlpx"

	"github.com/ethereum/go-ethereum/accounts/keystore"
)

var Logger = NewSimpleLogger("zkPoD-node")

var ETHKey *keystore.Key
var BConf BasicConfig
var ServerAddr *rlpx.Addr
var SellerNodeStart bool = false
var BuyerNodeStart bool = false
var SellerTxMap map[string]Transaction
var BuyerTxMap map[string]BuyerTransaction
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
		b := preparePOD(BConf.ECCBINPath)
		if !b {
			panic("failed to prepare ecc")
		}
		err := initDir(BConf.BuyerDir, BConf.SellerDir)
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
	http.HandleFunc("/s/withdraw/data", SellerWithdrawFromDataAPIHandler)
	http.HandleFunc("/s/withdraw/tx", SellerWithdrawFromTxAPIHandler)

	http.HandleFunc("/b/purchase", BuyerPurchaseDataAPIHandler)
	http.HandleFunc("/b/deposit", BuyerDepositETHAPIHandler)
	http.HandleFunc("/b/undeposit", BuyerUnDepositETHAPIHandler)
	http.HandleFunc("/b/withdraw", BuyerWithdrawETHAPIHandler)

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
		BuyerNodeStart = true
		fmt.Printf("\n====================(●'◡'●)ﾉ♥====================\n")
		fmt.Printf("buyer node start....\n\n")
		break
	}
	for {
		if BConf.NetIP == "" || GAdminAuth == nil || ZkPoDExchangeClient == nil {
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Printf("\n====================(づ｡◕‿‿◕｡)づ====================\n")
		fmt.Printf("seller node start....\n\n")
		err := SellerStartNode(BConf.NetIP, ETHKey, Log)
		if err != nil {
			Log.Errorf("failed to start node.")
			SellerNodeStart = false
		}
	}
}

//HandleCmdReq handles requests from command line.
func HandleCmdReq(config Config) {
	Log := Logger.NewSessionLogger()
	switch config.Operation {
	case OPERATION_SELLER_INITDATA:
		Log.Debugf("start initdata for publishing...config path=%v", config.InitPublishConfigPath)
		if config.InitPublishConfigPath == "" {
			Log.Warnf("parameter is incomplete. please input 'init'")
			panic("parameter is incomplete")
		}
		SellerInitDataNode(config.InitPublishConfigPath, Log)
	case OPERATION_SELLER_PUBLISH:
		Log.Debugf("start publish data...merkle root=%v, eth value=%v", config.MerkleRoot, config.ETHValue)
		if config.MerkleRoot == "" {
			Log.Warnf("parameter is incomplete. please input 'mkl' and 'eth'")
			panic("parameter is incomplete")
		}
		SellerPublishData(config.MerkleRoot, config.ETHValue, Log)
	case OPERATION_SELLER_CLOSE:
		Log.Debugf("start close published data...merkle root=%v", config.MerkleRoot)
		if config.MerkleRoot == "" {
			Log.Warnf("parameter is incomplete. please input 'mkl'")
			panic("parameter is incomplete")
		}
		SellerCloseData(config.MerkleRoot, Log)
	case OPERATION_WITHDRAW:
		if config.MerkleRoot != "" {
			Log.Debugf("start withdraw for seller...merkle root=%v", config.MerkleRoot)
			SellerWithdrawETHForData(config.MerkleRoot, Log)
		} else if config.SessionID != "" {
			Log.Debugf("start withdraw for seller...session id=%v", config.SessionID)
			SellerWithdrawETHForTx(config.SessionID, Log)
		} else if config.SellerAddress != "" {
			Log.Debugf("start withdraw for buyer...seller address=%v", config.SellerAddress)
			BuyerWithdrawETH(config.SellerAddress, Log)
		} else {
			Log.Warnf("parameters are incomplete. please input 'mkl' or 'addr'")
			panic("parameters are incomplete")
		}
	case OPERATION_BUYER_PURCHASE:
		Log.Debugf("start purchase data...config path=%v", config.PurchaseConfigPath)
		if config.PurchaseConfigPath == "" {
			Log.Warnf("parameter is incomplete. please input 'data'")
			panic("parameter is incomplete")
		}
		requestData, err := readRequestData(config.PurchaseConfigPath)
		if err != nil {
			panic(err)
		}
		BuyerPurchaseData(requestData, Log)
	case OPERATION_BUYER_DEPOSIT:
		Log.Debugf("start deposit for buyer...seller address=%v, eth value=%v", config.SellerAddress, config.ETHValue)
		if config.SellerAddress == "" || config.ETHValue == "" {
			Log.Warnf("parameters are incomplete. please input 'data'")
			panic("parameters are incomplete")
		}
		BuyerDepositETH(config.ETHValue, config.SellerAddress, Log)
	case OPERATION_BUYER_UNDEPOSIT:
		Log.Debugf("start undeposit for buyer...seller address=%v", config.SellerAddress)
		if config.SellerAddress == "" {
			Log.Warnf("parameter is incomplete. please input 'addr'")
			panic("parameter is incomplete")
		}
		BuyerUnDepositETH(config.SellerAddress, Log)
	default:
		Log.Warnf("invalid operation. operation=%v", config.Operation)
		panic("invalid operation")
	}
}
