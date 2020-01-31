package connection

type boltResult struct {
	metadata map[string]interface{}
}

func newBoltResult(metadata map[string]interface{}) *boltResult {
	return &boltResult{
		metadata: metadata,
	}
}

func (b *boltResult) GetStats() (map[string]interface{}, bool) {
	stats, ok := b.metadata["stats"]
	if !ok {
		return nil, false
	}

	statMap, ok := stats.(map[string]interface{})
	if !ok {
		return nil, false
	}

	return statMap, true
}

func (b *boltResult) GetNodesCreated() (int64, bool) {
	stats, ok := b.GetStats()
	if !ok {
		return -1, false
	}

	num, ok := stats["nodes-created"]
	return num.(int64), ok
}

func (b *boltResult) GetRelationshipsCreated() (int64, bool) {
	stats, ok := b.GetStats()
	if !ok {
		return -1, false
	}

	num, ok := stats["relationships-created"]
	return num.(int64), ok
}

func (b *boltResult) GetNodesDeleted() (int64, bool) {
	stats, ok := b.GetStats()
	if !ok {
		return -1, false
	}

	num, ok := stats["nodes-deleted"]
	return num.(int64), ok
}

func (b *boltResult) GetRelationshipsDeleted() (int64, bool) {
	stats, ok := b.GetStats()
	if !ok {
		return -1, false
	}

	num, ok := stats["relationships-deleted"]
	return num.(int64), ok
}

func (b *boltResult) Metadata() map[string]interface{} {
	return b.metadata
}
