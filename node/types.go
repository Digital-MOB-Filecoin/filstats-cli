package node

import proto "github.com/digital-mob-filecoin/filstats-proto"

type ChainHead struct {
	TipsetHeight int64
	Blocks       []Block
}

type Block struct {
	Cid              string
	ParentWeight     string
	Miner            string
	NumberOfMessages int
	Timestamp        uint64
}

func (h *ChainHead) ToChainHeadRequest() *proto.ChainHeadRequest {
	req := &proto.ChainHeadRequest{}
	req.TipsetHeight = h.TipsetHeight

	for _, b := range h.Blocks {
		req.Blocks = append(req.Blocks, &proto.ChainHeadBlock{
			Cid:              b.Cid,
			ParentWeight:     b.ParentWeight,
			Miner:            b.Miner,
			NumberOfMessages: int64(b.NumberOfMessages),
			Timestamp:        b.Timestamp,
		})
	}

	return req
}
