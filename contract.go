package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/miguelmota/go-solidity-sha3"
)

var ProviderURL = "https://ropsten.infura.io/v3/198fa8670c224bc49a56f4828fcad728"

var GAdminAuth *bind.TransactOpts
var ZkPoDExchangeClient *ZkPoDExchange

//ConnectToProvider connects to contract
func ConnectToProvider(key *keystore.Key, zkPODEXAddr string, Log ILogger) error {

	fmt.Println(key.Address.Hex())
	/**
	 * Connecting to provider
	 */
	var err error
	Log.Infof("Dial with provider...url=%v", ProviderURL)
	client, err := ethclient.Dial(ProviderURL)
	if err != nil {
		Log.Warnf("Failed to dial provider, err=%v", err)
		return err
	}

	auth := bind.NewKeyedTransactor(key.PrivateKey)

	zkPODEXContractAddr := common.HexToAddress(zkPODEXAddr)
	pClient, err := NewZkPoDExchange(zkPODEXContractAddr, client)
	if err != nil {
		return err
	}
	GAdminAuth = auth
	ZkPoDExchangeClient = pClient
	Log.Infof("connect to contract...contract addr=%v", zkPODEXAddr)
	return nil
}

func publishDataToContract(b Bulletin, value int64) (string, error) {

	size, _ := strconv.ParseUint(b.Size, 10, 64)
	s, _ := strconv.ParseUint(b.S, 10, 64)
	n, _ := strconv.ParseUint(b.N, 10, 64)
	mklrootInt := new(big.Int)
	mklroot, rs := mklrootInt.SetString(b.SigmaMKLRoot, 16)
	if !rs {
		return "", errors.New("failed to convert sessionId")
	}

	GAdminAuth.Value = big.NewInt(value)
	defer func() {
		GAdminAuth.Value = big.NewInt(0)
	}()

	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.Publish(GAdminAuth, size, s, n, mklroot, big.NewInt(0), big.NewInt(0))
	if err != nil {
		return "", fmt.Errorf("failed to Publish: %v", err)
	}
	return ctx.Hash().Hex(), nil
}

func closeDataAtContract(bltByte [32]byte) (string, bool, error) {

	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.UnPublish(GAdminAuth, bltByte)
	if err != nil {
		return "", false, fmt.Errorf("failed to close published data. err=%v", err)
	}
	return ctx.Hash().Hex(), true, nil
}

func readDataStatusAtContract(bltByte [32]byte) (string, error) {

	bulletins, err := ZkPoDExchangeClient.ZkPoDExchangeCaller.Bulletins(&bind.CallOpts{}, bltByte)
	if err != nil {
		return "", fmt.Errorf("failed to read bulletins. err=%v", err)
	}

	if bulletins.Stat == 0 {
		return "OK", nil
	} else if bulletins.Stat == 1 {
		return "CANCELING", nil
	} else if bulletins.Stat == 2 {
		return "CANCELED", nil
	} else {
		return "UNKNOWN", nil
	}
}

func withdrawAETHFromContract(bltByte [32]byte, Log ILogger) (string, error) {
	t := time.Now()
	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.WithdrawA(GAdminAuth, bltByte)
	if err != nil {
		Log.Warnf("failed to withdraw from closed data: %v", err)
		return "", errors.New("Failed to withdraw")
	}
	log.Printf("withdraw from close data...time cost=%v\n", time.Since(t))
	return ctx.Hash().Hex(), nil
}

func BobDeposit(value int64, AliceAddr string) (string, error) {

	GAdminAuth.Value = big.NewInt(value)
	defer func() {
		GAdminAuth.Value = big.NewInt(0)
	}()

	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.BobDeposit(GAdminAuth, common.HexToAddress(AliceAddr))
	if err != nil {
		return "", fmt.Errorf("failed to deposit to %v in contract. err=%v", AliceAddr, err)
	}
	return ctx.Hash().Hex(), nil
}

func BobUnDeposit(AliceAddr string) (string, error) {
	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.BobUnDeposit(GAdminAuth, common.HexToAddress(AliceAddr))
	if err != nil {
		return "", fmt.Errorf("failed to undeposit to  %v. err=%v", AliceAddr, err)
	}
	return ctx.Hash().Hex(), nil
}

