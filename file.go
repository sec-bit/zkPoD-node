package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/miguelmota/go-solidity-sha3"
)

//Batch1Secret is the struct of secret for batch1
type Batch1Secret struct {
	S string `json:"s"`
}

//Batch2Secret is the struct of secret for batch2
type Batch2Secret struct {
	S string `json:"s"`
}

//VRFSecret is the struct of secret for vrf
type VRFSecret struct {
	R string `json:"r"`
}

//Claim is the struct of claim
type Claim struct {
	I uint64   `json:"i"`
	J uint64   `json:"j"`
	K string   `json:"k"`
	M []string `json:"m"`
}

//Batch1Receipt is the struct of receipt for batch1
type Batch1Receipt struct {
	S string `json:"s"`
	K string `json:"k"`
	C uint64 `json:"c"`
}

//Batch2Receipt is the struct of receipt for batch2
type Batch2Receipt struct {
	S  string `json:"s"`
	VW string `json:"vw"`
	C  uint64 `json:"c"`
}

//VRFReceipt is the struct of receipt for vrf
type VRFReceipt struct {
	G string `json:"g"`
}

//Bulletin is the struct of bulletin.
type Bulletin struct {
	Mode         string `json:"mode"`
	Size         string `json:"size"`
	S            string `json:"s"`
	N            string `json:"n"`
	SigmaMKLRoot string `json:"sigma_mkl_root"`
}

func preBuyerTx(params BuyerConnParam, bulletinPath string, PubPath string, bulletin Bulletin, Log ILogger) (BuyerTransaction, error) {
	var tx BuyerTransaction
	dir := BConf.BuyerDir + "/transaction/" + params.SessionID
	err := saveBulletinAndPublic(dir, bulletinPath, PubPath, Log)
	if err != nil {
		Log.Warnf("failed to save bulletin for buyer. err=%v", err)
		return tx, errors.New("failed to save file")
	}
	Log.Debugf("[%v]success to save bulletin and public information...", params.SessionID)

	tx.SessionID = params.SessionID
	tx.Status = TRANSACTION_STATUS_START
	tx.Bulletin = bulletin
	tx.SellerIP = params.SellerIPAddr
	tx.SellerAddr = params.SellerAddr
	tx.Mode = params.Mode
	tx.SubMode = params.SubMode
	tx.OT = params.OT
	tx.UnitPrice = params.UnitPrice
	tx.BuyerAddr = fmt.Sprintf("%v", ETHKey.Address.Hex())
	return tx, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func copyKeyStore(dir string, keyStore string) error {
	rd, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(dir)
	}()
	exist := false
	for _, fi := range rd {
		if fi.IsDir() {
			continue
		} else {
			exist = true
			err = copyFile(dir+"/"+fi.Name(), keyStore)
			if err != nil {
				return err
			}
		}
	}
	if !exist {
		return fmt.Errorf("keystore file does not exist")
	}
	return nil
}

func savePublishFileForTransaction(sessionID string, mklroot string, Log ILogger) error {

	dir := BConf.SellerDir + "/publish/" + mklroot
	txDir := BConf.SellerDir + "/transaction"

	rs, err := pathExists(txDir)
	if err != nil {
		Log.Errorf("check path exist error. err=%v", err)
		return errors.New("save publish file error")
	}
	if !rs {
		err = os.Mkdir(txDir, os.ModePerm)
		if err != nil {
			Log.Errorf("create dictionary %v error. err=%v", txDir, err)
			return errors.New("save publish file error")
		}
		Log.Debugf("success to create dictionary. dir=%v", txDir)
	}

	txPath := txDir + "/" + sessionID
	err = os.Mkdir(txPath, os.ModePerm)
	if err != nil {
		Log.Errorf("create dictionary %v error. err=%v", txPath, err)
		return errors.New("save publish file error")
	}
	Log.Debugf("success to create dictionary. dir=%v", txPath)

	err = copyFile(dir+"/bulletin", txPath+"/bulletin")
	if err != nil {
		Log.Warnf("failed to copy bulletin file. err=%v", err)
		return errors.New("failed to copy bulletin file")
	}
	Log.Debugf("success to copy bulletin file to %v", txPath+"/bulletin")

	err = os.Mkdir(txPath+"/public", os.ModePerm)
	if err != nil {
		Log.Errorf("create dictionary %v error. err=%v", txPath+"/public", err)
		return errors.New("save public file error")
	}
	Log.Debugf("success to create dictionary. dir=%v", txPath+"/public")

	err = copyDir(dir+"/public", txPath+"/public")
	if err != nil {
		Log.Warnf("failed to copy public file. err=%v", err)
		return errors.New("Failed to copy public file")
	}
	Log.Debugf("success to copy public dictionary to %v", txPath+"/public")

	err = os.Mkdir(txPath+"/private", os.ModePerm)
	if err != nil {
		Log.Errorf("create dictionary %v error. err=%v", txPath+"/private", err)
		return errors.New("save private file error")
	}
	Log.Debugf("success to create dictionary. dir=%v", txPath+"/private")

	err = copyDir(dir+"/private", txPath+"/private")
	if err != nil {
		Log.Warnf("failed to copy private file. err=%v", err)
		return errors.New("Failed to copy private file")
	}
	Log.Debugf("success to copy private dictionary to %v", txPath+"/private")

	err = copyFile(dir+"/extra.json", txPath+"/extra.json")
	if err != nil {
		Log.Warnf("failed to copy extra file. err=%v", err)
		return errors.New("Failed to copy extra file")
	}
	Log.Debugf("success to copy extra file to %v", txPath+"/extra.json")

	return nil
}

