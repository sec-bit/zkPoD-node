package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/sec-bit/zkPoD-lib/pod_go/ecc"
	"github.com/urfave/cli"
)

//Config is the struct of config info
type Config struct {
	Operation             string      `json:"operation"`
	Password              string      `json:"password"`
	AliceAddress         string      `json:"Alice_address"`
	BasicConfig           BasicConfig `json:"basic_config"`
	RequestData           RequestData `json:"request_data"`
	PurchaseConfigPath    string      `json:"purchase_path"`
	InitPublishConfigPath string      `json:"init_path"`
	MerkleRoot            string      `json:"merkle_root"`
	ETHValue              string      `json:"value"`
	SessionID             string      `json:"session_id"`
}

type BasicConfig struct {
	ECCBINPath     string `json:"ecc_bin_path"`
	PublishBINPath string `json:"publish_bin_path"`
	ContractAddr   string `json:"contract_addr"`
	BobDir         string `json:"B_dir"`
	AliceDir      string `json:"A_dir"`
	KeyStoreFile   string `json:"keystore_file"`
	NetIP          string `json:"net_ip"`
	Port           string `json:"port"`
}

// RequestData is the struct of Bob's request data
type RequestData struct {
	MerkleRoot      string    `json:"merkle_root"`
	AliceIP        string    `json:"Alice_ip"`
	AliceAddr      string    `json:"Alice_addr"`
	PubPath         string    `json:"pub_path"`
	BulletinFile    string    `json:"bulletin_file"`
	SubMode         string    `json:"sub_mode"`
	OT              bool      `json:"ot"`
	Demands         []Demand  `json:"demands"`
	Phantoms        []Phantom `json:"phantoms"`
	KeyName         string    `json:"key_name"`
	KeyValue        []string  `json:"key_value"`
	PhantomKeyValue []string  `json:"phantom_key_value"`
	UnitPrice       int64     `json:"unit_price"`
}

// Demand is the struct of demand
type Demand struct {
	DemandStart uint64 `json:"demand_start"`
	DemandCount uint64 `json:"demand_count"`
}

// Phantom is the struct of phantom for ot mode
type Phantom struct {
	PhantomStart uint64 `json:"phantom_start"`
	PhantomCount uint64 `json:"phantom_count"`
}

// initCli reads command line inputs.
func initCli() (config Config) {

	app := cli.NewApp()

	app.Version = "0.0.1"
	app.Name = "secbit-pod-node"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "SECBIT Labs",
			Email: "hi@secbit.io",
		},
	}
	app.Copyright = "(c) 2019 SECBIT Labs"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "o",
			Value: "start",
			Usage: "operation type \n" +
				" start, initdata, publish, close, purchase, withdraw, deposit, undeposit.",
		},
		cli.StringFlag{
			Name:  "k",
			Value: "",
			Usage: "keystore file path.",
		},
		cli.StringFlag{
			Name:  "pass",
			Value: "",
			Usage: "the password of keystore file.",
		},
		cli.StringFlag{
			Name:  "ip",
			Value: "",
			Usage: "the Alice's node ip addr for pod net.",
		},
		cli.StringFlag{
			Name:  "port",
			Value: "",
			Usage: "the api server port.",
		},
		cli.StringFlag{
			Name:  "c",
			Value: "",
			Usage: "config file path while Bob's operation is 'purchase'.",
		},
		cli.StringFlag{
			Name:  "mkl",
			Value: "",
			Usage: "the sigma merkle root.",
		},
		cli.StringFlag{
			Name:  "eth",
			Value: "0",
			Usage: "the ETH's value(wei) for withdrawing or depositing.",
		},
		cli.StringFlag{
			Name:  "init",
			Value: "",
			Usage: "the Alice initializes a file for publishing.",
		},
		cli.StringFlag{
			Name:  "addr",
			Value: "",
			Usage: "the Alice's address.",
		},
		cli.StringFlag{
			Name:  "sid",
			Value: "",
			Usage: "the transaction's session Id.",
		},
	}
	var exit = true
	app.Action = func(c *cli.Context) error {
		exit = false
		config.Operation = c.String("o")
		config.BasicConfig.NetIP = c.String("ip")
		config.BasicConfig.Port = c.String("port")
		config.BasicConfig.KeyStoreFile = c.String("k")
		config.Password = c.String("pass")
		config.PurchaseConfigPath = c.String("c")
		config.MerkleRoot = c.String("mkl")
		config.ETHValue = c.String("eth")
		config.InitPublishConfigPath = c.String("init")
		config.AliceAddress = c.String("addr")
		config.SessionID = c.String("sid")
		// config.BasicConfig.SyncthingID = c.String("id")
		// config.BasicConfig.DataBase = c.String("db")

		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		return
	}
	if exit {
		os.Exit(1)
	}
	return
}