func BobWithdrawETHFromContract(AliceAddr string) (string, bool, error) {
	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.WithdrawB(GAdminAuth, common.HexToAddress(AliceAddr))
	if err != nil {
		return "", false, fmt.Errorf("failed to withdraw from closed data: %v", err)
	}
	return ctx.Hash().Hex(), true, nil
}

func submitScrtForComplaint(tx Transaction, receiptSign []byte, Log ILogger) (string, error) {
	seed0Path := BConf.AliceDir + "/transaction/" + tx.SessionID + "/secret"
	receiptPath := BConf.AliceDir + "/transaction/" + tx.SessionID + "/receipt"

	_, receipt, err := readReceiptForComplaint(receiptPath, Log)
	if err != nil {
		Log.Warnf("failed to read receipt. err=%v", err)
		return "", errors.New("failed to read receipt")
	}
	seed2Byte, err := hex.DecodeString(receipt.S)
	if err != nil {
		Log.Warnf("failed to decode seed2. seed2string=%v, err=%v", receipt.S, err)
		return "", errors.New("failed to decode seed2")
	}
	mklByte, err := hex.DecodeString(receipt.K)
	if err != nil {
		Log.Warnf("failed to decode mkl root. mklroot=%v, err=%v", receipt.K, err)
		return "", errors.New("failed to decode mkl root")
	}

	sessionInt := new(big.Int)
	sessionInt, rs := sessionInt.SetString(tx.SessionID, 10)
	if !rs {
		Log.Warnf("failed to convert sessionId.")
		return "", errors.New("failed to convert sessionId")
	}

	err = verifySigntureForComplaint(tx, sessionInt, receiptSign, receipt, Log)
	if err != nil {
		Log.Warnf("verify signature error! err=%v", err)
		return "", errors.New("verify signature error")
	}

	secret, err := readSeed0ForComplaint(seed0Path, Log)
	if err != nil {
		Log.Warnf("failed to read seed0. err=%v", err)
		return "", errors.New("failed to read seed0")
	}
	seedByte, err := hex.DecodeString(secret.S)
	if err != nil {
		Log.Warnf("failed to decode seed. seedstring=%v, err=%v", secret.S, err)
		return "", errors.New("failed to parse seed file")
	}

	bltKey, err := calcuBltKey(tx.Bulletin)
	if err != nil {
		Log.Warnf("failed to calculate blt key. tx.Bulletin=%v, err=%v", tx.Bulletin, err)
		return "", errors.New("failed to calculate blt key")
	}

	t := time.Now()
	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.SubmitProofComplaint(GAdminAuth, bltKey, *byte32(seedByte), sessionInt,
		common.HexToAddress(tx.BobAddr), *byte32(seed2Byte), *byte32(mklByte), receipt.C, big.NewInt(tx.Price), big.NewInt(tx.ExpireAt), receiptSign)
	if err != nil {
		Log.Warnf("failed to submit proof. err=%v", err)
		return "", errors.New("failed to submit proof")
	}
	Log.Debugf("POD submit proof successfully,  time cost=%v\n",
		time.Since(t))
	return ctx.Hash().Hex(), nil
}

func verifySigntureForComplaint(tx Transaction, sessionInt *big.Int, receiptSign []byte, receipt ComplaintReceipt, Log ILogger) error {

	if len(receiptSign) != 65 {
		Log.Warnf("invalid signature. sig=%v, len(sig)=%v", hexutil.Encode(receiptSign), len(receiptSign))
		return errors.New("invalid signature")
	}

	receiptHash := solsha3.SoliditySHA3( // types
		[]string{"uint256", "address", "bytes32", "bytes32", "uint64", "uint256", "uint256"},

		// values
		[]interface{}{
			sessionInt,
			common.HexToAddress(tx.BobAddr),
			"0x" + receipt.S,
			"0x" + receipt.K,
			receipt.C,
			big.NewInt(tx.UnitPrice * tx.Count),
			big.NewInt(tx.ExpireAt),
		})
	Log.Debugf("generate receipt receiptHash=0x%02x\n", receiptHash)

	receiptHash = solsha3.SoliditySHA3WithPrefix(receiptHash)

	sigPublicKeyECDSA, err := crypto.SigToPub(receiptHash, receiptSign)
	if err != nil {
		Log.Warnf("failed to generate pub key from signature. err=%v", err)
		return errors.New("verify signature error")
	}

	if crypto.PubkeyToAddress(*sigPublicKeyECDSA) != common.HexToAddress(tx.BobAddr) {
		Log.Warnf("failed to verify signature.")
		return errors.New("failed to verify signature")
	}

	return nil
}

