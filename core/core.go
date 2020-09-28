package core

import (
	"context"

	proto "github.com/digital-mob-filecoin/filstats-proto"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

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
		return c.sendMpoolSize(internalCtx)
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
