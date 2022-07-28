package main

import (
	"fmt"

	"github.com/everFinance/arsyncer"
	"github.com/everFinance/goar/types"
)

// only sync smart contract txs
func main() {
	swcFilterParams := arsyncer.FilterParams{
		Tags: []types.Tag{
			{Name: "App-Name", Value: "SmartWeaveAction"}, // smart contract tag
		},
	}
	startHeight := int64(472810)
	arNode := "https://arweave.net"
	concurrencyNumber := 50 // runtime concurrency number, default 10
	s := arsyncer.New(startHeight, swcFilterParams, arNode, concurrencyNumber, 15, "subscribe_tx")

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
