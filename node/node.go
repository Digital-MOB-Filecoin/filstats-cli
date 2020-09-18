package node

type Node interface {
	GetVersion() (string, error)
	GetPeers() (int, error)
}