func readScrtForComplaint(sessionID string, AliceAddr string, BobAddr string, Log ILogger) (ComplaintSecret, error) {

	var secret ComplaintSecret
	for i := 0; i < 20; i++ {
		Log.Debugf("round %v, read secret from contract. sessionId=%v", i+1, sessionID)
		sessionInt := new(big.Int)
		sessionInt, rs := sessionInt.SetString(sessionID, 10)
		if !rs {
			Log.Warnf("Failed to convert sessionId.")
			return secret, errors.New("Failed to convert sessionId")
		}
		t := time.Now()
		records, err := ZkPoDExchangeClient.ZkPoDExchangeCaller.GetRecordComplaint(&bind.CallOpts{}, common.HexToAddress(AliceAddr), common.HexToAddress(BobAddr), sessionInt)
		if err != nil {
			Log.Warnf("Failed to read session record: %v", err)
			return secret, errors.New("Failed to read session record")
		}
		Log.Debugf("POD reads secret from contract...time cost=%v\n",
			time.Since(t))
		if records.SubmitAt.Int64() == 0 {
			time.Sleep(30 * time.Second)
			continue
		}
		secret.S = hex.EncodeToString(records.Seed0[:])
		return secret, nil
	}

	return secret, errors.New("No secret to be read")
}

func claimToContractForComplaint(sessionID string, bulletin Bulletin, claimFile string, AliceAddr string, Log ILogger) (string, error) {

	sessionInt := new(big.Int)
	sessionInt, rs := sessionInt.SetString(sessionID, 10)
	if !rs {
		Log.Warnf("Failed to convert sessionId.")
		return "", errors.New("Failed to convert sessionId")
	}

	claim, err := readClaim(claimFile, Log)
	if err != nil {
		Log.Warnf("Failed to read claim. err=%v", err)
		return "", errors.New("Failed to read claim")
	}
	var mkls = make([][32]byte, len(claim.M))
	for i, v := range claim.M {
		mklbyte, err := hex.DecodeString(v)
		if err != nil {
			Log.Warnf("failed to decode merkle root. mkl=%v, err=%v", v, err)
			return "", errors.New("failed to decode merkle root")
		}
		mkls[i] = *byte32(mklbyte)
	}
	ks := strings.Split(claim.K, " ")
	if len(ks) < 3 {
		Log.Warnf("Invalid k in claim. err=%v", err)
		return "", errors.New("Invalid k in claim")
	}
	txInt := new(big.Int)
	tx, rs := txInt.SetString(ks[1], 10)
	if !rs {
		Log.Warnf("Failed to convert k[1]. ks[1]=%v", ks[1])
		return "", errors.New("Failed to convert k[1]")
	}
	tyInt := new(big.Int)
	ty, rs := tyInt.SetString(ks[2], 10)
	if !rs {
		Log.Warnf("Failed to convert k[2]. ks[2]=%v", ks[2])
		return "", errors.New("Failed to convert k[2]")
	}
	Log.Debugf("Alice addr=%v, sessionInt=%v, i=%v, j=%v,tx=%v,ty=%v,merkle root length=%v",
		AliceAddr, sessionInt, claim.I, claim.J, tx, ty, len(mkls))
	_sCnt, err := strconv.ParseUint(bulletin.S, 10, 64)
	if err != nil {
		Log.Warnf("failed to convert sCnt. sCnt=%v, err=%v", _sCnt, err)
		return "", errors.New("failed to convert sCnt")
	}
	t := time.Now()
	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.ClaimComplaint(GAdminAuth, common.HexToAddress(AliceAddr), sessionInt,
		claim.I, claim.J, tx, ty, mkls, _sCnt)
	if err != nil {
		Log.Warnf("Failed to submit claim. err=%v", err)
		return "", errors.New("Failed to submit claim")
	}
	Log.Debugf("POD submit claim successfully,  time cost=%v\n",
		time.Since(t))
	return ctx.Hash().Hex(), nil
}

