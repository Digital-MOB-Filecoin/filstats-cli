package core

import (
	"context"
	"time"

	proto "github.com/digital-mob-filecoin/filstats-proto"
)

func (c *Core) sendPeers(ctx context.Context) error {
	var oldPeers int64
	first := true
	log := c.logger.WithField("_req", "Peers")

	consecutiveFails := 0

	return c.intervalRunner(ctx, func() error {
		log.Debug("[⇢] outgoing request")
		start := time.Now()

		peers, err := c.node.GetPeers()
		if err != nil {
			// todo: allow multiple fails then crash
			c.logger.Error(err)

			consecutiveFails++
			if consecutiveFails >= 5 {
				return err
			}

			return nil
		}

		consecutiveFails = 0

		if peers == oldPeers && !first {
			log.Debug("[⇎] nothing new to send")
			return nil
		}

		first = false
		oldPeers = peers

		_, err = c.filstatsTelemetry.Peers(c.contextWithToken(ctx), &proto.PeersRequest{
			Peers: peers,
		})

		log.WithField("duration", time.Since(start)).Debug("[⇠] finalized request")

		return err
	}, PeersInterval)
}
