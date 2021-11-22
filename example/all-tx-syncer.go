package main

import (
	"fmt"
	"github.com/everFinance/arsyncer"
)

// sync all arweave tx
func main() {
	nullFilterParams := arsyncer.FilterParams{} // non-file params
	startHeight := int64(0)                     // from genesis block start
	arNode := "https://arweave.net"
	concurrencyNumber := 100 // runtime concurrency number, default 10
	s := arsyncer.New(startHeight, nullFilterParams, arNode, concurrencyNumber, 15)

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
