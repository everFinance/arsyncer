package main

import (
	"fmt"

	"github.com/everFinance/arsyncer"
)

// sync all arweave tx
func main() {
	nullFilterParams := arsyncer.FilterParams{} // non-file params
	startHeight := int64(879220)
	arNode := "https://arweave.net"
	concurrencyNumber := 10 // runtime concurrency number, default 10
	s := arsyncer.New(startHeight, nullFilterParams, arNode, concurrencyNumber, 15, true)

	// run
	s.Run()

	// subscribe tx
	for {
		select {
		case sTx := <-s.SubscribeTxCh():
			// process synced txs
			fmt.Println("Tx", sTx[0].ID)
		case sBlock := <-s.SubscribeBlockCh():
			fmt.Println("Block", sBlock.Height, sBlock.IndepHash)
		}

	}
}
