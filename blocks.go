package arsyncer

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	"gopkg.in/h2non/gentleman.v2"
	"strings"
)

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
	// get block hash_list from gateway3
	spiltList, err := GetBlockHashListByGateway3(startHeight, endHeight)
	if err != nil {
		log.Error("GetBlockHashListByGateway3(startHeight, endHeight)", "err", err)
		// get block hash_list from trust node
		spiltList, err = GetBlockHashList(arCli, startHeight, endHeight)
		if err != nil {
			log.Error("GetBlockHashList(arCli,startHeight,endHeight)", "err", err)
			// get block hash_list from no-trust node
			spiltList, err = GetBlockHashListFromPeers(arCli, startHeight, endHeight, 5)
		}
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

func GetBlockHashList(arCli *goar.Client, startHeight, endHeight int64) ([]string, error) {
	list, err := arCli.GetBlockHashList()
	if err != nil {
		return nil, err
	}
	curHeight := int64(len(list) - 1)
	if curHeight < endHeight {
		return nil, fmt.Errorf("curHeight must >= endHeight; curHeight:%d,endHeight:%d", curHeight, endHeight)
	}
	// todo bug fix
	spiltList := list[curHeight-endHeight : curHeight-startHeight+1]
	log.Debug("success get block hash_list from arweave gateway", "start", startHeight, "end", endHeight)
	return spiltList, nil
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
		hashList, err := pNode.GetBlockHashList()
		if err != nil {
			log.Error("pNode.GetBlockHashList()", "err", err)
			continue
		}
		curHeight := int64(len(hashList) - 1)
		if curHeight < endHeight {
			continue
		}

		spiltList := hashList[curHeight-endHeight : curHeight-startHeight+1]
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

// TODO Lev suggest Temporary Solutions
func getBlockHashListByHeightRange(startHeight, endHeight int64) ([]string, error) {
	var gateway3Cli = gentleman.New().URL("http://gateway-3.arweave.net:1984")
	gateway3Cli.AddPath(fmt.Sprintf("/hash_list/%d/%d", startHeight, endHeight))

	req := gateway3Cli.Request()
	resp, err := req.Send()
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, errors.New("resp is not ok")
	}

	defer resp.Close()
	list := make([]string, 0, endHeight-startHeight+1)
	if err = resp.JSON(&list); err != nil {
		return nil, err
	}

	return list, nil
}

func GetBlockHashListByGateway3(startHeight, endHeight int64) ([]string, error) {
	if startHeight > endHeight {
		return nil, fmt.Errorf("startHeight:%d must <= endHeight:%d", startHeight, endHeight)
	}
	list, err := getBlockHashListByHeightRange(startHeight, endHeight)
	if err != nil {
		return nil, err
	}
	// verify
	if len(list) != int(endHeight-startHeight+1) {
		return nil, errors.New("get list incorrect")
	}
	log.Debug("success get block hash_list from gateway3.", "start", startHeight, "end", endHeight)
	return list, nil
}
