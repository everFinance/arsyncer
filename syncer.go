package arsyncer

import (
	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	"github.com/panjf2000/ants/v2"
	"sync"
	"sync/atomic"
	"time"
)

var log = NewLog("syncer")

type Syncer struct {
	curHeight int64
	FilterParams
	blockChan            chan *types.Block
	txChan               chan SubscribeTx
	SubscribeChan        chan SubscribeTx
	arClient             *goar.Client
	nextSubscribeTxBlock int64
	conNum               int64 // concurrency of number
}

func New(startHeight int64, filterParams FilterParams, arNode string, conNum int) *Syncer {
	if conNum <= 0 {
		conNum = 10 // default concurrency of number is 10
	}
	return &Syncer{
		curHeight:            startHeight,
		FilterParams:         filterParams,
		blockChan:            make(chan *types.Block, 5*conNum),
		txChan:               make(chan SubscribeTx, 1000),
		SubscribeChan:        make(chan SubscribeTx, 1000),
		arClient:             goar.NewClient(arNode),
		nextSubscribeTxBlock: startHeight,
		conNum:               int64(conNum),
	}
}

func (s *Syncer) Run() {
	go s.pollingBlock()
	go s.pollingTx()
	go s.filterTx()
}

func (s *Syncer) Close() (subscribeHeight int64) {
	close(s.blockChan)
	close(s.txChan)
	close(s.SubscribeChan)
	return s.nextSubscribeTxBlock - 1
}

func (s *Syncer) SubscribeTxCh() <-chan SubscribeTx {
	return s.SubscribeChan
}

func (s *Syncer) GetSyncedHeight() int64 {
	return atomic.LoadInt64(&s.nextSubscribeTxBlock)
}

func (s *Syncer) pollingBlock() {
	for {
		info, err := s.arClient.GetInfo()
		if err != nil {
			log.Error("get info", "err", err)
			time.Sleep(10 * time.Second)
			continue
		}
		stableHeight := info.Height - 15 // stable height must low 15
		log.Debug("stable block", "height", stableHeight)

		if s.curHeight >= stableHeight {
			log.Debug("curHeight more than on chain stableHeight", "curHeight", s.curHeight, "stableHeight", stableHeight)
			time.Sleep(5 * time.Minute)
			continue
		}

		for s.curHeight <= stableHeight {
			start := s.curHeight
			end := start + s.conNum
			if end > stableHeight {
				end = stableHeight
			}
			blocks := mustGetBlocks(start, end, s.arClient, int(s.conNum))
			log.Info("get blocks success", "start", start, "end", end)

			s.curHeight = end + 1
			// add chan
			for _, b := range blocks {
				s.blockChan <- b
			}
		}
	}
}

func (s *Syncer) pollingTx() {
	for {
		select {
		case b := <-s.blockChan:
			bHeight := b.Height
			for {
				if bHeight-atomic.LoadInt64(&s.nextSubscribeTxBlock) < s.conNum {
					break
				}
				log.Debug("wait for pollingTxs", "wait block height", b.Height, "nextSubscribeTxBlock Height", s.nextSubscribeTxBlock)
				time.Sleep(5 * time.Second)
			}

			go s.getTxs(*b)
		}
	}
}

func (s *Syncer) getTxs(b types.Block) {
	txs := mustGetTxs(b.Height, b.Txs, s.arClient, int(s.conNum))

	// subscribe txs
	for {
		if b.Height == atomic.LoadInt64(&s.nextSubscribeTxBlock) {
			for _, tx := range txs {
				sTx := SubscribeTx{
					Transaction:    tx,
					BlockHeight:    b.Height,
					BlockId:        b.IndepHash,
					BlockTimestamp: b.Timestamp,
				}
				s.txChan <- sTx
			}
			log.Info("polling one block txs success", "height", b.Height, "txNum", len(txs))
			atomic.AddInt64(&s.nextSubscribeTxBlock, 1)
			break
		}
		time.Sleep(2 * time.Second)
	}
}

func (s *Syncer) filterTx() {
	for {
		select {
		case tx := <-s.txChan:
			if filter(s.FilterParams, tx.Transaction) {
				continue
			}
			s.SubscribeChan <- tx
		}
	}
}

