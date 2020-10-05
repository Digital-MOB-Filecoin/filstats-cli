package lotus

import (
	"context"
	"net/http"
	"time"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
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

func (n Node) Type() string {
	return "lotus"
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
				receivedAt := time.Now()

			headChangeLoop:
				for _, hc := range headChange {

					if hc.Val != nil {
						var head node.ChainHead
						head.TipsetHeight = int64(hc.Val.Height())
						head.ReceivedAt = &receivedAt

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

func (n Node) PeerID() (string, error) {
	data, err := n.api.ID(context.Background())
	if err != nil {
		return "", errors.Wrap(err, "could not get peer id")
	}

	return data.String(), nil
}

func (n Node) MpoolSize() (int64, error) {
	data, err := n.api.MpoolPending(context.Background(), types.EmptyTSK)
	if err != nil {
		return 0, errors.Wrap(err, "could not call MpoolPending")
	}

	return int64(len(data)), nil
}

func (n Node) Syncing() (bool, error) {
	data, err := n.api.SyncState(context.Background())
	if err != nil {
		return false, errors.Wrap(err, "could not call SyncState")
	}

	var isSyncing bool

	for _, s := range data.ActiveSyncs {
		if s.Stage != api.StageSyncComplete {
			isSyncing = true
		}
	}

	return isSyncing, nil
}

func (n Node) NetworkStoragePower() (string, error) {
	ch, err := n.api.ChainHead(context.Background())
	if err != nil {
		return "", errors.Wrap(err, "could not call ChainHead")
	}

	blocks := ch.Blocks()
	if len(blocks) == 0 {
		return "", errors.New("no blocks in current tipset")
	}

	p, err := n.api.StateMinerPower(context.Background(), blocks[0].Miner, ch.Key())
	if err != nil {
		return "", errors.Wrap(err, "could not call StateMinerPower")
	}

	return p.TotalPower.QualityAdjPower.String(), nil
}

func (n Node) Network() (string, error) {
	data, err := n.api.StateNetworkName(context.Background())
	if err != nil {
		return "", errors.Wrap(err, "could not call StateNetworkName")
	}

	return string(data), nil
}

func (n Node) Close() {
	n.closer()
}