func settleDealForComplaint(sessionID string, AliceAddr string, BobAddr string) (string, error) {
	sessionInt := new(big.Int)
	sessionInt, rs := sessionInt.SetString(sessionID, 10)
	if !rs {
		return "", errors.New("failed to convert sessionId")
	}
	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.SettleComplaintDeal(GAdminAuth, common.HexToAddress(AliceAddr), common.HexToAddress(BobAddr), sessionInt)
	if err != nil {
		return "", fmt.Errorf("failed to settle deal. sessionID=%v, err=%v", sessionID, err)
	}
	return ctx.Hash().Hex(), nil
}

func submitScrtForAtomicSwap(tx Transaction, receiptSign []byte, Log ILogger) (string, error) {
	seed0Path := BConf.AliceDir + "/transaction/" + tx.SessionID + "/secret"
	receiptPath := BConf.AliceDir + "/transaction/" + tx.SessionID + "/receipt"

	sessionInt := new(big.Int)
	sessionInt, rs := sessionInt.SetString(tx.SessionID, 10)
	if !rs {
		Log.Warnf("failed to convert sessionId.")
		return "", errors.New("failed to convert sessionId")
	}

	_, receipt, err := readReceiptForAtomicSwap(receiptPath, Log)
	if err != nil {
		Log.Warnf("failed to read receipt. err=%v", err)
		return "", errors.New("failed to read receipt")
	}
	seed2Byte, err := hex.DecodeString(receipt.S)
	if err != nil {
		Log.Warnf("failed to decode seed2. seed2string=%v, err=%v", receipt.S, err)
		return "", errors.New("failed to decode seed2")
	}

	vwInt := new(big.Int)
	vwInt, rs = vwInt.SetString(receipt.VW, 10)
	if !rs {
		Log.Warnf("failed to convert vw. vw=%v", receipt.VW)
		return "", errors.New("failed to convert vw")
	}
	Log.Debugf(" _sessionId=%v, Bob address=%v, seed2=%v, vwInt=%v, _count=%v,",
		sessionInt, tx.BobAddr, seed2Byte, vwInt, receipt.C)

	err = verifySigntureForAtomicSwap(tx, sessionInt, vwInt, receiptSign, receipt, Log)
	if err != nil {
		Log.Warnf("verify signature error! err=%v", err)
		return "", errors.New("verify signature error")
	}

	secret, err := readSeed0ForAtomicSwap(seed0Path, Log)
	if err != nil {
		Log.Warnf("failed to read seed0. err=%v", err)
		return "", errors.New("failed to read seed0")
	}
	seedByte, err := hex.DecodeString(secret.S)
	if err != nil {
		Log.Warnf("failed to decode seed s. seedstring=%v, err=%v", secret.S, err)
		return "", errors.New("failed to parse seed2 file")
	}

	_sCnt, err := strconv.ParseUint(tx.Bulletin.S, 10, 64)
	if err != nil {
		Log.Warnf("failed to convert sCnt. sCnt=%v, err=%v", _sCnt, err)
		return "", errors.New("failed to convert sCnt")
	}
	Log.Debugf("seedByte=%v, _sCnt=%v, sessionInt=%v, BobAddr=%v,seed2Byte=%v, vwInt=%v, receipt.C=%v, tx.Price=%v, tx.ExpireAt=%v, receiptSign=%v",
		*byte32(seedByte), _sCnt, sessionInt, tx.BobAddr, *byte32(seed2Byte), vwInt, receipt.C, big.NewInt(tx.Price), big.NewInt(tx.ExpireAt), receiptSign)

	t := time.Now()
	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.SubmitProofAtomicSwap(GAdminAuth, *byte32(seedByte), _sCnt, sessionInt,
		common.HexToAddress(tx.BobAddr), *byte32(seed2Byte), vwInt, receipt.C,
		big.NewInt(tx.Price), big.NewInt(tx.ExpireAt), receiptSign)
	if err != nil {
		Log.Warnf("failed to submit proof. err=%v", err)
		return "", errors.New("failed to submit proof")
	}
	Log.Debugf("POD submit proof successfully,  time cost=%v\n",
		time.Since(t))
	return ctx.Hash().Hex(), nil
}

