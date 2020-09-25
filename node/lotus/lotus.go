package lotus

import (
	"context"
	"net/http"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/digital-mob-filecoin/filstats-cli/node"
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

func (n Node) GetPeers() (int64, error) {
	data, err := n.api.NetPeers(context.Background())
	if err != nil {
		return 0, errors.Wrap(err, "could not call NetPeers")
	}

	return int64(len(data)), nil
}

func (n Node) GetChainHead() (*node.ChainHead, error) {
	data, err := n.api.ChainHead(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "could not get chain head")
	}

	var blocks []node.Block

	lotusBlocks := data.Blocks()

	for _, b := range lotusBlocks {
		msgs, err := n.api.ChainGetBlockMessages(context.Background(), b.Cid())
		if err != nil {
			return nil, errors.Wrap(err, "could not fetch block messages")
		}

		blocks = append(blocks, node.Block{
			Cid:              b.Cid().String(),
			ParentWeight:     b.ParentWeight.String(),
			Miner:            b.Miner.String(),
			NumberOfMessages: len(msgs.Cids),
			Timestamp:        b.Timestamp,
		})
	}

	return &node.ChainHead{
		TipsetHeight: int64(data.Height()),
		Blocks:       blocks,
	}, nil
}

func (n Node) SubscribeNewHeads(ctx context.Context) (<-chan node.ChainHead, error) {
	n.logger.Info("setting up ChainNotify")

	ch, err := n.api.ChainNotify(context.Background())
	if err != nil {
		return nil, err
	}

	n.logger.Info("done setting up ChainNotify")

	headChan := make(chan node.ChainHead)

	go func() {
		for {
			select {
			case headChange := <-ch:
				n.logger.Trace("got head change")
			headChangeLoop:
				for _, hc := range headChange {

					if hc.Val != nil {
						var head node.ChainHead
						head.TipsetHeight = int64(hc.Val.Height())

						for _, b := range hc.Val.Blocks() {
							msgs, err := n.api.ChainGetBlockMessages(context.Background(), b.Cid())
							if err != nil {
								n.logger.Error(err)
								continue headChangeLoop
							}

							head.Blocks = append(head.Blocks, node.Block{
								Cid:              b.Cid().String(),
								ParentWeight:     b.ParentWeight.String(),
								Miner:            b.Miner.String(),
								NumberOfMessages: len(msgs.Cids),
								Timestamp:        b.Timestamp,
							})

							headChan <- head
						}
					}
				}

			case <-ctx.Done():
				close(headChan)
				return
			}
		}
	}()

	return headChan, nil
}

func (n Node) Close() {
	n.closer()
}