func saveBulletinAndPublic(dir string, bulletin string, PubPath string, Log ILogger) error {

	rs, err := pathExists(dir)
	if err != nil {
		Log.Errorf("check dictionary %v error. err=%v", dir, err)
		return errors.New("Save bulletin file error")
	}
	if !rs {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			Log.Errorf("create dictionary %v error. err=%v", dir, err)
			return errors.New("Save bulletin file error")
		}
	}

	err = copyFile(bulletin, dir+"/bulletin")
	if err != nil {
		Log.Warnf("Failed to copy bulletin file. err=%v", err)
		return errors.New("Failed to copy bulletin file")
	}
	Log.Debugf("success to save bulletin. path=%v", dir+"/bulletin")

	err = os.Mkdir(dir+"/public", os.ModePerm)
	if err != nil {
		Log.Errorf("create dictionary %v error. err=%v", dir+"/public", err)
		return errors.New("Save bulletin file error")
	}
	err = copyDir(PubPath, dir+"/public")
	if err != nil {
		Log.Warnf("Failed to copy public file. err=%v", err)
		return errors.New("Failed to copy public file")
	}
	Log.Debugf("success to save public file. path=%v", dir+"/public")

	return nil
}

func readBulletinFile(filePath string, Log ILogger) (Bulletin, error) {

	var b Bulletin
	bl, err := ioutil.ReadFile(filePath)
	if err != nil {
		Log.Warnf("failed to open bulletin file. err=%v", err)
		return b, errors.New("failed to open bulletin file")
	}
	Log.Debugf("success to read bulletin info from file %v", filePath)

	err = json.Unmarshal(bl, &b)
	if err != nil {
		Log.Warnf("failed to parse bulletin file. err=%v", err)
		return b, errors.New("failed to parse bulletin file")
	}
	return b, nil
}

func readBatch1Receipt(filePath string, Log ILogger) ([]byte, Batch1Receipt, error) {
	var r Batch1Receipt
	rf, err := ioutil.ReadFile(filePath)
	if err != nil {
		Log.Warnf("failed to read receipt file. err=%v", err)
		return rf, r, errors.New("failed to read receipt file")
	}
	Log.Debugf("success to read receipt info from file %v", filePath)

	err = json.Unmarshal(rf, &r)
	if err != nil {
		Log.Warnf("failed to parse receipt file. err=%v", err)
		return rf, r, errors.New("failed to parse receipt file")
	}
	Log.Debugf("success to parse receipt info")
	return rf, r, nil
}

func signRecptForBatch1(key *keystore.Key, sessionID string, receipt Batch1Receipt, price int64, expireAt int64, Log ILogger) ([]byte, error) {

	sessionInt := new(big.Int)
	sessionInt, rs := sessionInt.SetString(sessionID, 10)
	if !rs {
		Log.Warnf("failed to convert sessionId.")
		return nil, errors.New("convert sessionId error")
	}

	receiptHash := solsha3.SoliditySHA3( // types
		[]string{"uint256", "address", "bytes32", "bytes32", "uint64", "uint256", "uint256"},

		// values
		[]interface{}{
			sessionInt,
			fmt.Sprintf("%v", key.Address.Hex()),
			"0x" + receipt.S,
			"0x" + receipt.K,
			receipt.C,
			big.NewInt(price),
			big.NewInt(expireAt),
		})

	Log.Debugf("generate receipt receiptHash=0x%02x\n", receiptHash)

	receiptHash = solsha3.SoliditySHA3WithPrefix(receiptHash)

	// Buyer sign the receipt
	sig, err := crypto.Sign(receiptHash, key.PrivateKey)
	if err != nil {
		Log.Warnf("generate signature for receipt hash error. err=%v", err)
		return nil, errors.New("signature error")
	}
	Log.Debugf("success to generate signature")
	return sig, nil
}

