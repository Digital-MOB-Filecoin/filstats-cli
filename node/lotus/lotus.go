package lotus

import (
	"context"
	"net/http"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Url   string
	Token string
}

type Node struct {
	config Config
	logger *logrus.Entry

	closer jsonrpc.ClientCloser
	api    apistruct.FullNodeStruct
}

func New(config Config) *Node {
	n := &Node{
		config: config,
		logger: logrus.WithField("module", "lotus"),
	}

	headers := http.Header{"Authorization": []string{"Bearer " + config.Token}}

	var api apistruct.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+config.Url+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
	if err != nil {
		n.logger.Fatalf("connecting with lotus failed: %s", err)
	}

	n.closer = closer
	n.api = api

	return n
}

func (n Node) GetVersion() (string, error) {
	version, err := n.api.Version(context.Background())
	if err != nil {
		return "", err
	}

	return version.Version, nil
}

func (n Node) GetPeers() (int, error) {
	data, err := n.api.NetPeers(context.Background())
	if err != nil {
		return 0, errors.Wrap(err, "could not call NetPeers")
	}

	return len(data), nil
}
