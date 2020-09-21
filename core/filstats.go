package core

import (
	"crypto/tls"
	"time"

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

	c.logger.Info("done initializing Filstats client")

	return nil
}

// Call the Register function on the Filstats server
func (c *Core) filstatsRegister() error {
	c.logger.Info("outgoing request: Register")

	start := time.Now()
	defer func() {
		c.logger.WithField("duration", time.Since(start)).Info("done Filstats register")
	}()

	version, err := c.node.GetVersion()
	if err != nil {
		return err
	}

	resp, err := c.filstatsServer.Register(c.contextWithToken(), &proto.RegisterRequest{
		Name:    c.config.Filstats.ClientName,
		Version: version,
	})
	if err != nil {
		return errors.Wrap(err, "could not execute Register request")
	}

	if resp.Status != proto.Status_OK {
		return errors.New("expected status OK from filstats server; got error")
	}

	c.token = resp.Token

	// persist the token received from server
	err = c.writeToken(resp.Token)
	if err != nil {
		return errors.Wrap(err, "could not persist token")
	}

	return nil
}