func readBasicFile(basic BasicConfig) (BasicConfig, error) {
	basicFile := DEFAULT_BASIC_CONFGI_FILE
	rs, err := pathExists(basicFile)
	if err != nil {
		return basic, fmt.Errorf("failed to read basic config file. config file=%v, err=%v", basicFile, err)
	}
	if rs {
		conf, err := ioutil.ReadFile(basicFile)
		if err != nil {
			return basic, fmt.Errorf("failed to read config file. config file=%v, err=%v", basicFile, err)
		}
		var preBasic BasicConfig
		err = json.Unmarshal(conf, &preBasic)
		if err != nil {
			return basic, fmt.Errorf("failed to parse basic config file. err=%v", err)
		}
		if basic.NetIP == "" {
			basic.NetIP = preBasic.NetIP
		}
		if basic.Port == "" {
			basic.Port = preBasic.Port
		}
		if basic.KeyStoreFile == "" {
			basic.KeyStoreFile = preBasic.KeyStoreFile
		}
		basic.ECCBINPath = preBasic.ECCBINPath
		basic.PublishBINPath = preBasic.PublishBINPath
		basic.BobDir = preBasic.BobDir
		basic.AliceDir = preBasic.AliceDir
		basic.ContractAddr = preBasic.ContractAddr
	}
	if basic.Port == "" {
		basic.Port = DEFAULT_SERVER_PORT
	}
	if basic.KeyStoreFile == "" {
		basic.KeyStoreFile = DEFAULT_KEYSTORE_FILE
	}
	if basic.NetIP == "" {
		basic.NetIP = DEFAULT_NET_IP
	}
	if basic.ECCBINPath == "" {
		basic.ECCBINPath = DEFAULT_ECC_FILE
	}
	if basic.PublishBINPath == "" {
		basic.PublishBINPath = DEFAULT_PUBLISH_BIN_FILE
	}
	if basic.BobDir == "" {
		basic.BobDir = DEFAULT_BOB_DIR
	}
	if basic.AliceDir == "" {
		basic.AliceDir = DEFAULT_ALICE_DIR
	}
	if basic.ContractAddr == "" {
		basic.ContractAddr = DEFAULT_PODEX_CONTRACT_ADDRESS
	}
	return basic, nil
}

func saveBasicFile(basic BasicConfig) error {

	basicFile := DEFAULT_BASIC_CONFGI_FILE
	err := saveBasicConfig(basic, basicFile)
	if err != nil {
		return err
	}
	return nil
}

// readKeyStore reads keystore file.
func readKeyStore(config BasicConfig, password string) (*keystore.Key, error) {

	Log := Logger.NewSessionLogger()

	var key *keystore.Key

	if password == "" {
		Log.Warnf("no password to read!")
		return nil, fmt.Errorf("no keystore password!")
	}
	rs, err := pathExists(config.KeyStoreFile)
	if err != nil {
		Log.Errorf("failed to check key store file. err=%v", err)
		return nil, errors.New("failed to check key store file")
	}

	if !rs {
		var tmpKeystore = "./tmp" + time.Now().String()
		ks := keystore.NewKeyStore(tmpKeystore, keystore.StandardScryptN, keystore.StandardScryptP)
		account, err := ks.NewAccount(password)
		if err != nil {
			Log.Warnf("failed to create a account. err=%v", err)
			return nil, errors.New("failed to create a account")
		}
		err = copyKeyStore(tmpKeystore, config.KeyStoreFile)
		if err != nil {
			Log.Errorf("failed to save key store file. err=%v", err)
			return nil, errors.New("failed to save key store file")
		}

		Log.Infof("create a new account finish. keystore file=%v, ethaddr=%v.", config.KeyStoreFile, account.Address.Hex())
	}

	key, err = initKeyStore(config.KeyStoreFile, password, Log)
	if err != nil {
		Log.Errorf("failed to initialize key store file. err=%v", err)
		return nil, errors.New("invalid key store file")
	}
	Log.Infof("initialize key store finish. ethaddr=%v.", key.Address.Hex())

	if config.ContractAddr == "" {
		Log.Warnf("invalid contract addr. contract=%v", config.ContractAddr)
		return nil, errors.New("invalid contract addr")
	}
	err = ConnectToProvider(key, config.ContractAddr, Log)
	if err != nil {
		Log.Warnf("failed to connect to provider for contract. err=%v", err)
		return nil, errors.New("failed to connect to provider for contract")
	}
	Log.Infof("success to connect to provider for contract")

	return key, nil
}

