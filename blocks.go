package arsyncer

import (
	"errors"
	"fmt"
	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
)

type BlockIdxs struct {
	StartHeight  int64
	EndHeight    int64
	IndepHashMap map[string]struct{}
}

func GetBlockIdxs(startHeight int64, arCli *goar.Client) (*BlockIdxs, error) {
	list, err := arCli.GetBlockHashList()
	if err != nil {
		return nil, err
	}
	endHeight := int64(len(list) - 1)
	if startHeight > endHeight {
		return nil, fmt.Errorf("startHeight:%d must less than endHeight:%d", startHeight, endHeight)
	}

	spiltList := list[:endHeight-startHeight+1]

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
