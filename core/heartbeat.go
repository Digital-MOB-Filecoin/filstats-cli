package core

import (
	"context"
	"time"

	proto "github.com/digital-mob-filecoin/filstats-proto"
)

// sendHeartbeat sends a request to the filstats server every HeartbeatInterval
// this is used to establish the online/offline status of the filstats-cli
func (c *Core) sendHeartbeat(ctx context.Context) error {
	log := c.logger.WithField("_req", "Heartbeat")

	return c.intervalRunner(ctx, func() error {
		log.Debug("[⇢] outgoing request")
		start := time.Now()
		defer func() {
			log.WithField("duration", time.Since(start)).Debug("[⇠] finalized request")
		}()

		_, err := c.filstatsServer.Heartbeat(c.contextWithToken(ctx), &proto.HeartbeatRequest{})

		return err
	}, HeartbeatInterval)
}
