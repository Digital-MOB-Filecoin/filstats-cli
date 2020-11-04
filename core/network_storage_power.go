package core

import (
	"context"
	"time"

	proto "github.com/digital-mob-filecoin/filstats-proto"
)

func (c *Core) sendNetworkStoragePower(ctx context.Context) error {
	var oldValue string
	first := true
	log := c.logger.WithField("_req", "NetworkStoragePower")

	consecutiveFails := 0

	return c.intervalRunner(ctx, func() error {
		log.Debug("[⇢] outgoing request")
		start := time.Now()

		nsp, err := c.node.NetworkStoragePower()
		if err != nil {
			c.logger.Error(err)

			consecutiveFails++
			if consecutiveFails >= 5 {
				return err
			}

			return nil
		}

		consecutiveFails = 0

		if nsp == oldValue && !first {
			log.Debug("[⇎] nothing new to send")
			return nil
		}

		first = false
		oldValue = nsp

		_, err = c.filstatsTelemetry.NetworkStoragePower(c.contextWithToken(ctx), &proto.NSPRequest{
			Power: nsp,
		})

		log.WithField("duration", time.Since(start)).Debug("[⇠] finalized request")

		return err
	}, NetworkStoragePowerInterval)
}
