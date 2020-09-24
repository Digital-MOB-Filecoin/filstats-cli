package core

import (
	"context"
	"time"

	"github.com/bep/debounce"
	proto "github.com/digital-mob-filecoin/filstats-proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/digital-mob-filecoin/filstats-cli/node"
)

type FilstatsConfig struct {
	ServerAddr string
	TLS        bool
	ClientName string
}

type Config struct {
	Filstats   FilstatsConfig
	DataFolder string
}

type Core struct {
	config Config
	token  string

	logger *logrus.Entry

	filstatsServer    proto.FilstatsClient
	filstatsTelemetry proto.TelemetryClient

	node node.Node
}

func New(config Config, node node.Node) (*Core, error) {
	c := &Core{
		config: config,
		node:   node,
		logger: logrus.WithField("module", "core"),
	}

	err := c.initServerConnection()
	if err != nil {
		return nil, err
	}

	err = c.searchToken()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Core) sendHeartbeat(ctx context.Context) error {
	ticker := time.NewTicker(HeartbeatInterval)

	for {
		select {
		case <-ticker.C:
			c.logger.Trace("sending heartbeat")

			_, err := c.filstatsServer.Heartbeat(c.contextWithToken(ctx), &proto.HeartbeatRequest{})
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					switch st.Code() {
					case codes.Unauthenticated:
						return errors.New("un-registered from Filstats server; triggering reconnect")
					case codes.Unavailable:
						c.logger.Error("could not reach Filstats server")
						continue
					default:
						c.logger.Error(err)
					}
				} else {
					c.logger.Errorf("could not send heartbeat, got: %s", err)

					continue
				}
			}

			c.logger.Trace("done sending heartbeat")
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

func (c *Core) sendPeers(ctx context.Context) error {
	ticker := time.NewTicker(PeersInterval)

	for {
		select {
		case <-ticker.C:
			c.logger.Trace("sending peers")

			peers, err := c.node.GetPeers()
			if err != nil {
				// todo: allow multiple fails then crash
				c.logger.Error(err)
				continue
			}

			_, err = c.filstatsTelemetry.Peers(c.contextWithToken(ctx), &proto.PeersRequest{
				Peers: peers,
			})
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					switch st.Code() {
					case codes.Unauthenticated:
						return errors.New("un-registered from Filstats server; triggering reconnect")
					case codes.Unavailable:
						c.logger.Error("could not reach Filstats server")
						continue
					default:
						c.logger.Error(err)
					}
				} else {
					c.logger.Errorf("could not send peers, got: %s", err)

					continue
				}
			}

			c.logger.Trace("done sending peers")
		case <-ctx.Done():
			ticker.Stop()
			return nil
		}
	}
}

func (c *Core) filstatsChainHead(ctx context.Context) error {
	// send initial chain head
	head, err := c.node.GetChainHead()
	if err != nil {
		return err
	}

	if head == nil {
		return errors.New("got nil chain head")
	}

	_, err = c.filstatsServer.ChainHead(c.contextWithToken(ctx), head.ToChainHeadRequest())
	if err != nil {
		return errors.Wrap(err, "could not call Filstats.ChainHead")
	}

	return nil
}

func (c *Core) watchNewHeads(ctx context.Context) error {
	ch, err := c.node.SubscribeNewHeads(ctx)
	if err != nil {
		return errors.Wrap(err, "could not subscribe to new heads")
	}

	var latestHead node.ChainHead

	d := debounce.New(time.Second)

	var globalErrors []error

	for head := range ch {
		c.logger.Trace("processing head")
		latestHead = head

		if len(globalErrors) > 0 {
			return globalErrors[0]
		}

		d(func() {
			c.logger.Debug("outgoing request: ChainHead")
			_, err = c.filstatsServer.ChainHead(c.contextWithToken(ctx), latestHead.ToChainHeadRequest())
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					switch st.Code() {
					case codes.Unauthenticated:
						globalErrors = append(globalErrors, errors.New("un-registered from Filstats server; triggering reconnect"))
					case codes.Unavailable:
						c.logger.Error("could not reach Filstats server")
					default:
						c.logger.Error(err)
					}
				} else {
					c.logger.Errorf("could not send ChainHead, got: %s", err)
				}
			}
		})
	}

	if ctx.Err() == context.Canceled {
		return nil
	}

	return errors.New("unexpected close of new heads subscription")
}

func (c *Core) Run(ctx context.Context) error {
	err := c.filstatsRegister(ctx)
	if err != nil {
		return err
	}

	err = c.filstatsChainHead(ctx)
	if err != nil {
		return err
	}

	g, internalCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return c.sendPeers(internalCtx)
	})

	g.Go(func() error {
		return c.sendHeartbeat(internalCtx)
	})

	g.Go(func() error {
		return c.watchNewHeads(internalCtx)
	})

	err = g.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (c *Core) Close() {
	c.logger.Info("Got stop signal")
}
