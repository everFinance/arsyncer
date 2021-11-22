package main

import (
	"fmt"
	syncer "github.com/everFinance/ar-syncer"
	"github.com/everFinance/ar-syncer/common"
)

// syncer target tx
func main() {
	ownerFilterParams := common.FilterParams{
		Target: "cSYOy8-p1QFenktkDBFyRM3cwZSTrQ_J4EsELLho_UE", // arTx target address
	}

	startHeight := int64(811484)
	arNode := "https://arweave.net"
	concurrencyNumber := 100 // runtime concurrency number, default 10
	s := syncer.New(startHeight, ownerFilterParams, arNode, concurrencyNumber)

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
