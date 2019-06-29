package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/miguelmota/go-solidity-sha3"
)

// ComplaintRequest is the structure of request for complaint request
type ComplaintRequest struct {
	S string   `json:"s"`
	D []string `json:"d"`
}

// OtComplaintRequest is the structure of request for ot complaint request
type OtComplaintRequest struct {
	S string   `json:"s"`
	P []string `json:"p"`
}

// AtomicSwapRequest is the structure of request for atomic_swap request
type AtomicSwapRequest struct {
	S string   `json:"s"`
	P []string `json:"p"`
}

// AtomicSwapVcRequest is the structure of request for atomic_swap_vc request
type AtomicSwapVcRequest struct {
	S string   `json:"s"`
	P []string `json:"p"`
}

//ComplaintSecret is the structure of secret for complaint mode
type ComplaintSecret struct {
	S string `json:"s"`
}

//AtomicSwapSecret is the structure of secret for atomic swap mode
type AtomicSwapSecret struct {
	S string `json:"s"`
}

//AtomicSwapVcSecret is the structure of secret for atomic swap mode
type AtomicSwapVcSecret struct {
	S string `json:"s"`
	R string `json:"r"`
}

//VRFSecret is the structure of secret for vrf mode
type VRFSecret struct {
	R string `json:"r"`
}

//Claim is the structure of claim
type Claim struct {
	I uint64   `json:"i"`
	J uint64   `json:"j"`
	K string   `json:"k"`
	M []string `json:"m"`
}

//ComplaintReceipt is the struct of receipt for complaint mode
type ComplaintReceipt struct {
	S string `json:"s"`
	K string `json:"k"`
	C uint64 `json:"c"`
}

//AtomicSwapReceipt is the struct of receipt for atomic swap mode
type AtomicSwapReceipt struct {
	S  string `json:"s"`
	VW string `json:"vw"`
	C  uint64 `json:"c"`
}

//AtomicSwapVcReceipt is the struct of receipt for atomic swap mode
type AtomicSwapVcReceipt struct {
	D string `json:"d"`
}

//VRFReceipt is the struct of receipt for vrf mode
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

