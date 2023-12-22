package task

import (
	"context"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	btcmempool "github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas_committer/config"
	"github.com/mapprotocol/atlas_committer/utils"
	"github.com/mapprotocol/atlas_committer/utils/mempool"
	"math/big"
	"time"
)

func getCheckPointFromCfg() (*utils.CheckPoint, error) {
	return nil, nil
}

func makeTx(content []byte, feerate int64, sender, receiverAddress btcutil.Address,
	outList []*PrevOutPoint, senderPriv *btcec.PrivateKey, btcApiClient *mempool.MempoolClient) (*wire.MsgTx, error) {

	commitTx := wire.NewMsgTx(wire.TxVersion)
	totalSenderAmount := btcutil.Amount(0)
	TxPrevOutputFetcher := txscript.NewMultiPrevOutFetcher(nil)

	for _, out := range outList {
		txOut, err := getTxOutByOutPoint(out.Outpoint, btcApiClient)
		if err != nil {
			return nil, err
		}
		TxPrevOutputFetcher.AddPrevOut(*out.Outpoint, txOut)
		in := wire.NewTxIn(out.Outpoint, nil, nil)
		in.Sequence = defaultSequenceNum
		commitTx.AddTxIn(in)

		totalSenderAmount += btcutil.Amount(out.Value)
	}

	// make OP_RETURN output
	opReturnScript, err := txscript.NullDataScript(content)
	//opReturnScript, err := txscript.NewScriptBuilder().
	//	AddOp(txscript.OP_RETURN).
	//	AddData(content).
	//	Script()
	if err != nil {
		return nil, err
	}

	commitTx.AddTxOut(&wire.TxOut{
		PkScript: opReturnScript,
		Value:    0,
	})
	changePkScript, err := txscript.PayToAddrScript(sender)
	if err != nil {
		return nil, err
	}
	// make the change
	commitTx.AddTxOut(wire.NewTxOut(0, changePkScript))
	txsize := btcmempool.GetTxVirtualSize(btcutil.NewTx(commitTx))
	fee := btcutil.Amount(txsize) * btcutil.Amount(feerate)
	changeAmount := totalSenderAmount - fee

	if changeAmount > 0 {
		commitTx.TxOut[len(commitTx.TxOut)-1].Value = int64(changeAmount)
	} else {
		return nil, errors.New("not enough fees")
	}
	// make the signature
	witnessList := make([]wire.TxWitness, len(commitTx.TxIn))
	for i := range commitTx.TxIn {
		txOut := TxPrevOutputFetcher.FetchPrevOutput(commitTx.TxIn[i].PreviousOutPoint)
		witness, err := txscript.TaprootWitnessSignature(commitTx, txscript.NewTxSigHashes(commitTx, TxPrevOutputFetcher),
			i, txOut.Value, txOut.PkScript, txscript.SigHashDefault, senderPriv)
		if err != nil {
			return nil, err
		}
		witnessList[i] = witness
	}
	for i := range witnessList {
		commitTx.TxIn[i].Witness = witnessList[i]
	}
	return commitTx, nil
}

func blockNumber(url string) (uint64, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return 0, err
	}
	return client.BlockNumber(context.Background())
}
func getMetadataFromAtlas(url string) ([]*utils.CheckPoint, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	bHeight, err := client.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}
	log.Info("getMetadataFromAtlas", "current", CurrentCommitHeight, "atlas height", bHeight)
	datas := []*utils.CheckPoint{}
	c := CurrentCommitHeight + commitLength
	for ; c < bHeight; c += commitLength {
		h, err := client.HeaderByNumber(context.Background(), big.NewInt(int64(c)))
		if err != nil {
			return nil, err
		}
		datas = append(datas, &utils.CheckPoint{
			Root:   h.Hash().String(),
			Height: big.NewInt(int64(c)),
		})
	}
	CurrentCommitHeight = c
	return datas, nil
}
func getMetadataByHeight(height uint64, url string) (*utils.CheckPoint, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	bHeight, err := client.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}
	if bHeight < height {
		log.Info("get metadata failed,cause the wrong height", "height", height, "current height", bHeight)
		return nil, errors.New("invalid height")
	}
	h, err := client.HeaderByNumber(context.Background(), big.NewInt(int64(height)))
	if err != nil {
		return nil, err
	}

	return &utils.CheckPoint{
		Root:   h.Hash().String(),
		Height: big.NewInt(int64(height)),
	}, nil
}

func timingCheck(txhash string, client *mempool.MempoolClient) bool {

	for {
		exist, err := checkTxOnChain(txhash, client)
		if exist {
			return exist
		} else {
			log.Error("checkTxOnChain error", "error", err)
		}
		log.Info("checkTxOnChain......")
		time.Sleep(5 * time.Minute)
	}
}

