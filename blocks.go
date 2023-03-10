package arsyncer

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	"strings"
)

const BlockMaxCount = 10000

type BlockIdxs struct {
	StartHeight  int64
	EndHeight    int64
	IndepHashMap map[string]struct{}
}

func GetBlockIdxs(startHeight int64, arCli *goar.Client) (*BlockIdxs, error) {
	info, err := arCli.GetInfo()
	if err != nil {
		log.Error("arCli.GetInfo()", "err", err)
		return nil, err
	}
	endHeight := info.Height
	if endHeight-startHeight >= BlockMaxCount {
		endHeight = startHeight + BlockMaxCount - 1
	}
	// get block hash_list from trust node
	spiltList, err := arCli.GetBlockHashList(int(startHeight), int(endHeight))
	if err != nil {
		log.Error("GetBlockHashList(arCli,startHeight,endHeight)", "err", err)
		// get block hash_list from no-trust node
		spiltList, err = GetBlockHashListFromPeers(arCli, startHeight, endHeight, 5)
	}

	if err != nil {
		return nil, err
	}

	hashMap := make(map[string]struct{})
	for _, h := range spiltList {
		hashMap[h] = struct{}{}
	}
	return &BlockIdxs{
		StartHeight:  startHeight,
		EndHeight:    endHeight,
		IndepHashMap: hashMap,
	}, nil
}

func (l *BlockIdxs) existBlock(b types.Block) bool {
	if b.Height < l.StartHeight || b.Height > l.EndHeight {
		log.Error("block height is outside", "height", b.Height, "startHeight", l.StartHeight, "endHeight", l.EndHeight)
		return false
	}
	_, ok := l.IndepHashMap[b.IndepHash]
	return ok
}

func (l *BlockIdxs) VerifyBlock(b types.Block) error {
	/*
		 2.6 is out - https://github.com/ArweaveTeam/arweave/releases/tag/N.2.6.0.
		The fork activates at height 1132210, approximately 2023-03-06 14:00 UTC.
		You will need to make sure you have upgraded your miner before this time to connect to the network.
		You can find more information in the release notes and the updated mining guide.
	*/
	if b.Height >= 1132210 { // not verify 2.6 block
		return nil
	}

	if !l.existBlock(b) {
		log.Warn("block indepHash not exist blockIdxs", "blockHeight", b.Height, "blockIndepHash", b.IndepHash)
		return errors.New("block indepHash not exist blockIdxs")
	}
	indepHash := utils.GenerateIndepHash(b)

	if indepHash != b.IndepHash {
		return fmt.Errorf("generateIndepHash not equal; b.IndepHash: %s", b.IndepHash)
	}
	return nil
}

func GetBlockHashListFromPeers(c *goar.Client, startHeight, endHeight int64, checkNum int, peers ...string) ([]string, error) {
	var err error
	if len(peers) == 0 {
		peers, err = c.GetPeers()
		if err != nil {
			return nil, err
		}
	}

	checkSum := ""
	successCount := 0
	failedCount := 0
	pNode := goar.NewTempConn()
	for _, peer := range peers {
		pNode.SetTempConnUrl("http://" + peer)
		spiltList, err := pNode.GetBlockHashList(int(startHeight), int(endHeight))
		if err != nil {
			log.Error("pNode.GetBlockHashList()", "err", err)
			continue
		}

		sum := strArrCheckSum(spiltList)
		if checkSum == "" {
			checkSum = sum
		}

		if checkSum == sum {
			fmt.Printf("success get block hash_list; peer: %s\n", peer)
			successCount++
		} else {
			fmt.Printf("failed get block hash_list, checksum failed; peer:%s\n", peer)
			failedCount++
		}

		if successCount >= checkNum {
			log.Debug("success get block hash_list from peers", "start", startHeight, "end", endHeight)
			return spiltList, nil
		}
		if failedCount >= checkNum/2 {
			return nil, errors.New("get hash_list from peers failed")
		}
	}
	return nil, errors.New("get hash_list from peers failed")
}

func strArrCheckSum(ss []string) string {
	hash := sha256.Sum256([]byte(strings.Join(ss, "")))
	return string(hash[:])
}