func verifySigntureForAtomicSwap(tx Transaction, sessionInt *big.Int, vwInt *big.Int, receiptSign []byte, receipt AtomicSwapReceipt, Log ILogger) error {

	if len(receiptSign) != 65 {
		Log.Warnf("invalid signature. sig=%v, len(sig)=%v", hexutil.Encode(receiptSign), len(receiptSign))
		return errors.New("invalid signature")
	}

	receiptHash := solsha3.SoliditySHA3( // types
		[]string{"uint256", "address", "bytes32", "uint256", "uint64", "uint256", "uint256"},

		// values
		[]interface{}{
			sessionInt,
			common.HexToAddress(tx.BobAddr),
			"0x" + receipt.S,
			vwInt,
			receipt.C,
			big.NewInt(tx.UnitPrice * tx.Count),
			big.NewInt(tx.ExpireAt),
		})
	Log.Debugf("generate receipt receiptHash=0x%02x\n", receiptHash)

	receiptHash = solsha3.SoliditySHA3WithPrefix(receiptHash)

	sigPublicKeyECDSA, err := crypto.SigToPub(receiptHash, receiptSign)
	if err != nil {
		Log.Warnf("failed to generate pub key from signature. err=%v", err)
		return errors.New("verify signature error")
	}

	if crypto.PubkeyToAddress(*sigPublicKeyECDSA) != common.HexToAddress(tx.BobAddr) {
		Log.Warnf("failed to verify signature.")
		return errors.New("failed to verify signature")
	}

	return nil
}

func readScrtForAtomicSwap(sessionID string, AliceAddr string, BobAddr string, Log ILogger) (AtomicSwapSecret, error) {

	var secret AtomicSwapSecret
	for i := 0; i < 20; i++ {
		Log.Debugf("round %v, read secret from contract. sessionId=%v", i+1, sessionID)
		sessionInt := new(big.Int)
		sessionInt, rs := sessionInt.SetString(sessionID, 10)
		if !rs {
			Log.Warnf("Failed to convert sessionId.")
			return secret, errors.New("Failed to convert sessionId")
		}
		t := time.Now()
		records, err := ZkPoDExchangeClient.ZkPoDExchangeCaller.GetRecordAtomicSwap(&bind.CallOpts{}, common.HexToAddress(AliceAddr), common.HexToAddress(BobAddr), sessionInt)
		if err != nil {
			Log.Warnf("Failed to read session record: %v", err)
			return secret, errors.New("Failed to read session record")
		}
		Log.Debugf("POD reads secret from contract...time cost=%v\n",
			time.Since(t))
		if records.SubmitAt.Int64() == 0 {
			time.Sleep(30 * time.Second)
			continue
		}

		secret.S = hex.EncodeToString(records.Seed0[:])
		return secret, nil
	}

	return secret, errors.New("No secret to be read")
}

