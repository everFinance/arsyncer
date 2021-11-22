package arsyncer

import "github.com/everFinance/goar/types"

type SubscribeTx struct {
	types.Transaction

	BlockHeight    int64
	BlockId        string
	BlockTimestamp int64
}

type FilterParams struct {
	Tags         []types.Tag
	OwnerAddress string
	Target       string
}
