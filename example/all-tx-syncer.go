package main

import (
	"fmt"
	syncer "github.com/everFinance/ar-syncer"
	"github.com/everFinance/ar-syncer/common"
)

// sync all arweave tx
func main() {
	nullFilterParams := common.FilterParams{} // non-file params
	startHeight := int64(0)                   // from genesis block start
	arNode := "https://arweave.net"
	concurrencyNumber := 100 // runtime concurrency number, default 10
	s := syncer.New(startHeight, nullFilterParams, arNode, concurrencyNumber)

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