func readReceiptForComplaint(filePath string, Log ILogger) ([]byte, ComplaintReceipt, error) {
	var r ComplaintReceipt
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

func signRecptForComplaint(key *keystore.Key, sessionID string, receipt ComplaintReceipt, price int64, expireAt int64, Log ILogger) ([]byte, error) {

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

	// Bob sign the receipt
	sig, err := crypto.Sign(receiptHash, key.PrivateKey)
	if err != nil {
		Log.Warnf("generate signature for receipt hash error. err=%v", err)
		return nil, errors.New("signature error")
	}
	Log.Debugf("success to generate signature")
	return sig, nil
}

func readReceiptForAtomicSwap(filePath string, Log ILogger) ([]byte, AtomicSwapReceipt, error) {
	var r AtomicSwapReceipt
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

func signRecptForAtomicSwap(key *keystore.Key, sessionID string, receipt AtomicSwapReceipt, price int64, expireAt int64, Log ILogger) ([]byte, error) {

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

	// Bob sign the receipt
	sig, err := crypto.Sign(receiptHash, key.PrivateKey)
	if err != nil {
		Log.Warnf("generate signature for receipt hash error. err=%v", err)
		return nil, errors.New("signature error")
	}
	Log.Debugf("generate receipt signature=%v", hexutil.Encode(sig))

	return sig, nil
}

func readReceiptForAtomicSwapVc(filePath string, Log ILogger) ([]byte, AtomicSwapVcReceipt, error) {
	var r AtomicSwapVcReceipt
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

func signRecptForAtomicSwapVc(key *keystore.Key, sessionID string, receipt AtomicSwapVcReceipt, price int64, expireAt int64, Log ILogger) ([]byte, error) {

	sessionInt := new(big.Int)
	sessionInt, rs := sessionInt.SetString(sessionID, 10)
	if !rs {
		Log.Warnf("failed to convert sessionId.")
		return nil, errors.New("convert sessionId error")
	}
	digestInt := new(big.Int)
	digestInt, rs = digestInt.SetString(receipt.D, 10)
	if !rs {
		Log.Warnf("failed to convert digestInt. receipt.D=%v", receipt.D)
		return nil, errors.New("failed to convert digestInt")
	}

	receiptHash := solsha3.SoliditySHA3( // types
		[]string{"uint256", "address", "uint256", "uint256", "uint256"},

		// values
		[]interface{}{
			sessionInt,
			fmt.Sprintf("%v", key.Address.Hex()),
			digestInt,
			big.NewInt(price),
			big.NewInt(expireAt),
		})

	Log.Debugf("generate receipt receiptHash=0x%02x\n", receiptHash)

	receiptHash = solsha3.SoliditySHA3WithPrefix(receiptHash)

	// Bob sign the receipt
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

	// Bob sign the receipt
	sig, err := crypto.Sign(receiptHash, key.PrivateKey)
	if err != nil {
		Log.Warnf("generate signature for receipt hash error. err=%v", err)
		return nil, errors.New("signature error")
	}
	return sig, nil
}

func readSeed0ForComplaint(filePath string, Log ILogger) (s ComplaintSecret, err error) {

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

func BobSaveSecretForComplaint(secret ComplaintSecret, filePath string, Log ILogger) error {
	sByte, err := json.Marshal(&secret)
	if err != nil {
		Log.Warnf("Failed to save secret for Bob. err=%v")
		return errors.New("Failed to save secret for Bob")
	}
	err = ioutil.WriteFile(filePath, sByte, 0644)
	if err != nil {
		Log.Errorf("Failed to save bulletin file. err=%v", err)
		return errors.New("Save bulletin file error")
	}
	return nil
}

func readSeed0ForAtomicSwap(filePath string, Log ILogger) (s AtomicSwapSecret, err error) {

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

func BobSaveSecretForAtomicSwap(secret AtomicSwapSecret, filePath string, Log ILogger) error {
	sByte, err := json.Marshal(&secret)
	if err != nil {
		Log.Warnf("Failed to save secret for Bob. err=%v")
		return errors.New("Failed to save secret for Bob")
	}
	err = ioutil.WriteFile(filePath, sByte, 0644)
	if err != nil {
		Log.Errorf("Failed to save bulletin file. err=%v", err)
		return errors.New("Save bulletin file error")
	}
	return nil
}

func readSeed0ForAtomicSwapVc(filePath string, Log ILogger) (s AtomicSwapVcSecret, err error) {

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

func BobSaveSecretForAtomicSwapVc(secret AtomicSwapVcSecret, filePath string, Log ILogger) error {
	sByte, err := json.Marshal(&secret)
	if err != nil {
		Log.Warnf("failed to save secret for Bob. err=%v")
		return errors.New("failed to save secret for Bob")
	}
	err = ioutil.WriteFile(filePath, sByte, 0644)
	if err != nil {
		Log.Errorf("failed to save bulletin file. err=%v", err)
		return errors.New("save bulletin file error")
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

func BobSaveSecretForVRFQ(secret VRFSecret, filePath string, Log ILogger) error {
	sByte, err := json.Marshal(&secret)
	if err != nil {
		Log.Warnf("Failed to save secret for Bob. err=%v")
		return errors.New("Failed to save secret for Bob")
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

func savePublishRawFile(fileByte []byte, AlicePublishDir string, dataFile string, Log ILogger) (string, error) {
	ts := fmt.Sprintf("%v", time.Now().UnixNano())

	rs, err := pathExists(AlicePublishDir)
	if err != nil {
		Log.Errorf("check path exist error. err=%v", err)
		return "", errors.New("save publish file error")
	}
	if !rs {
		err = os.Mkdir(AlicePublishDir, os.ModePerm)
		if err != nil {
			Log.Errorf("create dictionary %v error. err=%v", AlicePublishDir, err)
			return "", errors.New("save publish file error")
		}
		Log.Debugf("success to create dictionary. dir=%v", AlicePublishDir)
	}
	dir := AlicePublishDir + "/" + ts
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
		cmd = exec.Command(BConf.PublishBINPath, "-m", mode, "-f", filePath+"/"+fileName, "-o", filePath, "-c", column)
	} else {
		fs := strings.Split(fileName, ".")
		parameters := []string{"-m", mode, "-f", filePath + "/" + fileName, "-o", filePath, "-t", fs[len(fs)-1], "-k"}
		for _, k := range keys {
			parameters = append(parameters, fmt.Sprintf("%v", k))
		}
		cmd = exec.Command(BConf.PublishBINPath, parameters...)
	}
	// fmt.Printf("exec command: %v\n", cmd)

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

func calcuCntforComplaint(requestFile string) (int64, error) {
	var req ComplaintRequest
	request, err := ioutil.ReadFile(requestFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read request file. err=%v", err)
	}

	err = json.Unmarshal(request, &req)
	if err != nil {
		return 0, fmt.Errorf("failed to read request file. err=%v", err)
	}

	var count int64
	for _, d := range req.D {
		cs := strings.Split(d, "-")
		if len(cs) != 2 {
			return 0, fmt.Errorf("invalid demand. %v", d)
		}
		c, err := strconv.ParseInt(cs[len(cs)-1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid demand. %v", d)
		}
		count += c
	}
	return count, nil
}

func calcuCntforOtComplaint(requestFile string) (int64, error) {
	var req OtComplaintRequest
	request, err := ioutil.ReadFile(requestFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read request file. err=%v", err)
	}

	err = json.Unmarshal(request, &req)
	if err != nil {
		return 0, fmt.Errorf("failed to read request file. err=%v", err)
	}

	var count int64
	for _, p := range req.P {
		cs := strings.Split(p, "-")
		if len(cs) != 2 {
			return 0, fmt.Errorf("invalid phantoms. %v", p)
		}
		c, err := strconv.ParseInt(cs[len(cs)-1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid phantoms. %v", p)
		}
		count += c
	}
	return count, nil
}

func calcuCntforAS(requestFile string) (int64, error) {
	var req AtomicSwapRequest
	request, err := ioutil.ReadFile(requestFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read request file. err=%v", err)
	}

	err = json.Unmarshal(request, &req)
	if err != nil {
		return 0, fmt.Errorf("failed to read request file. err=%v", err)
	}

	var count int64
	for _, p := range req.P {
		cs := strings.Split(p, "-")
		if len(cs) != 2 {
			return 0, fmt.Errorf("invalid demands. %v", p)
		}
		c, err := strconv.ParseInt(cs[len(cs)-1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid demands. %v", p)
		}
		count += c
	}
	return count, nil
}

func calcuCntforASVC(requestFile string) (int64, error) {
	var req AtomicSwapRequest
	request, err := ioutil.ReadFile(requestFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read request file. err=%v", err)
	}

	err = json.Unmarshal(request, &req)
	if err != nil {
		return 0, fmt.Errorf("failed to read request file. err=%v", err)
	}

	var count int64
	for _, p := range req.P {
		cs := strings.Split(p, "-")
		if len(cs) != 2 {
			return 0, fmt.Errorf("invalid demands. %v", p)
		}
		c, err := strconv.ParseInt(cs[len(cs)-1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid demands. %v", p)
		}
		count += c
	}
	return count, nil
}
