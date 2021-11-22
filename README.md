# arsyncer

---
### Install
```go
go get github.com/everFinance/ar-syncer
```

### Introduction
arsyncer is the high performance arweave network transaction synchronisation component.

### Principle
Based on goar's [getBlockFromPeers](https://github.com/everFinance/goar/blob/main/client_broadcast.go#L55) and [getTxFromPeers](https://github.com/everFinance/goar/blob/main/client_broadcast.go#L75) methods.    
it is possible to achieve 100% pull to each block.   
then parse the tx_list in the block.   
and pull 100% of each tx(not return tx_data).   
and all transactions are returned in on-chain order.    

### New
```go
syncer := New(startHeight, filterParams, arNode, conNum)
```
`startHeight` block height at the start of the sync   
`filterParams` filter tx   
`arNode` arweave node url   
`conNum` runtime concurrency number, default 10   

### Run
```go
syncer.Run()
```

### SubscribeTx
```go
for {
     select{
	 case sTx := <-s.SubscribeTxCh():
	 // process sTx
	    ...
    }
}
```


### Example
1. Get all transactions example: [./example/all-tx-syncer.go](https://github.com/everFinance/ar-syncer/blob/v1.0.0/example/all-tx-syncer.go)
2. Get the transactions for the specified sender: [./example/from-tx-syncer.go](https://github.com/everFinance/ar-syncer/blob/v1.0.0/example/from-tx-syncer.go)
3. Get the transactions for the specified target: [./example/target-tx-syncer.go](https://github.com/everFinance/ar-syncer/blob/v1.0.0/example/target-tx-syncer.go)
4. Get all smart contract transactions: [./example/swc-syncer.go](https://github.com/everFinance/ar-syncer/blob/v1.0.0/example/swc-syncer.go)