func readBatch2Receipt(filePath string, Log ILogger) ([]byte, Batch2Receipt, error) {
	var r Batch2Receipt
	rf, err := ioutil.ReadFile(filePath)
	if err != nil {
		Log.Warnf("failed to read receipt file. err=%v", err)
		return rf, r, errors.New("failed to read receipt file")
	}
	// config := string(conf)

	err = json.Unmarshal(rf, &r)
	if err != nil {
		Log.Warnf("failed to parse receipt file. err=%v", err)
		return rf, r, errors.New("failed to parse receipt file")
	}

	return rf, r, nil
}

func signRecptForBatch2(key *keystore.Key, sessionID string, receipt Batch2Receipt, price int64, expireAt int64, Log ILogger) ([]byte, error) {

	sessionInt := new(big.Int)
	sessionInt, rs := sessionInt.SetString(sessionID, 10)
	if !rs {
		Log.Warnf("failed to convert sessionId.")
		return nil, errors.New("convert sessionId error")
	}

	vwInt := new(big.Int)
	vwInt, rs = vwInt.SetString(receipt.VW, 10)
	if !rs {
		Log.Warnf("failed to convert vw. vw=%v", receipt.VW)
		return nil, errors.New("failed to convert vw")
	}

	receiptHash := solsha3.SoliditySHA3( // types
		[]string{"uint256", "address", "bytes32", "uint256", "uint64", "uint256", "uint256"},

		// values
		[]interface{}{
			sessionInt,
			fmt.Sprintf("%v", key.Address.Hex()),
			"0x" + receipt.S,
			vwInt,
			receipt.C,
			big.NewInt(price),
			big.NewInt(expireAt),
		})

	// Log.Debugf("sessionInt=%v, addr=%v,S=%v,vwInt=%v,count=%v,price=%v,expireAt=%v",
	// 	sessionInt,
	// 	fmt.Sprintf("%v", key.Address.Hex()),
	// 	"0x"+receipt.S,
	// 	vwInt,
	// 	receipt.C,
	// 	big.NewInt(price),
	// 	big.NewInt(expireAt))
	Log.Debugf("generate receipt receiptHash=0x%02x\n", receiptHash)

	receiptHash = solsha3.SoliditySHA3WithPrefix(receiptHash)

	// Buyer sign the receipt
	sig, err := crypto.Sign(receiptHash, key.PrivateKey)
	if err != nil {
		Log.Warnf("generate signature for receipt hash error. err=%v", err)
		return nil, errors.New("signature error")
	}
	Log.Debugf("generate receipt signature=%v", hexutil.Encode(sig))

	return sig, nil
}

func readVRFReceipt(filePath string, Log ILogger) ([]byte, VRFReceipt, error) {
	var r VRFReceipt
	rf, err := ioutil.ReadFile(filePath)
	if err != nil {
		Log.Warnf("failed to read receipt file. err=%v", err)
		return rf, r, errors.New("failed to read receipt file")
	}
	// config := string(conf)

	err = json.Unmarshal(rf, &r)
	if err != nil {
		Log.Warnf("failed to parse receipt file. err=%v", err)
		return rf, r, errors.New("failed to parse receipt file")
	}

	return rf, r, nil
}