func mustGetTxs(blockHeight int64, blockTxs []string, arClient *goar.Client, conNum int) (txs []types.Transaction) {
	if len(blockTxs) == 0 {
		return
	}

	var (
		lock sync.Mutex
		wg   sync.WaitGroup
	)

	// for sort
	txIdxMap := make(map[string]int)
	for idx, txId := range blockTxs {
		txIdxMap[txId] = idx
	}
	txs = make([]types.Transaction, len(blockTxs))

	p, _ := ants.NewPoolWithFunc(conNum, func(i interface{}) {
		txId := i.(string)
		tx, err := getTxByIdRetry(blockHeight, arClient, txId)
		if err != nil {
			log.Error("get tx by id error", "txId", txId, "err", err)
			// notice: must return fetch failed tx
			tx = types.Transaction{ID: txId}
		}

		lock.Lock()
		idx := txIdxMap[tx.ID]
		txs[idx] = tx
		lock.Unlock()

		wg.Done()
	})

	defer p.Release()

	for _, txId := range blockTxs {
		wg.Add(1)
		_ = p.Invoke(txId)
	}
	wg.Wait()

	return
}

func mustGetBlocks(start, end int64, arClient *goar.Client, conNum int) (blocks []*types.Block) {
	if start > end {
		return
	}

	blocks = make([]*types.Block, end-start+1)
	var (
		lock sync.Mutex
		wg   sync.WaitGroup
	)

	p, _ := ants.NewPoolWithFunc(conNum, func(i interface{}) {
		height := i.(int64)
		b, err := getBlockByHeightRetry(arClient, height)
		if err != nil {
			log.Error("get block by height error", "height", height, "err", err)
			panic(err)
		}

		lock.Lock()
		blocks[height-start] = b
		lock.Unlock()

		wg.Done()
	})

	defer p.Release()

	for i := start; i <= end; i++ {
		wg.Add(1)
		_ = p.Invoke(i)
	}
	wg.Wait()

	return
}

func getTxByIdRetry(blockHeight int64, arCli *goar.Client, txId string) (types.Transaction, error) {
	count := 0
	for {
		// get from trust node
		tx, err := arCli.GetTransactionByID(txId)
		if err != nil {
			// get from non-trust nodes
			tx, err = arCli.GetTxFromPeers(txId)
		}

		if err == nil {
			// verify tx, ignore genesis block txs
			if blockHeight != 0 {
				err = utils.VerifyTransaction(*tx)
			}
		}

		if err == nil {
			return *tx, nil
		}

		if count == 2 {
			return types.Transaction{}, err
		}
		count++
		time.Sleep(2 * time.Second)
	}
}

func getBlockByHeightRetry(arCli *goar.Client, height int64) (*types.Block, error) {
	count := 0
	for {
		b, err := arCli.GetBlockByHeight(height)
		if err == nil {
			return b, nil
		}
		b, err = arCli.GetBlockFromPeers(height)
		if err == nil {
			return b, nil
		}
		if count == 5 {
			return nil, err
		}
		count++
		time.Sleep(2 * time.Second)
	}
}

func filter(params FilterParams, tx types.Transaction) bool {
	if tx.Owner == "" { // todo this tx is incorrect, need subscribe to coder debug
		return false
	}

	if params.OwnerAddress != "" {
		// filer owner address
		addr, err := utils.OwnerToAddress(tx.Owner)
		if err != nil {
			log.Error("utils.OwnerToAddress(tx.Owner) err", "err", err, "owner", tx.Owner)
			return true
		}
		if addr != params.OwnerAddress {
			return true
		}
	}

	if params.Target != "" {
		if params.Target != tx.Target {
			return true
		}
	}

	if len(params.Tags) > 0 {
		// filter tags
		filterTags := utils.TagsEncode(params.Tags)

		txTagsMap := make(map[string]string)
		for _, tg := range tx.Tags {
			txTagsMap[tg.Name] = tg.Value
		}

		for _, ftg := range filterTags {
			val, ok := txTagsMap[ftg.Name]
			if !ok {
				return true
			}
			if val != ftg.Value {
				return true
			}
		}
	}

	return false
}
