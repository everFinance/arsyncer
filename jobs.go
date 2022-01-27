package arsyncer

func (s *Syncer) runJobs() {
	s.scheduler.Every(30).Seconds().SingletonMode().Do(s.updateBlockHashList)

	s.scheduler.StartAsync()
}

func (s *Syncer) updateBlockHashList() {
	if s.blockIdxs.EndHeight-s.curHeight > s.stableDistance/2+1 {
		return
	}

	idxs, err := GetBlockIdxs(s.curHeight, s.arClient)
	if err != nil {
		log.Error("get blockIdxs failed", "err", err)
		return
	}
	s.blockIdxs = idxs
}