func signRecptForVRFQ(key *keystore.Key, sessionID string, receipt VRFReceipt, price int64, expireAt int64, Log ILogger) ([]byte, error) {

	sessionInt := new(big.Int)
	sessionInt, rs := sessionInt.SetString(sessionID, 10)
	if !rs {
		Log.Warnf("failed to convert sessionId.")
		return nil, errors.New("convert sessionId error")
	}

	gs := strings.Split(receipt.G, " ")
	if len(gs) != 3 {
		Log.Warnf("invalid g.  g=%v", receipt.G)
		return nil, errors.New("invalid g. ")
	}

	var _g_exp_r [2]*big.Int
	_g_exp_r[0] = new(big.Int)
	_g_exp_r[0], rs = _g_exp_r[0].SetString(gs[1], 10)
	if !rs {
		Log.Warnf("failed to convert g[1]. g[1] = %v", gs[1])
		return nil, errors.New("failed to convert g[1]")
	}
	_g_exp_r[1] = new(big.Int)
	_g_exp_r[1], rs = _g_exp_r[1].SetString(gs[2], 10)
	if !rs {
		Log.Warnf("failed to convert g[2]. g[2] = %v", gs[2])
		return nil, errors.New("failed to convert g[2]")
	}

	receiptHash := solsha3.SoliditySHA3( // types
		[]string{"uint256", "address", "uint256", "uint256", "uint256", "uint256"},

		// values
		[]interface{}{
			sessionInt,
			fmt.Sprintf("%v", key.Address.Hex()),
			_g_exp_r[0],
			_g_exp_r[1],
			big.NewInt(price),
			big.NewInt(expireAt),
		})

	Log.Debugf("generate receipt receiptHash=0x%02x\n", receiptHash)

	receiptHash = solsha3.SoliditySHA3WithPrefix(receiptHash)

	// Buyer sign the receipt
	sig, err := crypto.Sign(receiptHash, key.PrivateKey)
	if err != nil {
		Log.Warnf("generate signature for receipt hash error. err=%v", err)
		return nil, errors.New("signature error")
	}
	return sig, nil
}

func readSeed0ForBatch1(filePath string, Log ILogger) (s Batch1Secret, err error) {

	sf, err := ioutil.ReadFile(filePath)
	if err != nil {
		Log.Warnf("failed to open seed2 file. err=%v", err)
		return s, errors.New("failed to open seed2 file")
	}
	// config := string(conf)
	err = json.Unmarshal(sf, &s)
	if err != nil {
		Log.Warnf("failed to parse seed2 file. err=%v", err)
		return s, errors.New("failed to parse seed2 file")
	}

	return
}

func buyerSaveSecretForBatch1(secret Batch1Secret, filePath string, Log ILogger) error {
	sByte, err := json.Marshal(&secret)
	if err != nil {
		Log.Warnf("Failed to save secret for buyer. err=%v")
		return errors.New("Failed to save secret for buyer")
	}
	err = ioutil.WriteFile(filePath, sByte, 0644)
	if err != nil {
		Log.Errorf("Failed to save bulletin file. err=%v", err)
		return errors.New("Save bulletin file error")
	}
	return nil
}

func readSeed0ForBatch2(filePath string, Log ILogger) (s Batch2Secret, err error) {

	sf, err := ioutil.ReadFile(filePath)
	if err != nil {
		Log.Warnf("failed to open seed2 file. err=%v", err)
		return s, errors.New("failed to open seed2 file")
	}
	// config := string(conf)
	err = json.Unmarshal(sf, &s)
	if err != nil {
		Log.Warnf("failed to parse seed2 file. err=%v", err)
		return s, errors.New("failed to parse seed2 file")
	}

	return
}

func buyerSaveSecretForBatch2(secret Batch2Secret, filePath string, Log ILogger) error {
	sByte, err := json.Marshal(&secret)
	if err != nil {
		Log.Warnf("Failed to save secret for buyer. err=%v")
		return errors.New("Failed to save secret for buyer")
	}
	err = ioutil.WriteFile(filePath, sByte, 0644)
	if err != nil {
		Log.Errorf("Failed to save bulletin file. err=%v", err)
		return errors.New("Save bulletin file error")
	}
	return nil
}

func readSeed0ForVRFQ(filePath string, Log ILogger) (s VRFSecret, err error) {

	sf, err := ioutil.ReadFile(filePath)
	if err != nil {
		Log.Warnf("failed to open seed2 file. err=%v", err)
		return s, errors.New("failed to open seed2 file")
	}
	err = json.Unmarshal(sf, &s)
	if err != nil {
		Log.Warnf("failed to parse seed2 file. err=%v", err)
		return s, errors.New("failed to parse seed2 file")
	}

	return
}

func buyerSaveSecretForVRFQ(secret VRFSecret, filePath string, Log ILogger) error {
	sByte, err := json.Marshal(&secret)
	if err != nil {
		Log.Warnf("Failed to save secret for buyer. err=%v")
		return errors.New("Failed to save secret for buyer")
	}
	err = ioutil.WriteFile(filePath, sByte, 0644)
	if err != nil {
		Log.Errorf("Failed to save bulletin file. err=%v", err)
		return errors.New("Save bulletin file error")
	}
	return nil
}

