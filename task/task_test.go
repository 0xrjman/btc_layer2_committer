package task

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mapprotocol/btc_layer2_committer/utils"
	"github.com/mapprotocol/btc_layer2_committer/utils/mempool"
	"math/big"
	"strings"
	"testing"
	"time"
)

var (
	mapURL = "https://rpc.maplabs.io"
)

// =============================================================================
func getBlockHeader(height uint64, url string) (common.Hash, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return common.Hash{}, err
	}

	h, err := client.HeaderByNumber(context.Background(), big.NewInt(int64(height)))
	if err != nil {
		return common.Hash{}, err
	}
	printHead(h)
	return h.Hash(), nil
}
func printHead(hh *types.Header) {
	fmt.Println("---------------------------------------------------")
	fmt.Println("ParentHash", hh.ParentHash.String())
	fmt.Println("UncleHash", hh.UncleHash.String())
	fmt.Println("Coinbase", hh.Coinbase.String())
	fmt.Println("Root", hh.Root.String())
	fmt.Println("TxHash", hh.TxHash.String())
	fmt.Println("ReceiptHash", hh.ReceiptHash.String())
	fmt.Println("Bloom", hexutil.Encode(hh.Bloom.Bytes()))
	fmt.Println("Difficulty", hh.Difficulty.String())
	fmt.Println("Number", hh.Number.String())
	fmt.Println("GasLimit", hh.GasLimit)
	fmt.Println("GasUsed", hh.GasUsed)
	fmt.Println("Time", hh.Time)
	fmt.Println("Extra", hexutil.Encode(hh.Extra))
	fmt.Println("MixDigest", hh.MixDigest.String())
	fmt.Println("Nonce", hh.Nonce.Uint64())
	fmt.Println("BaseFee", hh.BaseFee.String())
	fmt.Println("WithdrawalsHash", hh.WithdrawalsHash)
	fmt.Println("BlobGasUsed", hh.BlobGasUsed)
	fmt.Println("ExcessBlobGas", hh.ExcessBlobGas)
	fmt.Println("ParentBeaconRoot", hh.ParentBeaconRoot)
	fmt.Println("---------------------------------------------------")
}

// =============================================================================
func TestGeneratePrivateKey(t *testing.T) {
	testnet := true
	netParams := &chaincfg.MainNetParams
	if testnet {
		netParams = &chaincfg.TestNet3Params
	}
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	privateKeyBytes := privateKey.Serialize()
	privateKeyString := hex.EncodeToString(privateKeyBytes)
	t.Logf("private key: %s", privateKeyString)

	privateKeyBytes, err = hex.DecodeString(privateKeyString)
	if err != nil {
		t.Fatal(err)
	}
	privateKey, _ = btcec.PrivKeyFromBytes(privateKeyBytes)
	privateKeyBytes = privateKey.Serialize()
	privateKeyString = hex.EncodeToString(privateKeyBytes)
	t.Logf("private key: %s", privateKeyString)

	sender, err := makeTpAddress(privateKey, netParams)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("sender: ", sender.String())
}

func Test_getSender(t *testing.T) {
	netParams := &chaincfg.TestNet3Params
	str := "552f6f8c8e0493ad9db3bf75ab4940239089f4d0ac14efc7f9ab3a2b49e523c77b7e11d671792599dba305bea5beed4d572f37acb7722af3764a831621dc31a0"
	data, err := hex.DecodeString(str)
	if err != nil {
		fmt.Println(err)
		return
	}
	tt, addrs, c, err := txscript.ExtractPkScriptAddrs(data, netParams)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tt, c)
	for i, addr := range addrs {
		fmt.Println(i, addr.String())
	}
}
func Test_SplitString(t *testing.T) {
	inputString := "OP_PUSHNUM_1 OP_PUSHBYTES_32 db8361a242956c8feefd388d6409c8b6afcf1a5035a6f68d1f2acc0b44e02900"

	result := strings.Split(inputString, " ")

	fmt.Printf("Original String: %s\n", inputString)
	fmt.Printf("Split Result: %v\n", result)
	fmt.Println("last string", result[len(result)-1])
}

func verifyBlockWithCheckPoint(ck *utils.CheckPoint) error {
	url := "https://rpc.maplabs.io"
	ck0, err := getMetadataByHeight(ck.Height, url)
	if err != nil {
		return err
	}
	if ck0.Equal(ck) {
		return nil
	}
	return errors.New("")
}

func Test_CheckPoint(t *testing.T) {
	testnet := true
	netParams := &chaincfg.MainNetParams
	if testnet {
		netParams = &chaincfg.TestNet3Params
	}
	senderStr := "tb1pkepxd60wx4z33qdgz5vad5dvtus6syv3m5m6xc3kthdfar9jmmvq3a8mp7"
	sender, err := btcutil.DecodeAddress(senderStr, netParams)
	if err != nil {
		fmt.Println(err)
		return
	}
	checkpoint, err := fetchLatestCheckPoint(sender, nil, netParams)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = verifyBlockWithCheckPoint(checkpoint)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("finish....")
}

func Test_GetFeeRecommend(t *testing.T) {
	testnet := true
	netParams := &chaincfg.MainNetParams
	if testnet {
		netParams = &chaincfg.TestNet3Params
	}
	client := mempool.NewClient(netParams)
	fees, err := client.RecommendedFees()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fees.EconomyFee, fees.HourFee, fees.HalfHourFee, fees.FastestFee)
}

func Test_CheckPoint2(t *testing.T) {
	//ck := &utils.CheckPoint{
	//	Root:   "0xa92e4bb5581e9f54dc5326bde49480fd0faa2412d84485303c1ce684b395f3ea",
	//	Height: uint64(9400000),
	//}
	ck, err := getMetadataByHeight(uint64(9400000), mapURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	b0, err := ToBytes(ck)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(b0), hexutil.Encode(b0))

	ck1, err := FromBytes(b0)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !ck1.Equal(ck) {
		fmt.Println("invalid check point")
		return
	}
	fmt.Println("finish")
}
func Test_CheckPoint3(t *testing.T) {
	ck, err := getMetadataByHeight(uint64(9400000), mapURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	content, err := ToBytes(ck)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("content", len(content), hexutil.Encode(content))
	opReturnScript, err := txscript.NullDataScript(content)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("opReturnScript", len(opReturnScript), hexutil.Encode(opReturnScript))
	if !txscript.IsNullData(opReturnScript) {
		fmt.Println("is a op_return tx")
	}
	content1 := opReturnScript[2:]
	ck1, err := FromBytes(content1)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !ck1.Equal(ck) {
		fmt.Println("invalid check point")
		return
	}

	fmt.Println("finish")
}

func Test_getBlockInfosFromAtlas(t *testing.T) {
	count, epoch := 2, 50000
	for i := 0; i < count; i++ {
		h := uint64(9000000 + i*epoch)
		h0, err := getBlockHeader(h, mapURL)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("index", h, "root", h0.String())
		time.Sleep(2 * time.Second)
	}
}
func Test_getMatedata(t *testing.T) {
	count := 3
	for i := 0; i < count; i++ {
		ck, err := getMetadataByHeight(uint64(9400000+i), mapURL)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("index", i, "root", ck.Root)
		time.Sleep(5 * time.Second)
	}
}
