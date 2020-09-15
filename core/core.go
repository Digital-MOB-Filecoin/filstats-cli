package core

import (
	"context"

	proto "github.com/digital-mob-filecoin/filstats-proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/digital-mob-filecoin/filstats-client/node"
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

	filstatsServer proto.FilstatsClient

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
	err := c.filstatsRegister()
	if err != nil {
		return err
	}

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
		}

	}

	return nil
}

func (c *Core) Close() {
	c.logger.Info("Got stop signal")
}

func (c *Core) contextWithToken() context.Context {
	ctx := context.Background()

	// if we found any token persisted, use that to identify the client with the server
	if c.token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "token", c.token)
	}

	return ctx
}
