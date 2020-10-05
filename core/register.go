package core

import (
	"context"
	"runtime"
	"time"

	proto "github.com/digital-mob-filecoin/filstats-proto"
	"github.com/pkg/errors"
)

// Call the Register function on the Filstats server
func (c *Core) filstatsRegister(ctx context.Context) error {
	log := c.logger.WithField("_req", "Register")

	log.Info("[⇢] outgoing request")
	start := time.Now()
	defer func() {
		log.WithField("duration", time.Since(start)).Info("[⇠] finalized request")
	}()

	version, err := c.node.GetVersion()
	if err != nil {
		return err
	}

	nodeType := c.node.Type()

	peerId, err := c.node.PeerID()
	if err != nil {
		return err
	}

	networkName, err := c.node.Network()
	if err != nil {
		return err
	}

	resp, err := c.filstatsServer.Register(c.contextWithToken(ctx), &proto.RegisterRequest{
		Name:        c.config.Filstats.ClientName,
		Version:     version,
		PeerId:      peerId,
		Os:          runtime.GOOS + "_" + runtime.GOARCH,
		NetworkName: networkName,
		Type:        nodeType,
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
