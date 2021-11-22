package main

import (
	"fmt"
	syncer "github.com/everFinance/ar-syncer"
	"github.com/everFinance/ar-syncer/common"
	"github.com/everFinance/goar/types"
)

// only sync smart contract txs
func main() {
	swcFilterParams := common.FilterParams{
		Tags: []types.Tag{
			{Name: "App-Name", Value: "SmartWeaveAction"}, // smart contract tag
		},
	}
	startHeight := int64(472810)
	arNode := "https://arweave.net"
	concurrencyNumber := 50 // runtime concurrency number, default 10
	s := syncer.New(startHeight, swcFilterParams, arNode, concurrencyNumber)

	// run
	s.Run()

	// subscribe tx
	for {
		select {
		case sTx := <-s.SubscribeTxCh():
			// process synced txs
			fmt.Println(sTx)
		}
	}
}