func readClaim(filePath string, Log ILogger) (Claim, error) {
	var c Claim
	clamin, err := ioutil.ReadFile(filePath)
	if err != nil {
		Log.Warnf("failed to read claim file. err=%v", err)
		return c, errors.New("failed to read claim file")
	}

	err = json.Unmarshal(clamin, &c)
	if err != nil {
		Log.Warnf("failed to parse claim file. err=%v", err)
		return c, errors.New("failed to parse claim file")
	}
	return c, nil
}

func copyFile(src string, dest string) error {
	inputSrc, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, inputSrc, 0644)
	if err != nil {
		return err
	}
	return nil
}

func copyDir(src string, dest string) error {

	err := filepath.Walk(src, func(src string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {

			err = copyDir(f.Name(), dest+"/"+f.Name())
			if err != nil {
				return err
			}
		} else {

			dest_new := dest + "/" + f.Name()
			copyFile(src, dest_new)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func saveBasicConfig(config BasicConfig, basicFile string) error {

	configByte, err := json.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal config, err=%v", err)
	}

	f, err := os.Create(basicFile)
	if err != nil {
		return fmt.Errorf("failed to open basic file=%v. err=%v", basicFile, err)
	}
	defer f.Close()

	_, err = f.Write(configByte)
	if err != nil {
		return fmt.Errorf("failed to write config to basic file. file=%v, err=%v", basicFile, err)
	}
	return nil
}

func savePublishRawFile(fileByte []byte, sellerPublishDir string, dataFile string, Log ILogger) (string, error) {
	ts := fmt.Sprintf("%v", time.Now().UnixNano())

	rs, err := pathExists(sellerPublishDir)
	if err != nil {
		Log.Errorf("check path exist error. err=%v", err)
		return "", errors.New("save publish file error")
	}
	if !rs {
		err = os.Mkdir(sellerPublishDir, os.ModePerm)
		if err != nil {
			Log.Errorf("create dictionary %v error. err=%v", sellerPublishDir, err)
			return "", errors.New("save publish file error")
		}
		Log.Debugf("success to create dictionary. dir=%v", sellerPublishDir)
	}
	dir := sellerPublishDir + "/" + ts
	err = os.Mkdir(dir, os.ModePerm)
	if err != nil {
		Log.Errorf("create dictionary %v error. err=%v", dir, err)
		return dir, errors.New("save publish file error")
	}
	Log.Debugf("success to create dictionary. dir=%v", dir)

	filePath := dir + "/" + dataFile
	f, err := os.Create(filePath)
	if err != nil {
		return dir, fmt.Errorf("failed to open publish data file=%v. err=%v", filePath, err)
	}
	defer f.Close()

	_, err = f.Write(fileByte)
	if err != nil {
		return dir, fmt.Errorf("failed to write file to publish data file. file=%v, err=%v", filePath, err)
	}
	return dir, nil
}

func publishRawData(filePath string, fileName string, mode string, column string, keys []int) error {

	var cmd *exec.Cmd
	if mode == TRANSACTION_MODE_PLAIN_POD {
		cmd = exec.Command(BConf.PublishBINPath, "-e", BConf.ECCBINPath, "-m", mode, "-f", filePath+"/"+fileName, "-o", filePath, "-c", column)
	} else {
		fs := strings.Split(fileName, ".")
		parameters := []string{"-e", BConf.ECCBINPath, "-m", mode, "-f", filePath + "/" + fileName, "-o", filePath, "-t", fs[len(fs)-1], "-k"}
		for _, k := range keys {
			parameters = append(parameters, fmt.Sprintf("%v", k))
		}
		cmd = exec.Command(BConf.PublishBINPath, parameters...)
	}
	fmt.Printf("exec command: %v\n", cmd)

	rs, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("publish execution error! %v, %v", err, string(rs))
	}
	return nil
}

func reNameFolder(filePath string, newPath string) error {
	err := os.Rename(filePath, newPath)
	if err != nil {
		return err
	}
	return nil
}

func savePublishExtraInfo(extra PublishExtraInfo, path string) error {

	extraByte, err := json.Marshal(&extra)
	if err != nil {
		return fmt.Errorf("failed to marshal publish extra info, err=%v", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create extra file=%v. err=%v", path, err)
	}
	defer f.Close()

	_, err = f.Write(extraByte)
	if err != nil {
		return fmt.Errorf("failed to write config to extra file. file=%v, err=%v", path, err)
	}
	return nil
}

func readExtraFile(filepath string) (extra PublishExtraInfo, err error) {

	ex, err := ioutil.ReadFile(filepath)
	if err != nil {
		return
	}

	err = json.Unmarshal(ex, &extra)
	if err != nil {
		return
	}
	return
}