func submitScrtForAtomicSwapVc(tx Transaction, receiptSign []byte, Log ILogger) (string, error) {
	seed0Path := BConf.AliceDir + "/transaction/" + tx.SessionID + "/secret"
	receiptPath := BConf.AliceDir + "/transaction/" + tx.SessionID + "/receipt"

	sessionInt := new(big.Int)
	sessionInt, rs := sessionInt.SetString(tx.SessionID, 10)
	if !rs {
		Log.Warnf("failed to convert sessionId.")
		return "", errors.New("failed to convert sessionId")
	}

	_, receipt, err := readReceiptForAtomicSwapVc(receiptPath, Log)
	if err != nil {
		Log.Warnf("failed to read receipt. err=%v", err)
		return "", errors.New("failed to read receipt")
	}

	digestInt := new(big.Int)
	digestInt, rs = digestInt.SetString(receipt.D, 10)
	if !rs {
		Log.Warnf("failed to convert digestInt. receipt.D=%v", receipt.D)
		return "", errors.New("failed to convert digestInt")
	}
	Log.Debugf(" _sessionId=%v, Bob address=%v, digestInt=%v", sessionInt, tx.BobAddr, digestInt)

	err = verifySigntureForAtomicSwapVc(tx, sessionInt, receiptSign, digestInt, Log)
	if err != nil {
		Log.Warnf("verify signature error! err=%v", err)
		return "", errors.New("verify signature error")
	}

	secret, err := readSeed0ForAtomicSwapVc(seed0Path, Log)
	if err != nil {
		Log.Warnf("failed to read seed0. err=%v", err)
		return "", errors.New("failed to read seed0")
	}

	seed0Int := new(big.Int)
	seed0Int, rs = seed0Int.SetString(secret.S, 10)
	if !rs {
		Log.Warnf("failed to convert seed0Int. secret.S=%v", secret.S)
		return "", errors.New("failed to convert seed0Int")
	}

	randInt := new(big.Int)
	randInt, rs = randInt.SetString(secret.R, 10)
	if !rs {
		Log.Warnf("failed to convert digestInt. secret.R=%v", secret.R)
		return "", errors.New("failed to convert digestInt")
	}
	Log.Debugf("seedByte=%v, randByte=%v, sessionInt=%v, BobAddr=%v, receipt.D=%v, tx.Price=%v, tx.ExpireAt=%v, receiptSign=%v",
		seed0Int, randInt, sessionInt, tx.BobAddr, receipt.D, big.NewInt(tx.Price), big.NewInt(tx.ExpireAt), receiptSign)

	t := time.Now()
	// SubmitProofAtomicSwapVC func(opts *bind.TransactOpts, _seed0 *big.Int, _seed0_rand *big.Int,
	// 	_sid *big.Int, _b common.Address, _seed0_mimc3_digest *big.Int,
	// 	_price *big.Int, _expireAt *big.Int, _sig []byte)
	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.SubmitProofAtomicSwapVC(GAdminAuth, seed0Int, randInt,
		sessionInt, common.HexToAddress(tx.BobAddr), digestInt,
		big.NewInt(tx.Price), big.NewInt(tx.ExpireAt), receiptSign)
	if err != nil {
		Log.Warnf("failed to submit proof. err=%v", err)
		return "", errors.New("failed to submit proof")
	}
	Log.Debugf("POD submit proof successfully,  time cost=%v\n",
		time.Since(t))
	return ctx.Hash().Hex(), nil
}

func verifySigntureForAtomicSwapVc(tx Transaction, sessionInt *big.Int, receiptSign []byte, digestInt *big.Int, Log ILogger) error {
	if len(receiptSign) != 65 {
		Log.Warnf("invalid signature. sig=%v, len(sig)=%v", hexutil.Encode(receiptSign), len(receiptSign))
		return errors.New("invalid signature")
	}

	receiptHash := solsha3.SoliditySHA3( // types
		[]string{"uint256", "address", "uint256", "uint256", "uint256"},

		// values
		[]interface{}{
			sessionInt,
			common.HexToAddress(tx.BobAddr),
			digestInt,
			big.NewInt(tx.UnitPrice * tx.Count),
			big.NewInt(tx.ExpireAt),
		})
	Log.Debugf("generate receipt receiptHash=0x%02x\n", receiptHash)

	receiptHash = solsha3.SoliditySHA3WithPrefix(receiptHash)

	sigPublicKeyECDSA, err := crypto.SigToPub(receiptHash, receiptSign)
	if err != nil {
		Log.Warnf("failed to generate pub key from signature. err=%v", err)
		return errors.New("verify signature error")
	}

	if crypto.PubkeyToAddress(*sigPublicKeyECDSA) != common.HexToAddress(tx.BobAddr) {
		Log.Warnf("failed to verify signature.")
		return errors.New("failed to verify signature")
	}

	return nil
}

func readScrtForAtomicSwapVc(sessionID string, AliceAddr string, BobAddr string, Log ILogger) (AtomicSwapVcSecret, error) {
	var secret AtomicSwapVcSecret
	for i := 0; i < 20; i++ {
		Log.Debugf("round %v, read secret from contract. sessionId=%v", i+1, sessionID)
		sessionInt := new(big.Int)
		sessionInt, rs := sessionInt.SetString(sessionID, 10)
		if !rs {
			Log.Warnf("Failed to convert sessionId.")
			return secret, errors.New("Failed to convert sessionId")
		}
		t := time.Now()
		records, err := ZkPoDExchangeClient.ZkPoDExchangeCaller.GetRecordAtomicSwapVC(&bind.CallOpts{}, common.HexToAddress(AliceAddr), common.HexToAddress(BobAddr), sessionInt)
		if err != nil {
			Log.Warnf("Failed to read session record: %v", err)
			return secret, errors.New("Failed to read session record")
		}
		Log.Debugf("POD reads secret from contract...time cost=%v\n",
			time.Since(t))
		if records.SubmitAt.Int64() == 0 {
			time.Sleep(30 * time.Second)
			continue
		}
		secret.S = records.Seed0.String()
		secret.R = records.Seed0Rand.String()
		return secret, nil
	}

	return secret, errors.New("No secret to be read")
}

