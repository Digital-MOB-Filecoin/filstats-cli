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
	log.WithField("server-addr", c.config.Filstats.ServerAddr).Info("setting up server connection")

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

	log.Info("connection successful; initializing Filstats client")

	c.filstatsServer = proto.NewFilstatsClient(conn)

	log.Info("done initializing Filstats client")

	return nil
}

func (c *Core) filstatsRegister() error {
	log.Info("outgoing request: Register")

	start := time.Now()
	defer func() {
		log.WithField("duration", time.Since(start)).Info("done Filstats register")
	}()

	resp, err := c.filstatsServer.Register(c.contextWithToken(), &proto.RegisterRequest{
		Name:    c.config.Filstats.ClientName,
		Version: "",
	})
	if err != nil {
		return errors.Wrap(err, "could not execute Register request")
	}

	if resp.Status != proto.Status_OK {
		return errors.New("expected status OK from filstats server; got error")
	}

	return nil
}
