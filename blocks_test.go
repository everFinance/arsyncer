package arsyncer

import (
	"github.com/everFinance/goar"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBlockHashListByHeightRange(t *testing.T) {
	start := int64(878791)
	end := int64(878793)
	list, err := GetBlockHashListByGateway3(start, end)
	assert.NoError(t, err)
	t.Log(list)
	arCli := goar.NewClient("https://arweave.net")
	list, err = GetBlockHashList(arCli, start, end)
	t.Log(list)
}
