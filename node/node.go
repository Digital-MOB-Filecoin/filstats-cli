package node

import "context"

type Node interface {
	GetVersion() (string, error)
	GetPeers() (int64, error)
	GetChainHead() (*ChainHead, error)
	SubscribeNewHeads(ctx context.Context) (<-chan ChainHead, error)
	Close()
}