func submitScrtForVRFQ(tx Transaction, receiptSign []byte, Log ILogger) (string, error) {
	seed0Path := BConf.AliceDir + "/transaction/" + tx.SessionID + "/secret"
	receiptPath := BConf.AliceDir + "/transaction/" + tx.SessionID + "/receipt"

	var rs bool
	sessionInt := new(big.Int)
	sessionInt, rs = sessionInt.SetString(tx.SessionID, 10)
	if !rs {
		Log.Warnf("failed to convert sessionId.")
		return "", errors.New("failed to convert sessionId")
	}

	_, receipt, err := readVRFReceipt(receiptPath, Log)
	if err != nil {
		Log.Warnf("failed to read receipt. err=%v", err)
		return "", errors.New("failed to read receipt")
	}
	gs := strings.Split(receipt.G, " ")
	if len(gs) != 3 {
		Log.Warnf("invalid g. g=%v", receipt.G)
		return "", errors.New("invalid g. ")
	}
	var _g_exp_r [2]*big.Int
	_g_exp_r[0] = new(big.Int)
	_g_exp_r[0], rs = _g_exp_r[0].SetString(gs[1], 10)
	if !rs {
		Log.Warnf("failed to convert g[1]. g[1] = %v", gs[1])
		return "", errors.New("failed to convert g[1]")
	}
	_g_exp_r[1] = new(big.Int)
	_g_exp_r[1], rs = _g_exp_r[1].SetString(gs[2], 10)
	if !rs {
		Log.Warnf("failed to convert g[2]. g[2] = %v", gs[2])
		return "", errors.New("failed to convert g[2]")
	}

	err = verifySigntureForVRFQ(tx, sessionInt, _g_exp_r, receiptSign, receipt, Log)
	if err != nil {
		Log.Warnf("verify signature error! err=%v", err)
		return "", errors.New("verify signature error")
	}

	secret, err := readSeed0ForVRFQ(seed0Path, Log)
	if err != nil {
		Log.Warnf("Failed to read seed0. err=%v", err)
		return "", errors.New("Failed to read seed0")
	}
	srInt := new(big.Int)
	srInt, rs = srInt.SetString(secret.R, 10)
	if !rs {
		Log.Warnf("Failed to convert s_r. s_r=%v", secret.R)
		return "", errors.New("Failed to convert s_r")
	}

	t := time.Now()
	ctx, err := ZkPoDExchangeClient.ZkPoDExchangeTransactor.SubmitProofVRF(GAdminAuth, srInt, sessionInt,
		common.HexToAddress(tx.BobAddr), _g_exp_r, big.NewInt(tx.Price), big.NewInt(tx.ExpireAt), receiptSign)
	if err != nil {
		Log.Warnf("Failed to submit proof. err=%v", err)
		return "", errors.New("Failed to submit proof")
	}
	Log.Debugf("POD submit proof successfully, time cost=%v\n",
		time.Since(t))
	return ctx.Hash().Hex(), nil
}

func verifySigntureForVRFQ(tx Transaction, sessionInt *big.Int, _g_exp_r [2]*big.Int, receiptSign []byte, receipt VRFReceipt, Log ILogger) error {

	if len(receiptSign) != 65 {
		Log.Warnf("invalid signature. sig=%v, len(sig)=%v", hexutil.Encode(receiptSign), len(receiptSign))
		return errors.New("invalid signature")
	}

	receiptHash := solsha3.SoliditySHA3( // types
		[]string{"uint256", "address", "uint256", "uint256", "uint256", "uint256"},

		// values
		[]interface{}{
			sessionInt,
			common.HexToAddress(tx.BobAddr),
			_g_exp_r[0],
			_g_exp_r[1],
			big.NewInt(tx.UnitPrice * tx.Count),
			big.NewInt(tx.ExpireAt),
		})
	Log.Debugf("generate receipt receiptHash=0x%02x\n", receiptHash)

	receiptHash = solsha3.SoliditySHA3WithPrefix(receiptHash)

	sigPublicKeyECDSA, err := crypto.SigToPub(receiptHash, receiptSign)
	if err != nil {
		Log.Warnf("failed to generate pub key from signature. err=%v", err)
		return errors.New("verify signature error")
	}

	if crypto.PubkeyToAddress(*sigPublicKeyECDSA) != common.HexToAddress(tx.BobAddr) {
		Log.Warnf("failed to verify signature.")
		return errors.New("failed to verify signature")
	}

	return nil
}

