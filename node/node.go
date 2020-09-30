package node

import "context"

type Node interface {
	GetVersion() (string, error)
	GetPeers() (int64, error)
	GetChainHead() (*ChainHead, error)
	SubscribeNewHeads(ctx context.Context) (<-chan ChainHead, error)
	PeerID() (string, error)
	MpoolSize() (int64, error)
	Syncing() (bool, error)
	NetworkStoragePower() (string, error)
	Close()
}