func HandleCommitTxProc(height, feerate uint64, atlasURL string, sender btcutil.Address,
	priv *btcec.PrivateKey, network *chaincfg.Params) error {
	// get the matedata
	datas, err := getMetadataFromAtlas(atlasURL)
	if err != nil {
		return err
	}
	if len(datas) == 0 {
		return nil
	}

	client := mempool.NewClient(network)

	for _, d := range datas {
		content, err := ToBytes(d)
		if err != nil {
			log.Error("make checkpoint tx failed", "err", err, "checkpoint", d)
			return err
		}

		OutPointList, err := gatherUtxo(client, sender)
		if err != nil {
			log.Error("fetch utxo failed ", "sender", sender.String(), "error", err)
			return err
		}

		commitTx, err := makeTx(content, int64(feerate), sender, sender, OutPointList, priv, client)
		if err != nil {
			log.Error("make checkpoint tx failed ", "error", err)
			return err
		}

		txHash, err := client.BroadcastTx(commitTx)
		if err != nil {
			log.Error("Broadcast checkpoint failed", "error", err, "txhash", txHash.String())
			return err
		}

		log.Info("Broadcast checkpoint success", "checkpoint", d)

		if len(commitTx.TxOut) > 0 {
			vout := len(commitTx.TxOut) - 1
			setPrevOutPoint(&wire.OutPoint{
				Hash:  *txHash,
				Index: uint32(vout),
			}, commitTx.TxOut[vout].Value)
		}
		// check the tx was on chain
		exist := timingCheck(txHash.String(), client)

		log.Info("commit checkpoint on the chain", "txhash", txHash.String(), "exist", exist)
	}
	return nil
}
func HandleCommitCheckPointTx(height, feerate uint64, atlasURL string, sender btcutil.Address,
	priv *btcec.PrivateKey, network *chaincfg.Params) (*utils.CheckPoint, error) {
	ck, err := getMetadataByHeight(height, atlasURL)
	if err != nil {
		return nil, err
	}

	content, err := ToBytes(ck)
	if err != nil {
		log.Error("encode the checkpoint failed", "err", err, "metadata", ck)
		return nil, err
	}

	client := mempool.NewClient(network)
	OutPointList, err := gatherUtxo(client, sender)
	if err != nil {
		log.Error("fetch utxo failed ", "sender", sender.String(), "error", err)
		return nil, err
	}

	commitTx, err := makeTx(content, int64(feerate), sender, sender, OutPointList, priv, client)
	if err != nil {
		log.Error("make checkpoint tx failed ", "error", err)
		return nil, err
	}

	txHash, err := client.BroadcastTx(commitTx)
	if err != nil {
		log.Error("Broadcast checkpoint failed", "error", err, "txhash", txHash.String())
		return nil, err
	}

	log.Info("Broadcast checkpoint success", "checkpoint", ck)

	if len(commitTx.TxOut) > 0 {
		vout := len(commitTx.TxOut) - 1
		setPrevOutPoint(&wire.OutPoint{
			Hash:  *txHash,
			Index: uint32(vout),
		}, commitTx.TxOut[vout].Value)
	}
	// check the tx was on chain
	exist := timingCheck(txHash.String(), client)
	log.Info("commit checkpoint on the chain", "txhash", txHash.String(), "exist", exist)

	return ck, nil
}
func HandleGetFeeRate(network *chaincfg.Params) {
	timeloop := 5 * time.Minute
	log.Info("first get the fee rate from mempool")
	for {
		client := mempool.NewClient(network)
		fees, err := client.RecommendedFees()
		if err != nil {
			log.Error("get recommended fees from mempool failed", "error", err)
			time.Sleep(timeloop)
			continue
		}
		GlobalFeeRate = fees.FastestFee
		time.Sleep(timeloop)
		continue
	}
}

func Run() {
	netParams := &chaincfg.MainNetParams
	if config.CfgParams.TestNet {
		netParams = &chaincfg.TestNet3Params
	}

	looptimeout := 5 * time.Minute
	go HandleGetFeeRate(netParams)

	privateKeyBytes, err := hex.DecodeString(config.CfgParams.Sender)
	if err != nil {
		log.Error("invalid sender private key")
		panic(err)
	}
	senderKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

	sender, err := makeTpAddress(senderKey, netParams)
	if err != nil {
		log.Error("invalid sender address")
		panic(err)
	}

	log.Info("fetch the latest checkpoint...")
	// get the checkpoint on the chain and update it
	checkpoint, err := fetchLatestCheckPoint(sender, netParams)
	if config.CfgParams.LatestCheckPoint.Height.Uint64() < checkpoint.Height.Uint64() {
		config.CfgParams.LatestCheckPoint = checkpoint
	}

	for {
		currentBlockNumber, err := blockNumber(config.CfgParams.AtlasURL)
		if err != nil {
			log.Error("get atlas block number failed,wait for again...", "error", err)
			time.Sleep(looptimeout)
			continue
		}
		if config.CfgParams.LatestCheckPoint.Height.Uint64()+commitLength < currentBlockNumber {
			feerate := uint64(GlobalFeeRate)
			cur := config.CfgParams.LatestCheckPoint.Height.Uint64() + commitLength
			ck, err := HandleCommitCheckPointTx(cur, feerate, config.CfgParams.AtlasURL, sender, senderKey, netParams)
			if err != nil {
				log.Error("commit checkpoint failed,wait for again...", "height", cur, "error", err)
				time.Sleep(looptimeout)
				continue
			}
			config.CfgParams.LatestCheckPoint = ck
		}
		time.Sleep(looptimeout)
	}
}
