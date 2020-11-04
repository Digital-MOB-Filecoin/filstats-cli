package core

import (
	"context"
	"time"

	proto "github.com/digital-mob-filecoin/filstats-proto"
)

func (c *Core) sendMpoolSize(ctx context.Context) error {
	var oldSize int64
	first := true
	log := c.logger.WithField("_req", "MpoolSize")

	consecutiveFails := 0

	return c.intervalRunner(ctx, func() error {
		log.Debug("[⇢] outgoing request")
		start := time.Now()

		mpoolSize, err := c.node.MpoolSize()
		if err != nil {
			c.logger.Error(err)

			consecutiveFails++
			if consecutiveFails >= 5 {
				return err
			}

			return nil
		}

		consecutiveFails = 0

		if mpoolSize == oldSize && !first {
			log.Debug("[⇎] nothing new to send")
			return nil
		}

		first = false
		oldSize = mpoolSize

		_, err = c.filstatsTelemetry.MpoolSize(c.contextWithToken(ctx), &proto.MpoolSizeRequest{
			Size: mpoolSize,
		})

		log.WithField("duration", time.Since(start)).Debug("[⇠] finalized request")

		return err
	}, MpoolSizeInterval)
}
