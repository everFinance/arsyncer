package arsyncer

import (
	"github.com/everFinance/goar"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBlockHashListByHeightRange(t *testing.T) {
	start := int64(1095756)

	arCli := goar.NewClient("https://arweave.net")
	// info, err := arCli.GetInfo()
	// assert.NoError(t, err)
	// end := info.Height
	// t.Log(start,end)
	//
	// list, err := arCli.GetBlockHashList(int(start),int(end))
	// assert.NoError(t, err)
	// t.Log(list)

	idx, err := GetBlockIdxs(start, arCli)
	assert.NoError(t, err)
	t.Log(idx)
}
