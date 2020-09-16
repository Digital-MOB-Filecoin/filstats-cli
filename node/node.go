package node

type Node interface {
	GetVersion() (string, error)
}
