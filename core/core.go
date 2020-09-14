package core

import (
	"context"

	proto "github.com/digital-mob-filecoin/filstats-proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

var log = logrus.WithField("module", "core")

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

	filstatsServer proto.FilstatsClient
}

func New(config Config) *Core {
	c := &Core{
		config: config,
	}

	c.initServerConnection()

	return c
}

func (c *Core) Run() {
	c.filstatsRegister()
}

func (c *Core) Close() {
	log.Info("Got stop signal")
}

func (c *Core) contextWithToken() context.Context {
	ctx := context.Background()

	// if we found any token persisted, use that to identify the client with the server
	if c.token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "token", c.token)
	}

	return ctx
}
