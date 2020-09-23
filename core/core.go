package core

import (
	"context"
	"time"

	proto "github.com/digital-mob-filecoin/filstats-proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

func (c *Core) sendHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(HeartbeatInterval)

	for {
		select {
		case <-ticker.C:
			c.logger.Trace("sending heartbeat")

			_, err := c.filstatsServer.Heartbeat(c.contextWithToken(), &proto.HeartbeatRequest{})
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					switch st.Code() {
					case codes.Unauthenticated:
						// this could happen if the server crashed and lost the authenticated clients
						err2 := c.filstatsRegister()
						if err2 != nil {
							c.logger.Fatal(errors.Wrap(err2, "could not re-register after Unauthenticated code"))
						}

						continue
					case codes.Unavailable:
						c.logger.Error("could not reach Filstats server")
						continue
					default:

					}
				} else {
					c.logger.Errorf("could not send heartbeat, got: %s", err)

					continue
				}
			}

			c.logger.Trace("done sending heartbeat")
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (c *Core) Run(ctx context.Context) error {
	err := c.filstatsRegister()
	if err != nil {
		return err
	}

	c.sendHeartbeat(ctx)

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