func readScrtForVRFQ(sessionID string, AliceAddr string, BobAddr string, Log ILogger) (VRFSecret, error) {

	var secret VRFSecret
	for i := 0; i < 20; i++ {
		Log.Debugf("round %v, read secret from contract. sessionId=%v", i+1, sessionID)
		sessionInt := new(big.Int)
		sessionInt, rs := sessionInt.SetString(sessionID, 10)
		if !rs {
			Log.Warnf("Failed to convert sessionId.")
			return secret, errors.New("Failed to convert sessionId")
		}
		t := time.Now()
		records, err := ZkPoDExchangeClient.ZkPoDExchangeCaller.GetRecordVRF(&bind.CallOpts{}, common.HexToAddress(AliceAddr), common.HexToAddress(BobAddr), sessionInt)
		if err != nil {
			Log.Warnf("Failed to read session record: %v", err)
			return secret, errors.New("Failed to read session record")
		}
		Log.Debugf("POD reads secret from contract...time cost=%v\n",
			time.Since(t))
		if records.SubmitAt.Int64() == 0 {
			time.Sleep(30 * time.Second)
			continue
		}
		secret.R = records.R.String()
		return secret, nil
	}

	return secret, errors.New("No secret to be read")
}

func calcuDeposit(AliceAddr string, BobAddr string, value int64) (bool, error) {
	dpst, err := ZkPoDExchangeClient.ZkPoDExchangeCaller.BobDeposits(&bind.CallOpts{}, common.HexToAddress(BobAddr), common.HexToAddress(AliceAddr))
	if err != nil {
		return false, fmt.Errorf("failed to read deposit. err=%v", err)
	}
	if value+DepositLockMap[AliceAddr+BobAddr] > dpst.Value.Int64() {
		return false, nil
	}
	if dpst.Stat == 1 {
		return false, nil
	}
	if dpst.Stat == 2 && time.Now().Unix()+600 > dpst.UnDepositAt.Int64()+8*60*60 {
		return false, nil
	}
	DepositLockMap[AliceAddr+BobAddr] += value
	return true, nil
}

func checkDeposit(AliceAddr string, BobAddr string, value int64) (bool, error) {
	dpst, err := ZkPoDExchangeClient.ZkPoDExchangeCaller.BobDeposits(&bind.CallOpts{}, common.HexToAddress(BobAddr), common.HexToAddress(AliceAddr))
	if err != nil {
		return false, fmt.Errorf("failed to read deposit. err=%v", err)
	}
	if dpst.Value.Int64() < value {
		return false, nil
	}
	if dpst.UnDepositAt.Int64() != 0 && dpst.UnDepositAt.Int64()+28800 < time.Now().Unix()+600 && dpst.Stat == 1 {
		return false, nil
	}
	return true, nil
}

func calcuBltKey(b Bulletin) ([32]byte, error) {
	size, _ := strconv.ParseUint(b.Size, 10, 64)
	s, _ := strconv.ParseUint(b.S, 10, 64)
	n, _ := strconv.ParseUint(b.N, 10, 64)
	mklrootInt := new(big.Int)
	mklroot, rs := mklrootInt.SetString(fmt.Sprintf("%x", b.SigmaMKLRoot), 10)
	if !rs {
		return *byte32(nil), fmt.Errorf("failed to convert sessionId. b.SigmaMKLRoot=%v", b.SigmaMKLRoot)
	}

	bltByte := solsha3.SoliditySHA3(
		// types
		[]string{"uint64", "uint64", "uint64", "uint256"},

		// values
		[]interface{}{
			size,
			s,
			n,
			mklroot,
		},
	)
	return *byte32(bltByte), nil
}

func byte32(s []byte) (a *[32]byte) {
	if len(a) <= len(s) {
		a = (*[len(a)]byte)(unsafe.Pointer(&s[0]))
	}
	return a
}
