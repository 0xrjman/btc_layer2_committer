package task

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/btc_layer2_committer/utils"
	"github.com/mapprotocol/btc_layer2_committer/utils/mempool"
	"strings"
)

const (
	defaultSequenceNum    = wire.MaxTxInSequenceNum - 10
	defaultCommitOutValue = int64(610)
	commitLength          = uint64(200000)
)

var (
	GlobalFeeRate                          = int64(100)
	PrevAdminOutPoint        *PrevOutPoint = nil
	MinPreAdminOutPointValue               = int64(20000)
	CurrentCommitHeight                    = uint64(0)
)

type PrevOutPoint struct {
	Outpoint *wire.OutPoint
	Value    int64
}

func ToBytes(m *utils.CheckPoint) ([]byte, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, m.Height)
	h, err := hexutil.Decode(m.Root)
	if err != nil {
		return nil, err
	}
	if len(h) != 32 {
		panic("invalid root length in to bytes")
	}
	return append(b, h...), nil
}
func FromBytes(data []byte) (*utils.CheckPoint, error) {
	if len(data) != 40 {
		return nil, errors.New("invalid data length in from bytes")
	}
	v := &utils.CheckPoint{}
	v.Height = binary.LittleEndian.Uint64(data[0:8])
	v.Root = hexutil.Encode(data[8:])
	return v, nil
}

func setPrevOutPoint(outpoint *wire.OutPoint, val int64) {
	tmp := &PrevOutPoint{
		Outpoint: outpoint,
		Value:    val,
	}
	PrevAdminOutPoint = tmp
}
func gatherUtxo(client *mempool.MempoolClient, sender btcutil.Address) ([]*PrevOutPoint, error) {
	outPointList := make([]*PrevOutPoint, 0)

	if PrevAdminOutPoint != nil {
		if PrevAdminOutPoint.Value > MinPreAdminOutPointValue {
			outPointList = append(outPointList, PrevAdminOutPoint)
			return outPointList, nil
		}
	}

	unspentList, err := client.ListUnspent(sender)
	if err != nil {
		return nil, err
	}

	if len(unspentList) == 0 {
		err = fmt.Errorf("no utxo for %s", sender)
		return nil, err
	}

	for i := range unspentList {
		if unspentList[i].Output.Value < 5000 {
			continue
		}
		outPointList = append(outPointList, &PrevOutPoint{
			Outpoint: unspentList[i].Outpoint,
			Value:    unspentList[i].Output.Value,
		})
	}
	return outPointList, nil
}
func getTxOutByOutPoint(outPoint *wire.OutPoint, btcClient *mempool.MempoolClient) (*wire.TxOut, error) {
	tx, err := btcClient.GetRawTransaction(&outPoint.Hash)
	if err != nil {
		return nil, err
	}
	if int(outPoint.Index) >= len(tx.TxOut) {
		return nil, errors.New("err out point")
	}
	return tx.TxOut[outPoint.Index], nil
}
func checkTxOnChain(txHash string, btcClient *mempool.MempoolClient) (bool, error) {
	txId, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		log.Error("checkTxOnChain failed", "decode hash error", err)
		return false, err
	}

	ret, err := btcClient.TransactionStatus(txId)
	if err != nil {
		log.Error("checkTxOnChain failed", "rpc error", err)
		return false, err
	}

	return ret.Confirmed, nil
}

func fetchLatestCheckPoint(sender btcutil.Address, cCheckPoint *utils.CheckPoint, network *chaincfg.Params) (*utils.CheckPoint, error) {
	client := mempool.NewClient(network)

	simTxs, err := client.GetTxsFromAddress(sender)
	if err != nil {
		return nil, err
	}
	// check latest checkpoint match with the config checkpoint
	for i := range simTxs {
		tx := simTxs[len(simTxs)-i-1]
		if sender.String() == tx.Sender && len(tx.OutPuts) == 2 {
			str := tx.OutPuts[0].Scriptpubkey_asm
			script, err := hex.DecodeString(tx.OutPuts[0].Scriptpubkey)
			if err != nil {
				log.Error("decode the Scriptpubkey failed", "err", err, "txid", tx.Txid.String())
				continue
			}
			if !txscript.IsNullData(script) {
				fmt.Println(tx.Txid.String(), "is a op_return tx")
			}
			cc, err := checkPointFromAsm(str)
			if err != nil {
				log.Error("not a OP_RETURN tx", "txhash", tx.Txid)
				continue
			}
			if cCheckPoint != nil {
				if cc.Height < cCheckPoint.Height {
					continue
				}
			}
			cCheckPoint = cc
		} else {
			log.Info("fetch the latest checkpoint, invalid tx", "txid", tx.Txid)
		}
	}
	//tx := wire.MsgTx{}
	//txscript.IsNullData(tx.TxOut[0].PkScript)
	return cCheckPoint, nil
}

func checkPointFromAsm(str string) (*utils.CheckPoint, error) {
	result := strings.Split(str, " ")
	// op_return op_pushbytes data
	if len(result) != 3 {
		return nil, errors.New("invalid script length")
	}
	strScript := result[len(result)-1]
	b0, err := hex.DecodeString(strScript)
	if err != nil {
		return nil, err
	}
	return FromBytes(b0)
}

func makeTpAddress(privKey *btcec.PrivateKey, network *chaincfg.Params) (btcutil.Address, error) {
	tapKey := txscript.ComputeTaprootKeyNoScript(privKey.PubKey())

	address, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(tapKey),
		network,
	)
	if err != nil {
		return nil, err
	}
	return address, nil
}
