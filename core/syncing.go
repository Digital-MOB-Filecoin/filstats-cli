package core

import (
	"context"
	"time"

	proto "github.com/digital-mob-filecoin/filstats-proto"
)

func (c *Core) sendSyncing(ctx context.Context) error {
	var oldValue bool
	first := true
	log := c.logger.WithField("_req", "Syncing")

	return c.intervalRunner(ctx, func() error {
		log.Debug("[⇢] outgoing request")
		start := time.Now()

		isSyncing, err := c.node.Syncing()
		if err != nil {
			c.logger.Error(err)
			return nil
		}

		if isSyncing == oldValue && !first {
			log.Debug("[⇎] nothing new to send")
			return nil
		}

		first = false
		oldValue = isSyncing

		_, err = c.filstatsTelemetry.Syncing(c.contextWithToken(ctx), &proto.SyncingRequest{
			IsSyncing: isSyncing,
		})

		log.WithField("duration", time.Since(start)).Debug("[⇠] finalized request")

		return err
	}, SyncingInterval)
}
