package core

import (
	"crypto/tls"

	proto "github.com/digital-mob-filecoin/filstats-proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func (c *Core) initServerConnection() error {
	c.logger.WithField("server-addr", c.config.Filstats.ServerAddr).Info("setting up server connection")

	var conn *grpc.ClientConn
	var err error

	if c.config.Filstats.TLS {
		tlsConfig := &tls.Config{}
		conn, err = grpc.Dial(
			c.config.Filstats.ServerAddr,
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		)
	} else {
		conn, err = grpc.Dial(c.config.Filstats.ServerAddr,
			grpc.WithInsecure(),
		)
	}
	if err != nil {
		return errors.Wrap(err, "could not connect to server")
	}

	c.logger.Info("connection successful; initializing Filstats client")

	c.filstatsServer = proto.NewFilstatsClient(conn)
	c.filstatsTelemetry = proto.NewTelemetryClient(conn)

	c.logger.Info("done initializing Filstats client")

	return nil
}
