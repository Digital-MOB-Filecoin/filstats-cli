package node

import (
	"time"

	proto "github.com/digital-mob-filecoin/filstats-proto"
)

type ChainHead struct {
	TipsetHeight int64
	Blocks       []Block
	ReceivedAt   *time.Time
}

type Block struct {
	Cid              string
	ParentWeight     string
	CurrentWeight    string
	Miner            string
	NumberOfMessages int
	Timestamp        uint64
}

func (h *ChainHead) ToChainHeadRequest() *proto.ChainHeadRequest {
	req := &proto.ChainHeadRequest{}
	req.TipsetHeight = h.TipsetHeight

	if h.ReceivedAt != nil {
		req.ReceivedAt = h.ReceivedAt.Format(time.RFC3339Nano)
	}

	for _, b := range h.Blocks {
		req.Blocks = append(req.Blocks, &proto.ChainHeadBlock{
			Cid:              b.Cid,
			ParentWeight:     b.ParentWeight,
			Miner:            b.Miner,
			NumberOfMessages: int64(b.NumberOfMessages),
			Timestamp:        b.Timestamp,
			CurrentWeight:    b.CurrentWeight,
		})
	}

	return req
}