// initKeyStore inits key store for ethereum account.
func initKeyStore(keystoreFile string, password string, Log ILogger) (key *keystore.Key, err error) {

	var jsonBytes []byte
	jsonBytes, err = ioutil.ReadFile(keystoreFile)
	if err != nil {
		Log.Warnf("failed to read keystore file. err=%v", err)
		return nil, errors.New("failed to read keystore file")
	}
	key, err = keystore.DecryptKey(jsonBytes, password)
	if err != nil {
		Log.Warnf("failed to read keystore file. err=%v", err)
		return nil, errors.New("failed to read keystore file")
	}
	return
}

func readRequestData(configFile string) (RequestData, error) {

	var requestData RequestData
	if configFile == "" {
		configFile = DEFAULT_CONFIG_FILE
	}
	rs, err := pathExists(configFile)
	if err != nil {
		return requestData, fmt.Errorf("failed to read config file. config file=%v, err=%v", configFile, err)
	}
	if !rs {
		return requestData, fmt.Errorf("the config file does not exist. config file=%v", configFile)
	}

	conf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return requestData, fmt.Errorf("failed to read config file. config file=%v, err=%v", configFile, err)
	}

	err = json.Unmarshal(conf, &requestData)
	if err != nil {
		return requestData, fmt.Errorf("failed to parse config file. config path=%v, err=%v", string(conf), err)
	}
	return requestData, nil
}

//initDir initializes node and check dictionary
func initDir(BobDir string, AliceDir string) error {

	if BobDir == "" || AliceDir == "" {
		fmt.Errorf("invalid dictionary.  Bob dictionary=%v, Alice dictionary=%v", BobDir, AliceDir)
	}
	BobTxDir := BobDir + "/transaction"
	AliceTxDir := AliceDir + "/transaction"
	AlicePublishDir := AliceDir + "/publish"

	rs, err := pathExists(BobDir)
	if err != nil {
		return fmt.Errorf("failed to check dictionary. spath=%v, err=%v", BobDir, err)
	}
	if !rs {
		err = os.Mkdir(BobDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("create dictionary %v error. err=%v", BobDir, err)
		}
		fmt.Printf("success to create dictionary=%v.\n", BobDir)
	}

	rs, err = pathExists(BobTxDir)
	if err != nil {
		return fmt.Errorf("check dictonary exist error. path=%v, err=%v", BobTxDir, err)
	}
	if !rs {
		err = os.Mkdir(BobTxDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("create dictionary %v error. err=%v", BobTxDir, err)
		}
		fmt.Printf("success to create dictionary=%v.\n", BobTxDir)
	}

	rs, err = pathExists(AliceDir)
	if err != nil {
		return fmt.Errorf("check dictonary error. path=%v, err=%v", AliceDir, err)
	}
	if !rs {
		err = os.Mkdir(AliceDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("create dictionary %v error. err=%v", AliceDir, err)
		}
		fmt.Printf("success to create dictionary=%v.\n", AliceDir)
	}

	rs, err = pathExists(AliceTxDir)
	if err != nil {
		return fmt.Errorf("check dictonary error. path=%v, err=%v", AliceTxDir, err)
	}
	if !rs {
		err = os.Mkdir(AliceTxDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("create dictionary %v error. err=%v", AliceTxDir, err)
		}
		fmt.Printf("success to create dictionary=%v.\n", AliceTxDir)
	}

	rs, err = pathExists(AlicePublishDir)
	if err != nil {
		return fmt.Errorf("check dictonary error. path=%v, err=%v", AlicePublishDir, err)
	}
	if !rs {
		err = os.Mkdir(AlicePublishDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("create dictionary %v error. err=%v", AlicePublishDir, err)
		}
		fmt.Printf("success to create dictionary=%v.\n", AlicePublishDir)
	}
	return nil
}

func initMap() {
	AliceTxMap = make(map[string]Transaction)
	BobTxMap = make(map[string]BobTransaction)
	DepositLockMap = make(map[string]int64)
}

func preparePOD(EccPubFile string) bool {

	ecc.Init()
	return ecc.Load(EccPubFile)
}
