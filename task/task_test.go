package task

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/mapprotocol/btc_layer2_committer/utils"
	"strings"
	"testing"
)

func TestGeneratePrivateKey(t *testing.T) {
	//privateKey, err := btcec.NewPrivateKey()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//privateKeyBytes := privateKey.Serialize()
	//privateKeyString := hex.EncodeToString(privateKeyBytes)
	//t.Logf("private key: %s", privateKeyString)
	//
	//privateKeyBytes, err = hex.DecodeString(privateKeyString)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//privateKey, _ = btcec.PrivKeyFromBytes(privateKeyBytes)
	//privateKeyBytes = privateKey.Serialize()
	//privateKeyString = hex.EncodeToString(privateKeyBytes)
	//t.Logf("private key: %s", privateKeyString)
	//
	//sender, err := logic.makeTpAddress(privateKey, &chaincfg.TestNet3Params)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Log("sender: ", sender.String())
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
	checkpoint, err := getMetadataByHeight(ck.Height.Uint64(), url)
	if err != nil {
		return err
	}
	if checkpoint.Equal(ck) {
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
	senderStr := ""
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
}
